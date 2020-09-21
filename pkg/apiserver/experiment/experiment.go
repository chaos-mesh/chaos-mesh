// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package experiment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/sync/errgroup"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = ctrl.Log.WithName("experiment api")

// ChaosState defines the number of chaos experiments of each phase
type ChaosState struct {
	Total    int `json:"Total"`
	Running  int `json:"Running"`
	Waiting  int `json:"Waiting"`
	Paused   int `json:"Paused"`
	Failed   int `json:"Failed"`
	Finished int `json:"Finished"`
}

// Service defines a handler service for experiments.
type Service struct {
	conf    *config.ChaosDashboardConfig
	kubeCli client.Client
	archive core.ExperimentStore
	event   core.EventStore
}

// NewService returns an experiment service instance.
func NewService(
	conf *config.ChaosDashboardConfig,
	cli client.Client,
	archive core.ExperimentStore,
	event core.EventStore,
) *Service {
	return &Service{
		conf:    conf,
		kubeCli: cli,
		archive: archive,
		event:   event,
	}
}

// Register mounts our HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/experiments")

	// TODO: add more api handlers
	endpoint.GET("", s.listExperiments)
	endpoint.POST("/new", s.createExperiment)
	endpoint.GET("/detail/:uid", s.getExperimentDetail)
	endpoint.DELETE("/:uid", s.deleteExperiment)
	endpoint.PUT("/update", s.updateExperiment)
	endpoint.PUT("/pause/:uid", s.pauseExperiment)
	endpoint.PUT("/start/:uid", s.startExperiment)
	endpoint.GET("/state", s.state)
}

// Experiment defines the basic information of an experiment
type Experiment struct {
	ExperimentBase
	Created string `json:"created"`
	Status  string `json:"status"`
	UID     string `json:"uid"`
}

// ExperimentBase is used to identify the unique experiment from API request.
type ExperimentBase struct {
	Kind      string `uri:"kind" binding:"required,oneof=PodChaos NetworkChaos IoChaos StressChaos TimeChaos KernelChaos" json:"kind"`
	Namespace string `uri:"namespace" binding:"required,NameValid" json:"namespace"`
	Name      string `uri:"name" binding:"required,NameValid" json:"name"`
}

// ExperimentDetail represents an experiment instance.
type ExperimentDetail struct {
	Experiment
	ExperimentInfo core.ExperimentInfo `json:"experiment_info"`
}

type actionFunc func(info *core.ExperimentInfo) error

// @Summary Create a new chaos experiment.
// @Description Create a new chaos experiment.
// @Tags experiments
// @Produce json
// @Param request body core.ExperimentInfo true "Request body"
// @Success 200 "create ok"
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/new [post]
func (s *Service) createExperiment(c *gin.Context) {
	exp := &core.ExperimentInfo{}
	if err := c.ShouldBindJSON(exp); err != nil {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	createFuncs := map[string]actionFunc{
		v1alpha1.KindPodChaos:     s.createPodChaos,
		v1alpha1.KindNetworkChaos: s.createNetworkChaos,
		v1alpha1.KindIOChaos:      s.createIOChaos,
		v1alpha1.KindStressChaos:  s.createStressChaos,
		v1alpha1.KindTimeChaos:    s.createTimeChaos,
		v1alpha1.KindKernelChaos:  s.createKernelChaos,
	}

	f, ok := createFuncs[exp.Target.Kind]
	if !ok {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New(exp.Target.Kind + " is not supported"))
		return
	}

	if err := f(exp); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (s *Service) createPodChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.PodChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector:      exp.Scope.ParseSelector(),
			Action:        v1alpha1.PodChaosAction(exp.Target.PodChaos.Action),
			Mode:          v1alpha1.PodMode(exp.Scope.Mode),
			Value:         exp.Scope.Value,
			ContainerName: exp.Target.PodChaos.ContainerName,
		},
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return s.kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createNetworkChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.NetworkChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.NetworkChaosSpec{
			Selector: exp.Scope.ParseSelector(),
			Action:   v1alpha1.NetworkChaosAction(exp.Target.NetworkChaos.Action),
			Mode:     v1alpha1.PodMode(exp.Scope.Mode),
			Value:    exp.Scope.Value,
			TcParameter: v1alpha1.TcParameter{
				Delay:     exp.Target.NetworkChaos.Delay,
				Loss:      exp.Target.NetworkChaos.Loss,
				Duplicate: exp.Target.NetworkChaos.Duplicate,
				Corrupt:   exp.Target.NetworkChaos.Corrupt,
			},
		},
	}

	if exp.Target.NetworkChaos.Action == string(v1alpha1.BandwidthAction) {
		chaos.Spec.Bandwidth = exp.Target.NetworkChaos.Bandwidth
		chaos.Spec.Direction = v1alpha1.Direction(exp.Target.NetworkChaos.Direction)
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	if exp.Target.NetworkChaos.TargetScope != nil {
		chaos.Spec.Target = &v1alpha1.Target{
			TargetSelector: exp.Target.NetworkChaos.TargetScope.ParseSelector(),
			TargetMode:     v1alpha1.PodMode(exp.Target.NetworkChaos.TargetScope.Mode),
			TargetValue:    exp.Target.NetworkChaos.TargetScope.Value,
		}
	}

	return s.kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createIOChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.IoChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.IoChaosSpec{
			Selector: exp.Scope.ParseSelector(),
			Mode:     v1alpha1.PodMode(exp.Scope.Mode),
			Value:    exp.Scope.Value,
			Action:   v1alpha1.IoChaosType(exp.Target.IOChaos.Action),
			// TODO: don't hardcode after we support other layers
			Delay:      exp.Target.IOChaos.Delay,
			Errno:      exp.Target.IOChaos.Errno,
			Attr:       exp.Target.IOChaos.Attr,
			Path:       exp.Target.IOChaos.Path,
			Methods:    exp.Target.IOChaos.Methods,
			Percent:    exp.Target.IOChaos.Percent,
			VolumePath: exp.Target.IOChaos.VolumePath,
		},
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return s.kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createTimeChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.TimeChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.TimeChaosSpec{
			Selector:       exp.Scope.ParseSelector(),
			Mode:           v1alpha1.PodMode(exp.Scope.Mode),
			Value:          exp.Scope.Value,
			TimeOffset:     exp.Target.TimeChaos.TimeOffset,
			ClockIds:       exp.Target.TimeChaos.ClockIDs,
			ContainerNames: exp.Target.TimeChaos.ContainerNames,
		},
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return s.kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createKernelChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.KernelChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.KernelChaosSpec{
			Selector:        exp.Scope.ParseSelector(),
			Mode:            v1alpha1.PodMode(exp.Scope.Mode),
			Value:           exp.Scope.Value,
			FailKernRequest: exp.Target.KernelChaos.FailKernRequest,
		},
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return s.kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createStressChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.StressChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.StressChaosSpec{
			Selector:          exp.Scope.ParseSelector(),
			Mode:              v1alpha1.PodMode(exp.Scope.Mode),
			Value:             exp.Scope.Value,
			Stressors:         exp.Target.StressChaos.Stressors,
			StressngStressors: exp.Target.StressChaos.StressngStressors,
		},
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	if exp.Target.StressChaos.ContainerName != nil {
		chaos.Spec.ContainerName = exp.Target.StressChaos.ContainerName
	}

	return s.kubeCli.Create(context.Background(), chaos)
}

func (s *Service) getPodChaosDetail(namespace string, name string) (ExperimentDetail, error) {
	chaos := &v1alpha1.PodChaos{}
	ctx := context.TODO()
	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := s.kubeCli.Get(ctx, chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return ExperimentDetail{}, utils.ErrNotFound.NewWithNoMessage()
		}
		return ExperimentDetail{}, err
	}
	info := core.ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: core.ScopeInfo{
			SelectorInfo: core.SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: core.TargetInfo{
			Kind: v1alpha1.KindPodChaos,
			PodChaos: &core.PodChaosInfo{
				Action:        string(chaos.Spec.Action),
				ContainerName: chaos.Spec.ContainerName,
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	return ExperimentDetail{
		Experiment: Experiment{
			ExperimentBase: ExperimentBase{
				Kind:      chaos.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  chaos.GetChaos().Status,
			UID:     chaos.GetChaos().UID,
		},
		ExperimentInfo: info,
	}, nil
}

func (s *Service) getIoChaosDetail(namespace string, name string) (ExperimentDetail, error) {
	chaos := &v1alpha1.IoChaos{}
	ctx := context.TODO()
	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := s.kubeCli.Get(ctx, chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return ExperimentDetail{}, utils.ErrNotFound.NewWithNoMessage()
		}
		return ExperimentDetail{}, err
	}

	info := core.ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: core.ScopeInfo{
			SelectorInfo: core.SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: core.TargetInfo{
			Kind: v1alpha1.KindIOChaos,
			IOChaos: &core.IOChaosInfo{
				Action:     string(chaos.Spec.Action),
				Delay:      chaos.Spec.Delay,
				Errno:      chaos.Spec.Errno,
				Attr:       chaos.Spec.Attr,
				Path:       chaos.Spec.Path,
				Percent:    chaos.Spec.Percent,
				Methods:    chaos.Spec.Methods,
				VolumePath: chaos.Spec.VolumePath,
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	return ExperimentDetail{
		Experiment: Experiment{
			ExperimentBase: ExperimentBase{
				Kind:      chaos.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  chaos.GetChaos().Status,
			UID:     chaos.GetChaos().UID,
		},
		ExperimentInfo: info,
	}, nil
}

func (s *Service) getNetworkChaosDetail(namespace string, name string) (ExperimentDetail, error) {
	chaos := &v1alpha1.NetworkChaos{}
	ctx := context.TODO()
	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := s.kubeCli.Get(ctx, chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return ExperimentDetail{}, utils.ErrNotFound.NewWithNoMessage()
		}
		return ExperimentDetail{}, err
	}
	info := core.ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: core.ScopeInfo{
			SelectorInfo: core.SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: core.TargetInfo{
			Kind: v1alpha1.KindNetworkChaos,
			NetworkChaos: &core.NetworkChaosInfo{
				Action:    string(chaos.Spec.Action),
				Delay:     chaos.Spec.Delay,
				Loss:      chaos.Spec.Loss,
				Duplicate: chaos.Spec.Duplicate,
				Corrupt:   chaos.Spec.Corrupt,
				Bandwidth: chaos.Spec.Bandwidth,
				Direction: string(chaos.Spec.Direction),
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	if chaos.Spec.Target != nil {
		info.Target.NetworkChaos.TargetScope.SelectorInfo = core.SelectorInfo{
			NamespaceSelectors:  chaos.Spec.Target.TargetSelector.Namespaces,
			LabelSelectors:      chaos.Spec.Target.TargetSelector.LabelSelectors,
			AnnotationSelectors: chaos.Spec.Target.TargetSelector.AnnotationSelectors,
			FieldSelectors:      chaos.Spec.Target.TargetSelector.FieldSelectors,
			PhaseSelector:       chaos.Spec.Target.TargetSelector.PodPhaseSelectors,
		}
		info.Target.NetworkChaos.TargetScope.Mode = string(chaos.Spec.Target.TargetMode)
		info.Target.NetworkChaos.TargetScope.Value = chaos.Spec.Target.TargetValue
	}

	return ExperimentDetail{
		Experiment: Experiment{
			ExperimentBase: ExperimentBase{
				Kind:      chaos.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  chaos.GetChaos().Status,
			UID:     chaos.GetChaos().UID,
		},
		ExperimentInfo: info,
	}, nil
}

func (s *Service) getTimeChaosDetail(namespace string, name string) (ExperimentDetail, error) {
	chaos := &v1alpha1.TimeChaos{}
	ctx := context.TODO()
	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := s.kubeCli.Get(ctx, chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return ExperimentDetail{}, utils.ErrNotFound.NewWithNoMessage()
		}
		return ExperimentDetail{}, err
	}
	info := core.ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: core.ScopeInfo{
			SelectorInfo: core.SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: core.TargetInfo{
			Kind: v1alpha1.KindTimeChaos,
			TimeChaos: &core.TimeChaosInfo{
				TimeOffset:     chaos.Spec.TimeOffset,
				ClockIDs:       chaos.Spec.ClockIds,
				ContainerNames: chaos.Spec.ContainerNames,
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	return ExperimentDetail{
		Experiment: Experiment{
			ExperimentBase: ExperimentBase{
				Kind:      chaos.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  chaos.GetChaos().Status,
			UID:     chaos.GetChaos().UID,
		},
		ExperimentInfo: info,
	}, nil
}

func (s *Service) getKernelChaosDetail(namespace string, name string) (ExperimentDetail, error) {
	chaos := &v1alpha1.KernelChaos{}
	ctx := context.TODO()
	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := s.kubeCli.Get(ctx, chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return ExperimentDetail{}, utils.ErrNotFound.NewWithNoMessage()
		}
		return ExperimentDetail{}, err
	}
	info := core.ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: core.ScopeInfo{
			SelectorInfo: core.SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: core.TargetInfo{
			Kind: v1alpha1.KindKernelChaos,
			KernelChaos: &core.KernelChaosInfo{
				FailKernRequest: chaos.Spec.FailKernRequest,
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	return ExperimentDetail{
		Experiment: Experiment{
			ExperimentBase: ExperimentBase{
				Kind:      chaos.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  chaos.GetChaos().Status,
			UID:     chaos.GetChaos().UID,
		},
		ExperimentInfo: info,
	}, nil
}

func (s *Service) getStressChaosDetail(namespace string, name string) (ExperimentDetail, error) {
	chaos := &v1alpha1.StressChaos{}
	ctx := context.TODO()
	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := s.kubeCli.Get(ctx, chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return ExperimentDetail{}, utils.ErrNotFound.NewWithNoMessage()
		}
		return ExperimentDetail{}, err
	}
	info := core.ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: core.ScopeInfo{
			SelectorInfo: core.SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: core.TargetInfo{
			Kind: v1alpha1.KindStressChaos,
			StressChaos: &core.StressChaosInfo{
				Stressors:         chaos.Spec.Stressors,
				StressngStressors: chaos.Spec.StressngStressors,
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	if chaos.Spec.ContainerName != nil {
		info.Target.StressChaos.ContainerName = chaos.Spec.ContainerName
	}

	return ExperimentDetail{
		Experiment: Experiment{
			ExperimentBase: ExperimentBase{
				Kind:      chaos.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  chaos.GetChaos().Status,
			UID:     chaos.GetChaos().UID,
		},
		ExperimentInfo: info,
	}, nil
}

// @Summary Get chaos experiments from Kubernetes cluster.
// @Description Get chaos experiments from Kubernetes cluster.
// @Tags experiments
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Param kind query string false "kind" Enums(PodChaos, IoChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Param status query string false "status" Enums(Running, Paused, Failed, Finished)
// @Success 200 {array} Experiment
// @Router /experiments [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listExperiments(c *gin.Context) {
	kind := c.Query("kind")
	name := c.Query("name")
	ns := c.Query("namespace")
	status := c.Query("status")

	data := make([]*Experiment, 0)
	for key, list := range v1alpha1.AllKinds() {
		if kind != "" && key != kind {
			continue
		}
		if err := s.kubeCli.List(context.Background(), list.ChaosList, &client.ListOptions{Namespace: ns}); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
		for _, chaos := range list.ListChaos() {
			if name != "" && chaos.Name != name {
				continue
			}
			if status != "" && chaos.Status != status {
				continue
			}
			data = append(data, &Experiment{
				ExperimentBase: ExperimentBase{
					Name:      chaos.Name,
					Namespace: chaos.Namespace,
					Kind:      chaos.Kind,
				},
				Created: chaos.StartTime.Format(time.RFC3339),
				Status:  chaos.Status,
				UID:     chaos.UID,
			})
		}
	}

	c.JSON(http.StatusOK, data)
}

// @Summary Get detailed information about the specified chaos experiment.
// @Description Get detailed information about the specified chaos experiment.
// @Tags experiments
// @Produce json
// @Param uid path string true "uid"
// @Router /experiments/detail/{uid} [GET]
// @Success 200 {object} ExperimentDetail
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
func (s *Service) getExperimentDetail(c *gin.Context) {
	var (
		err       error
		exp       *core.ArchiveExperiment
		expDetail ExperimentDetail
	)

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the experiment is not found"))
		}
	}

	kind := exp.Kind
	ns := exp.Namespace
	name := exp.Name

	switch kind {
	case v1alpha1.KindPodChaos:
		expDetail, err = s.getPodChaosDetail(ns, name)
	case v1alpha1.KindIOChaos:
		expDetail, err = s.getIoChaosDetail(ns, name)
	case v1alpha1.KindNetworkChaos:
		expDetail, err = s.getNetworkChaosDetail(ns, name)
	case v1alpha1.KindTimeChaos:
		expDetail, err = s.getTimeChaosDetail(ns, name)
	case v1alpha1.KindKernelChaos:
		expDetail, err = s.getKernelChaosDetail(ns, name)
	case v1alpha1.KindStressChaos:
		expDetail, err = s.getStressChaosDetail(ns, name)
	}
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, expDetail)
}

// @Summary Delete the specified chaos experiment.
// @Description Delete the specified chaos experiment.
// @Tags experiments
// @Produce json
// @Param uid path string true "uid"
// @Param force query string true "force" Enums(true, false)
// @Success 200 "delete ok"
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/{uid} [delete]
func (s *Service) deleteExperiment(c *gin.Context) {
	var (
		chaosKind *v1alpha1.ChaosKind
		chaosMeta metav1.Object
		ok        bool
		err       error
		exp       *core.ArchiveExperiment
	)

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the experiment is not found"))
		}
	}

	kind := exp.Kind
	ns := exp.Namespace
	name := exp.Name
	force := c.DefaultQuery("force", "false")

	ctx := context.TODO()
	chaosKey := types.NamespacedName{Namespace: ns, Name: name}

	if chaosKind, ok = v1alpha1.AllKinds()[kind]; !ok {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New(kind + " is not supported"))
		return
	}
	if err := s.kubeCli.Get(ctx, chaosKey, chaosKind.Chaos); err != nil {
		if apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrNotFound.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}
		return
	}

	if force == "true" {
		if chaosMeta, ok = chaosKind.Chaos.(metav1.Object); !ok {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("failed to get chaos meta information")))
			return
		}

		annotations := chaosMeta.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[common.AnnotationCleanFinalizer] = common.AnnotationCleanFinalizerForced
		chaosMeta.SetAnnotations(annotations)
		if err := s.kubeCli.Update(context.Background(), chaosKind.Chaos); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("forced deletion of chaos failed, because update chaos annotation error")))
			return
		}
	}

	if err := s.kubeCli.Delete(ctx, chaosKind.Chaos, &client.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrNotFound.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}
		return
	}

	c.JSON(http.StatusOK, nil)
}

// @Summary Get chaos experiments state from Kubernetes cluster.
// @Description Get chaos experiments state from Kubernetes cluster.
// @Tags experiments
// @Produce json
// @Success 200 {object} ChaosState
// @Router /experiments/state [get]
// @Failure 500 {object} utils.APIError
func (s *Service) state(c *gin.Context) {
	data := new(ChaosState)

	g, ctx := errgroup.WithContext(context.Background())
	m := &sync.Mutex{}
	kinds := v1alpha1.AllKinds()
	for index := range kinds {
		list := kinds[index]
		g.Go(func() error {
			if err := s.kubeCli.List(ctx, list.ChaosList); err != nil {
				return err
			}
			m.Lock()
			for _, chaos := range list.ListChaos() {
				switch chaos.Status {
				case string(v1alpha1.ExperimentPhaseRunning):
					data.Running++
				case string(v1alpha1.ExperimentPhaseWaiting):
					data.Waiting++
				case string(v1alpha1.ExperimentPhasePaused):
					data.Paused++
				case string(v1alpha1.ExperimentPhaseFailed):
					data.Failed++
				case string(v1alpha1.ExperimentPhaseFinished):
					data.Finished++
				}
				data.Total++
			}
			m.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, data)
}

// @Summary Pause chaos experiment by API
// @Description Pause chaos experiment by API
// @Tags experiments
// @Produce json
// @Param uid path string true "uid"
// @Success 200 "pause ok"
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/pause/{uid} [put]
func (s *Service) pauseExperiment(c *gin.Context) {
	var (
		err        error
		experiment *core.ArchiveExperiment
	)

	uid := c.Param("uid")
	if experiment, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the experiment is not found"))
		}
	}

	exp := &ExperimentBase{
		Kind:      experiment.Kind,
		Name:      experiment.Name,
		Namespace: experiment.Namespace,
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "true",
	}
	if err := s.patchExperiment(exp, annotations); err != nil {
		if apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrNotFound.WrapWithNoMessage(err))
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, nil)
}

// @Summary Start the paused chaos experiment by API
// @Description Start the paused chaos experiment by API
// @Tags experiments
// @Produce json
// @Param uid path string true "uid"
// @Success 200 "start ok"
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/start/{uid} [put]
func (s *Service) startExperiment(c *gin.Context) {
	var (
		err        error
		experiment *core.ArchiveExperiment
	)

	uid := c.Param("uid")
	if experiment, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the experiment is not found"))
		}
	}

	exp := &ExperimentBase{
		Kind:      experiment.Kind,
		Name:      experiment.Name,
		Namespace: experiment.Namespace,
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "false",
	}
	if err := s.patchExperiment(exp, annotations); err != nil {
		if apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrNotFound.WrapWithNoMessage(err))
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (s *Service) patchExperiment(exp *ExperimentBase, annotations map[string]string) error {
	var (
		chaosKind *v1alpha1.ChaosKind
		ok        bool
	)

	if chaosKind, ok = v1alpha1.AllKinds()[exp.Kind]; !ok {
		return fmt.Errorf("%s is not supported", exp.Kind)
	}

	key := types.NamespacedName{Namespace: exp.Namespace, Name: exp.Name}
	if err := s.kubeCli.Get(context.Background(), key, chaosKind.Chaos); err != nil {
		return err
	}

	var mergePatch []byte
	mergePatch, _ = json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": annotations,
		},
	})

	return s.kubeCli.Patch(context.Background(),
		chaosKind.Chaos,
		client.ConstantPatch(types.MergePatchType, mergePatch))
}

// @Summary Update the chaos experiment by API
// @Description Update the chaos experiment by API
// @Tags experiments
// @Produce json
// @Param request body core.ExperimentInfo true "Request body"
// @Success 200 "update ok"
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/update [put]
func (s *Service) updateExperiment(c *gin.Context) {
	exp := &core.ExperimentInfo{}
	if err := c.ShouldBindJSON(exp); err != nil {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	updateFuncs := map[string]actionFunc{
		v1alpha1.KindPodChaos:     s.updatePodChaos,
		v1alpha1.KindNetworkChaos: s.updateNetworkChaos,
		v1alpha1.KindIOChaos:      s.updateIOChaos,
		v1alpha1.KindStressChaos:  s.updateStressChaos,
		v1alpha1.KindTimeChaos:    s.updateTimeChaos,
		v1alpha1.KindKernelChaos:  s.updateKernelChaos,
	}

	f, ok := updateFuncs[exp.Target.Kind]
	if !ok {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New(exp.Target.Kind + " is not supported"))
		return
	}

	if err := f(exp); err != nil {
		if apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrNotFound.WrapWithNoMessage(err))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (s *Service) updatePodChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.PodChaos{}
	key := types.NamespacedName{Namespace: exp.Namespace, Name: exp.Name}

	if err := s.kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(exp.Labels)
	chaos.SetAnnotations(exp.Annotations)
	chaos.Spec = v1alpha1.PodChaosSpec{
		Selector:      exp.Scope.ParseSelector(),
		Action:        v1alpha1.PodChaosAction(exp.Target.PodChaos.Action),
		Mode:          v1alpha1.PodMode(exp.Scope.Mode),
		Value:         exp.Scope.Value,
		ContainerName: exp.Target.PodChaos.ContainerName,
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return s.kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateNetworkChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.NetworkChaos{}
	key := types.NamespacedName{Namespace: exp.Namespace, Name: exp.Name}

	if err := s.kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(exp.Labels)
	chaos.SetAnnotations(exp.Annotations)
	chaos.Spec = v1alpha1.NetworkChaosSpec{
		Selector: exp.Scope.ParseSelector(),
		Action:   v1alpha1.NetworkChaosAction(exp.Target.NetworkChaos.Action),
		Mode:     v1alpha1.PodMode(exp.Scope.Mode),
		Value:    exp.Scope.Value,
		TcParameter: v1alpha1.TcParameter{
			Delay:     exp.Target.NetworkChaos.Delay,
			Loss:      exp.Target.NetworkChaos.Loss,
			Duplicate: exp.Target.NetworkChaos.Duplicate,
			Corrupt:   exp.Target.NetworkChaos.Corrupt,
			Bandwidth: exp.Target.NetworkChaos.Bandwidth,
		},
		Direction: v1alpha1.Direction(exp.Target.NetworkChaos.Direction),
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	if exp.Target.NetworkChaos.TargetScope != nil {
		chaos.Spec.Target = &v1alpha1.Target{
			TargetSelector: exp.Target.NetworkChaos.TargetScope.ParseSelector(),
			TargetMode:     v1alpha1.PodMode(exp.Target.NetworkChaos.TargetScope.Mode),
			TargetValue:    exp.Target.NetworkChaos.TargetScope.Value,
		}
	}

	return s.kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateIOChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.IoChaos{}
	key := types.NamespacedName{Namespace: exp.Namespace, Name: exp.Name}

	if err := s.kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(exp.Labels)
	chaos.SetAnnotations(exp.Annotations)

	chaos.Spec = v1alpha1.IoChaosSpec{
		Selector:   exp.Scope.ParseSelector(),
		Action:     v1alpha1.IoChaosType(exp.Target.IOChaos.Action),
		Mode:       v1alpha1.PodMode(exp.Scope.Mode),
		Value:      exp.Scope.Value,
		Delay:      exp.Target.IOChaos.Delay,
		Errno:      exp.Target.IOChaos.Errno,
		Path:       exp.Target.IOChaos.Path,
		Percent:    exp.Target.IOChaos.Percent,
		Methods:    exp.Target.IOChaos.Methods,
		VolumePath: exp.Target.IOChaos.VolumePath,
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return s.kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateKernelChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.KernelChaos{}
	key := types.NamespacedName{Namespace: exp.Namespace, Name: exp.Name}

	if err := s.kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(exp.Labels)
	chaos.SetAnnotations(exp.Annotations)
	chaos.Spec = v1alpha1.KernelChaosSpec{
		Selector:        exp.Scope.ParseSelector(),
		Mode:            v1alpha1.PodMode(exp.Scope.Mode),
		Value:           exp.Scope.Value,
		FailKernRequest: exp.Target.KernelChaos.FailKernRequest,
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return s.kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateTimeChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.TimeChaos{}
	key := types.NamespacedName{Namespace: exp.Namespace, Name: exp.Name}

	if err := s.kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(exp.Labels)
	chaos.SetAnnotations(exp.Annotations)
	chaos.Spec = v1alpha1.TimeChaosSpec{
		Selector:       exp.Scope.ParseSelector(),
		Mode:           v1alpha1.PodMode(exp.Scope.Mode),
		Value:          exp.Scope.Value,
		TimeOffset:     exp.Target.TimeChaos.TimeOffset,
		ClockIds:       exp.Target.TimeChaos.ClockIDs,
		ContainerNames: exp.Target.TimeChaos.ContainerNames,
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return s.kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateStressChaos(exp *core.ExperimentInfo) error {
	chaos := &v1alpha1.StressChaos{}
	key := types.NamespacedName{Namespace: exp.Namespace, Name: exp.Name}

	if err := s.kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(exp.Labels)
	chaos.SetAnnotations(exp.Annotations)
	chaos.Spec = v1alpha1.StressChaosSpec{
		Selector:          exp.Scope.ParseSelector(),
		Mode:              v1alpha1.PodMode(exp.Scope.Mode),
		Value:             exp.Scope.Value,
		Stressors:         exp.Target.StressChaos.Stressors,
		StressngStressors: exp.Target.StressChaos.StressngStressors,
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	if exp.Target.StressChaos.ContainerName != nil {
		chaos.Spec.ContainerName = exp.Target.StressChaos.ContainerName
	}

	return s.kubeCli.Update(context.Background(), chaos)
}

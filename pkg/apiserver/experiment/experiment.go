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
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/sync/errgroup"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/finalizers"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	dashboardconfig "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
)

var log = ctrl.Log.WithName("experiment api")

// Service defines a handler service for experiments.
type Service struct {
	archive core.ExperimentStore
	event   core.EventStore
	conf    *dashboardconfig.ChaosDashboardConfig
	scheme  *runtime.Scheme
}

// NewService returns an experiment service instance.
func NewService(
	archive core.ExperimentStore,
	event core.EventStore,
	conf *dashboardconfig.ChaosDashboardConfig,
	scheme *runtime.Scheme,
) *Service {
	return &Service{
		archive: archive,
		event:   event,
		conf:    conf,
		scheme:  scheme,
	}
}

// Register mounts HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/experiments")

	endpoint.GET("", s.listExperiments)
	endpoint.POST("/new", s.createExperiment)
	endpoint.GET("/detail/:uid", s.getExperimentDetail)
	endpoint.DELETE("/:uid", s.deleteExperiment)
	endpoint.DELETE("/", s.batchDeleteExperiment)
	endpoint.PUT("/update", s.updateExperiment)
	endpoint.PUT("/pause/:uid", s.pauseExperiment)
	endpoint.PUT("/start/:uid", s.startExperiment)
	endpoint.GET("/state", s.state)
}

// ChaosState defines the number of chaos experiments of each phase
type ChaosState struct {
	Injecting int `json:"injecting"`
	Running   int `json:"running"`
	Finished  int `json:"finished"`
	Paused    int `json:"paused"`
}

// Base represents the base info of an experiment.
type Base struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// Experiment defines the basic information of an experiment
type Experiment struct {
	Base
	UID           string `json:"uid"`
	Created       string `json:"created_at"`
	Status        string `json:"status"`
	FailedMessage string `json:"failed_message,omitempty"`
}

// Detail represents an experiment instance.
type Detail struct {
	Experiment
	KubeObject core.KubeObjectDesc `json:"kube_object"`
}

type createExperimentFunc func(*core.ExperimentInfo, client.Client) error
type updateExperimentFunc func(*core.KubeObjectDesc, client.Client) error

// StatusResponse defines a common status struct.
type StatusResponse struct {
	Status string `json:"status"`
}

// @Summary Create a new chaos experiment.
// @Description Create a new chaos experiment.
// @Tags experiments
// @Produce json
// @Param request body core.ExperimentInfo true "Request body"
// @Success 200 {object} core.ExperimentInfo
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/new [post]
func (s *Service) createExperiment(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	exp := &core.ExperimentInfo{}
	if err := c.ShouldBindJSON(exp); err != nil {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	createFuncs := map[string]createExperimentFunc{
		v1alpha1.KindPodChaos:     s.createPodChaos,
		v1alpha1.KindNetworkChaos: s.createNetworkChaos,
		v1alpha1.KindIOChaos:      s.createIOChaos,
		v1alpha1.KindStressChaos:  s.createStressChaos,
		v1alpha1.KindTimeChaos:    s.createTimeChaos,
		v1alpha1.KindKernelChaos:  s.createKernelChaos,
		v1alpha1.KindDNSChaos:     s.createDNSChaos,
		v1alpha1.KindAwsChaos:     s.createAwsChaos,
		v1alpha1.KindGcpChaos:     s.createGcpChaos,
	}

	f, ok := createFuncs[exp.Target.Kind]
	if !ok {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New(exp.Target.Kind + " is not supported"))
		return
	}

	if err := f(exp, kubeCli); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, exp)
}

func (s *Service) createPodChaos(exp *core.ExperimentInfo, kubeCli client.Client) error {
	chaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.PodChaosSpec{
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: exp.Scope.ParseSelector(),
					Mode:     v1alpha1.PodMode(exp.Scope.Mode),
					Value:    exp.Scope.Value,
				},
				ContainerNames: exp.Target.PodChaos.ContainerNames,
			},
			Action:      v1alpha1.PodChaosAction(exp.Target.PodChaos.Action),
			GracePeriod: exp.Target.PodChaos.GracePeriod,
		},
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createNetworkChaos(exp *core.ExperimentInfo, kubeCli client.Client) error {
	chaos := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.NetworkChaosSpec{
			PodSelector: v1alpha1.PodSelector{
				Selector: exp.Scope.ParseSelector(),
				Mode:     v1alpha1.PodMode(exp.Scope.Mode),
				Value:    exp.Scope.Value,
			},
			Action: v1alpha1.NetworkChaosAction(exp.Target.NetworkChaos.Action),
			TcParameter: v1alpha1.TcParameter{
				Delay:     exp.Target.NetworkChaos.Delay,
				Loss:      exp.Target.NetworkChaos.Loss,
				Duplicate: exp.Target.NetworkChaos.Duplicate,
				Corrupt:   exp.Target.NetworkChaos.Corrupt,
				Bandwidth: exp.Target.NetworkChaos.Bandwidth,
			},
			Direction:       v1alpha1.Direction(exp.Target.NetworkChaos.Direction),
			ExternalTargets: exp.Target.NetworkChaos.ExternalTargets,
		},
	}

	if exp.Target.NetworkChaos.TargetScope != nil {
		chaos.Spec.Target = &v1alpha1.PodSelector{
			Selector: exp.Target.NetworkChaos.TargetScope.ParseSelector(),
			Mode:     v1alpha1.PodMode(exp.Target.NetworkChaos.TargetScope.Mode),
			Value:    exp.Target.NetworkChaos.TargetScope.Value,
		}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createIOChaos(exp *core.ExperimentInfo, kubeCli client.Client) error {
	chaos := &v1alpha1.IOChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.IOChaosSpec{
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: exp.Scope.ParseSelector(),
					Mode:     v1alpha1.PodMode(exp.Scope.Mode),
					Value:    exp.Scope.Value,
				},
				ContainerNames: []string{exp.Target.IOChaos.ContainerName},
			},
			Action:     v1alpha1.IOChaosType(exp.Target.IOChaos.Action),
			Delay:      exp.Target.IOChaos.Delay,
			Errno:      exp.Target.IOChaos.Errno,
			Attr:       exp.Target.IOChaos.Attr,
			Mistake:    exp.Target.IOChaos.Mistake,
			Path:       exp.Target.IOChaos.Path,
			Methods:    exp.Target.IOChaos.Methods,
			Percent:    exp.Target.IOChaos.Percent,
			VolumePath: exp.Target.IOChaos.VolumePath,
		},
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createTimeChaos(exp *core.ExperimentInfo, kubeCli client.Client) error {
	chaos := &v1alpha1.TimeChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.TimeChaosSpec{
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: exp.Scope.ParseSelector(),
					Mode:     v1alpha1.PodMode(exp.Scope.Mode),
					Value:    exp.Scope.Value,
				},
				ContainerNames: exp.Target.TimeChaos.ContainerNames,
			},
			TimeOffset: exp.Target.TimeChaos.TimeOffset,
			ClockIds:   exp.Target.TimeChaos.ClockIDs,
		},
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createKernelChaos(exp *core.ExperimentInfo, kubeCli client.Client) error {
	chaos := &v1alpha1.KernelChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.KernelChaosSpec{
			PodSelector: v1alpha1.PodSelector{
				Selector: exp.Scope.ParseSelector(),
				Mode:     v1alpha1.PodMode(exp.Scope.Mode),
				Value:    exp.Scope.Value,
			},
			FailKernRequest: exp.Target.KernelChaos.FailKernRequest,
		},
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createStressChaos(exp *core.ExperimentInfo, kubeCli client.Client) error {
	var stressors *v1alpha1.Stressors

	// Error checking
	if exp.Target.StressChaos.Stressors.CPUStressor.Workers <= 0 && exp.Target.StressChaos.Stressors.MemoryStressor.Workers > 0 {
		stressors = &v1alpha1.Stressors{
			MemoryStressor: exp.Target.StressChaos.Stressors.MemoryStressor,
		}
	} else if exp.Target.StressChaos.Stressors.MemoryStressor.Workers <= 0 && exp.Target.StressChaos.Stressors.CPUStressor.Workers > 0 {
		stressors = &v1alpha1.Stressors{
			CPUStressor: exp.Target.StressChaos.Stressors.CPUStressor,
		}
	} else {
		stressors = exp.Target.StressChaos.Stressors
	}

	chaos := &v1alpha1.StressChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.StressChaosSpec{
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: exp.Scope.ParseSelector(),
					Mode:     v1alpha1.PodMode(exp.Scope.Mode),
					Value:    exp.Scope.Value,
				},
				ContainerNames: []string{*exp.Target.StressChaos.ContainerName},
			},
			Stressors:         stressors,
			StressngStressors: exp.Target.StressChaos.StressngStressors,
		},
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createDNSChaos(exp *core.ExperimentInfo, kubeCli client.Client) error {
	chaos := &v1alpha1.DNSChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.DNSChaosSpec{
			Action: v1alpha1.DNSChaosAction(exp.Target.DNSChaos.Action),
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: exp.Scope.ParseSelector(),
					Mode:     v1alpha1.PodMode(exp.Scope.Mode),
					Value:    exp.Scope.Value,
				},
				ContainerNames: exp.Target.DNSChaos.ContainerNames,
			},
			DomainNamePatterns: exp.Target.DNSChaos.DomainNamePatterns,
		},
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createAwsChaos(exp *core.ExperimentInfo, kubeCli client.Client) error {
	chaos := &v1alpha1.AwsChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.AwsChaosSpec{
			Action:     v1alpha1.AwsChaosAction(exp.Target.AwsChaos.Action),
			SecretName: exp.Target.AwsChaos.SecretName,
			AwsSelector: v1alpha1.AwsSelector{
				AwsRegion:   exp.Target.AwsChaos.AwsRegion,
				Ec2Instance: exp.Target.AwsChaos.Ec2Instance,
				EbsVolume:   exp.Target.AwsChaos.EbsVolume,
				DeviceName:  exp.Target.AwsChaos.DeviceName,
			},
		},
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createGcpChaos(exp *core.ExperimentInfo, kubeCli client.Client) error {
	chaos := &v1alpha1.GcpChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.GcpChaosSpec{
			Action:     v1alpha1.GcpChaosAction(exp.Target.GcpChaos.Action),
			SecretName: exp.Target.GcpChaos.SecretName,
			GcpSelector: v1alpha1.GcpSelector{
				Project:     exp.Target.GcpChaos.Project,
				Zone:        exp.Target.GcpChaos.Zone,
				Instance:    exp.Target.GcpChaos.Instance,
				DeviceNames: exp.Target.GcpChaos.DeviceNames,
			},
		},
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	return kubeCli.Create(context.Background(), chaos)
}

func (s *Service) getPodChaosDetail(namespace string, name string, kubeCli client.Client) (Detail, error) {
	chaos := &v1alpha1.PodChaos{}

	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := kubeCli.Get(context.Background(), chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return Detail{}, utils.ErrNotFound.NewWithNoMessage()
		}

		return Detail{}, err
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		return Detail{}, err
	}

	return Detail{
		Experiment: Experiment{
			Base: Base{
				Kind:      gvk.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			UID:     chaos.GetChaos().UID,
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  string(utils.GetChaosState(chaos)),
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        chaos.Name,
				Namespace:   chaos.Namespace,
				Labels:      chaos.Labels,
				Annotations: chaos.Annotations,
			},
			Spec: chaos.Spec,
		},
	}, nil
}

func (s *Service) getIOChaosDetail(namespace string, name string, kubeCli client.Client) (Detail, error) {
	chaos := &v1alpha1.IOChaos{}

	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := kubeCli.Get(context.Background(), chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return Detail{}, utils.ErrNotFound.NewWithNoMessage()
		}

		return Detail{}, err
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		return Detail{}, err
	}

	return Detail{
		Experiment: Experiment{
			Base: Base{
				Kind:      gvk.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			UID:     chaos.GetChaos().UID,
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  string(utils.GetChaosState(chaos)),
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        chaos.Name,
				Namespace:   chaos.Namespace,
				Labels:      chaos.Labels,
				Annotations: chaos.Annotations,
			},
			Spec: chaos.Spec,
		},
	}, nil
}

func (s *Service) getNetworkChaosDetail(namespace string, name string, kubeCli client.Client) (Detail, error) {
	chaos := &v1alpha1.NetworkChaos{}

	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := kubeCli.Get(context.Background(), chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return Detail{}, utils.ErrNotFound.NewWithNoMessage()
		}

		return Detail{}, err
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		return Detail{}, err
	}

	return Detail{
		Experiment: Experiment{
			Base: Base{
				Kind:      gvk.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			UID:     chaos.GetChaos().UID,
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  string(utils.GetChaosState(chaos)),
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        chaos.Name,
				Namespace:   chaos.Namespace,
				Labels:      chaos.Labels,
				Annotations: chaos.Annotations,
			},
			Spec: chaos.Spec,
		},
	}, nil
}

func (s *Service) getTimeChaosDetail(namespace string, name string, kubeCli client.Client) (Detail, error) {
	chaos := &v1alpha1.TimeChaos{}

	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := kubeCli.Get(context.Background(), chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return Detail{}, utils.ErrNotFound.NewWithNoMessage()
		}

		return Detail{}, err
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		return Detail{}, err
	}

	return Detail{
		Experiment: Experiment{
			Base: Base{
				Kind:      gvk.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  string(utils.GetChaosState(chaos)),
			UID:     chaos.GetChaos().UID,
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        chaos.Name,
				Namespace:   chaos.Namespace,
				Labels:      chaos.Labels,
				Annotations: chaos.Annotations,
			},
			Spec: chaos.Spec,
		},
	}, nil
}

func (s *Service) getKernelChaosDetail(namespace string, name string, kubeCli client.Client) (Detail, error) {
	chaos := &v1alpha1.KernelChaos{}

	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := kubeCli.Get(context.Background(), chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return Detail{}, utils.ErrNotFound.NewWithNoMessage()
		}

		return Detail{}, err
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		return Detail{}, err
	}

	return Detail{
		Experiment: Experiment{
			Base: Base{
				Kind:      gvk.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  string(utils.GetChaosState(chaos)),
			UID:     chaos.GetChaos().UID,
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        chaos.Name,
				Namespace:   chaos.Namespace,
				Labels:      chaos.Labels,
				Annotations: chaos.Annotations,
			},
			Spec: chaos.Spec,
		},
	}, nil
}

func (s *Service) getStressChaosDetail(namespace string, name string, kubeCli client.Client) (Detail, error) {
	chaos := &v1alpha1.StressChaos{}

	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := kubeCli.Get(context.Background(), chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return Detail{}, utils.ErrNotFound.NewWithNoMessage()
		}

		return Detail{}, err
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		return Detail{}, err
	}

	return Detail{
		Experiment: Experiment{
			Base: Base{
				Kind:      gvk.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  string(utils.GetChaosState(chaos)),
			UID:     chaos.GetChaos().UID,
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        chaos.Name,
				Namespace:   chaos.Namespace,
				Labels:      chaos.Labels,
				Annotations: chaos.Annotations,
			},
			Spec: chaos.Spec,
		},
	}, nil
}

func (s *Service) getDNSChaosDetail(namespace string, name string, kubeCli client.Client) (Detail, error) {
	chaos := &v1alpha1.DNSChaos{}

	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := kubeCli.Get(context.Background(), chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return Detail{}, utils.ErrNotFound.NewWithNoMessage()
		}

		return Detail{}, err
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		return Detail{}, err
	}

	return Detail{
		Experiment: Experiment{
			Base: Base{
				Kind:      gvk.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  string(utils.GetChaosState(chaos)),
			UID:     chaos.GetChaos().UID,
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        chaos.Name,
				Namespace:   chaos.Namespace,
				Labels:      chaos.Labels,
				Annotations: chaos.Annotations,
			},
			Spec: chaos.Spec,
		},
	}, nil
}

func (s *Service) getAwsChaosDetail(namespace string, name string, kubeCli client.Client) (Detail, error) {
	chaos := &v1alpha1.AwsChaos{}

	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := kubeCli.Get(context.Background(), chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return Detail{}, utils.ErrNotFound.NewWithNoMessage()
		}

		return Detail{}, err
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		return Detail{}, err
	}

	return Detail{
		Experiment: Experiment{
			Base: Base{
				Kind:      gvk.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  string(utils.GetChaosState(chaos)),
			UID:     chaos.GetChaos().UID,
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        chaos.Name,
				Namespace:   chaos.Namespace,
				Labels:      chaos.Labels,
				Annotations: chaos.Annotations,
			},
			Spec: chaos.Spec,
		},
	}, nil
}

func (s *Service) getGcpChaosDetail(namespace string, name string, kubeCli client.Client) (Detail, error) {
	chaos := &v1alpha1.GcpChaos{}

	chaosKey := types.NamespacedName{Namespace: namespace, Name: name}
	if err := kubeCli.Get(context.Background(), chaosKey, chaos); err != nil {
		if apierrors.IsNotFound(err) {
			return Detail{}, utils.ErrNotFound.NewWithNoMessage()
		}

		return Detail{}, err
	}

	gvk, err := apiutil.GVKForObject(chaos, s.scheme)
	if err != nil {
		return Detail{}, err
	}

	return Detail{
		Experiment: Experiment{
			Base: Base{
				Kind:      gvk.Kind,
				Namespace: chaos.Namespace,
				Name:      chaos.Name,
			},
			Created: chaos.GetChaos().StartTime.Format(time.RFC3339),
			Status:  string(utils.GetChaosState(chaos)),
			UID:     chaos.GetChaos().UID,
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        chaos.Name,
				Namespace:   chaos.Namespace,
				Labels:      chaos.Labels,
				Annotations: chaos.Annotations,
			},
			Spec: chaos.Spec,
		},
	}, nil
}

// @Summary Get chaos experiments from Kubernetes cluster.
// @Description Get chaos experiments from Kubernetes cluster.
// @Tags experiments
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Param kind query string false "kind" Enums(PodChaos, IOChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Param status query string false "status" Enums(Running, Paused, Failed, Finished)
// @Success 200 {array} Experiment
// @Router /experiments [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listExperiments(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	kind := c.Query("kind")
	name := c.Query("name")
	ns := c.Query("namespace")

	if len(ns) == 0 && !s.conf.ClusterScoped &&
		len(s.conf.TargetNamespace) != 0 {
		ns = s.conf.TargetNamespace
	}

	exps := make([]*Experiment, 0)
	for key, list := range v1alpha1.AllKinds() {
		if kind != "" && key != kind {
			continue
		}
		if err := kubeCli.List(context.Background(), list.ChaosList, &client.ListOptions{Namespace: ns}); err != nil {
			c.Status(http.StatusInternalServerError)
			utils.SetErrorForGinCtx(c, err)
			return
		}

		items := reflect.ValueOf(list.ChaosList).Elem().FieldByName("Items")
		for i := 0; i < items.Len(); i++ {
			item := items.Index(i).Addr().Interface().(v1alpha1.InnerObject)
			chaos := item.GetChaos()
			if name != "" && chaos.Name != name {
				continue
			}
			status := utils.GetChaosState(item)
			exps = append(exps, &Experiment{
				Base: Base{
					Name:      chaos.Name,
					Namespace: chaos.Namespace,
					Kind:      chaos.Kind,
				},
				Created: chaos.StartTime.Format(time.RFC3339),
				Status:  string(status),
				UID:     chaos.UID,
			})
		}
	}

	sort.Slice(exps, func(i, j int) bool {
		return exps[i].Created > exps[j].Created
	})

	c.JSON(http.StatusOK, exps)
}

// @Summary Get detailed information about the specified chaos experiment.
// @Description Get detailed information about the specified chaos experiment.
// @Tags experiments
// @Produce json
// @Param uid path string true "uid"
// @Router /experiments/detail/{uid} [GET]
// @Success 200 {object} Detail
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
func (s *Service) getExperimentDetail(c *gin.Context) {
	var (
		exp       *core.Experiment
		expDetail Detail
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the experiment is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	kind := exp.Kind
	ns := exp.Namespace
	name := exp.Name

	switch kind {
	case v1alpha1.KindPodChaos:
		expDetail, err = s.getPodChaosDetail(ns, name, kubeCli)
	case v1alpha1.KindIOChaos:
		expDetail, err = s.getIOChaosDetail(ns, name, kubeCli)
	case v1alpha1.KindNetworkChaos:
		expDetail, err = s.getNetworkChaosDetail(ns, name, kubeCli)
	case v1alpha1.KindTimeChaos:
		expDetail, err = s.getTimeChaosDetail(ns, name, kubeCli)
	case v1alpha1.KindKernelChaos:
		expDetail, err = s.getKernelChaosDetail(ns, name, kubeCli)
	case v1alpha1.KindStressChaos:
		expDetail, err = s.getStressChaosDetail(ns, name, kubeCli)
	case v1alpha1.KindDNSChaos:
		expDetail, err = s.getDNSChaosDetail(ns, name, kubeCli)
	case v1alpha1.KindAwsChaos:
		expDetail, err = s.getAwsChaosDetail(ns, name, kubeCli)
	case v1alpha1.KindGcpChaos:
		expDetail, err = s.getGcpChaosDetail(ns, name, kubeCli)
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
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/{uid} [delete]
func (s *Service) deleteExperiment(c *gin.Context) {
	var (
		chaosKind *v1alpha1.ChaosKind
		ok        bool
		exp       *core.Experiment
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	uid := c.Param("uid")
	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the experiment is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		}
		return
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
	if err := kubeCli.Get(ctx, chaosKey, chaosKind.Chaos); err != nil {
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
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return setAnnotation(kubeCli, kind, ns, name)
		})
		if err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("forced deletion of chaos failed, because update chaos annotation error")))
			return
		}
	}

	if err := kubeCli.Delete(ctx, chaosKind.Chaos, &client.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrNotFound.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}
		return
	}

	c.JSON(http.StatusOK, StatusResponse{Status: "success"})
}

// @Summary Delete the specified chaos experiment.
// @Description Delete the specified chaos experiment.
// @Tags experiments
// @Produce json
// @Param uids query string true "uids"
// @Param force query string true "force" Enums(true, false)
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments [delete]
func (s *Service) batchDeleteExperiment(c *gin.Context) {
	var (
		chaosKind *v1alpha1.ChaosKind
		ok        bool
		exp       *core.Experiment
		errFlag   bool
		uidSlice  []string
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	uids := c.Query("uids")
	if uids == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("uids cannot be empty")))
		return
	}
	force := c.DefaultQuery("force", "false")
	uidSlice = strings.Split(uids, ",")
	errFlag = false

	if len(uidSlice) > 100 {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("too many uids, please reduce the number of uids")))
		return
	}

	for _, uid := range uidSlice {
		if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
			if gorm.IsRecordNotFoundError(err) {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because the experiment is not found", uid)))
			} else {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because %s", uid, err.Error())))
			}
			errFlag = true
			continue
		}

		kind := exp.Kind
		ns := exp.Namespace
		name := exp.Name

		ctx := context.TODO()
		chaosKey := types.NamespacedName{Namespace: ns, Name: name}

		if chaosKind, ok = v1alpha1.AllKinds()[kind]; !ok {
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because kind (%s) is not supported", uid, kind)))
			errFlag = true
			continue
		}
		if err := kubeCli.Get(ctx, chaosKey, chaosKind.Chaos); err != nil {
			if apierrors.IsNotFound(err) {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because the chaos is not found", uid)))
			} else {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because %s", uid, err.Error())))
			}
			errFlag = true
			continue
		}

		if force == "true" {
			err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return setAnnotation(kubeCli, kind, ns, name)
			})
			if err != nil {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("forced delete experiment uid (%s) error, because update chaos annotation error", uid)))
				errFlag = true
				continue
			}
		}

		if err := kubeCli.Delete(ctx, chaosKind.Chaos, &client.DeleteOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because the chaos is not found", uid)))
			} else {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because %s", uid, err.Error())))
			}
			errFlag = true
			continue
		}
	}
	if errFlag {
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, StatusResponse{Status: "success"})
	}
}

// @Summary Pause a chaos experiment.
// @Description Pause a chaos experiment.
// @Tags experiments
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/pause/{uid} [put]
func (s *Service) pauseExperiment(c *gin.Context) {
	var experiment *core.Experiment

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	uid := c.Param("uid")
	if experiment, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the experiment is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	exp := &Base{
		Kind:      experiment.Kind,
		Name:      experiment.Name,
		Namespace: experiment.Namespace,
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "true",
	}
	if err := s.patchExperiment(exp, annotations, kubeCli); err != nil {
		if apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrNotFound.WrapWithNoMessage(err))
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, StatusResponse{Status: "success"})
}

// @Summary Start a chaos experiment.
// @Description Start a chaos experiment.
// @Tags experiments
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/start/{uid} [put]
func (s *Service) startExperiment(c *gin.Context) {
	var experiment *core.Experiment

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	uid := c.Param("uid")
	if experiment, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the experiment is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	exp := &Base{
		Kind:      experiment.Kind,
		Name:      experiment.Name,
		Namespace: experiment.Namespace,
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "false",
	}

	if err := s.patchExperiment(exp, annotations, kubeCli); err != nil {
		if apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrNotFound.WrapWithNoMessage(err))
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, StatusResponse{Status: "success"})
}

func (s *Service) patchExperiment(exp *Base, annotations map[string]string, kubeCli client.Client) error {
	var (
		chaosKind *v1alpha1.ChaosKind
		ok        bool
	)

	if chaosKind, ok = v1alpha1.AllKinds()[exp.Kind]; !ok {
		return fmt.Errorf("%s is not supported", exp.Kind)
	}

	key := types.NamespacedName{Namespace: exp.Namespace, Name: exp.Name}
	if err := kubeCli.Get(context.Background(), key, chaosKind.Chaos); err != nil {
		return err
	}

	var mergePatch []byte
	mergePatch, _ = json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": annotations,
		},
	})

	return kubeCli.Patch(context.Background(),
		chaosKind.Chaos,
		client.ConstantPatch(types.MergePatchType, mergePatch))
}

// @Summary Get chaos experiments state from Kubernetes cluster.
// @Description Get chaos experiments state from Kubernetes cluster.
// @Tags experiments
// @Produce json
// @Param namespace query string false "namespace"
// @Success 200 {object} ChaosState
// @Router /experiments/state [get]
// @Failure 500 {object} utils.APIError
func (s *Service) state(c *gin.Context) {
	var (
		err error
	)
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	namespace := c.Query("namespace")
	if len(namespace) == 0 && !s.conf.ClusterScoped &&
		len(s.conf.TargetNamespace) != 0 {
		namespace = s.conf.TargetNamespace
	}

	states := new(ChaosState)

	g, ctx := errgroup.WithContext(context.Background())
	m := &sync.Mutex{}
	kinds := v1alpha1.AllKinds()

	var listOptions []client.ListOption
	if !s.conf.ClusterScoped {
		listOptions = append(listOptions, &client.ListOptions{Namespace: s.conf.TargetNamespace})
	} else if len(namespace) != 0 {
		listOptions = append(listOptions, &client.ListOptions{Namespace: namespace})
	}

	for index := range kinds {
		list := kinds[index]
		g.Go(func() error {
			if err := kubeCli.List(ctx, list.ChaosList, listOptions...); err != nil {
				return err
			}
			m.Lock()

			items := reflect.ValueOf(list.ChaosList).Elem().FieldByName("Items")
			for i := 0; i < items.Len(); i++ {
				item := items.Index(i).Addr().Interface().(v1alpha1.InnerObject)
				state := utils.GetChaosState(item)
				if err != nil {
					c.Status(http.StatusInternalServerError)
					_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
					return err
				}
				switch state {
				case utils.Paused:
					states.Paused++
				case utils.Running:
					states.Running++
				case utils.Injecting:
					states.Injecting++
				case utils.Finished:
					states.Finished++
				}
			}

			m.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		c.Status(http.StatusInternalServerError)
		utils.SetErrorForGinCtx(c, err)
		return
	}

	c.JSON(http.StatusOK, states)
}

// @Summary Update a chaos experiment.
// @Description Update a chaos experiment.
// @Tags experiments
// @Produce json
// @Param request body core.KubeObjectDesc true "Request body"
// @Success 200 {object} core.KubeObjectDesc
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/update [put]
func (s *Service) updateExperiment(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	exp := &core.KubeObjectDesc{}
	if err := c.ShouldBindJSON(exp); err != nil {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	updateFuncs := map[string]updateExperimentFunc{
		v1alpha1.KindPodChaos:     s.updatePodChaos,
		v1alpha1.KindNetworkChaos: s.updateNetworkChaos,
		v1alpha1.KindIOChaos:      s.updateIOChaos,
		v1alpha1.KindStressChaos:  s.updateStressChaos,
		v1alpha1.KindTimeChaos:    s.updateTimeChaos,
		v1alpha1.KindKernelChaos:  s.updateKernelChaos,
		v1alpha1.KindDNSChaos:     s.updateDNSChaos,
		v1alpha1.KindAwsChaos:     s.updateAwsChaos,
		v1alpha1.KindGcpChaos:     s.updateGcpChaos,
	}

	f, ok := updateFuncs[exp.Kind]
	if !ok {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New(exp.Kind + " is not supported"))
		return
	}
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return f(exp, kubeCli)
	})
	if err != nil {
		if apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrNotFound.WrapWithNoMessage(err))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}
		return
	}
	c.JSON(http.StatusOK, exp)
}

func (s *Service) updatePodChaos(exp *core.KubeObjectDesc, kubeCli client.Client) error {
	chaos := &v1alpha1.PodChaos{}
	meta := &exp.Meta
	key := types.NamespacedName{Namespace: meta.Namespace, Name: meta.Name}

	if err := kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(meta.Labels)
	chaos.SetAnnotations(meta.Annotations)

	// TODO: update chaos

	return kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateNetworkChaos(exp *core.KubeObjectDesc, kubeCli client.Client) error {
	chaos := &v1alpha1.NetworkChaos{}
	meta := &exp.Meta
	key := types.NamespacedName{Namespace: meta.Namespace, Name: meta.Name}

	if err := kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(meta.Labels)
	chaos.SetAnnotations(meta.Annotations)

	// TODO: update chaos

	return kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateIOChaos(exp *core.KubeObjectDesc, kubeCli client.Client) error {
	chaos := &v1alpha1.IOChaos{}
	meta := &exp.Meta
	key := types.NamespacedName{Namespace: meta.Namespace, Name: meta.Name}

	if err := kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(meta.Labels)
	chaos.SetAnnotations(meta.Annotations)

	// TODO: update chaos

	return kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateKernelChaos(exp *core.KubeObjectDesc, kubeCli client.Client) error {
	chaos := &v1alpha1.KernelChaos{}
	meta := &exp.Meta
	key := types.NamespacedName{Namespace: meta.Namespace, Name: meta.Name}

	if err := kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(meta.Labels)
	chaos.SetAnnotations(meta.Annotations)

	// TODO: update chaos

	return kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateTimeChaos(exp *core.KubeObjectDesc, kubeCli client.Client) error {
	chaos := &v1alpha1.TimeChaos{}
	meta := &exp.Meta
	key := types.NamespacedName{Namespace: meta.Namespace, Name: meta.Name}

	if err := kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(meta.Labels)
	chaos.SetAnnotations(meta.Annotations)

	// TODO: update chaos

	return kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateStressChaos(exp *core.KubeObjectDesc, kubeCli client.Client) error {
	chaos := &v1alpha1.StressChaos{}
	meta := &exp.Meta
	key := types.NamespacedName{Namespace: meta.Namespace, Name: meta.Name}

	if err := kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(meta.Labels)
	chaos.SetAnnotations(meta.Annotations)

	// TODO: update chaos

	return kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateDNSChaos(exp *core.KubeObjectDesc, kubeCli client.Client) error {
	chaos := &v1alpha1.DNSChaos{}
	meta := &exp.Meta
	key := types.NamespacedName{Namespace: meta.Namespace, Name: meta.Name}

	if err := kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(meta.Labels)
	chaos.SetAnnotations(meta.Annotations)

	// TODO: update chaos

	return kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateAwsChaos(exp *core.KubeObjectDesc, kubeCli client.Client) error {
	chaos := &v1alpha1.AwsChaos{}
	meta := &exp.Meta
	key := types.NamespacedName{Namespace: meta.Namespace, Name: meta.Name}

	if err := kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(meta.Labels)
	chaos.SetAnnotations(meta.Annotations)

	// TODO: update chaos

	return kubeCli.Update(context.Background(), chaos)
}

func (s *Service) updateGcpChaos(exp *core.KubeObjectDesc, kubeCli client.Client) error {
	chaos := &v1alpha1.AwsChaos{}
	meta := &exp.Meta
	key := types.NamespacedName{Namespace: meta.Namespace, Name: meta.Name}

	if err := kubeCli.Get(context.Background(), key, chaos); err != nil {
		return err
	}

	chaos.SetLabels(meta.Labels)
	chaos.SetAnnotations(meta.Annotations)

	// TODO: update chaos

	return kubeCli.Update(context.Background(), chaos)
}

func setAnnotation(kubeCli client.Client, kind string, ns string, name string) error {
	var (
		chaosKind *v1alpha1.ChaosKind
		chaosMeta metav1.Object
		ok        bool
	)

	if chaosKind, ok = v1alpha1.AllKinds()[kind]; !ok {
		return fmt.Errorf(kind + " is not supported")
	}
	ctx := context.TODO()
	chaosKey := types.NamespacedName{Namespace: ns, Name: name}
	if err := kubeCli.Get(ctx, chaosKey, chaosKind.Chaos); err != nil {
		return err
	}

	if chaosMeta, ok = chaosKind.Chaos.(metav1.Object); !ok {
		return fmt.Errorf("failed to get chaos meta information")
	}
	annotations := chaosMeta.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[finalizers.AnnotationCleanFinalizer] = finalizers.AnnotationCleanFinalizerForced
	chaosMeta.SetAnnotations(annotations)

	return kubeCli.Update(context.Background(), chaosKind.Chaos)
}

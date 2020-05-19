// Copyright 2020 PingCAP, Inc.
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
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	statuscode "github.com/pingcap/chaos-mesh/pkg/apiserver/status_code"
	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/core"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = ctrl.Log.WithName("experiment api")

// Service defines a handler service for experiments.
type Service struct {
	conf    *config.ChaosServerConfig
	kubeCli client.Client
	archive core.ExperimentStore
	event   core.EventStore
}

// NewService returns an experiment service instance.
func NewService(
	conf *config.ChaosServerConfig,
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
	endpoint := r.Group("/experiment")

	// TODO: add more api handlers
	endpoint.GET("/all", s.listExperiments)
	endpoint.POST("/new", s.createExperiment)
	endpoint.DELETE("/delete/:ns/:name", s.deleteExperiment)
	endpoint.GET("/detail/:name", s.getExperimentDetail)
	endpoint.GET("/state", s.state)
}

// TODO: need to be implemented
func (s *Service) listExperiments(c *gin.Context) {}

// TODO: need to be implemented
func (s *Service) getExperimentDetail(c *gin.Context) {}

// TODO: need to be implemented
func (s *Service) deleteExperiment(c *gin.Context) {}

// ExperimentInfo defines a form data of Experiment from API.
type ExperimentInfo struct {
	Name      string        `json:"name" binding:"required,NameValid"`
	Namespace string        `json:"namespace" binding:"required,NameValid"`
	Scope     ScopeInfo     `json:"scope"`
	Target    TargetInfo    `json:"target"`
	Scheduler SchedulerInfo `json:"scheduler"`
}

// ScopeInfo defines the scope of the Experiment.
type ScopeInfo struct {
	NamespaceSelectors  []string          `json:"namespace_selectors" binding:"NamespaceSelectorsValid"`
	LabelSelectors      map[string]string `json:"label_selectors" binding:"MapSelectorsValid"`
	AnnotationSelectors map[string]string `json:"annotation_selectors" binding:"MapSelectorsValid"`
	FieldSelectors      map[string]string `json:"field_selectors" binding:"MapSelectorsValid"`
	PhaseSelector       []string          `json:"phase_selectors" binding:"PhaseSelectorsValid"`

	Mode string `json:"mode" binding:"required,oneof=one all fixed fixed-percent random-max-percent"`
}

func (s *ScopeInfo) parseSelector() v1alpha1.SelectorSpec {
	selector := v1alpha1.SelectorSpec{}

	for _, ns := range s.NamespaceSelectors {
		selector.Namespaces = append(selector.Namespaces, ns)
	}

	for key, val := range s.LabelSelectors {
		selector.LabelSelectors = make(map[string]string)
		selector.LabelSelectors[key] = val
	}

	for key, val := range s.AnnotationSelectors {
		selector.AnnotationSelectors = make(map[string]string)
		selector.AnnotationSelectors[key] = val
	}

	for key, val := range s.FieldSelectors {
		selector.FieldSelectors = make(map[string]string)
		selector.FieldSelectors[key] = val
	}

	for _, ph := range s.PhaseSelector {
		selector.PodPhaseSelectors = append(selector.PodPhaseSelectors, ph)
	}

	return selector
}

// TargetInfo defines the information of target objects.
type TargetInfo struct {
	Kind         string       `json:"kind" binding:"required,oneof=PodChaos NetworkChaos IOChaos KernelChaos TimeChaos StressChaos"`
	PodChaos     PodChaosInfo `json:"pod_chaos"`
	NetworkChaos NetworkChaosInfo
	IOChaos      IOChaosInfo
	KernelChaos  KernelChaosInfo
	TimeChaos    TimeChaosInfo
	StressChaos  StressChaosInfo
}

// SchedulerInfo defines the scheduler information.
type SchedulerInfo struct {
	Cron string `json:"cron" binding:"CronValid"`
}

// PodChaosInfo defines the basic information of pod chaos.
type PodChaosInfo struct {
	Action        string `json:"action" binding:"oneof=pod-kill pod-failure container-kill"`
	ContainerName string `json:"container_name"`
}

// TODO: implement these structs
type NetworkChaosInfo struct{}
type IOChaosInfo struct{}
type KernelChaosInfo struct{}
type TimeChaosInfo struct{}
type StressChaosInfo struct{}

func (s *Service) createExperiment(c *gin.Context) {
	exp := &ExperimentInfo{}
	if err := c.ShouldBindJSON(exp); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"status":  statuscode.InvalidParameter,
		})
		return
	}

	switch exp.Target.Kind {
	case "PodChaos":
		if err := s.createPodChaos(exp); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"message": err.Error(),
				"status":  statuscode.CreateExperimentFailed,
			})
			return
		}
	default:
		c.JSON(http.StatusOK, gin.H{
			"message": "Target Kind is not supported!",
			"status":  statuscode.InvalidParameter,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Create Experiment successfully",
		"status":  statuscode.Success,
	})
}

func (s *Service) createPodChaos(exp *ExperimentInfo) error {
	chaos := &v1alpha1.PodChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:      exp.Name,
			Namespace: exp.Namespace,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector: exp.Scope.parseSelector(),
			Action:   v1alpha1.PodChaosAction(exp.Target.PodChaos.Action),
			Mode:     v1alpha1.PodMode(exp.Scope.Mode),
		},
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	return s.kubeCli.Create(context.Background(), chaos)
}

// getPodChaosState returns the state of PodChaos
func (s *Service) getPodChaosState(stateInfo map[string]int) error {
	var chaosList v1alpha1.PodChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo["Total"]++
		switch chaos.Status.ChaosStatus.Experiment.Phase {
		case v1alpha1.ExperimentPhaseRunning:
			stateInfo["Running"]++
		case v1alpha1.ExperimentPhasePaused:
			stateInfo["Paused"]++
		case v1alpha1.ExperimentPhaseFailed:
			stateInfo["Failed"]++
		case v1alpha1.ExperimentPhaseFinished:
			stateInfo["Finished"]++
		}
	}
	return nil
}

// getIoChaosState returns the state of IoChaos
func (s *Service) getIoChaosState(stateInfo map[string]int) error {
	var chaosList v1alpha1.IoChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo["Total"]++
		switch chaos.Status.ChaosStatus.Experiment.Phase {
		case v1alpha1.ExperimentPhaseRunning:
			stateInfo["Running"]++
		case v1alpha1.ExperimentPhasePaused:
			stateInfo["Paused"]++
		case v1alpha1.ExperimentPhaseFailed:
			stateInfo["Failed"]++
		case v1alpha1.ExperimentPhaseFinished:
			stateInfo["Finished"]++
		}
	}
	return nil
}

// getNetworkChaosState returns the state of NetworkChaos
func (s *Service) getNetworkChaosState(stateInfo map[string]int) error {
	var chaosList v1alpha1.NetworkChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo["Total"]++
		switch chaos.Status.ChaosStatus.Experiment.Phase {
		case v1alpha1.ExperimentPhaseRunning:
			stateInfo["Running"]++
		case v1alpha1.ExperimentPhasePaused:
			stateInfo["Paused"]++
		case v1alpha1.ExperimentPhaseFailed:
			stateInfo["Failed"]++
		case v1alpha1.ExperimentPhaseFinished:
			stateInfo["Finished"]++
		}
	}
	return nil
}

// getTimeChaosState returns the state of TimeChaos
func (s *Service) getTimeChaosState(stateInfo map[string]int) error {
	var chaosList v1alpha1.TimeChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo["Total"]++
		switch chaos.Status.ChaosStatus.Experiment.Phase {
		case v1alpha1.ExperimentPhaseRunning:
			stateInfo["Running"]++
		case v1alpha1.ExperimentPhasePaused:
			stateInfo["Paused"]++
		case v1alpha1.ExperimentPhaseFailed:
			stateInfo["Failed"]++
		case v1alpha1.ExperimentPhaseFinished:
			stateInfo["Finished"]++
		}
	}
	return nil
}

// getKernelChaosState returns the state of KernelChaos
func (s *Service) getKernelChaosState(stateInfo map[string]int) error {
	var chaosList v1alpha1.KernelChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo[string(chaos.Status.ChaosStatus.Experiment.Phase)]++
	}
	stateInfo["Total"] += len(chaosList.Items)

	return nil
}

// getStressChaosState returns the state of StressChaos
func (s *Service) getStressChaosState(stateInfo map[string]int) error {
	var chaosList v1alpha1.StressChaosList
	err := s.kubeCli.List(context.Background(), &chaosList)
	if err != nil {
		return err
	}
	for _, chaos := range chaosList.Items {
		stateInfo["Total"]++
		switch chaos.Status.ChaosStatus.Experiment.Phase {
		case v1alpha1.ExperimentPhaseRunning:
			stateInfo["Running"]++
		case v1alpha1.ExperimentPhasePaused:
			stateInfo["Paused"]++
		case v1alpha1.ExperimentPhaseFailed:
			stateInfo["Failed"]++
		case v1alpha1.ExperimentPhaseFinished:
			stateInfo["Finished"]++
		}
	}
	return nil
}

func (s *Service) state(c *gin.Context) {
	data := make(map[string]int)
	data["Total"] = 0
	data["Running"] = 0
	data["Paused"] = 0
	data["Failed"] = 0
	data["Finished"] = 0
	getChaosWrong := gin.H{
		"status":  statuscode.GetResourcesWrong,
		"message": "failed to get chaos state",
		"data":    make(map[string]int),
	}

	err := s.getPodChaosState(data)
	if err != nil {
		c.JSON(http.StatusOK, getChaosWrong)
		return
	}
	err = s.getIoChaosState(data)
	if err != nil {
		c.JSON(http.StatusOK, getChaosWrong)
		return
	}
	err = s.getNetworkChaosState(data)
	if err != nil {
		c.JSON(http.StatusOK, getChaosWrong)
		return
	}
	err = s.getTimeChaosState(data)
	if err != nil {
		c.JSON(http.StatusOK, getChaosWrong)
		return
	}
	err = s.getKernelChaosState(data)
	if err != nil {
		c.JSON(http.StatusOK, getChaosWrong)
		return
	}
	err = s.getStressChaosState(data)
	if err != nil {
		c.JSON(http.StatusOK, getChaosWrong)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  statuscode.Success,
		"message": "success",
		"data":    data,
	})
}

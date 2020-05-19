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
	"github.com/gin-gonic/gin/binding"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	statuscode "github.com/pingcap/chaos-mesh/pkg/apiserver/status_code"
	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/core"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
	endpoint.GET("/", s.listExperiments)
	endpoint.GET("/detail", s.getExperimentDetail)
	endpoint.DELETE("/delete/:ns/:name", s.deleteExperiment)
	endpoint.POST("/new", s.createExperiment)
}

// TODO: need to be implemented
func (s *Service) listExperiments(c *gin.Context) {}

// TODO: need to be implemented
func (s *Service) getExperimentDetail(c *gin.Context) {}

// TODO: need to be implemented
func (s *Service) deleteExperiment(c *gin.Context) {}

// ExperimentInfo defines a form data of Experiment from API.
type ExperimentInfo struct {
	Name      string        `form:"name" binding:"required,NameValid"`
	Namespace string        `form:"namespace" binding:"required,NameValid"`
	Scope     ScopeInfo     `form:"scope"`
	Target    TargetInfo    `form:"target"`
	Scheduler SchedulerInfo `form:"scheduler"`
}

// Scope defines the scope of the Experiment.
type ScopeInfo struct {
	NamespaceSelectors  []string          `form:"namespace_selectors" binding:"NamespaceSelectorsValid"`
	LabelSelectors      map[string]string `form:"label_selectors" binding:"MapSelectorsValid"`
	AnnotationSelectors map[string]string `form:"annotation_selectors" binding:"MapSelectorsValid"`
	FieldSelectors      map[string]string `form:"field_selectors" binding:"MapSelectorsValid"`
	PhaseSelector       []string          `form:"phase_selectors" binding:"PhaseSelectorsValid"`

	Mode string `form:"mode" binding:"required,oneof=one all fixed fixed-percent random-max-percent"`
}

// TargetInfo defines the information of target objects.
type TargetInfo struct {
	Kind         string `form:"kind" binding:"required,oneof=PodChaos NetworkChaos IOChaos KernelChaos TimeChaos StressChaos"`
	PodChaos     PodChaosInfo
	NetworkChaos PodChaosInfo
	IOChaos      IOChaosInfo
	KernelChaos  KernelChaosInfo
	TimeChaos    TimeChaosInfo
	StressChaos  StressChaosInfo
}

// PodChaosInfo defines the basic information of pod chaos.
type PodChaosInfo struct {
	Action        string `form:"action" binding:"oneof=pod-kill pod-failure container-kill"`
	ContainerName string `form:"container_name"`
}

// TODO: implement these structs
type NetworkChaosInfo struct{}
type IOChaosInfo struct{}
type KernelChaosInfo struct{}
type TimeChaosInfo struct{}
type StressChaosInfo struct{}

type SchedulerInfo struct {
	Cron string `form:"cron" binding:"CronValid"`
}

func (s *Service) createExperiment(c *gin.Context) {
	exp := &ExperimentInfo{}
	if err := c.ShouldBindWith(exp, binding.Query); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Experiment dates are invalid!",
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
			Selector: s.parseSelector(exp.Scope),
			Action:   v1alpha1.PodChaosAction(exp.Target.PodChaos.Action),
			Mode:     v1alpha1.PodMode(exp.Scope.Mode),
		},
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if err := s.kubeCli.Create(context.Background(), chaos); err != nil {
		return err
	}

	return nil
}

func (s *Service) parseSelector(scope ScopeInfo) v1alpha1.SelectorSpec {
	selector := v1alpha1.SelectorSpec{}

	for _, ns := range scope.NamespaceSelectors {
		selector.Namespaces = append(selector.Namespaces, ns)
	}

	for key, val := range scope.LabelSelectors {
		selector.LabelSelectors = make(map[string]string)
		selector.LabelSelectors[key] = val
	}

	for key, val := range scope.AnnotationSelectors {
		selector.AnnotationSelectors = make(map[string]string)
		selector.AnnotationSelectors[key] = val
	}

	for key, val := range scope.FieldSelectors {
		selector.FieldSelectors = make(map[string]string)
		selector.FieldSelectors[key] = val
	}

	for _, ph := range scope.PhaseSelector {
		selector.PodPhaseSelectors = append(selector.PodPhaseSelectors, ph)
	}

	return selector
}

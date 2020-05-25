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
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/pkg/apiserver/utils"
	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/core"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Experiment defines the basic information of an experiment
type Experiment struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Kind      string `json:"kind"`
	Created   string `json:"created"`
	Status    string `json:"status"`
}

// ChaosState defines the number of chaos experiments of each phase
type ChaosState struct {
	Total    int `json:"total"`
	Running  int `json:"running"`
	Paused   int `json:"paused"`
	Failed   int `json:"failed"`
	Finished int `json:"finished"`
}

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
	endpoint := r.Group("/experiments")

	// TODO: add more api handlers
	endpoint.GET("", s.listExperiments)
	endpoint.POST("/new", s.createExperiment)
	endpoint.GET("/detail/:kind/:namespace/:name", s.getExperimentDetail)
	endpoint.DELETE("/:kind/:namespace/:name", s.deleteExperiment)
	endpoint.GET("/state", s.state)
}

// ExperimentInfo defines a form data of Experiment from API.
type ExperimentInfo struct {
	Name        string            `json:"name" binding:"required,NameValid"`
	Namespace   string            `json:"namespace" binding:"required,NameValid"`
	Labels      map[string]string `json:"labels" binding:"MapSelectorsValid"`
	Annotations map[string]string `json:"annotations" binding:"MapSelectorsValid"`
	Scope       ScopeInfo         `json:"scope"`
	Target      TargetInfo        `json:"target"`
	Scheduler   SchedulerInfo     `json:"scheduler"`
}

// ScopeInfo defines the scope of the Experiment.
type ScopeInfo struct {
	NamespaceSelectors  []string          `json:"namespace_selectors" binding:"NamespaceSelectorsValid"`
	LabelSelectors      map[string]string `json:"label_selectors" binding:"MapSelectorsValid"`
	AnnotationSelectors map[string]string `json:"annotation_selectors" binding:"MapSelectorsValid"`
	FieldSelectors      map[string]string `json:"field_selectors" binding:"MapSelectorsValid"`
	PhaseSelector       []string          `json:"phase_selectors" binding:"PhaseSelectorsValid"`

	Mode  string `json:"mode" binding:"oneof='' 'one' 'all' 'fixed' 'fixed' 'fixed-percent' 'random-max-percent'"`
	Value string `json:"value" binding:"ValueValid"`
}

func (s *ScopeInfo) parseSelector() v1alpha1.SelectorSpec {
	selector := v1alpha1.SelectorSpec{}

	for _, ns := range s.NamespaceSelectors {
		selector.Namespaces = append(selector.Namespaces, ns)
	}

	selector.LabelSelectors = make(map[string]string)
	for key, val := range s.LabelSelectors {
		selector.LabelSelectors[key] = val
	}

	selector.AnnotationSelectors = make(map[string]string)
	for key, val := range s.AnnotationSelectors {
		selector.AnnotationSelectors[key] = val
	}

	selector.FieldSelectors = make(map[string]string)
	for key, val := range s.FieldSelectors {
		selector.FieldSelectors[key] = val
	}

	for _, ph := range s.PhaseSelector {
		selector.PodPhaseSelectors = append(selector.PodPhaseSelectors, ph)
	}

	return selector
}

// TargetInfo defines the information of target objects.
type TargetInfo struct {
	Kind         string           `json:"kind" binding:"required,oneof=PodChaos NetworkChaos IoChaos KernelChaos TimeChaos StressChaos"`
	PodChaos     PodChaosInfo     `json:"pod_chaos"`
	NetworkChaos NetworkChaosInfo `json:"network_chaos"`
	IOChaos      IOChaosInfo      `json:"io_chaos"`
	KernelChaos  KernelChaosInfo  `json:"kernel_chaos"`
	TimeChaos    TimeChaosInfo    `json:"time_chaos"`
	StressChaos  StressChaosInfo  `json:"stress_chaos"`
}

// SchedulerInfo defines the scheduler information.
type SchedulerInfo struct {
	Cron     string `json:"cron" binding:"CronValid"`
	Duration string `json:"duration" binding:"DurationValid"`
}

// PodChaosInfo defines the basic information of pod chaos for creating a new PodChaos.
type PodChaosInfo struct {
	Action        string `json:"action" binding:"oneof='' 'pod-kill' 'pod-failure' 'container-kill'"`
	ContainerName string `json:"container_name"`
}

// PodChaosInfo defines the basic information of network chaos for creating a new NetworkChaos.
type NetworkChaosInfo struct {
	Action      string                  `json:"action" binding:"oneof='' 'netem' 'delay' 'loss' 'duplicate' 'corrupt' 'partition' 'bandwidth'"`
	Delay       *v1alpha1.DelaySpec     `json:"delay"`
	Loss        *v1alpha1.LossSpec      `json:"loss"`
	Duplicate   *v1alpha1.DuplicateSpec `json:"duplicate"`
	Corrupt     *v1alpha1.CorruptSpec   `json:"corrupt"`
	Bandwidth   *v1alpha1.BandwidthSpec `json:"bandwidth"`
	Direction   string                  `json:"direction" binding:"oneof='' 'to' 'from' 'both'"`
	TargetScope *ScopeInfo              `json:"target_scope"`
}

// IOChaosInfo defines the basic information of io chaos for creating a new IOChaos.
type IOChaosInfo struct {
	Action  string   `json:"action" binding:"oneof='' 'delay' 'errno' 'mixed'"`
	Addr    string   `json:"addr"`
	Delay   string   `json:"delay"`
	Errno   string   `json:"errno"`
	Path    string   `json:"path"`
	Percent string   `json:"percent"`
	Methods []string `json:"methods"`
}

// KernelChaosInfo defines the basic information of kernel chaos for creating a new KernelChaos.
type KernelChaosInfo struct {
	FailKernRequest v1alpha1.FailKernRequest `json:"fail_kernel_req"`
}

// TimeChaosInfo defines the basic information of time chaos for creating a new TimeChaos.
type TimeChaosInfo struct {
	TimeOffset     string   `json:"offset"`
	ClockIDs       []string `json:"clock_ids"`
	ContainerNames []string `json:"container_names"`
}

// StressChaosInfo defines the basic information of stress chaos for creating a new StressChaos.
type StressChaosInfo struct {
	Stressors         *v1alpha1.Stressors `json:"stressors"`
	StressngStressors string              `json:"stressng_stressors,omitempty"`
}

// @Summary Create a nex chaos experiments.
// @Description Create a new chaos experiments.
// @Tags experiments
// @Produce json
// @Param request body ExperimentInfo true "Request body"
// @Success 200 "create ok"
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /api/experiments/new [post]
func (s *Service) createExperiment(c *gin.Context) {
	exp := &ExperimentInfo{}
	if err := c.ShouldBindJSON(exp); err != nil {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	switch exp.Target.Kind {
	case v1alpha1.KindPodChaos:
		if err := s.createPodChaos(exp); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
	case v1alpha1.KindNetworkChaos:
		if err := s.createNetworkChaos(exp); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
	case v1alpha1.KindIOChaos:
		if err := s.createIOChaos(exp); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
	case v1alpha1.KindTimeChaos:
		if err := s.createTimeChaos(exp); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
	case v1alpha1.KindKernelChaos:
		if err := s.createKernelChaos(exp); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
	case v1alpha1.KindStressChaos:
		if err := s.createStressChaos(exp); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
	default:
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("Target kind is not available"))
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (s *Service) createPodChaos(exp *ExperimentInfo) error {
	chaos := &v1alpha1.PodChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector:      exp.Scope.parseSelector(),
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

func (s *Service) createNetworkChaos(exp *ExperimentInfo) error {
	chaos := &v1alpha1.NetworkChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.NetworkChaosSpec{
			Selector:  exp.Scope.parseSelector(),
			Action:    v1alpha1.NetworkChaosAction(exp.Target.NetworkChaos.Action),
			Mode:      v1alpha1.PodMode(exp.Scope.Mode),
			Value:     exp.Scope.Value,
			Delay:     exp.Target.NetworkChaos.Delay,
			Loss:      exp.Target.NetworkChaos.Loss,
			Duplicate: exp.Target.NetworkChaos.Duplicate,
			Corrupt:   exp.Target.NetworkChaos.Corrupt,
			Bandwidth: exp.Target.NetworkChaos.Bandwidth,
			Direction: v1alpha1.Direction(exp.Target.NetworkChaos.Direction),
		},
	}

	if exp.Scheduler.Cron != "" {
		chaos.Spec.Scheduler = &v1alpha1.SchedulerSpec{Cron: exp.Scheduler.Cron}
	}

	if exp.Scheduler.Duration != "" {
		chaos.Spec.Duration = &exp.Scheduler.Duration
	}

	if exp.Target.NetworkChaos.TargetScope != nil {
		chaos.Spec.Target = &v1alpha1.Target{
			TargetSelector: exp.Target.NetworkChaos.TargetScope.parseSelector(),
			TargetMode:     v1alpha1.PodMode(exp.Target.NetworkChaos.TargetScope.Mode),
			TargetValue:    exp.Target.NetworkChaos.TargetScope.Value,
		}
	}

	return s.kubeCli.Create(context.Background(), chaos)
}

func (s *Service) createIOChaos(exp *ExperimentInfo) error {
	chaos := &v1alpha1.IoChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.IoChaosSpec{
			Selector: exp.Scope.parseSelector(),
			Action:   v1alpha1.IOChaosAction(exp.Target.IOChaos.Action),
			Mode:     v1alpha1.PodMode(exp.Scope.Mode),
			Value:    exp.Scope.Value,
			// TODO: don't hardcode after we support other layers
			Layer:   v1alpha1.FileSystemLayer,
			Addr:    exp.Target.IOChaos.Addr,
			Delay:   exp.Target.IOChaos.Delay,
			Errno:   exp.Target.IOChaos.Errno,
			Path:    exp.Target.IOChaos.Path,
			Percent: exp.Target.IOChaos.Percent,
			Methods: exp.Target.IOChaos.Methods,
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

func (s *Service) createTimeChaos(exp *ExperimentInfo) error {
	chaos := &v1alpha1.TimeChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.TimeChaosSpec{
			Selector:       exp.Scope.parseSelector(),
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

func (s *Service) createKernelChaos(exp *ExperimentInfo) error {
	chaos := &v1alpha1.KernelChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.KernelChaosSpec{
			Selector:        exp.Scope.parseSelector(),
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

func (s *Service) createStressChaos(exp *ExperimentInfo) error {
	chaos := &v1alpha1.StressChaos{
		ObjectMeta: v1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.StressChaosSpec{
			Selector:          exp.Scope.parseSelector(),
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

	return s.kubeCli.Create(context.Background(), chaos)
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
// @Router /api/experiments [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listExperiments(c *gin.Context) {
	kind := c.Query("kind")
	name := c.Query("name")
	ns := c.Query("namespace")
	status := c.Query("status")

	data := make([]*Experiment, 0)
	for key, list := range v1alpha1.Kinds {
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
				Name:      chaos.Name,
				Namespace: chaos.Namespace,
				Kind:      chaos.Kind,
				Created:   chaos.StartTime.Format(time.RFC3339),
				Status:    chaos.Status,
			})
		}
	}

	c.JSON(http.StatusOK, data)
}

// @Summary Get detailed information about the specified chaos experiment.
// @Description Get detailed information about the specified chaos experiment.
// @Tags experiments
// @Produce json
// @Param namespace path string true "namespace"
// @Param name path string true "name"
// @Param kind path string true "kind" Enums(PodChaos, IoChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Success 200 "delete ok"
// @Router /api/experiments/detail/{kind}/{namespace}/{name} [GET]
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
func (s *Service) getExperimentDetail(c *gin.Context) {
	kind := c.Param("kind")
	ns := c.Param("namespace")
	name := c.Param("name")

	ctx := context.TODO()
	chaosKey := types.NamespacedName{Namespace: ns, Name: name}

	var (
		chaosKind *v1alpha1.ChaosKind
		ok        bool
	)
	if chaosKind, ok = v1alpha1.Kinds[kind]; !ok {
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

	c.JSON(http.StatusOK, nil)
}

// @Summary Delete the specified chaos experiment.
// @Description Delete the specified chaos experiment.
// @Tags experiments
// @Produce json
// @Param namespace path string true "namespace"
// @Param name path string true "name"
// @Param kind path string true "kind" Enums(PodChaos, IoChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Success 200 "delete ok"
// @Router /api/experiments/{kind}/{namespace}/{name} [delete]
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
func (s *Service) deleteExperiment(c *gin.Context) {
	kind := c.Param("kind")
	ns := c.Param("namespace")
	name := c.Param("name")

	ctx := context.TODO()
	chaosKey := types.NamespacedName{Namespace: ns, Name: name}

	var (
		chaosKind *v1alpha1.ChaosKind
		ok        bool
	)
	if chaosKind, ok = v1alpha1.Kinds[kind]; !ok {
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
// @Router /api/experiments/state [get]
// @Failure 500 {object} utils.APIError
func (s *Service) state(c *gin.Context) {
	data := new(ChaosState)

	g, ctx := errgroup.WithContext(context.Background())
	m := &sync.Mutex{}
	for index := range v1alpha1.Kinds {
		list := v1alpha1.Kinds[index].ChaosList
		g.Go(func() error {
			if err := s.kubeCli.List(ctx, list); err != nil {
				return err
			}
			m.Lock()
			for _, chaos := range list.ListChaos() {
				switch chaos.Status {
				case string(v1alpha1.ExperimentPhaseRunning):
					data.Running++
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

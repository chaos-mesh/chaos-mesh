// Copyright 2021 Chaos Mesh Authors.
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

package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	dashboardconfig "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
)

var log = ctrl.Log.WithName("schedule api")

// Service defines a handler service for experiments.
type Service struct {
	schedule core.ScheduleStore
	event    core.EventStore
	conf     *dashboardconfig.ChaosDashboardConfig
	scheme   *runtime.Scheme
}

// NewService returns an experiment service instance.
func NewService(
	schedule core.ScheduleStore,
	event core.EventStore,
	conf *dashboardconfig.ChaosDashboardConfig,
	scheme *runtime.Scheme,
) *Service {
	return &Service{
		schedule: schedule,
		event:    event,
		conf:     conf,
		scheme:   scheme,
	}
}

// Register mounts HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/schedules")

	endpoint.GET("", s.listSchedules)
	endpoint.GET("/:uid", s.getScheduleDetail)
	endpoint.POST("/", s.createSchedule)
	endpoint.PUT("/", s.updateSchedule)
	endpoint.DELETE("/:uid", s.deleteSchedule)
	endpoint.DELETE("/", s.batchDeleteSchedule)
	endpoint.PUT("/pause/:uid", s.pauseSchedule)
	endpoint.PUT("/start/:uid", s.startSchedule)
}

// Base represents the base info of an experiment.
type Base struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// Schedule defines the basic information of a Schedule object
type Schedule struct {
	Base
	UID     string `json:"uid"`
	Created string `json:"created_at"`
	Status  string `json:"status"`
}

// Detail represents an experiment instance.
type Detail struct {
	Schedule
	YAML           core.KubeObjectDesc `json:"kube_object"`
	ExperimentUIDs []string            `json:"experiment_uids"`
}

type parseScheduleFunc func(*core.ScheduleInfo) v1alpha1.ScheduleItem

// StatusResponse defines a common status struct.
type StatusResponse struct {
	Status string `json:"status"`
}

type pauseFlag bool

const (
	PauseSchedule pauseFlag = true
	StartSchedule pauseFlag = false
)

// @Summary Create a new schedule experiment.
// @Description Create a new schedule experiment.
// @Tags schedules
// @Produce json
// @Param request body core.ScheduleInfo true "Request body"
// @Success 200 {object} core.ScheduleInfo
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules [post]
func (s *Service) createSchedule(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	exp := &core.ScheduleInfo{}
	if err := c.ShouldBindJSON(exp); err != nil {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	sch := &v1alpha1.Schedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:        exp.Name,
			Namespace:   exp.Namespace,
			Labels:      exp.Labels,
			Annotations: exp.Annotations,
		},
		Spec: v1alpha1.ScheduleSpec{
			Schedule:          exp.Schedule,
			ConcurrencyPolicy: exp.ConcurrencyPolicy,
			HistoryLimit:      exp.HistoryLimit,
			Type:              v1alpha1.ScheduleTemplateType(exp.Target.Kind),
		},
	}
	if exp.StartingDeadlineSeconds != nil {
		sch.Spec.StartingDeadlineSeconds = exp.StartingDeadlineSeconds
	}

	parseFuncs := map[string]parseScheduleFunc{
		v1alpha1.KindPodChaos:     parsePodChaos,
		v1alpha1.KindNetworkChaos: parseNetworkChaos,
		v1alpha1.KindIOChaos:      parseIOChaos,
		v1alpha1.KindStressChaos:  parseStressChaos,
		v1alpha1.KindTimeChaos:    parseTimeChaos,
		v1alpha1.KindKernelChaos:  parseKernelChaos,
		v1alpha1.KindDNSChaos:     parseDNSChaos,
		v1alpha1.KindAwsChaos:     parseAwsChaos,
		v1alpha1.KindGcpChaos:     parseGcpChaos,
	}

	f, ok := parseFuncs[exp.Target.Kind]
	if !ok {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New(exp.Target.Kind + " is not supported"))
		return
	}
	embedChaos := f(exp)
	sch.Spec.ScheduleItem = embedChaos

	if err := kubeCli.Create(context.Background(), sch); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, exp)
}

func parsePodChaos(exp *core.ScheduleInfo) v1alpha1.ScheduleItem {
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

	if exp.Duration != "" {
		chaos.Spec.Duration = &exp.Duration
	}

	return v1alpha1.ScheduleItem{
		EmbedChaos: v1alpha1.EmbedChaos{PodChaos: &chaos.Spec},
	}
}

func parseNetworkChaos(exp *core.ScheduleInfo) v1alpha1.ScheduleItem {
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

	if exp.Duration != "" {
		chaos.Spec.Duration = &exp.Duration
	}

	return v1alpha1.ScheduleItem{
		EmbedChaos: v1alpha1.EmbedChaos{NetworkChaos: &chaos.Spec},
	}
}

func parseIOChaos(exp *core.ScheduleInfo) v1alpha1.ScheduleItem {
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

	if exp.Duration != "" {
		chaos.Spec.Duration = &exp.Duration
	}

	return v1alpha1.ScheduleItem{
		EmbedChaos: v1alpha1.EmbedChaos{IOChaos: &chaos.Spec},
	}
}

func parseTimeChaos(exp *core.ScheduleInfo) v1alpha1.ScheduleItem {
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

	if exp.Duration != "" {
		chaos.Spec.Duration = &exp.Duration
	}

	return v1alpha1.ScheduleItem{
		EmbedChaos: v1alpha1.EmbedChaos{TimeChaos: &chaos.Spec},
	}
}

func parseKernelChaos(exp *core.ScheduleInfo) v1alpha1.ScheduleItem {
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

	if exp.Duration != "" {
		chaos.Spec.Duration = &exp.Duration
	}

	return v1alpha1.ScheduleItem{
		EmbedChaos: v1alpha1.EmbedChaos{KernelChaos: &chaos.Spec},
	}
}

func parseStressChaos(exp *core.ScheduleInfo) v1alpha1.ScheduleItem {
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
			},
			Stressors:         stressors,
			StressngStressors: exp.Target.StressChaos.StressngStressors,
		},
	}

	if exp.Target.StressChaos.ContainerName != nil {
		chaos.Spec.ContainerSelector.ContainerNames = []string{*exp.Target.StressChaos.ContainerName}
	}

	if exp.Duration != "" {
		chaos.Spec.Duration = &exp.Duration
	}

	return v1alpha1.ScheduleItem{
		EmbedChaos: v1alpha1.EmbedChaos{StressChaos: &chaos.Spec},
	}
}

func parseDNSChaos(exp *core.ScheduleInfo) v1alpha1.ScheduleItem {
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

	if exp.Duration != "" {
		chaos.Spec.Duration = &exp.Duration
	}

	return v1alpha1.ScheduleItem{
		EmbedChaos: v1alpha1.EmbedChaos{DNSChaos: &chaos.Spec},
	}
}

func parseAwsChaos(exp *core.ScheduleInfo) v1alpha1.ScheduleItem {
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

	if exp.Duration != "" {
		chaos.Spec.Duration = &exp.Duration
	}

	return v1alpha1.ScheduleItem{
		EmbedChaos: v1alpha1.EmbedChaos{AwsChaos: &chaos.Spec},
	}
}

func parseGcpChaos(exp *core.ScheduleInfo) v1alpha1.ScheduleItem {
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

	if exp.Duration != "" {
		chaos.Spec.Duration = &exp.Duration
	}

	return v1alpha1.ScheduleItem{
		EmbedChaos: v1alpha1.EmbedChaos{GcpChaos: &chaos.Spec},
	}
}

// @Summary Get chaos schedules from Kubernetes cluster.
// @Description Get chaos schedules from Kubernetes cluster.
// @Tags schedules
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Success 200 {array} Schedule
// @Router /Schedules [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listSchedules(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	name := c.Query("name")
	ns := c.Query("namespace")

	if !s.conf.ClusterScoped {
		log.Info("Overwrite namespace within namespace scoped mode", "origin", ns, "new", s.conf.TargetNamespace)
		ns = s.conf.TargetNamespace
	}

	ScheduleList := v1alpha1.ScheduleList{}
	sches := make([]*Schedule, 0)
	if err := kubeCli.List(context.Background(), &ScheduleList, &client.ListOptions{Namespace: ns}); err != nil {
		c.Status(http.StatusInternalServerError)
		utils.SetErrorForGinCtx(c, err)
		return
	}
	for _, schedule := range ScheduleList.Items {
		if name != "" && schedule.Name != name {
			continue
		}
		sches = append(sches, &Schedule{
			Base: Base{
				Kind:      string(schedule.Spec.Type),
				Namespace: schedule.Namespace,
				Name:      schedule.Name,
			},
			UID:     string(schedule.UID),
			Created: schedule.CreationTimestamp.Format(time.RFC3339),
			Status:  string(utils.GetScheduleState(schedule)),
		})
	}

	sort.Slice(sches, func(i, j int) bool {
		return sches[i].Created > sches[j].Created
	})

	c.JSON(http.StatusOK, sches)
}

// @Summary Get detailed information about the specified schedule experiment.
// @Description Get detailed information about the specified schedule experiment.
// @Tags experiments
// @Produce json
// @Param uid path string true "uid"
// @Router /schedules/{uid} [GET]
// @Success 200 {object} Detail
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
func (s *Service) getScheduleDetail(c *gin.Context) {
	var (
		sch       *core.Schedule
		schDetail Detail
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	uid := c.Param("uid")
	if sch, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the schedule is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	ns := sch.Namespace
	name := sch.Name

	if !s.conf.ClusterScoped && ns != s.conf.TargetNamespace {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("the namespace is not supported in cluster scoped mode"))
		return
	}

	schedule := &v1alpha1.Schedule{}

	scheduleKey := types.NamespacedName{Namespace: ns, Name: name}
	if err := kubeCli.Get(context.Background(), scheduleKey, schedule); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	gvk, err := apiutil.GVKForObject(schedule, s.scheme)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	UIDList := make([]string, 0)
	kind, ok := v1alpha1.AllScheduleItemKinds()[string(schedule.Spec.Type)]
	if !ok {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInvalidRequest.New("the kind is not supported"))
		return
	}
	list := kind.ChaosList.DeepCopyObject()
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"managed-by": schedule.Name},
	})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	err = kubeCli.List(context.Background(), list, &client.ListOptions{
		Namespace:     ns,
		LabelSelector: selector,
	})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	items := reflect.ValueOf(list).Elem().FieldByName("Items")
	for i := 0; i < items.Len(); i++ {
		if schedule.Spec.Type != v1alpha1.ScheduleTypeWorkflow {
			item := items.Index(i).Addr().Interface().(v1alpha1.InnerObject)
			UIDList = append(UIDList, item.GetChaos().UID)
		} else {
			workflow := items.Index(i).Addr().Interface().(*v1alpha1.Workflow)
			UIDList = append(UIDList, string(workflow.UID))
		}
	}

	schDetail = Detail{
		Schedule: Schedule{
			Base: Base{
				Kind:      string(schedule.Spec.Type),
				Namespace: schedule.Namespace,
				Name:      schedule.Name,
			},
			UID:     string(schedule.UID),
			Created: schedule.CreationTimestamp.Format(time.RFC3339),
			Status:  string(utils.GetScheduleState(*schedule)),
		},
		YAML: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				Kind:       gvk.Kind,
				APIVersion: gvk.GroupVersion().String(),
			},
			Meta: core.KubeObjectMeta{
				Name:        schedule.Name,
				Namespace:   schedule.Namespace,
				Labels:      schedule.Labels,
				Annotations: schedule.Annotations,
			},
			Spec: schedule.Spec,
		},
		ExperimentUIDs: UIDList,
	}

	c.JSON(http.StatusOK, schDetail)
}

// @Summary Delete the specified schedule experiment.
// @Description Delete the specified schedule experiment.
// @Tags schedules
// @Produce json
// @Param uid path string true "uid"
// @Param force query string true "force" Enums(true, false)
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /experiments/{uid} [delete]
func (s *Service) deleteSchedule(c *gin.Context) {
	var (
		exp *core.Schedule
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	uid := c.Param("uid")
	if exp, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the experiment is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	ns := exp.Namespace
	name := exp.Name

	ctx := context.TODO()
	scheduleKey := types.NamespacedName{Namespace: ns, Name: name}
	schedule := &v1alpha1.Schedule{}

	if err := kubeCli.Get(ctx, scheduleKey, schedule); err != nil {
		if apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrNotFound.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}
		return
	}

	if err := kubeCli.Delete(ctx, schedule, &client.DeleteOptions{}); err != nil {
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

// @Summary Update a schedule experiment.
// @Description Update a schedule experiment.
// @Tags schedules
// @Produce json
// @Param request body core.KubeObjectDesc true "Request body"
// @Success 200 {object} core.KubeObjectDesc
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules [put]
func (s *Service) updateSchedule(c *gin.Context) {
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

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return s.updateScheduleFun(exp, kubeCli)
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

func (s *Service) updateScheduleFun(exp *core.KubeObjectDesc, kubeCli client.Client) error {
	sch := &v1alpha1.Schedule{}
	meta := &exp.Meta
	key := types.NamespacedName{Namespace: meta.Namespace, Name: meta.Name}

	if err := kubeCli.Get(context.Background(), key, sch); err != nil {
		return err
	}

	sch.SetLabels(meta.Labels)
	sch.SetAnnotations(meta.Annotations)

	var spec v1alpha1.ScheduleSpec
	bytes, err := json.Marshal(exp.Spec)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, &spec); err != nil {
		return err
	}
	sch.Spec = spec

	return kubeCli.Update(context.Background(), sch)
}

// @Summary Delete the specified schedule experiment.
// @Description Delete the specified schedule experiment.
// @Tags schedules
// @Produce json
// @Param uids query string true "uids"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules [delete]
func (s *Service) batchDeleteSchedule(c *gin.Context) {
	var (
		exp      *core.Schedule
		errFlag  bool
		uidSlice []string
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
	uidSlice = strings.Split(uids, ",")
	errFlag = false

	for _, uid := range uidSlice {
		if exp, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
			if gorm.IsRecordNotFoundError(err) {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because the experiment is not found", uid)))
			} else {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because %s", uid, err.Error())))
			}
			errFlag = true
			continue
		}

		ns := exp.Namespace
		name := exp.Name

		ctx := context.TODO()
		scheduleKey := types.NamespacedName{Namespace: ns, Name: name}
		schedule := &v1alpha1.Schedule{}

		if err := kubeCli.Get(ctx, scheduleKey, schedule); err != nil {
			if apierrors.IsNotFound(err) {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because the chaos is not found", uid)))
			} else {
				_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("delete experiment uid (%s) error, because %s", uid, err.Error())))
			}
			errFlag = true
			continue
		}

		if err := kubeCli.Delete(ctx, schedule, &client.DeleteOptions{}); err != nil {
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

// @Summary Pause a schedule object.
// @Description Pause a schedule object.
// @Tags schedules
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules/pause/{uid} [put]
func (s *Service) pauseSchedule(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	uid := c.Param("uid")
	err = s.pauseOrStartSchedule(uid, PauseSchedule, kubeCli)

	if err != nil {
		if gorm.IsRecordNotFoundError(err) || apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrInvalidRequest.New("the schedule is not found"))
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, StatusResponse{Status: "success"})
}

// @Summary Start a schedule object.
// @Description Start a schedule object.
// @Tags schedules
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules/start/{uid} [put]
func (s *Service) startSchedule(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	uid := c.Param("uid")
	err = s.pauseOrStartSchedule(uid, StartSchedule, kubeCli)

	if err != nil {
		if gorm.IsRecordNotFoundError(err) || apierrors.IsNotFound(err) {
			c.Status(http.StatusNotFound)
			_ = c.Error(utils.ErrInvalidRequest.New("the schedule is not found"))
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, StatusResponse{Status: "success"})
}

func (s *Service) pauseOrStartSchedule(uid string, flag pauseFlag, kubeCli client.Client) error {
	var (
		err             error
		schedule        *core.Schedule
		pauseAnnotation string
	)

	if schedule, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
		return err
	}

	exp := &Base{
		Name:      schedule.Name,
		Namespace: schedule.Namespace,
	}

	if flag == PauseSchedule {
		pauseAnnotation = "true"
	} else {
		pauseAnnotation = "false"
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: pauseAnnotation,
	}

	return s.patchSchedule(exp, annotations, kubeCli)
}

func (s *Service) patchSchedule(exp *Base, annotations map[string]string, kubeCli client.Client) error {
	sch := &v1alpha1.Schedule{}
	key := types.NamespacedName{Namespace: exp.Namespace, Name: exp.Name}

	if err := kubeCli.Get(context.Background(), key, sch); err != nil {
		return err
	}

	var mergePatch []byte
	mergePatch, _ = json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": annotations,
		},
	})

	return kubeCli.Patch(context.Background(),
		sch,
		client.ConstantPatch(types.MergePatchType, mergePatch))
}

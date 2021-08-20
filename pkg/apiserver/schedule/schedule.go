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
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	u "github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/status"
)

var log = u.Log.WithName("schedules")

// Service defines a handler service for schedules.
type Service struct {
	schedule core.ScheduleStore
	event    core.EventStore
	conf     *config.ChaosDashboardConfig
}

func NewService(
	schedule core.ScheduleStore,
	event core.EventStore,
	conf *config.ChaosDashboardConfig,
) *Service {
	return &Service{
		schedule: schedule,
		event:    event,
		conf:     conf,
	}
}

// Register schedules RouterGroup.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/schedules")

	endpoint.GET("", s.list)
	endpoint.POST("", s.create)
	endpoint.GET("/:uid", s.get)
	endpoint.DELETE("/:uid", s.delete)
	endpoint.DELETE("", s.batchDelete)
	endpoint.PUT("", s.update)
	endpoint.PUT("/pause/:uid", s.pauseSchedule)
	endpoint.PUT("/start/:uid", s.startSchedule)
}

// Schedule defines the information of a schedule.
type Schedule struct {
	core.ObjectBase
	Status status.ScheduleStatus `json:"status"`
}

// Detail adds KubeObjectDesc on Schedule.
type Detail struct {
	Schedule
	ExperimentUIDs []string            `json:"experiment_uids"`
	KubeObject     core.KubeObjectDesc `json:"kube_object"`
}

// @Summary List chaos schedules.
// @Description Get chaos schedules from k8s cluster in real time.
// @Tags schedules
// @Produce json
// @Param namespace query string false "filter schedules by namespace"
// @Param name query string false "filter schedules by name"
// @Success 200 {array} Schedule
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules [get]
func (s *Service) list(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	ns, name := c.Query("namespace"), c.Query("name")

	if ns == "" && !s.conf.ClusterScoped && s.conf.TargetNamespace != "" {
		ns = s.conf.TargetNamespace

		log.V(1).Info("Replace query namespace with", ns)
	}

	ScheduleList := v1alpha1.ScheduleList{}
	if err = kubeCli.List(context.Background(), &ScheduleList, &client.ListOptions{Namespace: ns}); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}

	sches := make([]*Schedule, 0)
	for _, schedule := range ScheduleList.Items {
		if name != "" && schedule.Name != name {
			continue
		}

		sches = append(sches, &Schedule{
			ObjectBase: core.ObjectBase{
				Namespace: schedule.Namespace,
				Name:      schedule.Name,
				Kind:      string(schedule.Spec.Type),
				UID:       string(schedule.UID),
				Created:   schedule.CreationTimestamp.Format(time.RFC3339),
			},
			Status: status.GetScheduleStatus(schedule),
		})
	}

	sort.Slice(sches, func(i, j int) bool {
		return sches[i].Created > sches[j].Created
	})

	c.JSON(http.StatusOK, sches)
}

// @Summary Create a new schedule.
// @Description Pass a JSON object to create a new schedule. The schema for JSON is the same as the YAML schema for the Kubernetes object.
// @Tags schedules
// @Accept json
// @Produce json
// @Param schedule body v1alpha1.Schedule true "the schedule definition"
// @Success 200 {object} v1alpha1.Schedule
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules [post]
func (s *Service) create(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	var sch v1alpha1.Schedule
	if err = u.ShouldBindBodyWithJSON(c, &sch); err != nil {
		return
	}

	if err = kubeCli.Create(context.Background(), &sch); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, sch)
}

// @Summary Get a schedule.
// @Description Get the schedule's detail by uid.
// @Tags schedules
// @Produce json
// @Param uid path string true "the schedule uid"
// @Success 200 {object} Detail
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules/{uid} [get]
func (s *Service) get(c *gin.Context) {
	var (
		sch       *core.Schedule
		schDetail *Detail
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if sch, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrBadRequest.New("Schedule "+uid+"not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	ns, name := sch.Namespace, sch.Name
	schDetail = s.findScheduleInCluster(c, kubeCli, types.NamespacedName{Namespace: ns, Name: name})
	if schDetail == nil {
		return
	}

	c.JSON(http.StatusOK, schDetail)
}

func (s *Service) findScheduleInCluster(c *gin.Context, kubeCli client.Client, namespacedName types.NamespacedName) *Detail {
	var sch v1alpha1.Schedule

	if err := kubeCli.Get(context.Background(), namespacedName, &sch); err != nil {
		u.SetAPImachineryError(c, err)

		return nil
	}

	UIDList := make([]string, 0)
	schType := string(sch.Spec.Type)
	chaosKind, ok := v1alpha1.AllScheduleItemKinds()[schType]
	if !ok {
		u.SetAPIError(c, u.ErrInternalServer.New("Kind "+schType+" is not supported"))

		return nil
	}

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"managed-by": sch.Name},
	})
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))

		return nil
	}

	chaosList := chaosKind.GenericChaosList.DeepCopyObject()
	err = kubeCli.List(context.Background(), chaosList, &client.ListOptions{
		Namespace:     sch.Namespace,
		LabelSelector: selector,
	})
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))

		return nil
	}

	items := reflect.ValueOf(chaosList).Elem().FieldByName("Items")
	for i := 0; i < items.Len(); i++ {
		if sch.Spec.Type != v1alpha1.ScheduleTypeWorkflow {
			item := items.Index(i).Addr().Interface().(v1alpha1.InnerObject)
			UIDList = append(UIDList, string(item.GetUID()))
		} else {
			workflow := items.Index(i).Addr().Interface().(*v1alpha1.Workflow)
			UIDList = append(UIDList, string(workflow.UID))
		}
	}

	return &Detail{
		Schedule: Schedule{
			ObjectBase: core.ObjectBase{
				Namespace: sch.Namespace,
				Name:      sch.Name,
				Kind:      string(sch.Spec.Type),
				UID:       string(sch.UID),
				Created:   sch.CreationTimestamp.Format(time.RFC3339),
			},
			Status: status.GetScheduleStatus(sch),
		},
		ExperimentUIDs: UIDList,
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.GroupVersion.String(),
				Kind:       v1alpha1.KindSchedule,
			},
			Meta: core.KubeObjectMeta{
				Namespace:   sch.Namespace,
				Name:        sch.Name,
				Labels:      sch.Labels,
				Annotations: sch.Annotations,
			},
			Spec: sch.Spec,
		},
	}
}

// @Summary Delete a schedule.
// @Description Delete the schedule by uid.
// @Tags schedules
// @Produce json
// @Param uid path string true "the schedule uid"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules/{uid} [delete]
func (s *Service) delete(c *gin.Context) {
	var (
		sch *core.Schedule
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if sch, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Schedule "+uid+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	ns, name := sch.Namespace, sch.Name
	if err = checkAndDeleteSchedule(c, kubeCli, types.NamespacedName{Namespace: ns, Name: name}); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

// @Summary Batch delete schedules.
// @Description Batch delete schedules by uids.
// @Tags schedules
// @Produce json
// @Param uids query string true "the schedule uids, split with comma. Example: ?uids=uid1,uid2"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules [delete]
func (s *Service) batchDelete(c *gin.Context) {
	var (
		sch *core.Schedule
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uids := c.Query("uids")
	if uids == "" {
		u.SetAPIError(c, u.ErrInternalServer.New("The uids cannot be empty"))

		return
	}

	uidSlice := strings.Split(uids, ",")

	if len(uidSlice) > 100 {
		u.SetAPIError(c, u.ErrInternalServer.New("Too many uids, please delete less than 100 at a time"))

		return
	}

	for _, uid := range uidSlice {
		if sch, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
			if gorm.IsRecordNotFoundError(err) {
				u.SetAPIError(c, u.ErrNotFound.New("Experiment "+uid+" not found"))
			} else {
				u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
			}

			return
		}

		ns, name := sch.Namespace, sch.Name
		if err = checkAndDeleteSchedule(c, kubeCli, types.NamespacedName{Namespace: ns, Name: name}); err != nil {
			u.SetAPImachineryError(c, err)

			return
		}

	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

func checkAndDeleteSchedule(c *gin.Context, kubeCli client.Client, namespacedName types.NamespacedName) (err error) {
	ctx := context.Background()
	var sch v1alpha1.Schedule

	if err = kubeCli.Get(ctx, namespacedName, &sch); err != nil {
		return
	}

	if err = kubeCli.Delete(ctx, &sch); err != nil {
		return
	}

	return
}

// @Summary Update a schedule.
// @Description Update a schedule.
// @Tags schedules
// @Produce json
// @Param request body v1alpha1.Schedule true "Request body"
// @Success 200 {object} v1alpha1.Schedule
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules [put]
func (s *Service) update(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	var sch v1alpha1.Schedule
	if err = u.ShouldBindBodyWithJSON(c, &sch); err != nil {
		return
	}

	if err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return internalUpdate(kubeCli, &sch)
	}); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, sch)
}

func internalUpdate(kubeCli client.Client, sch *v1alpha1.Schedule) error {
	ns, name := sch.Namespace, sch.Name

	if err := kubeCli.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, sch); err != nil {
		return err
	}

	return kubeCli.Update(context.Background(), sch)
}

// @Summary Pause a schedule.
// @Description Pause a schedule.
// @Tags schedules
// @Produce json
// @Param uid path string true "the schedule uid"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules/pause/{uid} [put]
func (s *Service) pauseSchedule(c *gin.Context) {
	var sch *core.Schedule

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if sch, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment "+uid+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "true",
	}
	if err = patchSchedule(kubeCli, sch, annotations); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}
	c.JSON(http.StatusOK, u.ResponseSuccess)
}

// @Summary Start a schedule.
// @Description Start a schedule.
// @Tags schedules
// @Produce json
// @Param uid path string true "the schedule uid"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /schedules/start/{uid} [put]
func (s *Service) startSchedule(c *gin.Context) {
	var sch *core.Schedule

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if sch, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment "+uid+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	annotations := map[string]string{
		v1alpha1.PauseAnnotationKey: "false",
	}
	if err = patchSchedule(kubeCli, sch, annotations); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}
	c.JSON(http.StatusOK, u.ResponseSuccess)
}

func patchSchedule(kubeCli client.Client, sch *core.Schedule, annotations map[string]string) error {
	var tmp v1alpha1.Schedule

	if err := kubeCli.Get(context.Background(), types.NamespacedName{Namespace: sch.Namespace, Name: sch.Name}, &tmp); err != nil {
		return err
	}

	var mergePatch []byte
	mergePatch, _ = json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": annotations,
		},
	})

	return kubeCli.Patch(context.Background(), &tmp, client.RawPatch(types.MergePatchType, mergePatch))
}

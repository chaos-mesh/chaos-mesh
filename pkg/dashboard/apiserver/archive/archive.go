// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package archive

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/types"
	u "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

// Service defines a handler service for archives.
type Service struct {
	archive         core.ExperimentStore
	archiveSchedule core.ScheduleStore
	event           core.EventStore
	workflowStore   core.WorkflowStore
	conf            *config.ChaosDashboardConfig
}

func NewService(
	archive core.ExperimentStore,
	archiveSchedule core.ScheduleStore,
	event core.EventStore,
	workflowStore core.WorkflowStore,
	conf *config.ChaosDashboardConfig,
) *Service {
	return &Service{
		archive:         archive,
		archiveSchedule: archiveSchedule,
		event:           event,
		workflowStore:   workflowStore,
		conf:            conf,
	}
}

// Register archives RouterGroup.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/archives")
	endpoint.Use(func(c *gin.Context) {
		u.AuthMiddleware(c, s.conf)
	})

	endpoint.GET("", s.list)
	endpoint.GET("/:uid", s.get)
	endpoint.DELETE("/:uid", s.delete)
	endpoint.DELETE("", s.batchDelete)

	endpoint.GET("/schedules", s.listSchedule)
	endpoint.GET("/schedules/:uid", s.detailSchedule)
	endpoint.DELETE("/schedules/:uid", s.deleteSchedule)
	endpoint.DELETE("/schedules", s.batchDeleteSchedule)

	endpoint.GET("/workflows", s.listWorkflow)
	endpoint.GET("/workflows/:uid", s.detailWorkflow)
	endpoint.DELETE("/workflows/:uid", s.deleteWorkflow)
	endpoint.DELETE("/workflows", s.batchDeleteWorkflow)
}

// @Summary Get archived chaos experiments.
// @Description Get archived chaos experiments.
// @Tags archives
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Param kind query string false "kind" Enums(PodChaos, IOChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Success 200 {array} types.Archive
// @Router /archives [get]
// @Failure 500 {object} u.APIError
func (s *Service) list(c *gin.Context) {
	kind := c.Query("kind")
	name := c.Query("name")
	ns := c.Query("namespace")
	if len(ns) == 0 && !s.conf.ClusterScoped &&
		len(s.conf.TargetNamespace) != 0 {
		ns = s.conf.TargetNamespace
	}

	metas, err := s.archive.ListMeta(context.Background(), kind, ns, name, true)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(u.ErrInternalServer.NewWithNoMessage())
		return
	}

	archives := make([]types.Archive, 0)

	for _, meta := range metas {
		archives = append(archives, types.Archive{
			UID:       meta.UID,
			Kind:      meta.Kind,
			Namespace: meta.Namespace,
			Name:      meta.Name,
			Created:   meta.StartTime.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, archives)
}

// @Summary Get an archived chaos experiment.
// @Description Get the archived chaos experiment's detail by uid.
// @Tags archives
// @Produce json
// @Param uid path string true "the archive uid"
// @Success 200 {object} types.ArchiveDetail
// @Failure 404 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /archives/{uid} [get]
func (s *Service) get(c *gin.Context) {
	uid := c.Param("uid")
	exp, err := s.archive.FindByUID(context.Background(), uid)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Experiment "+uid+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	chaos := v1alpha1.AllKinds()[exp.Kind].SpawnObject()
	_ = json.Unmarshal([]byte(exp.Experiment), chaos)

	c.JSON(http.StatusOK, &types.ArchiveDetail{
		Archive: types.Archive{
			UID:       exp.UID,
			Kind:      exp.Kind,
			Name:      exp.Name,
			Namespace: exp.Namespace,
			Created:   exp.StartTime.Format(time.RFC3339),
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: reflect.ValueOf(chaos).Elem().FieldByName("APIVersion").String(),
				Kind:       reflect.ValueOf(chaos).Elem().FieldByName("Kind").String(),
			},
			Meta: core.KubeObjectMeta{
				Namespace:   reflect.ValueOf(chaos).Elem().FieldByName("Namespace").String(),
				Name:        reflect.ValueOf(chaos).Elem().FieldByName("Name").String(),
				Labels:      reflect.ValueOf(chaos).Elem().FieldByName("Labels").Interface().(map[string]string),
				Annotations: reflect.ValueOf(chaos).Elem().FieldByName("Annotations").Interface().(map[string]string),
			},
			Spec: reflect.ValueOf(chaos).Elem().FieldByName("Spec").Interface(),
		},
	})
}

// @Summary Delete the specified archived experiment.
// @Description Delete the specified archived experiment.
// @Tags archives
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} u.Response
// @Failure 500 {object} u.APIError
// @Router /archives/{uid} [delete]
func (s *Service) delete(c *gin.Context) {
	var (
		err error
		exp *core.Experiment
	)

	uid := c.Param("uid")

	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrBadRequest.New("the archived experiment is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		}
		return
	}

	if err = s.archive.Delete(context.Background(), exp); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
	} else {
		if err = s.event.DeleteByUID(context.Background(), uid); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		} else {
			c.JSON(http.StatusOK, u.ResponseSuccess)
		}
	}
}

// @Summary Delete the specified archived experiment.
// @Description Delete the specified archived experiment.
// @Tags archives
// @Produce json
// @Param uids query string true "uids"
// @Success 200 {object} u.Response
// @Failure 500 {object} u.APIError
// @Router /archives [delete]
func (s *Service) batchDelete(c *gin.Context) {
	var (
		err      error
		uidSlice []string
	)

	uids := c.Query("uids")
	if uids == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(errors.New("uids cannot be empty")))
		return
	}
	uidSlice = strings.Split(uids, ",")

	if err = s.archive.DeleteByUIDs(context.Background(), uidSlice); err != nil {
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	if err = s.event.DeleteByUIDs(context.Background(), uidSlice); err != nil {
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

// @Summary Get archived schedule experiments.
// @Description Get archived schedule experiments.
// @Tags archives
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Success 200 {array} types.Archive
// @Router /archives/schedules [get]
// @Failure 500 {object} u.APIError
func (s *Service) listSchedule(c *gin.Context) {
	name := c.Query("name")
	ns := c.Query("namespace")

	metas, err := s.archiveSchedule.ListMeta(context.Background(), ns, name, true)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(u.ErrInternalServer.NewWithNoMessage())
		return
	}

	archives := make([]types.Archive, 0)

	for _, meta := range metas {
		archives = append(archives, types.Archive{
			UID:       meta.UID,
			Kind:      meta.Kind,
			Namespace: meta.Namespace,
			Name:      meta.Name,
			Created:   meta.StartTime.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, archives)
}

// @Summary Get the detail of an archived schedule experiment.
// @Description Get the detail of an archived schedule experiment.
// @Tags archives
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} types.ArchiveDetail
// @Failure 500 {object} u.APIError
// @Router /archives/schedules/{uid} [get]
func (s *Service) detailSchedule(c *gin.Context) {
	var (
		err    error
		detail types.ArchiveDetail
	)
	uid := c.Param("uid")

	if uid == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(u.ErrBadRequest.New("uid cannot be empty"))
		return
	}

	exp, err := s.archiveSchedule.FindByUID(context.Background(), uid)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrBadRequest.New("the archive schedule is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	sch := &v1alpha1.Schedule{}
	if err := json.Unmarshal([]byte(exp.Schedule), &sch); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	detail = types.ArchiveDetail{
		Archive: types.Archive{
			UID:       exp.UID,
			Kind:      exp.Kind,
			Name:      exp.Name,
			Namespace: exp.Namespace,
			Created:   exp.StartTime.Format(time.RFC3339),
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: sch.APIVersion,
				Kind:       sch.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        sch.Name,
				Namespace:   sch.Namespace,
				Labels:      sch.Labels,
				Annotations: sch.Annotations,
			},
			Spec: sch.Spec,
		},
	}

	c.JSON(http.StatusOK, detail)
}

// @Summary Delete the specified archived schedule.
// @Description Delete the specified archived schedule.
// @Tags archives
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} u.Response
// @Failure 500 {object} u.APIError
// @Router /archives/schedules/{uid} [delete]
func (s *Service) deleteSchedule(c *gin.Context) {
	var (
		err error
		exp *core.Schedule
	)

	uid := c.Param("uid")

	if exp, err = s.archiveSchedule.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrBadRequest.New("the archived schedule is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		}
		return
	}

	if err = s.archiveSchedule.Delete(context.Background(), exp); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
	} else {
		if err = s.event.DeleteByUID(context.Background(), uid); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		} else {
			c.JSON(http.StatusOK, u.ResponseSuccess)
		}
	}
}

// @Summary Delete the specified archived schedule.
// @Description Delete the specified archived schedule.
// @Tags archives
// @Produce json
// @Param uids query string true "uids"
// @Success 200 {object} u.Response
// @Failure 500 {object} u.APIError
// @Router /archives/schedules [delete]
func (s *Service) batchDeleteSchedule(c *gin.Context) {
	var (
		err      error
		uidSlice []string
	)

	uids := c.Query("uids")
	if uids == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(errors.New("uids cannot be empty")))
		return
	}
	uidSlice = strings.Split(uids, ",")

	if err = s.archiveSchedule.DeleteByUIDs(context.Background(), uidSlice); err != nil {
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	if err = s.event.DeleteByUIDs(context.Background(), uidSlice); err != nil {
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

// @Summary Get archived workflow.
// @Description Get archived workflow.
// @Tags archives
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Success 200 {array} types.Archive
// @Router /archives/workflows [get]
// @Failure 500 {object} u.APIError
func (s *Service) listWorkflow(c *gin.Context) {
	name := c.Query("name")
	ns := c.Query("namespace")

	metas, err := s.workflowStore.ListMeta(context.Background(), ns, name, true)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(u.ErrInternalServer.NewWithNoMessage())
		return
	}

	archives := make([]types.Archive, 0)

	for _, meta := range metas {
		archives = append(archives, types.Archive{
			UID:       meta.UID,
			Kind:      v1alpha1.KindWorkflow,
			Namespace: meta.Namespace,
			Name:      meta.Name,
			Created:   meta.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, archives)
}

// @Summary Get the detail of an archived workflow.
// @Description Get the detail of an archived workflow.
// @Tags archives
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} types.ArchiveDetail
// @Failure 500 {object} u.APIError
// @Router /archives/workflows/{uid} [get]
func (s *Service) detailWorkflow(c *gin.Context) {
	var (
		err    error
		detail types.ArchiveDetail
	)
	uid := c.Param("uid")

	if uid == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(u.ErrBadRequest.New("uid cannot be empty"))
		return
	}

	meta, err := s.workflowStore.FindByUID(context.Background(), uid)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrBadRequest.New("the archive schedule is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	workflow := &v1alpha1.Workflow{}
	if err := json.Unmarshal([]byte(meta.Workflow), &workflow); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	detail = types.ArchiveDetail{
		Archive: types.Archive{
			UID:       meta.UID,
			Kind:      v1alpha1.KindWorkflow,
			Name:      meta.Name,
			Namespace: meta.Namespace,
			Created:   meta.CreatedAt.Format(time.RFC3339),
		},
		KubeObject: core.KubeObjectDesc{
			TypeMeta: metav1.TypeMeta{
				APIVersion: workflow.APIVersion,
				Kind:       workflow.Kind,
			},
			Meta: core.KubeObjectMeta{
				Name:        workflow.Name,
				Namespace:   workflow.Namespace,
				Labels:      workflow.Labels,
				Annotations: workflow.Annotations,
			},
			Spec: workflow.Spec,
		},
	}

	c.JSON(http.StatusOK, detail)
}

// @Summary Delete the specified archived workflow.
// @Description Delete the specified archived workflow.
// @Tags archives
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} u.Response
// @Failure 500 {object} u.APIError
// @Router /archives/workflows/{uid} [delete]
func (s *Service) deleteWorkflow(c *gin.Context) {
	var (
		err error
	)

	uid := c.Param("uid")

	if err = s.workflowStore.DeleteByUID(context.Background(), uid); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
	} else {
		if err = s.event.DeleteByUID(context.Background(), uid); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		} else {
			c.JSON(http.StatusOK, u.ResponseSuccess)
		}
	}
}

// @Summary Delete the specified archived workflows.
// @Description Delete the specified archived workflows.
// @Tags archives
// @Produce json
// @Param uids query string true "uids"
// @Success 200 {object} u.Response
// @Failure 500 {object} u.APIError
// @Router /archives/workflows [delete]
func (s *Service) batchDeleteWorkflow(c *gin.Context) {
	var (
		err      error
		uidSlice []string
	)

	uids := c.Query("uids")
	if uids == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(errors.New("uids cannot be empty")))
		return
	}
	uidSlice = strings.Split(uids, ",")

	if err = s.workflowStore.DeleteByUIDs(context.Background(), uidSlice); err != nil {
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	if err = s.event.DeleteByUIDs(context.Background(), uidSlice); err != nil {
		_ = c.Error(u.ErrInternalServer.WrapWithNoMessage(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

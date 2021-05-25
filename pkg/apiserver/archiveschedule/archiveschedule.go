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

package archiveschedule

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
)

// Service defines a handler service for archive experiments.
type Service struct {
	archive core.ScheduleStore
	event   core.EventStore
}

// NewService returns an archive experiment service instance.
func NewService(
	archive core.ScheduleStore,
	event core.EventStore,
) *Service {
	return &Service{
		archive: archive,
		event:   event,
	}
}

// StatusResponse defines a common status struct.
type StatusResponse struct {
	Status string `json:"status"`
}

// Register mounts our HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/archiveSchedules")
	endpoint.Use(utils.AuthRequired)

	endpoint.GET("", s.list)
	endpoint.GET("/:uid", s.detail)
	endpoint.DELETE("/:uid", s.delete)
	endpoint.DELETE("/", s.batchDelete)
}

// Archive defines the basic information of an archive.
type Archive struct {
	UID        string    `json:"uid"`
	Kind       string    `json:"kind"`
	Namespace  string    `json:"namespace"`
	Name       string    `json:"name"`
	Action     string    `json:"action"`
	StartTime  time.Time `json:"start_time"`
	FinishTime time.Time `json:"finish_time"`
}

// Detail represents an archive instance.
type Detail struct {
	Archive
	KubeObject core.KubeObjectDesc `json:"kube_object"`
}

// @Summary Get archived schedule experiments.
// @Description Get archived schedule experiments.
// @Tags archiveSchedules
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Success 200 {array} ArchiveSchedules
// @Router /archiveSchedules [get]
// @Failure 500 {object} utils.APIError
func (s *Service) list(c *gin.Context) {
	name := c.Query("name")
	ns := c.Query("namespace")

	metas, err := s.archive.ListMeta(context.Background(), ns, name, true)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		return
	}

	archives := make([]Archive, 0)

	for _, meta := range metas {
		archives = append(archives, Archive{
			UID:        meta.UID,
			Kind:       meta.Kind,
			Namespace:  meta.Namespace,
			Name:       meta.Name,
			Action:     meta.Action,
			StartTime:  meta.StartTime,
			FinishTime: meta.FinishTime,
		})
	}

	c.JSON(http.StatusOK, archives)
}

// @Summary Get the detail of an archived schedule experiment.
// @Description Get the detail of an archived schedule experiment.
// @Tags archiveSchedules
// @Produce json
// @Param uid query string true "uid"
// @Success 200 {object} Detail
// @Router /archiveSchedules/{uid} [get]
// @Failure 500 {object} utils.APIError
func (s *Service) detail(c *gin.Context) {
	var (
		err        error
		detail     Detail
	)
	uid := c.Param("uid")

	if uid == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("uid cannot be empty"))
		return
	}

	exp, err := s.archive.FindByUID(context.Background(), uid)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the archive schedule is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	sch := &v1alpha1.Schedule{}
	if err := json.Unmarshal([]byte(exp.Schedule), &sch); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	detail = Detail{
		Archive: Archive{
			UID:        exp.UID,
			Kind:       exp.Kind,
			Name:       exp.Name,
			Namespace:  exp.Namespace,
			Action:     exp.Action,
			StartTime:  exp.StartTime,
			FinishTime: exp.FinishTime,
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
// @Tags archiveSchedules
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} StatusResponse
// @Failure 500 {object} utils.APIError
// @Router /archiveSchedules/{uid} [delete]
func (s *Service) delete(c *gin.Context) {
	var (
		err error
		exp *core.Schedule
	)

	uid := c.Param("uid")

	if exp, err = s.archive.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the archived schedule is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		}
		return
	}

	if err = s.archive.Delete(context.Background(), exp); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
	} else {
		if err = s.event.DeleteByUID(context.Background(), uid); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		} else {
			c.JSON(http.StatusOK, StatusResponse{Status: "success"})
		}
	}
}

// @Summary Delete the specified archived schedule.
// @Description Delete the specified archived schedule.
// @Tags archiveSchedules
// @Produce json
// @Param uids query string true "uids"
// @Success 200 {object} StatusResponse
// @Failure 500 {object} utils.APIError
// @Router /archiveSchedules [delete]
func (s *Service) batchDelete(c *gin.Context) {
	var (
		err      error
		uidSlice []string
	)

	uids := c.Query("uids")
	if uids == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("uids cannot be empty")))
		return
	}
	uidSlice = strings.Split(uids, ",")

	if err = s.archive.DeleteByUIDs(context.Background(), uidSlice); err != nil {
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		c.Status(http.StatusInternalServerError)
		return
	}
	if err = s.event.DeleteByUIDs(context.Background(), uidSlice); err != nil {
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, StatusResponse{Status: "success"})
}

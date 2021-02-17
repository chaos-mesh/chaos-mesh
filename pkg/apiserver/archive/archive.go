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

package archive

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
)

// Service defines a handler service for archive experiments.
type Service struct {
	archive core.ExperimentStore
	event   core.EventStore
}

// NewService returns an archive experiment service instance.
func NewService(
	archive core.ExperimentStore,
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
	endpoint := r.Group("/archives")
	endpoint.Use(utils.AuthRequired)

	endpoint.GET("", s.list)
	endpoint.GET("/detail", s.detail)
	endpoint.GET("/report", s.report)
	endpoint.DELETE("/:uid", s.delete)
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
	YAML core.ExperimentYAMLDescription `json:"yaml"`
}

// Report defines the report of archive experiments.
type Report struct {
	Meta           *Archive      `json:"meta"`
	Events         []*core.Event `json:"events"`
	TotalTime      string        `json:"total_time"`
	TotalFaultTime string        `json:"total_fault_time"`
}

// @Summary Get archived chaos experiments.
// @Description Get archived chaos experiments.
// @Tags archives
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Param kind query string false "kind" Enums(PodChaos, IoChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Success 200 {array} Archive
// @Router /archives [get]
// @Failure 500 {object} utils.APIError
func (s *Service) list(c *gin.Context) {
	kind := c.Query("kind")
	name := c.Query("name")
	ns := c.Query("namespace")

	metas, err := s.archive.ListMeta(context.Background(), kind, ns, name, true)
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

// @Summary Get the detail of an archived chaos experiment.
// @Description Get the detail of an archived chaos experiment.
// @Tags archives
// @Produce json
// @Param uid query string true "uid"
// @Success 200 {object} Detail
// @Router /archives/detail [get]
// @Failure 500 {object} utils.APIError
func (s *Service) detail(c *gin.Context) {
	var (
		err    error
		yaml   core.ExperimentYAMLDescription
		detail Detail
	)
	uid := c.Query("uid")
	namespace := c.Query("namespace")

	if uid == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("uid cannot be empty"))
		return
	}

	exp, err := s.archive.FindByUID(context.Background(), uid)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the archive is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	if len(namespace) != 0 && exp.Namespace != namespace {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("exp %s belong to namespace %s but not namespace %s", uid, exp.Namespace, namespace))
		return
	}

	switch exp.Kind {
	case v1alpha1.KindPodChaos:
		yaml, err = exp.ParsePodChaos()
	case v1alpha1.KindIoChaos:
		yaml, err = exp.ParseIOChaos()
	case v1alpha1.KindNetworkChaos:
		yaml, err = exp.ParseNetworkChaos()
	case v1alpha1.KindTimeChaos:
		yaml, err = exp.ParseTimeChaos()
	case v1alpha1.KindKernelChaos:
		yaml, err = exp.ParseKernelChaos()
	case v1alpha1.KindStressChaos:
		yaml, err = exp.ParseStressChaos()
	case v1alpha1.KindDNSChaos:
		yaml, err = exp.ParseDNSChaos()
	default:
		err = fmt.Errorf("kind %s is not support", exp.Kind)
	}
	if err != nil {
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
		YAML: yaml,
	}

	c.JSON(http.StatusOK, detail)
}

// @Summary Get the report of an archived chaos experiment.
// @Description Get the report of an archived chaos experiment.
// @Tags archives
// @Produce json
// @Param uid query string true "uid"
// @Success 200 {array} Report
// @Router /archives/report [get]
// @Failure 500 {object} utils.APIError
func (s *Service) report(c *gin.Context) {
	var (
		err    error
		report Report
	)
	uid := c.Query("uid")
	namespace := c.Query("namespace")

	if uid == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("uid cannot be empty"))
		return
	}

	meta, err := s.archive.FindMetaByUID(context.Background(), uid)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the archive is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	if len(namespace) != 0 && meta.Namespace != namespace {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("exp %s belong to namespace %s but not namespace %s", uid, meta.Namespace, namespace))
		return
	}

	report.Meta = &Archive{
		UID:        meta.UID,
		Kind:       meta.Kind,
		Namespace:  meta.Namespace,
		Name:       meta.Name,
		Action:     meta.Action,
		StartTime:  meta.StartTime,
		FinishTime: meta.FinishTime,
	}

	report.Events, err = s.event.ListByUID(context.TODO(), uid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		return
	}

	report.TotalTime = report.Meta.FinishTime.Sub(report.Meta.StartTime).String()

	timeNow := time.Now()
	timeAfter := timeNow
	for _, et := range report.Events {
		timeAfter = timeAfter.Add(et.FinishTime.Sub(*et.StartTime))
	}
	report.TotalFaultTime = timeAfter.Sub(timeNow).String()

	c.JSON(http.StatusOK, report)
}

// @Summary Delete the specified archived experiment.
// @Description Delete the specified archived experiment.
// @Tags archives
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} StatusResponse
// @Failure 500 {object} utils.APIError
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
			_ = c.Error(utils.ErrInvalidRequest.New("the archived experiment is not found"))
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

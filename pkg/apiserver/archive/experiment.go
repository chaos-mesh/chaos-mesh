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
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Report defines the report of archive experiments.
type Report struct {
	Meta           *core.ArchiveExperimentMeta
	Events         []*core.Event
	TotalTime      string
	TotalFaultTime string
}

// Service defines a handler service for archive experiments.
type Service struct {
	conf    *config.ChaosDashboardConfig
	kubeCli client.Client
	archive core.ExperimentStore
	event   core.EventStore
}

// NewService returns an archive experiment service instance.
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
	endpoint := r.Group("/archives")

	// TODO: add more api handlers
	endpoint.GET("", s.listExperiments)
	endpoint.GET("/detail/search", s.experimentDetailSearch)
	endpoint.GET("/detail", s.experimentDetail)
	endpoint.GET("/report", s.experimentReport)
}

// ArchiveExperimentDetail represents an experiment instance.
type ArchiveExperimentDetail struct {
	core.ArchiveExperimentMeta
	ExperimentInfo core.ExperimentInfo `json:"experiment_info"`
}

// @Summary Get archived chaos experiments.
// @Description Get archived chaos experiments.
// @Tags archives
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Param kind query string false "kind" Enums(PodChaos, IoChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Success 200 {array} core.ArchiveExperimentMeta
// @Router /api/archives [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listExperiments(c *gin.Context) {
	kind := c.Query("kind")
	name := c.Query("name")
	ns := c.Query("namespace")

	data, err := s.archive.ListMeta(context.TODO(), kind, ns, name)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		return
	}

	c.JSON(http.StatusOK, data)
}

// @Summary Get the details of chaos experiment.
// @Description Get the details of chaos experiment.
// @Tags archives
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Param kind query string false "kind" Enums(PodChaos, IoChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Param uid query string false "uid"
// @Success 200 {array} ArchiveExperimentDetail
// @Router /api/archives/detail/search [get]
// @Failure 500 {object} utils.APIError
func (s *Service) experimentDetailSearch(c *gin.Context) {
	var (
		err           error
		info          core.ExperimentInfo
		expDetailList []ArchiveExperimentDetail
	)
	kind := c.Query("kind")
	name := c.Query("name")
	ns := c.Query("namespace")
	uid := c.Query("uid")

	datalist, err := s.archive.DetailList(context.TODO(), kind, ns, name, uid)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the archive is not found"))
		}
		return
	}

	for _, data := range datalist {
		switch data.Kind {
		case v1alpha1.KindPodChaos:
			info, err = data.ParsePodChaos()
		case v1alpha1.KindIOChaos:
			info, err = data.ParseIOChaos()
		case v1alpha1.KindNetworkChaos:
			info, err = data.ParseNetworkChaos()
		case v1alpha1.KindTimeChaos:
			info, err = data.ParseTimeChaos()
		case v1alpha1.KindKernelChaos:
			info, err = data.ParseKernelChaos()
		case v1alpha1.KindStressChaos:
			info, err = data.ParseStressChaos()
		default:
			err = fmt.Errorf("kind %s is not support", data.Kind)
		}
		if err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
		expDetailList = append(expDetailList, ArchiveExperimentDetail{
			ArchiveExperimentMeta: core.ArchiveExperimentMeta{
				ID:         data.ID,
				CreatedAt:  data.CreatedAt,
				UpdatedAt:  data.UpdatedAt,
				DeletedAt:  data.DeletedAt,
				Name:       data.Name,
				Namespace:  data.Namespace,
				Kind:       data.Kind,
				Action:     data.Action,
				UID:        data.UID,
				StartTime:  data.StartTime,
				FinishTime: data.FinishTime,
				Archived:   data.Archived,
			},
			ExperimentInfo: info,
		})
	}

	c.JSON(http.StatusOK, expDetailList)
}

// @Summary Get the details of chaos experiment.
// @Description Get the details of chaos experiment.
// @Tags archives
// @Produce json
// @Param uid query string true "uid"
// @Success 200 {object} ArchiveExperimentDetail
// @Router /api/archives/detail [get]
// @Failure 500 {object} utils.APIError
func (s *Service) experimentDetail(c *gin.Context) {
	var (
		err       error
		info      core.ExperimentInfo
		expDetail ArchiveExperimentDetail
	)
	uid := c.Query("uid")

	if uid == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("uid cannot be empty"))
		return
	}

	data, err := s.archive.FindByUID(context.TODO(), uid)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the archive is not found"))
		}
		return
	}

	switch data.Kind {
	case v1alpha1.KindPodChaos:
		info, err = data.ParsePodChaos()
	case v1alpha1.KindIOChaos:
		info, err = data.ParseIOChaos()
	case v1alpha1.KindNetworkChaos:
		info, err = data.ParseNetworkChaos()
	case v1alpha1.KindTimeChaos:
		info, err = data.ParseTimeChaos()
	case v1alpha1.KindKernelChaos:
		info, err = data.ParseKernelChaos()
	case v1alpha1.KindStressChaos:
		info, err = data.ParseStressChaos()
	default:
		err = fmt.Errorf("kind %s is not support", data.Kind)
	}
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	expDetail = ArchiveExperimentDetail{
		ArchiveExperimentMeta: core.ArchiveExperimentMeta{
			ID:         data.ID,
			CreatedAt:  data.CreatedAt,
			UpdatedAt:  data.UpdatedAt,
			DeletedAt:  data.DeletedAt,
			Name:       data.Name,
			Namespace:  data.Namespace,
			Kind:       data.Kind,
			Action:     data.Action,
			UID:        data.UID,
			StartTime:  data.StartTime,
			FinishTime: data.FinishTime,
			Archived:   data.Archived,
		},
		ExperimentInfo: info,
	}

	c.JSON(http.StatusOK, expDetail)
}

// @Summary Get the report of chaos experiment.
// @Description Get the report of chaos experiment.
// @Tags archives
// @Produce json
// @Param uid query string true "uid"
// @Success 200 {array} Report
// @Router /api/archives/report [get]
// @Failure 500 {object} utils.APIError
func (s *Service) experimentReport(c *gin.Context) {
	var (
		report    Report
		err       error
		timeNow   time.Time
		timeAfter time.Time
	)
	uid := c.Query("uid")

	if uid == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("uid cannot be empty"))
		return
	}

	report.Meta, err = s.archive.FindMetaByUID(context.TODO(), uid)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the archive is not found"))
		}
		return
	}
	report.TotalTime = report.Meta.FinishTime.Sub(report.Meta.StartTime).String()
	report.Events, err = s.event.ListByUID(context.TODO(), uid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		return
	}
	timeNow = time.Now()
	timeAfter = timeNow
	for _, et := range report.Events {
		timeAfter = timeAfter.Add(et.FinishTime.Sub(*et.StartTime))
	}
	report.TotalFaultTime = timeAfter.Sub(timeNow).String()

	c.JSON(http.StatusOK, report)
}

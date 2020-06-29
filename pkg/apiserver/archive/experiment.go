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

package archive

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/pingcap/chaos-mesh/pkg/apiserver/utils"
	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/core"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
	endpoint.GET("/detailSearch", s.experimentDetailSearch)
	endpoint.GET("/detail", s.experimentDetail)
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
// @Success 200 {array} core.ArchiveExperiment
// @Router /api/archives/detailSearch [get]
// @Failure 500 {object} utils.APIError
func (s *Service) experimentDetailSearch(c *gin.Context) {
	kind := c.Query("kind")
	name := c.Query("name")
	ns := c.Query("namespace")
	uid := c.Query("uid")

	data, err := s.archive.DetailList(context.TODO(), kind, ns, name, uid)
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

	c.JSON(http.StatusOK, data)
}

// @Summary Get the details of chaos experiment.
// @Description Get the details of chaos experiment.
// @Tags archives
// @Produce json
// @Param uid query string true "uid"
// @Success 200 {array} core.ArchiveExperiment
// @Router /api/archives/detail [get]
// @Failure 500 {object} utils.APIError
func (s *Service) experimentDetail(c *gin.Context) {

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

	c.JSON(http.StatusOK, data)
}

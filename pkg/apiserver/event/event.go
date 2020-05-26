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

package event

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pingcap/chaos-mesh/pkg/apiserver/utils"
	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/core"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service defines a handler service for events.
type Service struct {
	conf    *config.ChaosServerConfig
	kubeCli client.Client
	archive core.ExperimentStore
	event   core.EventStore
}

// NewService return an event service instance.
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
	endpoint := r.Group("/events")

	// TODO: add more api handlers
	endpoint.GET("", s.listEvents)
}

// @Summary Get all events from db.
// @Description Get all chaos events from db.
// @Tags events
// @Produce json
// @Success 200 {array} core.Event
// @Router /api/events [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listEvents(c *gin.Context) {
	podName := c.Query("podName")
	podNamespace := c.Query("podNamespace")
	startTimeStr := c.Query("startTime")
	experimentName := c.Query("experimentName")
	experimentNamespace := c.Query("experimentNamespace")

	if podName != "" && podNamespace == "" {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("when podName is not empty, podNamespace cannot be empty")))
		return
	}

	eventList, err := s.event.ListByFilter(context.Background(), podName, podNamespace, experimentName, experimentNamespace, startTimeStr)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, eventList)
}
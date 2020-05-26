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
	"strings"
	"time"

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

// NewService return a event service instance.
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
	endpoint.GET("/listbyexperiment", s.listEventsByExperiment)
}

// @Summary Get all events from db.
// @Description Get all chaos events from db.
// @Tags events
// @Produce json
// @Success 200 {array} core.Event
// @Router /api/events [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listEvents(c *gin.Context) {
	name := c.Query("podName")
	namespace := c.Query("podNamespace")
	//eventList := make([]*core.Event, 0)
	var eventList []*core.Event
	var err error

	if name != "" && namespace == "" {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("namespace is empty")))
		return
	} else if name == "" && namespace == "" {
		eventList, err = s.event.List(context.Background())
		if err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
	} else if name == "" && namespace != "" {
		eventList, err = s.event.ListByNamespace(context.Background(), namespace)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
	} else {
		eventList, err = s.event.ListByPod(context.Background(), namespace, name)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
	}
	c.JSON(http.StatusOK, eventList)
}

// @Summary Get the events from db according to the experiment.
// @Description Get the events from db according to the experiment.
// @Tags events
// @Produce json
// @Success 200 {array} core.Event
// @Router /api/events/listbyexperiment [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listEventsByExperiment(c *gin.Context) {
	startTimeStr := c.Query("startTime")
	name := c.Query("experimentName")
    namespace := c.Query("namespace")
	var eventList []*core.Event
	var err error
	if name == "" || namespace == "" {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("chaosName or chaosNamespace is empty")))
		return
	} else if startTimeStr == "" {
		eventList, err = s.event.ListByExperiment(context.Background(), namespace, name)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
	} else {
		//startTime, err := time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, time.Local)
		startTime, err := time.Parse(time.RFC3339, strings.Replace(startTimeStr, " ", "+", -1))
		if err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("the format of the time is wrong")))
			return
		}
		et, err := s.event.FindByExperimentAndStartTime(context.Background(), name, namespace, &startTime)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
		eventList = append(eventList, et)
	}
	c.JSON(http.StatusOK, eventList)
}
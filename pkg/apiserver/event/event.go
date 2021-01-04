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

package event

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
)

// Service defines a handler service for events.
type Service struct {
	conf    *config.ChaosDashboardConfig
	archive core.ExperimentStore
	event   core.EventStore
}

// NewService return an event service instance.
func NewService(
	conf *config.ChaosDashboardConfig,
	archive core.ExperimentStore,
	event core.EventStore,
) *Service {
	return &Service{
		conf:    conf,
		archive: archive,
		event:   event,
	}
}

// Register mounts our HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/events")
	endpoint.Use(utils.AuthRequired)

	// TODO: add more api handlers
	endpoint.GET("", s.listEvents)
	endpoint.GET("/dry", s.listDryEvents)
	endpoint.GET("/get", s.getEvent)
}

// @Summary Get the list of events from db.
// @Description Get the list of events from db.
// @Tags events
// @Produce json
// @Param podName query string false "The pod's name"
// @Param podNamespace query string false "The pod's namespace"
// @Param startTime query string false "The start time of events"
// @Param endTime query string false "The end time of events"
// @Param experimentName query string false "The name of the experiment"
// @Param experimentNamespace query string false "The namespace of the experiment"
// @Param uid query string false "The UID of the experiment"
// @Param kind query string false "kind" Enums(PodChaos, IoChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Param limit query string false "The max length of events list"
// @Success 200 {array} core.Event
// @Router /events [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listEvents(c *gin.Context) {
	filter := core.Filter{
		PodName:             c.Query("podName"),
		PodNamespace:        c.Query("podNamespace"),
		StartTimeStr:        c.Query("startTime"),
		FinishTimeStr:       c.Query("finishTime"),
		ExperimentName:      c.Query("experimentName"),
		ExperimentNamespace: c.Query("namespace"),
		UID:                 c.Query("uid"),
		Kind:                c.Query("kind"),
		LimitStr:            c.Query("limit"),
	}

	if filter.PodName != "" && filter.PodNamespace == "" {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("when podName is not empty, podNamespace cannot be empty")))
		return
	}

	eventList, err := s.event.ListByFilter(context.Background(), filter)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, eventList)
}

// @Summary Get the list of events without pod records from db.
// @Description Get the list of events without pod records from db.
// @Tags events
// @Produce json
// @Param startTime query string false "The start time of events"
// @Param endTime query string false "The end time of events"
// @Param experimentName query string false "The name of the experiment"
// @Param experimentNamespace query string false "The namespace of the experiment"
// @Param kind query string false "kind" Enums(PodChaos, IoChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Param limit query string false "The max length of events list"
// @Success 200 {array} core.Event
// @Router /events/dry [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listDryEvents(c *gin.Context) {
	filter := core.Filter{
		StartTimeStr:        c.Query("startTime"),
		FinishTimeStr:       c.Query("finishTime"),
		ExperimentName:      c.Query("experimentName"),
		ExperimentNamespace: c.Query("namespace"),
		Kind:                c.Query("kind"),
		LimitStr:            c.Query("limit"),
	}

	eventList, err := s.event.DryListByFilter(context.Background(), filter)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, eventList)
}

// @Summary Get the event from db by ID.
// @Description Get the event from db by ID.
// @Tags events
// @Produce json
// @Param id query uint true "The id of the event"
// @Success 200 {object} core.Event
// @Router /events/get [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getEvent(c *gin.Context) {
	idStr := c.Query("id")
	namespace := c.Query("namespace")

	if idStr == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("id cannot be empty"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("the format of id is wrong"))
		return
	}

	event, err := s.event.Find(context.Background(), uint(id))
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the event is not found"))
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		}
		return
	}

	if len(namespace) != 0 && event.Namespace != namespace {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("event %s belong to namespace %s but not namespace %s", idStr, event.Namespace, namespace))
		return
	}

	c.JSON(http.StatusOK, event)
}

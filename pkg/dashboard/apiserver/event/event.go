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
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	u "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

var log = u.Log.WithName("events")

// Service defines a handler service for events.
type Service struct {
	event core.EventStore
	conf  *config.ChaosDashboardConfig
}

func NewService(
	event core.EventStore,
	conf *config.ChaosDashboardConfig,
) *Service {
	return &Service{
		event: event,
		conf:  conf,
	}
}

// Register events RouterGroup.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/events")
	endpoint.Use(func(c *gin.Context) {
		u.AuthMiddleware(c, s.conf)
	})

	endpoint.GET("", s.list)
	endpoint.GET("/:id", s.get)
}

// @Summary list events.
// @Description Get events from db.
// @Tags events
// @Produce json
// @Param created_at query string false "The create time of events"
// @Param name query string false "The name of the object"
// @Param namespace query string false "The namespace of the object"
// @Param object_id query string false "The UID of the object"
// @Param kind query string false "kind" Enums(PodChaos, IOChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos, AWSChaos, GCPChaos, DNSChaos, Schedule)
// @Param limit query string false "The max length of events list"
// @Success 200 {array} core.Event
// @Failure 500 {object} utils.APIError
// @Router /events [get]
func (s *Service) list(c *gin.Context) {
	ns := c.Query("namespace")

	if ns == "" && !s.conf.ClusterScoped && s.conf.TargetNamespace != "" {
		ns = s.conf.TargetNamespace

		log.V(1).Info("Replace query namespace with", ns)
	}

	filter := core.Filter{
		ObjectID:  c.Query("object_id"),
		Start:     c.Query("start"),
		End:       c.Query("end"),
		Namespace: ns,
		Name:      c.Query("name"),
		Kind:      c.Query("kind"),
		Limit:     c.Query("limit"),
	}

	events, err := s.event.ListByFilter(context.Background(), filter)
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))

		return
	}

	c.JSON(http.StatusOK, events)
}

// @Summary Get an event.
// @Description Get the event from db by ID.
// @Tags events
// @Produce json
// @Param id path uint true "The event ID"
// @Success 200 {object} core.Event
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /events/{id} [get]
func (s *Service) get(c *gin.Context) {
	id, ns := c.Param("id"), c.Query("namespace")

	if id == "" {
		u.SetAPIError(c, u.ErrBadRequest.New("ID cannot be empty"))

		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.New("ID is not a number"))

		return
	}

	if ns == "" && !s.conf.ClusterScoped && s.conf.TargetNamespace != "" {
		ns = s.conf.TargetNamespace

		log.V(1).Info("Replace query namespace with", ns)
	}

	event, err := s.event.Find(context.Background(), uint(intID))
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Event "+id+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	if len(ns) != 0 && event.Namespace != ns {
		u.SetAPIError(c, u.ErrInternalServer.New("The namespace of event %s is %s instead of the %s in the request", id, event.Namespace, ns))

		return
	}

	c.JSON(http.StatusOK, event)
}

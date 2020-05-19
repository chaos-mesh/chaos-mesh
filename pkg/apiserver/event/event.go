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
	"net/http"

	"github.com/gin-gonic/gin"

	statuscode "github.com/pingcap/chaos-mesh/pkg/apiserver/status_code"
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
	endpoint.GET("/all", s.listEvents)
}

func (s *Service) listEvents(c *gin.Context) {
	name := c.Query("name")
	namespace := c.Query("namespace")
	eventList := make([]*core.Event, 0)

	if name == "" && namespace == "" {
		eventList, err := s.event.List(context.Background())
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": statuscode.GetResourcesFromDBWrong,
				"message": "get events wrong",
				"data": eventList,
			})
			return
		}
	} else if (name != "" && namespace == "") || (name == "" && namespace != "") {
		c.JSON(http.StatusOK, gin.H{
			"status": statuscode.IncompleteField,
			"message": "one of name and namespace is empty",
			"data": eventList,
		})
		return
	} else {
		eventList, err := s.event.ListByPod(context.Background(), namespace, name)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"status": statuscode.GetResourcesFromDBWrong,
				"message": "get events wrong",
				"data": eventList,
			})
			return
		}
	}

	// TODO: Return required fields in event
}

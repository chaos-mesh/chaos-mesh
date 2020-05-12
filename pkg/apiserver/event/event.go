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
	"github.com/gin-gonic/gin"

	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/core"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service defines a handler service for events.
type Service struct {
	conf      *config.ChaosServerConfig
	kubeCli   client.Client
	archive   core.ExperimentStore
	event     core.EventStore
	podRecord core.PodRecordStore
}

// NewService return a event service instance.
func NewService(
	conf *config.ChaosServerConfig,
	cli client.Client,
	archive core.ExperimentStore,
	event core.EventStore,
	podRecord core.PodRecordStore,
) *Service {
	return &Service{
		conf:      conf,
		kubeCli:   cli,
		archive:   archive,
		event:     event,
		podRecord: podRecord,
	}
}

// Register mounts our HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/event")

	// TODO: add more api handlers
	endpoint.GET("/all", s.listEvents)
}

// TODO: need to be implemented
func (s *Service) listEvents(c *gin.Context) {}

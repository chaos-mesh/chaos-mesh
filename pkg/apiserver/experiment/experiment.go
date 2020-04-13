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

package experiment

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/store/archive"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Service struct {
	conf    *config.ChaosServerConfig
	kubeCli client.Client
	store   archive.ArchiveStore
}

func NewService(conf *config.ChaosServerConfig, cli client.Client, store archive.ArchiveStore) *Service {
	return &Service{
		conf:    conf,
		kubeCli: cli,
		store:   store,
	}
}

func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/experiment")
	endpoint.GET("/all", s.listExperiments)
	endpoint.GET("/:name", s.getExperimentDetail)
	endpoint.DELETE("/delete/:ns/:name", s.deleteExperiment)
}

func (s *Service) listExperiments(c *gin.Context) {}

func (s *Service) getExperimentDetail(c *gin.Context) {}

func (s *Service) deleteExperiment(c *gin.Context) {}

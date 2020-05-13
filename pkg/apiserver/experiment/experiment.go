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
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	"time"

	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/core"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service defines a handler service for experiments.
type Service struct {
	conf    *config.ChaosServerConfig
	kubeCli client.Client
	archive core.ExperimentStore
	event   core.EventStore
}

// NewService returns an experiment service instance.
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
	endpoint := r.Group("/experiments")

	// TODO: add more api handlers
	endpoint.GET("/", s.listExperiments)
	endpoint.GET("/detail/:name", s.getExperimentDetail)
	endpoint.DELETE("/delete/:ns/:name", s.deleteExperiment)
	endpoint.GET("/test", s.test)
}

// TODO: need to be implemented
func (s *Service) listExperiments(c *gin.Context) {}

// TODO: need to be implemented
func (s *Service) getExperimentDetail(c *gin.Context) {}

// TODO: need to be implemented
func (s *Service) deleteExperiment(c *gin.Context) {}

func (s *Service) test(c *gin.Context) {

	chaosKey := types.NamespacedName{
		Namespace: "chaos-testing",
		Name:      "io-chaos",
	}


	chaos := &v1alpha1.IoChaos{}
	err := s.kubeCli.Get(context.Background(), chaosKey, chaos)


	err := s.kubeCli.List(context.Background(), &namespace)

	if err != nil {
		fmt.Println("!!!!!!!!!!!! ,eerrerer")
	} else {
		fmt.Println(chaos.Status)
	}

	c.JSON(200, gin.H{
		"Name": "testtest",
	})
}



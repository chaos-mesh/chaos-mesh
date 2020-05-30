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

package common

import (
	"context"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/pkg/apiserver/experiment"
	"github.com/pingcap/chaos-mesh/pkg/apiserver/utils"
	"github.com/pingcap/chaos-mesh/pkg/config"
	pkgutils "github.com/pingcap/chaos-mesh/pkg/utils"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Pod defines the basic information of a pod
type Pod struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Namespace string `json:"namespace"`
	State     string `json:"state"`
}

// Service defines a handler service for cluster common objects.
type Service struct {
	conf    *config.ChaosServerConfig
	kubeCli client.Client
}

// NewService returns an experiment service instance.
func NewService(
	conf *config.ChaosServerConfig,
	cli client.Client,
) *Service {
	return &Service{
		conf:    conf,
		kubeCli: cli,
	}
}

// Register mounts our HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/common")

	endpoint.POST("/pods", s.listPods)
	endpoint.GET("/namespaces", s.getNamespaces)
	endpoint.GET("/kinds", s.getKinds)

}

// @Summary Get pods from Kubernetes cluster.
// @Description Get pods from Kubernetes cluster.
// @Tags common
// @Produce json
// @Param request body experiment.SelectorInfo true "Request body"
// @Success 200 {array} Pod
// @Router /api/common/pods [post]
// @Failure 500 {object} utils.APIError
func (s *Service) listPods(c *gin.Context) {
	exp := &experiment.SelectorInfo{}
	if err := c.ShouldBindJSON(exp); err != nil {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}
	ctx := context.TODO()
	filteredPods, err := pkgutils.SelectPods(ctx, s.kubeCli, exp.ParseSelector())
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	pods := make([]Pod, 0, len(filteredPods))
	for _, pod := range filteredPods {
		pods = append(pods, Pod{
			Name:      pod.Name,
			IP:        pod.Status.PodIP,
			Namespace: pod.Namespace,
			State:     string(pod.Status.Phase),
		})
	}

	c.JSON(http.StatusOK, pods)
}

// @Summary Get all namespaces from Kubernetes cluster.
// @Description Get all namespaces from Kubernetes cluster.
// @Tags common
// @Produce json
// @Success 200 {array} string
// @Router /api/common/namespaces [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getNamespaces(c *gin.Context) {
	var nsList v1.NamespaceList
	if err := s.kubeCli.List(context.Background(), &nsList); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	namespaces := make(sort.StringSlice, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	sort.Sort(namespaces)
	c.JSON(http.StatusOK, namespaces)
}

// @Summary Get all chaos kinds from Kubernetes cluster.
// @Description Get all chaos kinds from Kubernetes cluster.
// @Tags common
// @Produce json
// @Success 200 {array} string
// @Router /api/common/kinds [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getKinds(c *gin.Context) {
	var kinds []string

	allKinds := v1alpha1.AllKinds()
	for name := range allKinds {
		kinds = append(kinds, name)
	}

	sort.Strings(kinds)
	c.JSON(http.StatusOK, kinds)
}

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
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/experiment"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	pkgutils "github.com/chaos-mesh/chaos-mesh/pkg/utils"

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
	conf    *config.ChaosDashboardConfig
	kubeCli client.Client
}

// NewService returns an experiment service instance.
func NewService(
	conf *config.ChaosDashboardConfig,
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
	endpoint.GET("/labels", s.getLabels)
	endpoint.GET("/annotations", s.getAnnotations)

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

// MapSlice defines a common map
type MapSlice map[string][]string

// @Summary Get the labels of the pods in the specified namespace from Kubernetes cluster.
// @Description Get the labels of the pods in the specified namespace from Kubernetes cluster.
// @Tags common
// @Produce json
// @Param podNamespaceList query string true "The pod's namespace list, split by ,"
// @Success 200 {object} MapSlice
// @Router /api/common/labels [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getLabels(c *gin.Context) {
	podNamespaceList := c.Query("podNamespaceList")

	if podNamespaceList == "" {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("podNamespaceList cannot be empty")))
		return
	}

	exp := &experiment.SelectorInfo{}
	nsList := strings.Split(podNamespaceList, ",")
	exp.NamespaceSelectors = nsList

	ctx := context.TODO()
	filteredPods, err := pkgutils.SelectPods(ctx, s.kubeCli, exp.ParseSelector())
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	labels := make(map[string][]string)

	for _, pod := range filteredPods {
		for k, v := range pod.Labels {
			if _, ok := labels[k]; ok {
				if !inSlice(v, labels[k]) {
					labels[k] = append(labels[k], v)
				}
			} else {
				labels[k] = []string{v}
			}
		}
	}
	c.JSON(http.StatusOK, labels)
}

// @Summary Get the annotations of the pods in the specified namespace from Kubernetes cluster.
// @Description Get the annotations of the pods in the specified namespace from Kubernetes cluster.
// @Tags common
// @Produce json
// @Param podNamespaceList query string true "The pod's namespace list, split by ,"
// @Success 200 {object} MapSlice
// @Router /api/common/annotations [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getAnnotations(c *gin.Context) {
	podNamespaceList := c.Query("podNamespaceList")

	if podNamespaceList == "" {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("podNamespaceList cannot be empty")))
		return
	}

	exp := &experiment.SelectorInfo{}
	nsList := strings.Split(podNamespaceList, ",")
	exp.NamespaceSelectors = nsList

	ctx := context.TODO()
	filteredPods, err := pkgutils.SelectPods(ctx, s.kubeCli, exp.ParseSelector())
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	annotations := make(map[string][]string)

	for _, pod := range filteredPods {
		for k, v := range pod.Annotations {
			if _, ok := annotations[k]; ok {
				if !inSlice(v, annotations[k]) {
					annotations[k] = append(annotations[k], v)
				}
			} else {
				annotations[k] = []string{v}
			}
		}
	}
	c.JSON(http.StatusOK, annotations)
}

// inSlice checks given string in string slice or not.
func inSlice(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

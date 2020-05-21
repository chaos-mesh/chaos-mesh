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
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/pingcap/chaos-mesh/pkg/apiserver/utils"
	"github.com/pingcap/chaos-mesh/pkg/config"

	v1 "k8s.io/api/core/v1"
	apicli "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Pod defines the basic information of a pod
type Pod struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Namespace string `json:"namespace"`
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

	endpoint.GET("/pods", s.getPods)
	endpoint.GET("/namespaces", s.getNamespaces)
	endpoint.GET("/kinds", s.getKinds)

}

// @Summary Get pods from Kubernetes cluster.
// @Description Get pods from Kubernetes cluster.
// @Tags common
// @Produce json
// @Param namespace query string false "namespace"
// @Success 200 {array} Pod
// @Router /api/common/pods [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getPods(c *gin.Context) {
	listOptions := &client.ListOptions{}
	ns := c.Query("namespace")
	if ns != "" {
		listOptions.Namespace = ns
	}

	var podList v1.PodList
	if err := s.kubeCli.List(context.Background(), &podList, listOptions); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	pods := make([]Pod, 0, len(podList.Items))
	for _, pod := range podList.Items {
		pods = append(pods, Pod{
			Name:      pod.Name,
			IP:        pod.Status.PodIP,
			Namespace: pod.Namespace,
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
	conf, _ := ctrlconfig.GetConfig()
	apiExtCli, _ := apicli.NewForConfig(conf)

	crdList, err := apiExtCli.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	kinds := make(sort.StringSlice, 0, len(crdList.Items))
	for _, crd := range crdList.Items {
		if strings.Contains(crd.Spec.Names.Kind, "Chaos") == true {
			kinds = append(kinds, crd.Spec.Names.Kind)
		}
	}
	sort.Sort(kinds)
	c.JSON(http.StatusOK, kinds)
}

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

// Pod defines the basic information of the pod
type Pod struct {
	Name      string
	IP        string
	Namespace string
}

// Service defines a handler service for experiments.
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
// @Produce json
// @Param ns query string false "namespace"
// @Success 200 {array} Pod
// @Router /api/common/pods [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getPods(c *gin.Context) {
	listOptions := &client.ListOptions{}
	ns := c.Query("ns")
	if ns != "" {
		listOptions.Namespace = ns
	}

	var podList v1.PodList
	err := s.kubeCli.List(context.Background(), &podList, listOptions)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	pods := make([]Pod, 0)
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
// @Produce json
// @Success 200 {array} string
// @Router /api/common/namespaces [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getNamespaces(c *gin.Context) {
	var namespace v1.NamespaceList
	err := s.kubeCli.List(context.Background(), &namespace)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	var namespaces []string
	for _, ns := range namespace.Items {
		namespaces = append(namespaces, ns.Name)
	}

	c.JSON(http.StatusOK, namespaces)
}

// @Summary Get all chaos kinds from Kubernetes cluster.
// @Description Get all chaos kinds from Kubernetes cluster.
// @Produce json
// @Success 200 {array} string
// @Router /api/common/kinds [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getKinds(c *gin.Context) {
	config, _ := ctrlconfig.GetConfig()
	apiExtCli, _ := apicli.NewForConfig(config)

	crdList, err := apiExtCli.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	var kinds []string
	for _, crd := range crdList.Items {
		if strings.Contains(crd.Spec.Names.Kind, "Chaos") == true {
			kinds = append(kinds, crd.Spec.Names.Kind)
		}
	}

	c.JSON(http.StatusOK, kinds)
}

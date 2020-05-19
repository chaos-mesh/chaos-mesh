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

	"github.com/pingcap/chaos-mesh/pkg/config"

	v1 "k8s.io/api/core/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Pod defines the basic information of the pod
type Pod struct {
	Name      string
	Namespace string
}

// Namespace defines the basic information of the namespace
type Namespace struct {
	Name string
}

// ChaosKind defines the kind information of the chaos
type ChaosKind struct {
	Name string
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

// @Summary Get all pods from Kubernetes cluster.
// @Description Get all pods from Kubernetes cluster.
// @Produce json
// @Success 200 {array} Pod
// @Router /common/pods [get]
// @Failure 401 {object} utils.APIError "Unauthorized failure"
func (s *Service) getPods(c *gin.Context) {
	pods := make([]Pod, 0)

	var podList v1.PodList
	err := s.kubeCli.List(context.Background(), &podList)

	if err != nil {
		_ = c.Error(err)
		return
	}

	for _, pod := range podList.Items {
		pods = append(pods, Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})
	}

	c.JSON(http.StatusOK, pods)
}

// @Summary Get all namespaces from Kubernetes cluster.
// @Description Get all namespaces from Kubernetes cluster.
// @Produce json
// @Success 200 {array} Namespace
// @Router /common/namespaces [get]
// @Failure 401 {object} utils.APIError "Unauthorized failure"
func (s *Service) getNamespaces(c *gin.Context) {

	var namespace v1.NamespaceList

	err := s.kubeCli.List(context.Background(), &namespace)

	if err != nil {
		_ = c.Error(err)
		return
	}

	namespaceList := make([]Namespace, 0, len(namespace.Items))
	for _, ns := range namespace.Items {
		namespaceList = append(namespaceList, Namespace{
			Name: ns.Name,
		})
	}

	c.JSON(http.StatusOK, namespaceList)
}

// @Summary Get all chaos kinds from Kubernetes cluster.
// @Description Get all chaos kinds from Kubernetes cluster.
// @Produce json
// @Success 200 {array} ChaosKind
// @Router /common/kinds [get]
// @Failure 401 {object} utils.APIError "Unauthorized failure"
func (s *Service) getKinds(c *gin.Context) {
	ChaosKindList := make([]ChaosKind, 0)

	config, _ := ctrlconfig.GetConfig()
	apiExtCli, _ := apiextensionsclientset.NewForConfig(config)

	crdList, err := apiExtCli.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		_ = c.Error(err)
		return
	}
	for _, crd := range crdList.Items {
		if strings.Contains(crd.Spec.Names.Kind, "Chaos") == true {
			ChaosKindList = append(ChaosKindList, ChaosKind{
				Name: crd.Spec.Names.Kind,
			})
		}
	}

	c.JSON(http.StatusOK, ChaosKindList)
}

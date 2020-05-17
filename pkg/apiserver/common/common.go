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

	"github.com/pingcap/chaos-mesh/pkg/apiserver"
	"github.com/pingcap/chaos-mesh/pkg/config"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/api/core/v1"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Pod defines the basic information of the pod
type Pod struct {
	Name string
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

	endpoint.GET("/pods/all", s.GetPods)
	endpoint.GET("/namespaces/all", s.GetNamespaces)
	endpoint.GET("/kinds/all", s.GetKinds)

}

// GetPods returns the list of pods
func (s *Service) GetPods(c *gin.Context) {
	pods :=  make([]Pod, 0)

	var podList v1.PodList
	err := s.kubeCli.List(context.Background(), &podList)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": apiserver.GetResourcesWrong,
			"message": "get pods wrong",
			"data": pods,
		})
		return
	}

	for _, pod := range podList.Items {
		pods = append(pods, Pod{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})
	}

	c.JSON(http.StatusOK, gin.H{
			"status": apiserver.Success,
			"message": "success",
			"data": pods,
	})
}

// GetNamespaces returns the list of namespaces
func (s *Service) GetNamespaces(c *gin.Context) {
	namespaceList :=  make([]Namespace, 0)

	var namespace v1.NamespaceList

	err := s.kubeCli.List(context.Background(), &namespace)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": apiserver.GetResourcesWrong,
			"message": "get namespaces wrong",
			"data": namespaceList,
		})
		return
	}

	for _, ns := range namespace.Items {
		namespaceList = append(namespaceList, Namespace{
			Name: ns.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": apiserver.Success,
		"message": "success",
		"data": namespaceList,
	})
}

// GetKinds returns the list of chaos kinds
func (s *Service) GetKinds(c *gin.Context) {
	ChaosKindList :=  make([]ChaosKind, 0)

	config, _ := ctrlconfig.GetConfig()
	apiExtCli, _ := apiextensionsclientset.NewForConfig(config)

	crdList, err := apiExtCli.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": apiserver.GetResourcesWrong,
			"message": "get CRDs wrong",
			"data": ChaosKindList,
		})
	}
	for _, crd := range crdList.Items {
		if strings.Contains(crd.Spec.Names.Kind, "Chaos") == true {
			ChaosKindList = append(ChaosKindList, ChaosKind{
				Name: crd.Spec.Names.Kind,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": apiserver.Success,
		"message": "success",
		"data": ChaosKindList,
	})
}
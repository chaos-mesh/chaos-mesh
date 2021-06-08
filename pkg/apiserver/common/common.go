// Copyright 2020 Chaos Mesh Authors.
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
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

const (
	roleManager = "manager"
	roleViewer  = "viewer"

	serviceAccountTemplate = `kind: ServiceAccount
apiVersion: v1
metadata:
  namespace: %s
  name: %s
`
	roleTemplate = `kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: %s
  name: %s
rules:
- apiGroups: [""]
  resources: ["pods", "namespaces"]
  verbs: ["get", "watch", "list"]
- apiGroups:
  - chaos-mesh.org
  resources: [ "*" ]
  verbs: [%s]
`
	clusterRoleTemplate = `kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: %s
rules:
- apiGroups: [""]
  resources: ["pods", "namespaces"]
  verbs: ["get", "watch", "list"]
- apiGroups:
  - chaos-mesh.org
  resources: [ "*" ]
  verbs: [%s]
`
	roleBindingTemplate = `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: %s
  namespace: %s
subjects:
- kind: ServiceAccount
  name: %s
  namespace: %s
roleRef:
  kind: Role
  name: %s
  apiGroup: rbac.authorization.k8s.io
`
	clusterRoleBindingTemplate = `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: %s
subjects:
- kind: ServiceAccount
  name: %s
  namespace: %s
roleRef:
  kind: ClusterRole
  name: %s
  apiGroup: rbac.authorization.k8s.io
`
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
	// this kubeCli use the local token, used for list namespace of the K8s cluster
	kubeCli client.Client
	conf    *config.ChaosDashboardConfig
}

// NewService returns an experiment service instance.
func NewService(
	conf *config.ChaosDashboardConfig,
	kubeCli client.Client,
) *Service {
	return &Service{
		conf:    conf,
		kubeCli: kubeCli,
	}
}

// Register mounts our HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/common")

	endpoint.POST("/pods", s.listPods)
	endpoint.GET("/namespaces", s.listNamespaces)
	endpoint.GET("/chaos-available-namespaces", s.getChaosAvailableNamespaces)
	endpoint.GET("/kinds", s.getKinds)
	endpoint.GET("/labels", s.getLabels)
	endpoint.GET("/annotations", s.getAnnotations)
	endpoint.GET("/config", s.getConfig)
	endpoint.GET("/rbac-config", s.getRbacConfig)
}

// @Summary Get pods from Kubernetes cluster.
// @Description Get pods from Kubernetes cluster.
// @Tags common
// @Produce json
// @Param request body core.SelectorInfo true "Request body"
// @Success 200 {array} Pod
// @Router /common/pods [post]
// @Failure 500 {object} utils.APIError
func (s *Service) listPods(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	exp := &core.SelectorInfo{}
	if err := c.ShouldBindJSON(exp); err != nil {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}
	ctx := context.TODO()
	filteredPods, err := pod.SelectPods(ctx, kubeCli, nil, exp.ParseSelector(), s.conf.ClusterScoped, s.conf.TargetNamespace, s.conf.EnableFilterNamespace)
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
// @Description Get all from Kubernetes cluster.
// @Deprecated This API only works within cluster scoped mode. Please use /common/chaos-available-namespaces instead.
// @Tags common
// @Produce json
// @Success 200 {array} string
// @Router /common/namespaces [get]
// @Failure 500 {object} utils.APIError
func (s *Service) listNamespaces(c *gin.Context) {

	var namespaces sort.StringSlice

	var nsList v1.NamespaceList
	if err := s.kubeCli.List(context.Background(), &nsList); err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	namespaces = make(sort.StringSlice, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	sort.Sort(namespaces)
	c.JSON(http.StatusOK, namespaces)
}

// @Summary Get all namespaces which could inject chaos(explosion scope) from Kubernetes cluster.
// @Description Get all namespaces which could inject chaos(explosion scope) from Kubernetes cluster.
// @Tags common
// @Produce json
// @Success 200 {array} string
// @Router /common/chaos-available-namespaces [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getChaosAvailableNamespaces(c *gin.Context) {

	var namespaces sort.StringSlice

	if s.conf.ClusterScoped {
		var nsList v1.NamespaceList
		if err := s.kubeCli.List(context.Background(), &nsList); err != nil {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
		namespaces = make(sort.StringSlice, 0, len(nsList.Items))
		for _, ns := range nsList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	} else {
		namespaces = append(namespaces, s.conf.TargetNamespace)
	}

	sort.Sort(namespaces)
	c.JSON(http.StatusOK, namespaces)
}

// @Summary Get all chaos kinds from Kubernetes cluster.
// @Description Get all chaos kinds from Kubernetes cluster.
// @Tags common
// @Produce json
// @Success 200 {array} string
// @Router /common/kinds [get]
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
// @Router /common/labels [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getLabels(c *gin.Context) {

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	podNamespaceList := c.Query("podNamespaceList")

	if podNamespaceList == "" {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("podNamespaceList cannot be empty")))
		return
	}

	exp := &core.SelectorInfo{}
	nsList := strings.Split(podNamespaceList, ",")
	exp.Namespaces = nsList

	ctx := context.TODO()
	filteredPods, err := pod.SelectPods(ctx, kubeCli, nil, exp.ParseSelector(), s.conf.ClusterScoped, s.conf.TargetNamespace, s.conf.EnableFilterNamespace)
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
// @Router /common/annotations [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getAnnotations(c *gin.Context) {

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	podNamespaceList := c.Query("podNamespaceList")

	if podNamespaceList == "" {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(fmt.Errorf("podNamespaceList cannot be empty")))
		return
	}

	exp := &core.SelectorInfo{}
	nsList := strings.Split(podNamespaceList, ",")
	exp.Namespaces = nsList

	ctx := context.TODO()
	filteredPods, err := pod.SelectPods(ctx, kubeCli, nil, exp.ParseSelector(), s.conf.ClusterScoped, s.conf.TargetNamespace, s.conf.EnableFilterNamespace)
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

// @Summary Get the config of Dashboard.
// @Description Get the config of Dashboard.
// @Tags common
// @Produce json
// @Success 200 {object} config.ChaosDashboardConfig
// @Router /common/config [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, s.conf)
}

// @Summary Get the rbac config according to the user's choice.
// @Description Get the rbac config according to the user's choice.
// @Tags common
// @Produce json
// @Success 200 {object} MapSlice
// @Router /common/rbac-config [get]
// @Failure 500 {object} utils.APIError
func (s *Service) getRbacConfig(c *gin.Context) {
	namespace := c.Query("namespace")
	roleType := c.Query("role")

	var serviceAccount, role, roleBinding, verbs string
	randomStr := randomStringWithCharset(5, charset)

	scope := namespace
	if len(namespace) == 0 {
		namespace = "default"
		scope = "cluster"
	}
	if roleType == roleManager {
		verbs = `"get", "list", "watch", "create", "delete", "patch", "update"`
	} else if roleType == roleViewer {
		verbs = `"get", "list", "watch"`
	} else {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(fmt.Errorf("roleType is neither manager nor viewer")))
		return
	}

	serviceAccountName := fmt.Sprintf("account-%s-%s-%s", scope, roleType, randomStr)
	roleName := fmt.Sprintf("role-%s-%s-%s", scope, roleType, randomStr)
	roleBindingName := fmt.Sprintf("bind-%s-%s-%s", scope, roleType, randomStr)

	serviceAccount = fmt.Sprintf(serviceAccountTemplate, namespace, serviceAccountName)
	if scope == "cluster" {
		role = fmt.Sprintf(clusterRoleTemplate, roleName, verbs)
		roleBinding = fmt.Sprintf(clusterRoleBindingTemplate, roleBindingName, serviceAccountName, namespace, roleName)
	} else {
		role = fmt.Sprintf(roleTemplate, namespace, roleName, verbs)
		roleBinding = fmt.Sprintf(roleBindingTemplate, roleBindingName, namespace, serviceAccountName, namespace, roleName)
	}

	rbacMap := make(map[string]string)
	rbacMap[serviceAccountName] = serviceAccount + "\n---\n" + role + "\n---\n" + roleBinding

	c.JSON(http.StatusOK, rbacMap)
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

const charset = "abcdefghijklmnopqrstuvwxyz"

func randomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

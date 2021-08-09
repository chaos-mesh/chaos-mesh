// Copyright 2021 Chaos Mesh Authors.
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

package workflow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
)

var log = ctrl.Log.WithName("workflow api")

// StatusResponse defines a common status struct.
type StatusResponse struct {
	Status string `json:"status"`
}

func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/workflows")
	endpoint.GET("", s.listWorkflows)
	endpoint.POST("", s.createWorkflow)
	endpoint.GET("/:uid", s.getWorkflowDetailByUID)
	endpoint.PUT("/:uid", s.updateWorkflow)
	endpoint.DELETE("/:uid", s.deleteWorkflow)
}

// Service defines a handler service for workflows.
type Service struct {
	conf  *config.ChaosDashboardConfig
	store core.WorkflowStore
}

func NewService(conf *config.ChaosDashboardConfig, store core.WorkflowStore) *Service {
	return &Service{conf: conf, store: store}
}

// @Summary List workflows from Kubernetes cluster.
// @Description List workflows from Kubernetes cluster.
// @Tags workflows
// @Produce json
// @Param namespace query string false "namespace, given empty string means list from all namespace"
// @Param status query string false "status" Enums(Initializing, Running, Errored, Finished)
// @Success 200 {array} core.WorkflowMeta
// @Router /workflows [get]
// @Failure 500 {object} utils.APIError
func (it *Service) listWorkflows(c *gin.Context) {
	namespace := c.Query("namespace")
	if len(namespace) == 0 && !it.conf.ClusterScoped &&
		len(it.conf.TargetNamespace) != 0 {
		namespace = it.conf.TargetNamespace
	}

	result := make([]core.WorkflowMeta, 0)

	kubeClient, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}
	repo := core.NewKubeWorkflowRepository(kubeClient)

	if namespace != "" {
		workflowFromNs, err := repo.ListByNamespace(c.Request.Context(), namespace)
		if err != nil {
			utils.SetErrorForGinCtx(c, err)
			return
		}
		result = append(result, workflowFromNs...)
	} else {
		allWorkflow, err := repo.List(c.Request.Context())
		if err != nil {
			utils.SetErrorForGinCtx(c, err)
			return
		}
		result = append(result, allWorkflow...)
	}

	// enriching with ID
	for index, item := range result {
		entity, err := it.store.FindByUID(c.Request.Context(), string(item.UID))
		if err != nil {
			log.Info("warning: workflow does not have a record in database",
				"namespaced name", fmt.Sprintf("%s/%s", item.Namespace, item.Name),
				"uid", item.UID,
			)
		}

		if entity != nil {
			result[index].ID = entity.ID
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[i].CreatedAt)
	})

	c.JSON(http.StatusOK, result)
}

// @Summary Get detailed information about the specified workflow.
// @Description Get detailed information about the specified workflow. If that object is not existed in kubernetes, it will only return ths persisted data in the database.
// @Tags workflows
// @Produce json
// @Param uid path string true "uid"
// @Router /workflows/{uid} [GET]
// @Success 200 {object} core.WorkflowDetail
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
func (it *Service) getWorkflowDetailByUID(c *gin.Context) {
	uid := c.Param("uid")

	entity, err := it.store.FindByUID(c.Request.Context(), uid)
	if err != nil {
		utils.SetErrorForGinCtx(c, err)
		return
	}

	namespace := entity.Namespace
	name := entity.Name

	kubeClient, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// if not exists in kubernetes anymore, return the persisted entity directly.
			workflowDetail, err := core.WorkflowEntity2WorkflowDetail(entity)
			if err != nil {
				utils.SetErrorForGinCtx(c, err)
				return
			}
			c.JSON(http.StatusOK, workflowDetail)
			return
		}
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	// enriching the topology and spec/status with CR in kubernetes
	repo := core.NewKubeWorkflowRepository(kubeClient)

	workflowCRInKubernetes, err := repo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		utils.SetErrorForGinCtx(c, err)
		return
	}
	result, err := core.WorkflowEntity2WorkflowDetail(entity)
	if err != nil {
		utils.SetErrorForGinCtx(c, err)
		return
	}
	result.Topology = workflowCRInKubernetes.Topology
	result.KubeObject = workflowCRInKubernetes.KubeObject

	c.JSON(http.StatusOK, result)
}

// @Summary Create a new workflow.
// @Description Create a new workflow.
// @Tags workflows
// @Produce json
// @Param request body v1alpha1.Workflow true "Request body"
// @Success 200 {object} core.WorkflowDetail
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /workflows/new [post]
func (it *Service) createWorkflow(c *gin.Context) {
	payload := v1alpha1.Workflow{}

	err := json.NewDecoder(c.Request.Body).Decode(&payload)
	if err != nil {
		_ = c.Error(utils.ErrInternalServer.Wrap(err, "failed to parse request body"))
		return
	}

	kubeClient, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	repo := core.NewKubeWorkflowRepository(kubeClient)

	result, err := repo.Create(c.Request.Context(), payload)
	if err != nil {
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary Delete the specified workflow.
// @Description Delete the specified workflow.
// @Tags workflows
// @Produce json
// @Param uid path string true "uid"
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /workflows/{uid} [delete]
func (it *Service) deleteWorkflow(c *gin.Context) {
	uid := c.Param("uid")

	entity, err := it.store.FindByUID(c.Request.Context(), uid)
	if err != nil {
		utils.SetErrorForGinCtx(c, err)
		return
	}

	namespace := entity.Namespace
	name := entity.Name

	kubeClient, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	repo := core.NewKubeWorkflowRepository(kubeClient)

	err = repo.Delete(c.Request.Context(), namespace, name)
	if err != nil {
		utils.SetErrorForGinCtx(c, err)
		return
	}
	c.JSON(http.StatusOK, StatusResponse{Status: "success"})
}

// @Summary Update a workflow.
// @Description Update a workflow.
// @Tags workflows
// @Produce json
// @Param uid path string true "uid"
// @Param request body v1alpha1.Workflow true "Request body"
// @Success 200 {object} core.WorkflowDetail
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /workflows/{uid} [put]
func (it *Service) updateWorkflow(c *gin.Context) {
	payload := v1alpha1.Workflow{}

	err := json.NewDecoder(c.Request.Body).Decode(&payload)
	if err != nil {
		_ = c.Error(utils.ErrInternalServer.Wrap(err, "failed to parse request body"))
		return
	}
	uid := c.Param("uid")
	entity, err := it.store.FindByUID(c.Request.Context(), uid)
	if err != nil {
		utils.SetErrorForGinCtx(c, err)
		return
	}

	namespace := entity.Namespace
	name := entity.Name

	if namespace != payload.Namespace {
		_ = c.Error(utils.ErrInvalidRequest.Wrap(err,
			"namespace is not consistent, pathParameter: %s, metaInRaw: %s",
			namespace,
			payload.Namespace),
		)
		return
	}
	if name != payload.Name {
		_ = c.Error(utils.ErrInvalidRequest.Wrap(err,
			"name is not consistent, pathParameter: %s, metaInRaw: %s",
			name,
			payload.Name),
		)
		return
	}

	kubeClient, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		_ = c.Error(utils.ErrInvalidRequest.WrapWithNoMessage(err))
		return
	}

	repo := core.NewKubeWorkflowRepository(kubeClient)

	result, err := repo.Update(c.Request.Context(), namespace, name, payload)
	if err != nil {
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

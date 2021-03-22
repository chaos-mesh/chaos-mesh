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
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	"github.com/gin-gonic/gin"
	"net/http"
)

// StatusResponse defines a common status struct.
type StatusResponse struct {
	Status string `json:"status"`
}

func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/workflows")
	endpoint.GET("", s.listWorkflows)
	endpoint.POST("/new", s.createWorkflow)
	endpoint.GET("/detail/:uid", s.getWorkflowDetail)
	endpoint.DELETE("/:uid", s.deleteWorkflow)
	endpoint.PUT("/:uid", s.updateWorkflow)
}

// Service defines a handler service for workflows.
type Service struct {
	repo core.WorkflowRepository
	conf *config.ChaosDashboardConfig
}

func NewService(repo core.WorkflowRepository, conf *config.ChaosDashboardConfig) *Service {
	return &Service{repo: repo, conf: conf}
}

// @Summary List workflows from Kubernetes cluster.
// @Description List workflows from Kubernetes cluster.
// @Tags workflows
// @Produce json
// @Param namespace query string false "namespace, given empty string means list from all namespace"
// @Param status query string false "status" Enums(Initializing, Running, Errored, Finished)
// @Success 200 {array} core.Workflow
// @Router /workflows [get]
// @Failure 500 {object} utils.APIError
func (it *Service) listWorkflows(c *gin.Context) {

	namespace := c.Query("namespace")
	var result []core.Workflow
	if namespace != "" {
		workflowFromNs, err := it.repo.ListWorkflowWithNamespace(c.Request.Context(), namespace)
		if err != nil {
			utils.SetErrorForGinCtx(c, err)
			return
		}
		result = workflowFromNs
	} else {
		allWorkflow, err := it.repo.ListWorkflowFromAllNamespace(c.Request.Context())
		if err != nil {
			utils.SetErrorForGinCtx(c, err)
			return
		}
		result = allWorkflow
	}

	c.JSON(http.StatusOK, result)
}

// @Summary Get detailed information about the specified workflow.
// @Description Get detailed information about the specified workflow.
// @Tags workflows
// @Produce json
// @Param namespace path string true "namespace"
// @Param name path string true "name"
// @Router /workflows/detail/{namespace}/{name} [GET]
// @Success 200 {object} core.Workflow
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
func (it *Service) getWorkflowDetail(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	result, err := it.repo.GetWorkflowByNamespacedName(c.Request.Context(), namespace, name)
	if err != nil {
		utils.SetErrorForGinCtx(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary Create a new workflow.
// @Description Create a new workflow.
// @Tags workflows
// @Produce json
// @Param request body core.Workflow true "Request body"
// @Success 200 {object} core.Workflow
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /workflows/new [post]
func (it *Service) createWorkflow(c *gin.Context) {
	panic("unimplemented")
}

// @Summary Delete the specified workflow.
// @Description Delete the specified workflow.
// @Tags workflows
// @Produce json
// @Param namespace path string true "namespace"
// @Param name path string true "name"
// @Param force query string true "force" Enums(true, false)
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /workflows/{namespace}/{name} [delete]
func (it *Service) deleteWorkflow(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	err := it.repo.DeleteWorkflowByNamespacedName(c.Request.Context(), namespace, name)
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
// @Param request body core.Workflow true "Request body"
// @Success 200 {object} core.Workflow
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /workflows/update [put]
func (it *Service) updateWorkflow(c *gin.Context) {
	panic("unimplemented")
}

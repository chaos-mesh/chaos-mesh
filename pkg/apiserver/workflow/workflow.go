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
	"github.com/gin-gonic/gin"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
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
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Param status query string false "status" Enums(Initializing, Running, Errored, Finished)
// @Success 200 {array} core.Workflow
// @Router /workflows [get]
// @Failure 500 {object} utils.APIError
func (it *Service) listWorkflows(context *gin.Context) {
	panic("unimplemented")
}

// @Summary Get detailed information about the specified workflow.
// @Description Get detailed information about the specified workflow.
// @Tags workflows
// @Produce json
// @Param uid path string true "uid"
// @Router /workflows/detail/{uid} [GET]
// @Success 200 {object} core.Workflow
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
func (it *Service) getWorkflowDetail(context *gin.Context) {
	panic("unimplemented")
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
func (it *Service) createWorkflow(context *gin.Context) {
	panic("unimplemented")
}

// @Summary Delete the specified workflow.
// @Description Delete the specified workflow.
// @Tags workflows
// @Produce json
// @Param uid path string true "uid"
// @Param force query string true "force" Enums(true, false)
// @Success 200 {object} StatusResponse
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /workflows/{uid} [delete]
func (it *Service) deleteWorkflow(context *gin.Context) {
	panic("unimplemented")
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
func (it *Service) updateWorkflow(context *gin.Context) {
	panic("unimplemented")
}

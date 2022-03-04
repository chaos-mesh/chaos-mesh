// Copyright Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package template

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	u "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/jinzhu/gorm"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"time"
)

// Service defines a handler service for cluster common objects.
type Service struct {
	conf   *config.ChaosDashboardConfig
	logger logr.Logger
}

func NewService(conf *config.ChaosDashboardConfig, logger logr.Logger) *Service {
	return &Service{conf: conf, logger: logger}
}

func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/templates")

	statusCheckEndpoint := endpoint.Group("/statuschecks")
	statusCheckEndpoint.GET("", s.listStatusCheckTemplate)
	statusCheckEndpoint.POST("", s.createStatusCheckTemplate)
	statusCheckEndpoint.GET("/:uid", s.getStatusCheckTemplateDetailByUID)
	statusCheckEndpoint.PUT("/:uid", s.updateStatusCheckTemplate)
	statusCheckEndpoint.DELETE("/:uid", s.deleteStatusCheckTemplate)
}

type StatusCheckTemplate struct {
	core.ObjectBase
}

// StatusCheckTemplateDetail adds KubeObjectDesc on StatusCheckTemplate.
type StatusCheckTemplateDetail struct {
	StatusCheckTemplate
	KubeObject core.KubeObjectDesc `json:"kube_object"`
}

// @Summary List status check templates.
// @Description Get status check templates from k8s cluster in real time.
// @Tags template
// @Produce json
// @Param namespace query string false "filter status check templates by namespace"
// @Param name query string false "filter status check templates by name"
// @Success 200 {array} StatusCheckTemplate
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /templates/statuschecks [get]
func (s *Service) listStatusCheckTemplate(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))
		return
	}

	ns, name := c.Query("namespace"), c.Query("name")
	if ns == "" && !s.conf.ClusterScoped && s.conf.TargetNamespace != "" {
		ns = s.conf.TargetNamespace
		s.logger.Info("Replace query namespace with", ns)
	}

	templateList := v1.ConfigMapList{}
	if err = kubeCli.List(context.Background(), &templateList,
		client.InNamespace(ns),
		client.MatchingLabels{"template-type": "status-check"},
	); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}

	templates := make([]*StatusCheckTemplate, 0)
	for _, template := range templateList.Items {
		if name != "" && template.Name != name {
			continue
		}

		templates = append(templates, &StatusCheckTemplate{
			ObjectBase: core.ObjectBase{
				Namespace: template.Namespace,
				Name:      template.Name,
				Kind:      string(template.Spec.Type),
				UID:       string(template.UID),
				Created:   template.CreationTimestamp.Format(time.RFC3339),
			},
		})
	}

	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Created > templates[j].Created
	})

	c.JSON(http.StatusOK, templates)
}

// @Summary Create a new status check template.
// @Description Pass a JSON object to create a new status check template. The schema for JSON is the same as the YAML schema for the Kubernetes object.
// @Tags templates
// @Accept json
// @Produce json
// @Param schedule body v1alpha1.Schedule true "the schedule definition"
// @Success 200 {object} v1alpha1.Schedule
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /templates/statuschecks [post]
func (s *Service) createStatusCheckTemplate(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	// TODO: mock
	// TODO: need a template store or not?
	var sch v1alpha1.Schedule
	if err = u.ShouldBindBodyWithJSON(c, &sch); err != nil {
		return
	}

	if err = kubeCli.Create(context.Background(), &sch); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}

	c.JSON(http.StatusOK, sch)
}

// @Summary Get a status check template.
// @Description Get the status check template's detail by uid.
// @Tags templates
// @Produce json
// @Param uid path string true "the status check template uid"
// @Success 200 {object} Detail
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /templates/statuschecks/{uid} [get]
func (s *Service) getStatusCheckTemplateDetailByUID(c *gin.Context) {
	var (
		sch       *core.Schedule
		schDetail *Detail
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	uid := c.Param("uid")
	if sch, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrBadRequest.New("Schedule "+uid+"not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	ns, name := sch.Namespace, sch.Name
	schDetail = s.findScheduleInCluster(c, kubeCli, types.NamespacedName{Namespace: ns, Name: name})
	if schDetail == nil {
		return
	}

	c.JSON(http.StatusOK, schDetail)
}

// @Summary Update a status check template.
// @Description Update a status check template.
// @Tags templates
// @Produce json
// @Param uid path string true "the status check template uid"
// @Param request body v1alpha1.Workflow true "Request body"
// @Success 200 {object} core.WorkflowDetail
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /templates/statuschecks/{uid} [put]
func (it *Service) updateStatusCheckTemplate(c *gin.Context) {
	payload := v1alpha1.Workflow{}

	err := json.NewDecoder(c.Request.Body).Decode(&payload)
	if err != nil {
		utils.SetAPIError(c, utils.ErrInternalServer.Wrap(err, "failed to parse request body"))
		return
	}
	uid := c.Param("uid")
	entity, err := it.store.FindByUID(c.Request.Context(), uid)
	if err != nil {
		utils.SetAPImachineryError(c, err)
		return
	}

	namespace := entity.Namespace
	name := entity.Name

	if namespace != payload.Namespace {
		utils.SetAPIError(c, utils.ErrBadRequest.Wrap(err,
			"namespace is not consistent, pathParameter: %s, metaInRaw: %s",
			namespace,
			payload.Namespace))
		return
	}
	if name != payload.Name {
		utils.SetAPIError(c, utils.ErrBadRequest.Wrap(err,
			"name is not consistent, pathParameter: %s, metaInRaw: %s",
			name,
			payload.Name))
		return
	}

	kubeClient, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		utils.SetAPImachineryError(c, err)
		return
	}

	repo := core.NewKubeWorkflowRepository(kubeClient)

	result, err := repo.Update(c.Request.Context(), namespace, name, payload)
	if err != nil {
		utils.SetAPImachineryError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Summary Delete a status check template.
// @Description Delete the status check template by uid.
// @Tags templates
// @Produce json
// @Param uid path string true "the status check template uid"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /templates/statuschecks/{uid} [delete]
func (s *Service) deleteStatusCheckTemplate(c *gin.Context) {
	var (
		sch *core.Schedule
	)

	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))
		return
	}

	uid := c.Param("uid")
	if sch, err = s.schedule.FindByUID(context.Background(), uid); err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Schedule "+uid+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	ns, name := sch.Namespace, sch.Name
	if err = checkAndDeleteSchedule(c, kubeCli, types.NamespacedName{Namespace: ns, Name: name}); err != nil {
		u.SetAPImachineryError(c, err)

		return
	}

	c.JSON(http.StatusOK, u.ResponseSuccess)
}

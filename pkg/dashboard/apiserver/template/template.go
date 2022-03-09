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
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	u "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
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
	statusCheckEndpoint.GET("/statuscheck", s.getStatusCheckTemplateDetail)
	statusCheckEndpoint.PUT("/statuscheck", s.updateStatusCheckTemplate)
	statusCheckEndpoint.DELETE("/statuscheck", s.deleteStatusCheckTemplate)
}

type StatusCheckTemplateBase struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	UID         string `json:"uid"`
	Description string `json:"description,omitempty"`
	Created     string `json:"created_at"`
}

// StatusCheckTemplateDetail represent details of StatusCheckTemplate.
type StatusCheckTemplateDetail struct {
	StatusCheckTemplateBase
	// TODO StatusCheckSpec
}

type StatusCheckTemplateSpec struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// TODO StatusCheckSpec
}

// @Summary List status check templates.
// @Description Get status check templates from k8s cluster in real time.
// @Tags template
// @Produce json
// @Param namespace query string false "filter status check templates by namespace"
// @Param name query string false "filter status check templates by name"
// @Success 200 {array} StatusCheckTemplateBase
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

	templates := make([]*StatusCheckTemplateBase, 0)
	for _, template := range templateList.Items {
		if name != "" && template.Name != name {
			continue
		}

		templates = append(templates, &StatusCheckTemplateBase{
			Namespace:   template.Namespace,
			Name:        template.Name, // TODO
			UID:         string(template.UID),
			Created:     template.CreationTimestamp.Format(time.RFC3339),
			Description: template.Annotations[""], // TODO
		})
	}

	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Created > templates[j].Created
	})

	c.JSON(http.StatusOK, templates)
}

// @Summary Create a new status check template.
// @Description Pass a JSON object to create a new status check template.
// @Tags templates
// @Accept json
// @Produce json
// @Param statuscheck body StatusCheckTemplateSpec true "the status check definition"
// @Success 200 {object} StatusCheckTemplateSpec
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /templates/statuschecks [post]
func (s *Service) createStatusCheckTemplate(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))
		return
	}

	var spec StatusCheckTemplateSpec
	if err = u.ShouldBindBodyWithJSON(c, &spec); err != nil {
		return
	}

	// TODO convert to StatusCheckSpec
	data, err := yaml.Marshal(spec)
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	// TODO
	template := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   spec.Namespace,
			Name:        spec.Name,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Data: map[string]string{"template": string(data)},
	}
	if err = kubeCli.Create(context.Background(), &template); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}

	c.JSON(http.StatusOK, spec)
}

// @Summary Get a status check template.
// @Description Get the status check template's detail by namespaced name.
// @Tags templates
// @Produce json
// @Param namespace query string true "the namespace of status check templates"
// @Param name query string true "the name of status check templates"
// @Success 200 {object} StatusCheckTemplateDetail
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /templates/statuschecks/statuscheck [get]
func (s *Service) getStatusCheckTemplateDetail(c *gin.Context) {
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

	var template v1.ConfigMap
	if err = kubeCli.Get(context.Background(),
		types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &template); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}

	var data StatusCheckTemplateSpec
	if err := yaml.Unmarshal([]byte(template.Data["template"]), &data); err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	detail := StatusCheckTemplateDetail{
		StatusCheckTemplateBase{
			Namespace: template.Namespace,
			Name:      template.Name,
			UID:       string(template.UID),
		},
	}
	c.JSON(http.StatusOK, detail)
}

// @Summary Update a status check template.
// @Description Update a status check template by namespaced name.
// @Tags templates
// @Produce json
// @Param request body StatusCheckTemplateSpec true "Request body"
// @Success 200 {object} StatusCheckTemplateSpec
// @Failure 400 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /templates/statuschecks/statuscheck [put]
func (it *Service) updateStatusCheckTemplate(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))
		return
	}

	var spec StatusCheckTemplateSpec
	if err = u.ShouldBindBodyWithJSON(c, &spec); err != nil {
		return
	}

	var template v1.ConfigMap
	if err = kubeCli.Get(context.Background(),
		types.NamespacedName{
			Namespace: spec.Namespace,
			Name:      spec.Name,
		}, &template); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}

	data, err := yaml.Marshal(spec)
	if err != nil {
		return
	}
	template.Data["template"] = string(data)
	if err := kubeCli.Update(context.Background(), &template); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}
	c.JSON(http.StatusOK, spec)
}

// @Summary Delete a status check template.
// @Description Delete the status check template by namespaced name.
// @Tags templates
// @Produce json
// @Param namespace query string true "the namespace of status check templates"
// @Param name query string true "the name of status check templates"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /templates/statuschecks/statuscheck [delete]
func (s *Service) deleteStatusCheckTemplate(c *gin.Context) {
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

	var template v1.ConfigMap
	if err = kubeCli.Get(context.Background(),
		types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &template); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}

	if err := kubeCli.Delete(context.Background(), &template); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}
	c.JSON(http.StatusOK, u.ResponseSuccess)
}

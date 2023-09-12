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

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	apiservertypes "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/types"
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

// @Summary List status check templates.
// @Description Get status check templates from k8s cluster in real time.
// @Tags template
// @Produce json
// @Param namespace query string false "filter status check templates by namespace"
// @Param name query string false "filter status check templates by name"
// @Success 200 {array} apiservertypes.StatusCheckTemplateBase
// @Failure 400 {object} u.APIError
// @Failure 500 {object} u.APIError
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

	configMapList := v1.ConfigMapList{}
	if err = kubeCli.List(context.Background(), &configMapList,
		client.InNamespace(ns),
		client.MatchingLabels{
			v1alpha1.TemplateTypeLabelKey: v1alpha1.KindStatusCheck,
			v1alpha1.ManagedByLabelKey:    v1alpha1.ManagedByLabelValue,
		},
	); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}

	templates := make([]*apiservertypes.StatusCheckTemplateBase, 0)
	for _, cm := range configMapList.Items {
		templateName := v1alpha1.GetTemplateName(cm)
		if templateName == "" {
			// skip illegal template
			continue
		}
		if name != "" && templateName != name {
			continue
		}

		templates = append(templates, &apiservertypes.StatusCheckTemplateBase{
			Namespace:   cm.Namespace,
			Name:        templateName,
			UID:         string(cm.UID),
			Created:     cm.CreationTimestamp.Format(time.RFC3339),
			Description: v1alpha1.GetTemplateDescription(cm),
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
// @Param statuscheck body apiservertypes.StatusCheckTemplate true "the status check definition"
// @Success 200 {object} apiservertypes.StatusCheckTemplate
// @Failure 400 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /templates/statuschecks [post]
func (s *Service) createStatusCheckTemplate(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))
		return
	}

	var template apiservertypes.StatusCheckTemplate
	if err = u.ShouldBindBodyWithJSON(c, &template); err != nil {
		return
	}
	template.Spec.Default()
	if _, err := template.Spec.Validate(); err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	spec, err := yaml.Marshal(template.Spec)
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: template.Namespace,
			Name:      v1alpha1.GenerateTemplateName(template.Name),
			Labels: map[string]string{
				v1alpha1.TemplateTypeLabelKey: v1alpha1.KindStatusCheck,
				v1alpha1.ManagedByLabelKey:    v1alpha1.ManagedByLabelValue,
			},
			Annotations: map[string]string{
				v1alpha1.TemplateNameAnnotationKey:        template.Name,
				v1alpha1.TemplateDescriptionAnnotationKey: template.Description,
			},
		},
		Data: map[string]string{v1alpha1.StatusCheckTemplateKey: string(spec)},
	}
	if err = kubeCli.Create(context.Background(), &cm); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}

	c.JSON(http.StatusOK, template)
}

// @Summary Get a status check template.
// @Description Get the status check template's detail by namespaced name.
// @Tags templates
// @Produce json
// @Param namespace query string true "the namespace of status check templates"
// @Param name query string true "the name of status check templates"
// @Success 200 {object} apiservertypes.StatusCheckTemplateDetail
// @Failure 400 {object} u.APIError
// @Failure 404 {object} u.APIError
// @Failure 500 {object} u.APIError
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
	if name == "" {
		u.SetAPIError(c, u.ErrBadRequest.New("name is required"))
		return
	}

	var cm v1.ConfigMap
	if err = kubeCli.Get(context.Background(),
		types.NamespacedName{
			Namespace: ns,
			Name:      v1alpha1.GenerateTemplateName(name),
		}, &cm); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}

	if !v1alpha1.IsStatusCheckTemplate(cm) {
		u.SetAPIError(c, u.ErrInternalServer.New("invalid status check template"))
		return
	}

	var spec v1alpha1.StatusCheckTemplate
	if err := yaml.Unmarshal([]byte(cm.Data[v1alpha1.StatusCheckTemplateKey]), &spec); err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	detail := apiservertypes.StatusCheckTemplateDetail{
		StatusCheckTemplateBase: apiservertypes.StatusCheckTemplateBase{
			Namespace:   cm.Namespace,
			Name:        v1alpha1.GetTemplateName(cm),
			UID:         string(cm.UID),
			Created:     cm.CreationTimestamp.Format(time.RFC3339),
			Description: v1alpha1.GetTemplateDescription(cm),
		},
		Spec: spec,
	}
	c.JSON(http.StatusOK, detail)
}

// @Summary Update a status check template.
// @Description Update a status check template by namespaced name.
// @Tags templates
// @Produce json
// @Param request body apiservertypes.StatusCheckTemplate true "Request body"
// @Success 200 {object} apiservertypes.StatusCheckTemplate
// @Failure 400 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /templates/statuschecks/statuscheck [put]
func (s *Service) updateStatusCheckTemplate(c *gin.Context) {
	kubeCli, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))
		return
	}

	var template apiservertypes.StatusCheckTemplate
	if err = u.ShouldBindBodyWithJSON(c, &template); err != nil {
		return
	}
	template.Spec.Default()
	if _, err := template.Spec.Validate(); err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	var cm v1.ConfigMap
	if err = kubeCli.Get(context.Background(),
		types.NamespacedName{
			Namespace: template.Namespace,
			Name:      v1alpha1.GenerateTemplateName(template.Name),
		}, &cm); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}
	if !v1alpha1.IsStatusCheckTemplate(cm) {
		u.SetAPIError(c, u.ErrInternalServer.New("invalid status check template"))
		return
	}

	spec, err := yaml.Marshal(template.Spec)
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	if cm.Data == nil {
		cm.Data = map[string]string{v1alpha1.StatusCheckTemplateKey: string(spec)}
	} else {
		cm.Data[v1alpha1.StatusCheckTemplateKey] = string(spec)
	}
	cm.Annotations[v1alpha1.TemplateDescriptionAnnotationKey] = template.Description

	if err := kubeCli.Update(context.Background(), &cm); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}
	c.JSON(http.StatusOK, template)
}

// @Summary Delete a status check template.
// @Description Delete the status check template by namespaced name.
// @Tags templates
// @Produce json
// @Param namespace query string true "the namespace of status check templates"
// @Param name query string true "the name of status check templates"
// @Success 200 {object} u.Response
// @Failure 400 {object} u.APIError
// @Failure 404 {object} u.APIError
// @Failure 500 {object} u.APIError
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
	if name == "" {
		u.SetAPIError(c, u.ErrBadRequest.New("name is required"))
		return
	}

	var cm v1.ConfigMap
	if err = kubeCli.Get(context.Background(),
		types.NamespacedName{
			Namespace: ns,
			Name:      v1alpha1.GenerateTemplateName(name),
		}, &cm); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}
	if !v1alpha1.IsStatusCheckTemplate(cm) {
		u.SetAPIError(c, u.ErrInternalServer.New("invalid status check template"))
		return
	}

	if err := kubeCli.Delete(context.Background(), &cm); err != nil {
		u.SetAPImachineryError(c, err)
		return
	}
	c.JSON(http.StatusOK, u.ResponseSuccess)
}

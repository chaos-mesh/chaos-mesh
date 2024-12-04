// Copyright 2021 Chaos Mesh Authors.
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

package event

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/jinzhu/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	u "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

// Service defines a handler service for events.
type Service struct {
	event         core.EventStore
	workflowStore core.WorkflowStore
	conf          *config.ChaosDashboardConfig
	logger        logr.Logger
}

func NewService(
	event core.EventStore,
	workflowStore core.WorkflowStore,
	conf *config.ChaosDashboardConfig,
	logger logr.Logger,
) *Service {
	return &Service{
		event:         event,
		workflowStore: workflowStore,
		conf:          conf,
		logger:        logger.WithName("events"),
	}
}

// Register events RouterGroup.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/events")
	endpoint.Use(func(c *gin.Context) {
		u.AuthMiddleware(c, s.conf)
	})

	endpoint.GET("", s.list)
	endpoint.GET("/:id", s.get)
	endpoint.GET("/workflow/:uid", s.cascadeFetchEventsForWorkflow)
}

const layout = "2006-01-02 15:04:05"

// @Summary list events.
// @Description Get events from db.
// @Tags events
// @Produce json
// @Param created_at query string false "The create time of events"
// @Param name query string false "The name of the object"
// @Param namespace query string false "The namespace of the object"
// @Param object_id query string false "The UID of the object"
// @Param kind query string false "kind" Enums(PodChaos, IOChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos, AWSChaos, GCPChaos, DNSChaos, Schedule)
// @Param limit query number false "The max length of events list"
// @Success 200 {array} core.Event
// @Failure 500 {object} u.APIError
// @Router /events [get]
func (s *Service) list(c *gin.Context) {
	ns := c.Query("namespace")

	if ns == "" && !s.conf.ClusterScoped && s.conf.TargetNamespace != "" {
		ns = s.conf.TargetNamespace

		s.logger.V(1).Info("Replace query namespace", "ns", ns)
	}

	start, _ := time.Parse(time.RFC3339, c.Query("start"))
	end, _ := time.Parse(time.RFC3339, c.Query("end"))

	filter := core.Filter{
		ObjectID:  c.Query("object_id"),
		Start:     start.UTC().Format(layout),
		End:       end.UTC().Format(layout),
		Namespace: ns,
		Name:      c.Query("name"),
		Kind:      c.Query("kind"),
		Limit:     c.Query("limit"),
	}

	events, err := s.event.ListByFilter(context.Background(), filter)
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))

		return
	}

	c.JSON(http.StatusOK, events)
}

// @Summary cascadeFetchEventsForWorkflow list all events for Workflow and related WorkflowNode.
// @Description list all events for Workflow and related WorkflowNode.
// @Tags events
// @Produce json
// @Param uid path string true "The UID of the Workflow"
// @Param namespace query string false "The namespace of the object"
// @Param limit query number false "The max length of events list"
// @Success 200 {array} core.Event
// @Failure 500 {object} u.APIError
// @Router /events/workflow/{uid} [get]
func (s *Service) cascadeFetchEventsForWorkflow(c *gin.Context) {
	ctx := c.Request.Context()
	ns := c.Query("namespace")
	uid := c.Param("uid")
	start, _ := time.Parse(time.RFC3339, c.Query("start"))
	end, _ := time.Parse(time.RFC3339, c.Query("end"))
	limit := 0
	limitString := c.Query("limit")
	if len(limitString) > 0 {
		parsedLimit, err := strconv.Atoi(limitString)
		if err != nil {
			u.SetAPIError(c, u.ErrBadRequest.Wrap(err, "parameter limit should be a integer"))
			return
		}
		limit = parsedLimit
	}

	if ns == "" && !s.conf.ClusterScoped && s.conf.TargetNamespace != "" {
		ns = s.conf.TargetNamespace

		s.logger.V(1).Info("Replace query namespace", "ns", ns)
	}

	// we should fetch the events for Workflow and related WorkflowNode, so we need namespaced name at first
	workflowEntity, err := s.workflowStore.FindByUID(ctx, uid)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.Wrap(err, "this requested workflow is not found, uid: %s", uid))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}
		return
	}

	// if workflow has been archived, the Workflow CR and WorkflowNode CR also has been deleted from kubernetes, it's
	// no way to "cascade fetch" anymore
	if workflowEntity.Archived {
		u.SetAPIError(c, u.ErrBadRequest.New("this requested workflow already been archived, can not list events for it"))
		return
	}

	kubeClient, err := clientpool.ExtractTokenAndGetClient(c.Request.Header)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))
		return
	}

	// fetch all related WorkflowNodes
	workflowNodeList := v1alpha1.WorkflowNodeList{}
	controlledByThisWorkflow, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: map[string]string{
		v1alpha1.LabelWorkflow: workflowEntity.Name,
	}})
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))
		return
	}
	err = kubeClient.List(ctx, &workflowNodeList, &client.ListOptions{
		Namespace:     ns,
		LabelSelector: controlledByThisWorkflow,
	})
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	result := make([]*core.Event, 0)
	// fetch events of Workflow
	eventsForWorkflow, err := s.event.ListByFilter(ctx, core.Filter{
		ObjectID:  uid,
		Namespace: ns,
		Start:     start.UTC().Format(layout),
		End:       end.UTC().Format(layout),
	})
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	result = append(result, eventsForWorkflow...)

	// fetch all events of WorkflowNodes
	for _, workflowNode := range workflowNodeList.Items {
		eventsForWorkflowNode, err := s.event.ListByFilter(ctx, core.Filter{
			Namespace: ns,
			Name:      workflowNode.GetName(),
			Start:     start.UTC().Format(layout),
			End:       end.UTC().Format(layout),
		})
		if err != nil {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
			return
		}
		result = append(result, eventsForWorkflowNode...)
	}

	// sort by CreatedAt
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.UnixNano() > result[j].CreatedAt.UnixNano()
	})

	if limit > 0 && len(result) > limit {
		c.JSON(http.StatusOK, result[:limit])
		return
	}
	c.JSON(http.StatusOK, result)
}

// @Summary Get an event.
// @Description Get the event from db by ID.
// @Tags events
// @Produce json
// @Param id path uint true "The event ID"
// @Success 200 {object} core.Event
// @Failure 400 {object} u.APIError
// @Failure 404 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /events/{id} [get]
func (s *Service) get(c *gin.Context) {
	id, ns := c.Param("id"), c.Query("namespace")

	if id == "" {
		u.SetAPIError(c, u.ErrBadRequest.New("ID cannot be empty"))

		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.New("ID is not a number"))

		return
	}

	if ns == "" && !s.conf.ClusterScoped && s.conf.TargetNamespace != "" {
		ns = s.conf.TargetNamespace

		s.logger.V(1).Info("Replace query namespace", "ns", ns)
	}

	event, err := s.event.Find(context.Background(), uint(intID))
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.SetAPIError(c, u.ErrNotFound.New("Event "+id+" not found"))
		} else {
			u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		}

		return
	}

	if len(ns) != 0 && event.Namespace != ns {
		u.SetAPIError(c, u.ErrInternalServer.New("The namespace of event %s is %s instead of the %s in the request", id, event.Namespace, ns))

		return
	}

	c.JSON(http.StatusOK, event)
}

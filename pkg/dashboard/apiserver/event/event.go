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
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	u "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

var extractTokenAndGetClient = clientpool.ExtractTokenAndGetClient

// Service defines a handler service for events.
type Service struct {
	eventStore    core.EventStore
	workflowStore core.WorkflowStore
	conf          *config.ChaosDashboardConfig
	logger        logr.Logger
}

func NewService(
	eventStore core.EventStore,
	workflowStore core.WorkflowStore,
	conf *config.ChaosDashboardConfig,
	logger logr.Logger,
) *Service {
	return &Service{
		eventStore:    eventStore,
		workflowStore: workflowStore,
		conf:          conf,
		logger:        logger.WithName("events"),
	}
}

// Register events RouterGroup.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/events")
	auth := func(c *gin.Context) { u.AuthMiddleware(c, s.conf) }
	endpoint.GET("", s.list)
	endpoint.GET("/page", s.listPage)
	endpoint.GET("/:id", auth, s.get)
	endpoint.GET("/workflow/:uid", auth, s.cascadeFetchEventsForWorkflow)
}

const layout = "2006-01-02 15:04:05"

// @Summary list events.
// @Description Returns an event array ordered by ID descending unless sort/order is supplied. Repeated namespace/name/kind/object_id values are ORed within each field; fields and inclusive time bounds are ANDed.
// @Description Requests without sort/order or repeated filters retain legacy time parsing and second precision. Enhanced requests accept RFC3339 time bounds. Comparison precision depends on the database. Sort ties use ID in the same direction. Pagination parameters are only supported by /events/page.
// @Tags events
// @Produce json
// @Param created_at query string false "Deprecated, ignored; use start and end"
// @Param name query []string false "Object names" collectionFormat(multi)
// @Param namespace query []string false "Object namespaces; every namespace must be authorized" collectionFormat(multi)
// @Param object_id query []string false "Object UIDs" collectionFormat(multi)
// @Param kind query []string false "Object kinds" collectionFormat(multi)
// @Param start query string false "Inclusive creation time lower bound (RFC3339)"
// @Param end query string false "Inclusive creation time upper bound (RFC3339)"
// @Param limit query integer false "Maximum rows; unbounded when omitted"
// @Param sort query string false "Sort field, default id" Enums(id, object_id, namespace, name, kind, created_at, message)
// @Param order query string false "Sort direction, default desc" Enums(asc, desc)
// @Success 200 {array} core.Event
// @Failure 400 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /events [get]
func (s *Service) list(c *gin.Context) {
	request, result := s.queryList(c, false)
	if request == nil {
		return
	}
	c.JSON(http.StatusOK, result.Events)
}

// @Summary list a page of events.
// @Description Returns a paginated event object. Page is one-based and required, with 12 rows by default and a maximum of 1000. Repeated namespace/name/kind/object_id values are ORed within each field; fields and inclusive time bounds are ANDed.
// @Description Time bounds accept RFC3339 timestamps. Comparison precision depends on the database. Sort ties use ID in the same direction. next_page is null when there are no more events. include_total=true additionally returns total and total_pages, with total_pages=0 for no matches. Count and page can change under concurrent writes; separate requests are not a fixed snapshot.
// @Tags events
// @Produce json
// @Param created_at query string false "Deprecated, ignored; use start and end"
// @Param name query []string false "Object names" collectionFormat(multi)
// @Param namespace query []string false "Object namespaces; every namespace must be authorized" collectionFormat(multi)
// @Param object_id query []string false "Object UIDs" collectionFormat(multi)
// @Param kind query []string false "Object kinds" collectionFormat(multi)
// @Param start query string false "Inclusive creation time lower bound (RFC3339)"
// @Param end query string false "Inclusive creation time upper bound (RFC3339)"
// @Param page query integer true "Page number; one-based" minimum(1)
// @Param limit query integer false "Maximum rows, default 12" minimum(1) maximum(1000)
// @Param include_total query boolean false "Include total and total_pages"
// @Param sort query string false "Sort field, default id" Enums(id, object_id, namespace, name, kind, created_at, message)
// @Param order query string false "Sort direction, default desc" Enums(asc, desc)
// @Success 200 {object} PageResponse
// @Failure 400 {object} u.APIError
// @Failure 500 {object} u.APIError
// @Router /events/page [get]
func (s *Service) listPage(c *gin.Context) {
	request, result := s.queryList(c, true)
	if request == nil {
		return
	}
	response := PageResponse{Events: result.Events, Page: request.Page, Limit: request.Limit, Total: result.Total}
	if result.HasMore {
		next := request.Page + 1
		response.NextPage = &next
	}
	if result.Total != nil {
		totalPages := *result.Total / int64(request.Limit)
		if *result.Total%int64(request.Limit) != 0 {
			totalPages++
		}
		response.TotalPages = &totalPages
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) queryList(c *gin.Context, paged bool) (*listRequest, core.EventResult) {
	values, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))
		return nil, core.EventResult{}
	}
	request, err := parseListRequest(values, s.conf, paged)
	if err != nil {
		u.SetAPIError(c, u.ErrBadRequest.WrapWithNoMessage(err))
		return nil, core.EventResult{}
	}
	if !u.AuthorizeNamespaces(c, request.Query.Namespaces) {
		return nil, core.EventResult{}
	}
	var result core.EventResult
	if request.Legacy != nil {
		result.Events, err = s.eventStore.ListByFilter(c.Request.Context(), *request.Legacy)
	} else {
		result, err = s.eventStore.Query(c.Request.Context(), request.Query)
	}
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return nil, core.EventResult{}
	}
	return &request, result
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	kubeClient, err := extractTokenAndGetClient(c.Request.Header)
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
	objectIDs := make([]string, 0, len(workflowNodeList.Items)+1)
	objectIDs = append(objectIDs, uid)
	for _, workflowNode := range workflowNodeList.Items {
		objectIDs = append(objectIDs, string(workflowNode.GetUID()))
	}

	result, err = s.eventStore.ListByUIDListWithFilter(ctx, objectIDs, core.Filter{
		Namespace: ns,
		Start:     start.UTC().Format(layout),
		End:       end.UTC().Format(layout),
	})
	if err != nil {
		u.SetAPIError(c, u.ErrInternalServer.WrapWithNoMessage(err))
		return
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

	event, err := s.eventStore.Find(c.Request.Context(), uint(intID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.SetAPIError(c, u.ErrNotFound.New("Event %s not found", id))
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

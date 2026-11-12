// Copyright 2026 Chaos Mesh Authors.
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

package event

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	authclient "k8s.io/client-go/kubernetes/typed/authorization/v1"
	clienttesting "k8s.io/client-go/testing"

	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type eventAuthClients struct {
	clientpool.Clients
	auth authclient.AuthorizationV1Interface
}

func (c *eventAuthClients) AuthClient(string) (authclient.AuthorizationV1Interface, error) {
	return c.auth, nil
}

func installEventAuth(review func(string) bool) {
	previous := clientpool.K8sClients
	DeferCleanup(func() { clientpool.K8sClients = previous })
	kube := fake.NewClientset()
	kube.PrependReactor("create", "selfsubjectaccessreviews", func(action clienttesting.Action) (bool, runtime.Object, error) {
		attrs := action.(clienttesting.CreateAction).GetObject().(*authorizationv1.SelfSubjectAccessReview).Spec.ResourceAttributes
		Expect(attrs.Verb).To(Equal("list"))
		Expect(attrs.Group).To(Equal("chaos-mesh.org"))
		Expect(attrs.Resource).To(Equal("*"))
		return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: review(attrs.Namespace)}}, nil
	})
	clientpool.K8sClients = &eventAuthClients{auth: kube.AuthorizationV1()}
}

type listEventStore struct {
	core.EventStore
	query  func(core.EventQuery) (core.EventResult, error)
	legacy func(core.Filter) ([]*core.Event, error)
}

func (s *listEventStore) Query(_ context.Context, query core.EventQuery) (core.EventResult, error) {
	return s.query(query)
}

func (s *listEventStore) ListByFilter(_ context.Context, filter core.Filter) ([]*core.Event, error) {
	return s.legacy(filter)
}

var _ = Describe("Event list requests", func() {
	var router *gin.Engine
	var conf *config.ChaosDashboardConfig
	var store *listEventStore
	var reviews []string
	var calls int
	request := func(query string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/events"+query, nil))
		return response
	}
	BeforeEach(func() {
		reviews = nil
		calls = 0
		installEventAuth(func(ns string) bool {
			reviews = append(reviews, ns)
			return ns != "denied"
		})
		conf = &config.ChaosDashboardConfig{ClusterScoped: true}
		store = &listEventStore{
			query: func(core.EventQuery) (core.EventResult, error) {
				calls++
				return core.EventResult{Events: []*core.Event{}}, nil
			},
			legacy: func(core.Filter) ([]*core.Event, error) {
				calls++
				return []*core.Event{}, nil
			},
		}
		router = gin.New()
		Register(router.Group("/api"), &Service{conf: conf, eventStore: store})
	})

	It("authorizes every normalized namespace before querying and returns an optional counted page", func() {
		store.query = func(query core.EventQuery) (core.EventResult, error) {
			calls++
			Expect(reviews).To(Equal([]string{"a", "b"}))
			Expect(query.Namespaces).To(Equal([]string{"a", "b"}))
			Expect(query.Names).To(Equal([]string{"x", "y"}))
			Expect(query.Kinds).To(Equal([]string{"PodChaos", "NetworkChaos"}))
			Expect(query.ObjectIDs).To(Equal([]string{"one", "two"}))
			Expect(query.Window).To(HaveValue(Equal(core.EventWindow{Limit: 12})))
			Expect(query.Sort).To(Equal(core.EventSortID))
			Expect(query.Order).To(Equal(core.EventOrderDesc))
			total := int64(13)
			return core.EventResult{Events: []*core.Event{{ID: 1}}, HasMore: true, Total: &total}, nil
		}
		response := request("/page?page=1&include_total=true&namespace=&namespace=a&namespace=a&namespace=b&name=x&name=y&kind=PodChaos&kind=NetworkChaos&object_id=one&object_id=two")
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(calls).To(Equal(1))
		var page PageResponse
		Expect(json.Unmarshal(response.Body.Bytes(), &page)).To(Succeed())
		Expect(page.Page).To(Equal(1))
		Expect(page.Limit).To(Equal(12))
		Expect(page.NextPage).To(HaveValue(Equal(2)))
		Expect(page.Total).To(HaveValue(Equal(int64(13))))
		Expect(page.TotalPages).To(HaveValue(Equal(int64(2))))
		Expect(page.Events).To(HaveLen(1))
	})

	It("denies any unauthorized namespace on both list routes before accessing the store", func() {
		for _, path := range []string{"?", "/page?page=1&"} {
			reviews = nil
			response := request(path + "namespace=allowed&namespace=denied")
			Expect(response.Code).To(Equal(http.StatusUnauthorized))
			Expect(reviews).To(Equal([]string{"allowed", "denied"}))
			Expect(calls).To(BeZero())
		}
	})

	It("retains authorization for ID and workflow routes", func() {
		for _, path := range []string{"/api/events/1?namespace=denied", "/api/events/workflow/uid?namespace=denied"} {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			Expect(response.Code).To(Equal(http.StatusUnauthorized))
		}
		Expect(reviews).To(Equal([]string{"denied", "denied"}))
		Expect(calls).To(BeZero())
	})

	It("uses the same target namespace or cluster scope for authorization and filtering", func() {
		for _, scoped := range []bool{true, false} {
			conf.ClusterScoped = !scoped
			conf.TargetNamespace = "target"
			expected := []string{}
			scope := ""
			if scoped {
				expected = []string{"target"}
				scope = "target"
			}
			reviews = nil
			store.query = func(query core.EventQuery) (core.EventResult, error) {
				Expect(query.Namespaces).To(Equal(expected))
				return core.EventResult{Events: []*core.Event{}}, nil
			}
			response := request("/page?page=2&namespace=&namespace=")
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(reviews).To(Equal([]string{scope}))
			Expect(response.Body.String()).To(MatchJSON(`{"events":[],"page":2,"limit":12,"next_page":null}`))
		}
	})

	It("returns zero and deep-page counts without rounding overflow", func() {
		for _, total := range []int64{0, math.MaxInt64} {
			store.query = func(query core.EventQuery) (core.EventResult, error) {
				Expect(query.Window).To(HaveValue(Equal(core.EventWindow{Offset: 24, Limit: 12})))
				return core.EventResult{Events: []*core.Event{}, Total: &total}, nil
			}
			response := request("/page?page=3&include_total=true")
			Expect(response.Code).To(Equal(http.StatusOK))
			var page PageResponse
			Expect(json.Unmarshal(response.Body.Bytes(), &page)).To(Succeed())
			Expect(page.NextPage).To(BeNil())
			expected := total / 12
			if total%12 != 0 {
				expected++
			}
			Expect(page.TotalPages).To(HaveValue(Equal(expected)))
		}
	})

	It("rejects invalid or duplicate controls before authorization or querying", func() {
		for _, query := range []string{
			"", "?limit=12", "?page=", "?page=0", "?page=-1", "?page=one", "?page=9223372036854775808",
			fmt.Sprintf("?page=%d&limit=2", math.MaxInt),
			"?page=1&limit=0", "?page=1&limit=-1", "?page=1&limit=1001", "?page=1&limit=",
			"?page=1&limit=one", "?include_total=false", "?page=1&include_total=maybe",
			"?page=1&sort=id;DROP+TABLE+events", "?page=1&sort=id%3BDROP+TABLE+events", "?page=1&sort=", "?page=1&order=sideways",
			"?page=1&start=bad", "?page=1&end=bad", "?page=1&start=",
			"?page=1&page=1", "?page=1&limit=12&limit=12", "?sort=id&sort=name",
			"?order=asc&order=desc", "?start=bad&start=bad", "?end=bad&end=bad",
			"?page=1&include_total=true&include_total=false",
		} {
			response := request("/page" + query)
			Expect(response.Code).To(Equal(http.StatusBadRequest), "%s: %s", query, response.Body)
		}
		Expect(reviews).To(BeEmpty())
		Expect(calls).To(BeZero())
	})

	It("rejects pagination controls on the array route without querying", func() {
		for _, query := range []string{"?page=1", "?include_total=false", "?page=1&include_total=true"} {
			response := request(query)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		}
		Expect(reviews).To(BeEmpty())
		Expect(calls).To(BeZero())
	})

	It("returns an empty array when the array route has no matches", func() {
		for _, query := range []string{"", "?namespace=a&namespace=b&sort=id"} {
			response := request(query)
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(MatchJSON(`[]`))
		}
	})

	It("preserves the legacy array path and its old time and limit values", func() {
		store.legacy = func(filter core.Filter) ([]*core.Event, error) {
			calls++
			Expect(filter.Start).To(Equal("2026-11-01 00:00:00"))
			Expect(filter.End).To(Equal("0001-01-01 00:00:00"))
			Expect(filter.Limit).To(Equal("999"))
			Expect(filter.ObjectID).To(Equal("old-caller"))
			return []*core.Event{{ID: 7}}, nil
		}
		response := request("?object_id=old-caller&limit=999&start=2026-11-01T08:00:00.123456789%2B08:00&end=ignored")
		Expect(response.Code).To(Equal(http.StatusOK))
		var events []*core.Event
		Expect(json.Unmarshal(response.Body.Bytes(), &events)).To(Succeed())
		Expect(events).To(HaveLen(1))
		Expect(calls).To(Equal(1))
		for _, limit := range []string{"", "-1", "0", "invalid"} {
			values := url.Values{"limit": []string{limit}}
			parsed, err := parseListRequest(values, conf, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.Legacy.Limit).To(Equal(limit))
		}
	})

	It("uses full-precision bounds with sort or repeated filters while still returning arrays without page", func() {
		store.query = func(query core.EventQuery) (core.EventResult, error) {
			calls++
			Expect(query.Start.Nanosecond()).To(Equal(123456789))
			Expect(query.Start.Equal(time.Date(2026, 11, 1, 0, 0, 0, 123456789, time.UTC))).To(BeTrue())
			Expect(query.Window).To(HaveValue(Equal(core.EventWindow{Limit: 2})))
			return core.EventResult{Events: []*core.Event{{ID: 7}}}, nil
		}
		for _, controls := range []string{"sort=created_at&order=asc", "name=x&name=y"} {
			response := request("?" + controls + "&limit=2&start=2026-11-01T08:00:00.123456789%2B08:00")
			Expect(response.Code).To(Equal(http.StatusOK))
			var events []*core.Event
			Expect(json.Unmarshal(response.Body.Bytes(), &events)).To(Succeed())
			Expect(events).To(HaveLen(1))
		}
		Expect(calls).To(Equal(2))
	})
})

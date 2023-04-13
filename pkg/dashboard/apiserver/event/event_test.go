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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	pkgmock "github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// MockEventService is a mock of core.EventStore
type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) List(context.Context) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) ListByExperiment(context.Context, string, string, string) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) ListByUID(context.Context, string) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) ListByUIDs(context.Context, []string) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) ListByFilter(ctx context.Context, filter core.Filter) ([]*core.Event, error) {
	var res []*core.Event
	var err error
	if filter.ObjectID == "testUID" {
		event := &core.Event{
			ID:        0,
			CreatedAt: time.Time{},
			Kind:      "testKind",
			Type:      "testType",
			Reason:    "testReason",
			Message:   "testMessage",
			Name:      "testName",
			Namespace: "testNamespace",
			ObjectID:  "testUID",
		}
		res = append(res, event)
	} else {
		err = errors.New("test err")
	}
	return res, err
}

func (m *MockEventService) Find(_ context.Context, id uint) (*core.Event, error) {
	var res *core.Event
	var err error
	if id == 0 {
		res = &core.Event{
			ID:        0,
			CreatedAt: time.Time{},
			Kind:      "testKind",
			Type:      "testType",
			Reason:    "testReason",
			Message:   "testMessage",
			Name:      "testName",
			Namespace: "testNamespace",
			ObjectID:  "testUID",
		}
	} else {
		if id == 1 {
			err = gorm.ErrRecordNotFound
		} else {
			err = errors.New("test err")
		}
	}
	return res, err
}

func (m *MockEventService) Create(context.Context, *core.Event) error {
	panic("implement me")
}

func (m *MockEventService) DeleteByUIDs(context.Context, []string) error {
	panic("implement me")
}

func (m *MockEventService) DeleteByCreateTime(context.Context, time.Duration) error {
	panic("implement me")
}

func (m *MockEventService) DeleteByUID(context.Context, string) error {
	panic("implement me")
}

func (m *MockEventService) DeleteByTime(context.Context, string, string) error {
	panic("implement me")
}

func (m *MockEventService) DeleteByDuration(context.Context, time.Duration) error {
	panic("implement me")
}

func TestEvent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Event Suite")
}

var _ = Describe("event", func() {
	var router *gin.Engine
	BeforeEach(func() {
		pkgmock.With("AuthMiddleware", true)

		mockes := new(MockEventService)

		var s = Service{
			conf: &config.ChaosDashboardConfig{
				ClusterScoped: true,
			},
			event: mockes,
		}
		router = gin.Default()
		r := router.Group("/api")
		endpoint := r.Group("/events")
		endpoint.GET("", s.list)
		endpoint.GET("/:id", s.get)
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
		pkgmock.Reset("AuthMiddleware")
	})

	Context("ListEvents", func() {
		It("success", func() {
			response := []*core.Event{
				{
					ID:        0,
					CreatedAt: time.Time{},
					Kind:      "testKind",
					Type:      "testType",
					Reason:    "testReason",
					Message:   "testMessage",
					Name:      "testName",
					Namespace: "testNamespace",
					ObjectID:  "testUID",
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events?object_id=testUID", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("test err", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events?object_id=err", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})
	})

	Context("GetEvent", func() {
		It("not found", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events/1", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusNotFound))
		})
	})
})

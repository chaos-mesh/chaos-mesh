// Copyright 2020 Chaos Mesh Authors.
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

package event

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

// MockEventService is a mock type for event.Service
type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) List(context.Context) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) ListByExperiment(context.Context, string, string) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) ListByNamespace(context.Context, string) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) ListByPod(context.Context, string, string) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) ListByUID(context.Context, string) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) DryListByFilter(context.Context, core.Filter) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) Find(context.Context, uint) (*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) FindByExperimentAndStartTime(context.Context, string, string, *time.Time) (*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) Create(context.Context, *core.Event) error {
	panic("implement me")
}

func (m *MockEventService) Update(context.Context, *core.Event) error {
	panic("implement me")
}

func (m *MockEventService) DeleteIncompleteEvents(context.Context) error {
	panic("implement me")
}

func (m *MockEventService) DeleteByFinishTime(context.Context, time.Duration) error {
	panic("implement me")
}

func (m *MockEventService) UpdateIncompleteEvents(context.Context, string, string) error {
	panic("implement me")
}

func TestEvent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Event Suite")
}

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func (m *MockEventService) ListByFilter(ctx context.Context, filter core.Filter) ([]*core.Event, error) {
	return nil, nil
	//ret := m.Called(ctx, filter)
	//return ret.Get(0).([]*core.Event), ret.Get(1).(error)
}

var _ = Describe("event", func() {
	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	Context("ListEvents", func() {
		It("empty podNamespace", func() {
			mockes := new(MockEventService)
			mockes.On("ListByFilter", mock.Anything, mock.Anything).Return(nil, nil)

			s := Service{
				conf:    nil,
				kubeCli: nil,
				archive: nil,
				event:   mockes,
			}
			router := gin.Default()
			r := router.Group("/api")
			endpoint := r.Group("/events")
			endpoint.GET("", s.ListEvents)
			rr := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, "/api/events?podName=testpodNamespace", nil)

			Expect(err).ShouldNot(HaveOccurred())
			router.ServeHTTP(rr, request)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})

		It("success", func() {
			mockes := new(MockEventService)
			mockes.On("ListByFilter", mock.Anything, mock.Anything).Return(nil, nil)

			s := Service{
				conf:    nil,
				kubeCli: nil,
				archive: nil,
				event:   mockes,
			}
			router := gin.Default()
			r := router.Group("/api")
			endpoint := r.Group("/events")
			endpoint.GET("", s.ListEvents)
			rr := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, "/api/events", nil)

			Expect(err).ShouldNot(HaveOccurred())

			router.ServeHTTP(rr, request)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Code).Should(Equal(http.StatusOK))
		})
	})
})

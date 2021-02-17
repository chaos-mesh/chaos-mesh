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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	pkgmock "github.com/chaos-mesh/chaos-mesh/pkg/mock"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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

func (m *MockEventService) ListByFilter(ctx context.Context, filter core.Filter) ([]*core.Event, error) {
	var res []*core.Event
	var err error
	if filter.UID == "testUID" {
		event := &core.Event{
			ID:           0,
			CreatedAt:    time.Time{},
			UpdatedAt:    time.Time{},
			DeletedAt:    nil,
			Experiment:   "testExperiment",
			Namespace:    "testNamespace",
			Kind:         "testKind",
			Message:      "test",
			StartTime:    nil,
			FinishTime:   nil,
			Duration:     "1m",
			Pods:         nil,
			ExperimentID: "testUID",
		}
		res = append(res, event)
	} else {
		err = fmt.Errorf("test err")
	}
	return res, err
}

func (m *MockEventService) DryListByFilter(_ context.Context, filter core.Filter) ([]*core.Event, error) {
	var res []*core.Event
	var err error
	if filter.Kind == "testKind" {
		event := &core.Event{
			ID:           0,
			CreatedAt:    time.Time{},
			UpdatedAt:    time.Time{},
			DeletedAt:    nil,
			Experiment:   "testExperiment",
			Namespace:    "testNamespace",
			Kind:         "testKind",
			Message:      "test",
			StartTime:    nil,
			FinishTime:   nil,
			Duration:     "1m",
			Pods:         nil,
			ExperimentID: "testUID",
		}
		res = append(res, event)
	} else {
		err = fmt.Errorf("test err")
	}
	return res, err
}

func (m *MockEventService) Find(_ context.Context, id uint) (*core.Event, error) {
	var res *core.Event
	var err error
	if id == 0 {
		res = &core.Event{
			ID:           0,
			CreatedAt:    time.Time{},
			UpdatedAt:    time.Time{},
			DeletedAt:    nil,
			Experiment:   "testExperiment",
			Namespace:    "testNamespace",
			Kind:         "testKind",
			Message:      "test",
			StartTime:    nil,
			FinishTime:   nil,
			Duration:     "1m",
			Pods:         nil,
			ExperimentID: "testUID",
		}
	} else {
		if id == 1 {
			err = gorm.ErrRecordNotFound
		} else {
			err = fmt.Errorf("test err")
		}
	}
	return res, err
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

func (m *MockEventService) DeleteByUID(context.Context, string) error {
	panic("implement me")
}

func (m *MockEventService) UpdateIncompleteEvents(context.Context, string, string) error {
	panic("implement me")
}

func TestEvent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Event Suite")
}

var _ = Describe("event", func() {
	var router *gin.Engine
	BeforeEach(func() {
		pkgmock.With("MockAuthRequired", true)

		mockes := new(MockEventService)

		s := Service{
			conf:    nil,
			archive: nil,
			event:   mockes,
		}
		router = gin.Default()
		r := router.Group("/api")
		endpoint := r.Group("/events")
		endpoint.GET("", s.listEvents)
		endpoint.GET("/dry", s.listDryEvents)
		endpoint.GET("/get", s.getEvent)
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
		pkgmock.Reset("MockAuthRequired")
	})

	Context("ListEvents", func() {
		It("empty podNamespace", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events?podName=testpodNamespace", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})

		It("success", func() {
			response := []*core.Event{
				&core.Event{
					ID:           0,
					CreatedAt:    time.Time{},
					UpdatedAt:    time.Time{},
					DeletedAt:    nil,
					Experiment:   "testExperiment",
					Namespace:    "testNamespace",
					Kind:         "testKind",
					Message:      "test",
					StartTime:    nil,
					FinishTime:   nil,
					Duration:     "1m",
					Pods:         nil,
					ExperimentID: "testUID",
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events?uid=testUID", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("test err", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events?uid=err", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})
	})

	Context("ListDryEvents", func() {
		It("success", func() {
			response := []*core.Event{
				&core.Event{
					ID:           0,
					CreatedAt:    time.Time{},
					UpdatedAt:    time.Time{},
					DeletedAt:    nil,
					Experiment:   "testExperiment",
					Namespace:    "testNamespace",
					Kind:         "testKind",
					Message:      "test",
					StartTime:    nil,
					FinishTime:   nil,
					Duration:     "1m",
					Pods:         nil,
					ExperimentID: "testUID",
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events/dry?kind=testKind", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("test err", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events/dry?kind=err", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})
	})

	Context("GetEvent", func() {
		It("success", func() {
			response := &core.Event{
				ID:           0,
				CreatedAt:    time.Time{},
				UpdatedAt:    time.Time{},
				DeletedAt:    nil,
				Experiment:   "testExperiment",
				Namespace:    "testNamespace",
				Kind:         "testKind",
				Message:      "test",
				StartTime:    nil,
				FinishTime:   nil,
				Duration:     "1m",
				Pods:         nil,
				ExperimentID: "testUID",
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events/get?id=0", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("empty id", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events/get", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusBadRequest))
		})

		It("bad id", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events/get?id=badID", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusBadRequest))
		})

		It("not found", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events/get?id=1", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})

		It("other err", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events/get?id=2", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})
	})
})

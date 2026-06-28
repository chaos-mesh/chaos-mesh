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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	pkgmock "github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// MockEventService is a mock of core.EventStore
type MockEventService struct {
	mock.Mock
}

type MockWorkflowStore struct {
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

func (m *MockEventService) ListByUIDList(context.Context, []string) ([]*core.Event, error) {
	panic("implement me")
}

func (m *MockEventService) ListByUIDListWithFilter(ctx context.Context, uids []string, filter core.Filter) ([]*core.Event, error) {
	args := m.Called(ctx, uids, filter)
	return args.Get(0).([]*core.Event), args.Error(1)
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

func (m *MockEventService) DeleteByUIDList(context.Context, []string) error {
	panic("implement me")
}

func (m *MockEventService) DeleteByCreateTime(context.Context, time.Duration) error {
	panic("implement me")
}

func (m *MockEventService) DeleteByUID(context.Context, string) error {
	panic("implement me")
}

func (m *MockEventService) DeleteByDuration(context.Context, time.Duration) error {
	panic("implement me")
}

func (m *MockWorkflowStore) List(context.Context, string, string, bool) ([]*core.WorkflowEntity, error) {
	panic("implement me")
}

func (m *MockWorkflowStore) ListMeta(context.Context, string, string, bool) ([]*core.WorkflowMeta, error) {
	panic("implement me")
}

func (m *MockWorkflowStore) FindByID(context.Context, uint) (*core.WorkflowEntity, error) {
	panic("implement me")
}

func (m *MockWorkflowStore) FindByUID(_ context.Context, uid string) (*core.WorkflowEntity, error) {
	args := m.Called(uid)
	if entity := args.Get(0); entity != nil {
		return entity.(*core.WorkflowEntity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockWorkflowStore) FindMetaByUID(context.Context, string) (*core.WorkflowMeta, error) {
	panic("implement me")
}

func (m *MockWorkflowStore) Save(context.Context, *core.WorkflowEntity) error {
	panic("implement me")
}

func (m *MockWorkflowStore) DeleteByUID(context.Context, string) error {
	panic("implement me")
}

func (m *MockWorkflowStore) DeleteByUIDs(context.Context, []string) error {
	panic("implement me")
}

func (m *MockWorkflowStore) DeleteByFinishTime(context.Context, time.Duration) error {
	panic("implement me")
}

func (m *MockWorkflowStore) MarkAsArchived(context.Context, string, string) error {
	panic("implement me")
}

func (m *MockWorkflowStore) MarkAsArchivedWithUID(context.Context, string) error {
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
		extractTokenAndGetClient = clientpool.ExtractTokenAndGetClient

		mockes := new(MockEventService)
		workflowStore := new(MockWorkflowStore)

		var s = Service{
			conf: &config.ChaosDashboardConfig{
				ClusterScoped: true,
			},
			event:         mockes,
			workflowStore: workflowStore,
		}
		router = gin.Default()
		r := router.Group("/api")
		endpoint := r.Group("/events")
		endpoint.GET("", s.list)
		endpoint.GET("/:id", s.get)
		endpoint.GET("/workflow/:uid", s.cascadeFetchEventsForWorkflow)
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
		pkgmock.Reset("AuthMiddleware")
		extractTokenAndGetClient = clientpool.ExtractTokenAndGetClient
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

	Context("CascadeFetchEventsForWorkflow", func() {
		It("sorts by CreatedAt and applies limit after batching", func() {
			eventStore := new(MockEventService)
			workflowStore := new(MockWorkflowStore)

			scheme := runtime.NewScheme()
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

			node1 := &v1alpha1.WorkflowNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "node-1",
					Namespace: "default",
					UID:       types.UID("node-uid-1"),
					Labels: map[string]string{
						v1alpha1.LabelWorkflow: "demo-workflow",
					},
				},
			}
			node2 := &v1alpha1.WorkflowNode{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "node-2",
					Namespace: "default",
					UID:       types.UID("node-uid-2"),
					Labels: map[string]string{
						v1alpha1.LabelWorkflow: "demo-workflow",
					},
				},
			}

			kubeClient := fakeclient.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(node1, node2).
				Build()

			extractTokenAndGetClient = func(http.Header) (client.Client, error) {
				return kubeClient, nil
			}

			workflowStore.On("FindByUID", "workflow-uid").Return(&core.WorkflowEntity{
				WorkflowMeta: core.WorkflowMeta{
					UID:       "workflow-uid",
					Namespace: "default",
					Name:      "demo-workflow",
					Archived:  false,
				},
			}, nil)

			oldest := &core.Event{
				ID:        1,
				ObjectID:  "workflow-uid",
				CreatedAt: time.Unix(10, 0),
				Namespace: "default",
				Name:      "demo-workflow",
			}
			newest := &core.Event{
				ID:        2,
				ObjectID:  "node-uid-2",
				CreatedAt: time.Unix(30, 0),
				Namespace: "default",
				Name:      "node-2",
			}
			middle := &core.Event{
				ID:        3,
				ObjectID:  "node-uid-1",
				CreatedAt: time.Unix(20, 0),
				Namespace: "default",
				Name:      "node-1",
			}

			eventStore.On("ListByUIDListWithFilter", mock.Anything, mock.AnythingOfType("[]string"), core.Filter{
				Namespace: "default",
				Start:     "0001-01-01 00:00:00",
				End:       "0001-01-01 00:00:00",
			}).Return([]*core.Event{oldest, middle, newest}, nil)

			service := Service{
				conf: &config.ChaosDashboardConfig{
					ClusterScoped: true,
				},
				event:         eventStore,
				workflowStore: workflowStore,
			}

			localRouter := gin.Default()
			r := localRouter.Group("/api")
			endpoint := r.Group("/events")
			endpoint.GET("/workflow/:uid", service.cascadeFetchEventsForWorkflow)

			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/events/workflow/workflow-uid?namespace=default&limit=2", nil)
			localRouter.ServeHTTP(rr, request)

			Expect(rr.Code).Should(Equal(http.StatusOK))

			var response []*core.Event
			Expect(json.Unmarshal(rr.Body.Bytes(), &response)).To(Succeed())
			Expect(response).To(HaveLen(2))
			Expect(response[0].ObjectID).To(Equal("node-uid-2"))
			Expect(response[1].ObjectID).To(Equal("node-uid-1"))

			eventStore.AssertExpectations(GinkgoT())
			workflowStore.AssertExpectations(GinkgoT())
		})
	})
})

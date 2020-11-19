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
	"testing"

	//"database/sql"
	"fmt"
	//"github.com/DATA-DOG/go-sqlmock"
	//"github.com/chaos-mesh/chaos-mesh/pkg/store/event"
	"github.com/gin-gonic/gin"
	//"github.com/jinzhu/gorm"
	//"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	//"testing"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"

	"github.com/stretchr/testify/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

func (m *MockEventService) ListByFilter (ctx context.Context, filter core.Filter) ([]*core.Event, error) {
	fmt.Println("?weishenme ")
	return nil,nil
	//ret := m.Called(ctx, filter)
	//return ret.Get(0).([]*core.Event), ret.Get(1).(error)
}

//func TestListEvents(t *testing.T) {
//	gin.SetMode(gin.TestMode)
//
//	t.Run("Success", func(t *testing.T) {
//
//
//		mockes := new(MockEventService)
//		//mockService.On("Get", mock.AnythingOfType("*gin.Context"), uid).Return(mockUserResp, nil)
//		mockes.On("ListByFilter", mock.Anything,mock.Anything).Return(nil, nil)
//
//		rr := httptest.NewRecorder()
//
//		s := Service{
//			conf:    nil,
//			kubeCli: nil,
//			archive: nil,
//			event:   mockes,
//		}
//
//
//		router := gin.Default()
//		r := router.Group("/api")
//		endpoint := r.Group("/events")
//		endpoint.GET("", s.ListEvents)
//
//		//_ = router.Run("127.0.0.1:2333")
//		//router.Use(func(c *gin.Context) {
//		//	c.Set("user", &model.User{
//		//		UID: uid,
//		//	},
//		//	)
//		//})
//
//		//NewHandler(&Config{
//		//	R:           router,
//		//	UserService: mockUserService,
//		//})
//
//
//		request, err := http.NewRequest(http.MethodGet, "/api/events", nil)
//		assert.NoError(t, err)
//
//		router.ServeHTTP(rr, request)
//		fmt.Println("!!!!!!?????")
//		fmt.Println(rr.Body)
//		fmt.Println("!!!!!!")
//		//respBody, err := json.Marshal(gin.H{
//		//	"user": mockUserResp,
//		//})
//
//		assert.NoError(t, err)
//		assert.Equal(t, 200, rr.Code)
//		//assert.Equal(t, respBody, rr.Body.Bytes())
//		//mockes.AssertExpectations(t)
//	})
//}


var _ = Describe("event", func() {
	//var (
	//	s          *Service
	//	mock       sqlmock.Sqlmock
	//
	//)
	//
	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
		//var db *sql.DB
		//var err error
		//db, mock, err = sqlmock.New()
		//Expect(err).ShouldNot(HaveOccurred())
		//
		//gdb, err := gorm.Open("sqlite3", db)
		//Expect(err).ShouldNot(HaveOccurred())
		//
		//xs := event.eventStore
		//
		//s = &Service{
		//	conf:    nil,
		//	kubeCli: nil,
		//	archive: nil,
		//	event: nil,
		//	//event:   &event.eventStore{db: &dbstore.DB{DB: gdb},
		//}
		////s = &Service{db: &dbstore.DB{DB: gdb}}

	})

	AfterEach(func() {
		// Add any setup steps that needs to be executed before each test
	})
	Context("ListEvents", func() {
		It("success", func() {
			fmt.Println("这里这里这里")
			mockes := new(MockEventService)
			//mockService.On("Get", mock.AnythingOfType("*gin.Context"), uid).Return(mockUserResp, nil)
			mockes.On("ListByFilter", mock.Anything,mock.Anything).Return(nil, nil)

			rr := httptest.NewRecorder()

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

			//_ = router.Run("127.0.0.1:2333")
			//router.Use(func(c *gin.Context) {
			//	c.Set("user", &model.User{
			//		UID: uid,
			//	},
			//	)
			//})

			//NewHandler(&Config{
			//	R:           router,
			//	UserService: mockUserService,
			//})


			request, err := http.NewRequest(http.MethodGet, "/api/events", nil)

			Expect(err).ShouldNot(HaveOccurred())

			router.ServeHTTP(rr, request)
			fmt.Println("!!!!!!?????")
			fmt.Println(rr.Body)
			fmt.Println("!!!!!!")
			//respBody, err := json.Marshal(gin.H{
			//	"user": mockUserResp,
			//})

			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Code).Should(Equal(200))
		})
	})
})


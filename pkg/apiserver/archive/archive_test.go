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

package archive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"
	pkgmock "github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// MockExperimentStore is a mock type for ExperimentStore
type MockExperimentStore struct {
	mock.Mock
}

// MockScheduleStore is a mock type for ScheduleStore
type MockScheduleStore struct {
	mock.Mock
}

func TestEvent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Archive Suite")
}

func (m *MockExperimentStore) ListMeta(ctx context.Context, kind, namespace, name string, archived bool) ([]*core.ExperimentMeta, error) {
	var res []*core.ExperimentMeta
	var err error
	if kind == "testKind" {
		expMeta := &core.ExperimentMeta{
			UID:        "testUID",
			Kind:       "testKind",
			Name:       "testName",
			Namespace:  "testNamespace",
			Action:     "testAction",
			StartTime:  time.Time{},
			FinishTime: time.Time{},
			Archived:   true,
		}
		res = append(res, expMeta)
	} else {
		err = fmt.Errorf("test err")
	}
	return res, err
}

func (m *MockExperimentStore) FindByUID(ctx context.Context, UID string) (*core.Experiment, error) {
	var res *core.Experiment
	var err error
	switch UID {
	case "testPodChaos":
		chaos := v1alpha1.PodChaos{}
		jsonStr, _ := json.Marshal(chaos)
		res = &core.Experiment{
			ExperimentMeta: core.ExperimentMeta{
				UID:        UID,
				Kind:       v1alpha1.KindPodChaos,
				Name:       "testName",
				Namespace:  "testNamespace",
				Action:     "testAction",
				StartTime:  time.Time{},
				FinishTime: time.Time{},
				Archived:   true,
			},
			Experiment: string(jsonStr),
		}
	case "testIOChaos":
		chaos := v1alpha1.IOChaos{}
		jsonStr, _ := json.Marshal(chaos)
		res = &core.Experiment{
			ExperimentMeta: core.ExperimentMeta{
				UID:        UID,
				Kind:       v1alpha1.KindIOChaos,
				Name:       "testName",
				Namespace:  "testNamespace",
				Action:     "testAction",
				StartTime:  time.Time{},
				FinishTime: time.Time{},
				Archived:   true,
			},
			Experiment: string(jsonStr),
		}
	case "testNetworkChaos":
		chaos := v1alpha1.NetworkChaos{}
		jsonStr, _ := json.Marshal(chaos)
		res = &core.Experiment{
			ExperimentMeta: core.ExperimentMeta{
				UID:        UID,
				Kind:       v1alpha1.KindNetworkChaos,
				Name:       "testName",
				Namespace:  "testNamespace",
				Action:     "testAction",
				StartTime:  time.Time{},
				FinishTime: time.Time{},
				Archived:   true,
			},
			Experiment: string(jsonStr),
		}
	case "testTimeChaos":
		chaos := v1alpha1.TimeChaos{}
		jsonStr, _ := json.Marshal(chaos)
		res = &core.Experiment{
			ExperimentMeta: core.ExperimentMeta{
				UID:        UID,
				Kind:       v1alpha1.KindTimeChaos,
				Name:       "testName",
				Namespace:  "testNamespace",
				Action:     "testAction",
				StartTime:  time.Time{},
				FinishTime: time.Time{},
				Archived:   true,
			},
			Experiment: string(jsonStr),
		}
	case "testKernelChaos":
		chaos := v1alpha1.KernelChaos{}
		jsonStr, _ := json.Marshal(chaos)
		res = &core.Experiment{
			ExperimentMeta: core.ExperimentMeta{
				UID:        UID,
				Kind:       v1alpha1.KindKernelChaos,
				Name:       "testName",
				Namespace:  "testNamespace",
				Action:     "testAction",
				StartTime:  time.Time{},
				FinishTime: time.Time{},
				Archived:   true,
			},
			Experiment: string(jsonStr),
		}
	case "testStressChaos":
		chaos := v1alpha1.StressChaos{}
		jsonStr, _ := json.Marshal(chaos)
		res = &core.Experiment{
			ExperimentMeta: core.ExperimentMeta{
				UID:        UID,
				Kind:       v1alpha1.KindStressChaos,
				Name:       "testName",
				Namespace:  "testNamespace",
				Action:     "testAction",
				StartTime:  time.Time{},
				FinishTime: time.Time{},
				Archived:   true,
			},
			Experiment: string(jsonStr),
		}
	case "testOtherChaos":
		res = &core.Experiment{
			ExperimentMeta: core.ExperimentMeta{
				UID:        UID,
				Kind:       "OtherChaos",
				Name:       "testName",
				Namespace:  "testNamespace",
				Action:     "testAction",
				StartTime:  time.Time{},
				FinishTime: time.Time{},
				Archived:   true,
			},
			Experiment: "",
		}
	case "testErrRecordNotFound":
		err = gorm.ErrRecordNotFound
	default:
		err = fmt.Errorf("test err")
	}
	return res, err
}

func (m *MockExperimentStore) FindMetaByUID(ctx context.Context, UID string) (*core.ExperimentMeta, error) {
	var res *core.ExperimentMeta
	var err error
	switch UID {
	case "tsetUID":
		res = &core.ExperimentMeta{
			UID:        "testUID",
			Kind:       "testKind",
			Name:       "testName",
			Namespace:  "testNamespace",
			Action:     "testAction",
			StartTime:  time.Time{},
			FinishTime: time.Time{},
			Archived:   true,
		}
	case "testErrRecordNotFound":
		err = gorm.ErrRecordNotFound
	default:
		err = fmt.Errorf("test err")
	}
	return res, err
}

func (m *MockExperimentStore) Set(context.Context, *core.Experiment) error {
	panic("implement me")
}

func (m *MockExperimentStore) Archive(ctx context.Context, namespace, name string) error {
	panic("implement me")
}

func (m *MockExperimentStore) Delete(context.Context, *core.Experiment) error {
	panic("implement me")
}

func (m *MockExperimentStore) DeleteByFinishTime(context.Context, time.Duration) error {
	panic("implement me")
}

func (m *MockExperimentStore) DeleteIncompleteExperiments(context.Context) error {
	panic("implement me")
}

func (m *MockExperimentStore) DeleteByUIDs(context.Context, []string) error {
	panic("implement me")
}

func (m *MockScheduleStore) ListMeta(ctx context.Context, namespace, name string, archived bool) ([]*core.ScheduleMeta, error) {
	var res []*core.ScheduleMeta
	var err error
	if name == "testScheduleName" {
		schMeta := &core.ScheduleMeta{
			UID:        "testUID",
			Kind:       "testKind",
			Name:       "testScheduleName",
			Namespace:  "testNamespace",
			Action:     "testAction",
			StartTime:  time.Time{},
			FinishTime: time.Time{},
			Archived:   true,
		}
		res = append(res, schMeta)
	} else {
		err = fmt.Errorf("test err")
	}
	return res, err
}

func (m *MockScheduleStore) FindByUID(ctx context.Context, UID string) (*core.Schedule, error) {
	var res *core.Schedule
	var err error
	switch UID {
	case "testPodChaos":
		sch := v1alpha1.Schedule{}
		jsonStr, _ := json.Marshal(sch)
		res = &core.Schedule{
			ScheduleMeta: core.ScheduleMeta{
				UID:        UID,
				Kind:       v1alpha1.KindPodChaos,
				Name:       "testName",
				Namespace:  "testNamespace",
				Action:     "testAction",
				StartTime:  time.Time{},
				FinishTime: time.Time{},
				Archived:   true,
			},
			Schedule: string(jsonStr),
		}
	case "testErrRecordNotFound":
		err = gorm.ErrRecordNotFound
	default:
		err = fmt.Errorf("test err")
	}
	return res, err
}

func (m *MockScheduleStore) FindMetaByUID(context.Context, string) (*core.ScheduleMeta, error) {
	panic("implement me")
}

func (m *MockScheduleStore) Set(context.Context, *core.Schedule) error {
	panic("implement me")
}

func (m *MockScheduleStore) Archive(ctx context.Context, namespace, name string) error {
	panic("implement me")
}

func (m *MockScheduleStore) Delete(context.Context, *core.Schedule) error {
	panic("implement me")
}

func (m *MockScheduleStore) DeleteByFinishTime(context.Context, time.Duration) error {
	panic("implement me")
}

func (m *MockScheduleStore) DeleteByUIDs(context.Context, []string) error {
	panic("implement me")
}

func (m *MockScheduleStore) DeleteIncompleteSchedules(context.Context) error {
	panic("implement me")
}

var _ = Describe("event", func() {
	var router *gin.Engine
	BeforeEach(func() {
		pkgmock.With("MockAuthRequired", true)

		mockExpStore := new(MockExperimentStore)
		mockSchStore := new(MockScheduleStore)

		s := Service{
			archive:         mockExpStore,
			archiveSchedule: mockSchStore,
			event:           nil,
			conf: &config.ChaosDashboardConfig{
				ClusterScoped: true,
			},
		}
		router = gin.Default()
		r := router.Group("/api")
		endpoint := r.Group("/archives")

		endpoint.GET("", s.list)
		endpoint.GET("/detail", s.detail)

		endpoint.GET("/schedules", s.listSchedule)
		endpoint.GET("/schedules/:uid", s.detailSchedule)
	})

	AfterEach(func() {
		// Add any setup steps that needs to be executed after each test
		pkgmock.Reset("MockAuthRequired")
	})

	Context("List", func() {
		It("success", func() {
			response := []Archive{
				Archive{
					UID:       "testUID",
					Kind:      "testKind",
					Namespace: "testNamespace",
					Name:      "testName",
					CreatedAt: time.Time{},
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives?kind=testKind", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("test err", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})
	})

	Context("Detail", func() {
		It("empty uid", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusBadRequest))
		})

		It("testPodChaos", func() {
			chaos := &v1alpha1.PodChaos{}
			response := Detail{
				Archive: Archive{
					UID:       "testPodChaos",
					Kind:      v1alpha1.KindPodChaos,
					Namespace: "testNamespace",
					Name:      "testName",
					CreatedAt: time.Time{},
				},
				KubeObject: core.KubeObjectDesc{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
						Kind:       "",
					},
					Meta: core.KubeObjectMeta{
						Name:        "",
						Namespace:   "",
						Labels:      nil,
						Annotations: nil,
					},
					Spec: chaos.Spec,
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail?uid=testPodChaos", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("testIOChaos", func() {
			chaos := &v1alpha1.IOChaos{}
			response := Detail{
				Archive: Archive{
					UID:       "testIOChaos",
					Kind:      v1alpha1.KindIOChaos,
					Namespace: "testNamespace",
					Name:      "testName",
					CreatedAt: time.Time{},
				},
				KubeObject: core.KubeObjectDesc{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
						Kind:       "",
					},
					Meta: core.KubeObjectMeta{
						Name:        "",
						Namespace:   "",
						Labels:      nil,
						Annotations: nil,
					},
					Spec: chaos.Spec,
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail?uid=testIOChaos", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("testNetworkChaos", func() {
			chaos := &v1alpha1.NetworkChaos{}
			response := Detail{
				Archive: Archive{
					UID:       "testNetworkChaos",
					Kind:      v1alpha1.KindNetworkChaos,
					Namespace: "testNamespace",
					Name:      "testName",
					CreatedAt: time.Time{},
				},
				KubeObject: core.KubeObjectDesc{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
						Kind:       "",
					},
					Meta: core.KubeObjectMeta{
						Name:        "",
						Namespace:   "",
						Labels:      nil,
						Annotations: nil,
					},
					Spec: chaos.Spec,
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail?uid=testNetworkChaos", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("testTimeChaos", func() {
			chaos := &v1alpha1.TimeChaos{}
			response := Detail{
				Archive: Archive{
					UID:       "testTimeChaos",
					Kind:      v1alpha1.KindTimeChaos,
					Namespace: "testNamespace",
					Name:      "testName",
					CreatedAt: time.Time{},
				},
				KubeObject: core.KubeObjectDesc{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
						Kind:       "",
					},
					Meta: core.KubeObjectMeta{
						Name:        "",
						Namespace:   "",
						Labels:      nil,
						Annotations: nil,
					},
					Spec: chaos.Spec,
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail?uid=testTimeChaos", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("testKernelChaos", func() {
			chaos := &v1alpha1.KernelChaos{}
			response := Detail{
				Archive: Archive{
					UID:       "testKernelChaos",
					Kind:      v1alpha1.KindKernelChaos,
					Namespace: "testNamespace",
					Name:      "testName",
					CreatedAt: time.Time{},
				},
				KubeObject: core.KubeObjectDesc{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
						Kind:       "",
					},
					Meta: core.KubeObjectMeta{
						Name:        "",
						Namespace:   "",
						Labels:      nil,
						Annotations: nil,
					},
					Spec: chaos.Spec,
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail?uid=testKernelChaos", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("testStressChaos", func() {
			chaos := &v1alpha1.StressChaos{}
			response := Detail{
				Archive: Archive{
					UID:       "testStressChaos",
					Kind:      v1alpha1.KindStressChaos,
					Namespace: "testNamespace",
					Name:      "testName",
					CreatedAt: time.Time{},
				},
				KubeObject: core.KubeObjectDesc{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
						Kind:       "",
					},
					Meta: core.KubeObjectMeta{
						Name:        "",
						Namespace:   "",
						Labels:      nil,
						Annotations: nil,
					},
					Spec: chaos.Spec,
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail?uid=testStressChaos", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("testOtherChaos", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail?uid=testOtherChaos", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})

		It("testErrRecordNotFound", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail?uid=testErrRecordNotFound", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})

		It("test err", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail?uid=testErr", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})
	})

	Context("ListSchedule", func() {
		It("success", func() {
			response := []Archive{
				Archive{
					UID:       "testUID",
					Kind:      "testKind",
					Namespace: "testNamespace",
					Name:      "testScheduleName",
					CreatedAt: time.Time{},
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/schedules?name=testScheduleName", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("test err", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/schedules", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})
	})

	Context("DetailSchedule", func() {
		It("empty uid", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/detail", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusBadRequest))
		})

		It("testPodChaos", func() {
			sch := &v1alpha1.Schedule{}
			response := Detail{
				Archive: Archive{
					UID:       "testPodChaos",
					Kind:      v1alpha1.KindPodChaos,
					Namespace: "testNamespace",
					Name:      "testName",
					CreatedAt: time.Time{},
				},
				KubeObject: core.KubeObjectDesc{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
						Kind:       "",
					},
					Meta: core.KubeObjectMeta{
						Name:        "",
						Namespace:   "",
						Labels:      nil,
						Annotations: nil,
					},
					Spec: sch.Spec,
				},
			}
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/schedules/testPodChaos", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusOK))
			responseBody, err := json.Marshal(response)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(rr.Body.Bytes()).Should(Equal(responseBody))
		})

		It("testErrRecordNotFound", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/schedules/testErrRecordNotFound", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})

		It("test err", func() {
			rr := httptest.NewRecorder()
			request, _ := http.NewRequest(http.MethodGet, "/api/archives/schedules/testErr", nil)
			router.ServeHTTP(rr, request)
			Expect(rr.Code).Should(Equal(http.StatusInternalServerError))
		})
	})
})

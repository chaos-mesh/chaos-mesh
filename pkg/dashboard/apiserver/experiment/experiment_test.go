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

package experiment

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/finalizers"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func TestExperimentAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Experiment API")
}

type experimentClientPool struct {
	clientpool.Clients
	kube client.Client
}

func (p experimentClientPool) Client(string) (client.Client, error) { return p.kube, nil }

type experimentLookup struct {
	core.ExperimentStore
	names map[string]string
}

func (s experimentLookup) FindByUID(_ context.Context, uid string) (*core.Experiment, error) {
	name, ok := s.names[uid]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &core.Experiment{ExperimentMeta: core.ExperimentMeta{Namespace: "test", Name: name, Kind: "PodChaos", UID: uid}}, nil
}

var _ = Describe("Deleting experiments", func() {
	ctx := context.Background()
	var (
		first         *v1alpha1.PodChaos
		second        *v1alpha1.PodChaos
		kube          client.WithWatch
		router        *gin.Engine
		updateFailure error
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		first = &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{
			Namespace: "test", Name: "first", UID: "one",
			Finalizers:  []string{finalizers.RecordFinalizer},
			Annotations: map[string]string{"owner": "team"},
		}}
		second = &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "second", UID: "two"}}
		updateFailure = nil
		kube = fake.NewClientBuilder().WithScheme(scheme).WithObjects(first, second).Build()
		previous := clientpool.K8sClients
		DeferCleanup(func() { clientpool.K8sClients = previous })
		clientpool.K8sClients = experimentClientPool{kube: interceptor.NewClient(kube, interceptor.Funcs{
			Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				if updateFailure != nil {
					return updateFailure
				}
				return c.Update(ctx, obj, opts...)
			},
		})}
		store := experimentLookup{names: map[string]string{"one": first.Name, "two": second.Name}}
		router = gin.New()
		Register(router.Group("/api"), NewService(store, nil, &config.ChaosDashboardConfig{}, scheme, logr.Discard()))
	})

	It("only requests force deletion after the cleanup annotation is persisted", func() {
		path := "/api/experiments/one?force=true"
		key := client.ObjectKeyFromObject(first)
		updateFailure = errors.New("update unavailable")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, path, nil))
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(response.Body.String()).To(ContainSubstring("Forced deletion failed"))
		Expect(kube.Get(ctx, key, first)).To(Succeed())
		Expect(first.DeletionTimestamp).To(BeNil())
		Expect(first.Annotations).To(Equal(map[string]string{"owner": "team"}))

		updateFailure = nil
		response = httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, path, nil))
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(kube.Get(ctx, key, first)).To(Succeed())
		Expect(first.DeletionTimestamp).NotTo(BeNil())
		Expect(first.Annotations).To(Equal(map[string]string{
			"owner":                             "team",
			finalizers.AnnotationCleanFinalizer: finalizers.AnnotationCleanFinalizerForced,
		}))
	})

	It("stops a batch at a missing experiment after accepting the previous deletion", func() {
		key := client.ObjectKeyFromObject(first)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/api/experiments?uids=one,missing,two", nil))
		Expect(response.Code).To(Equal(http.StatusNotFound))
		Expect(kube.Get(ctx, key, first)).To(Succeed())
		Expect(first.DeletionTimestamp).NotTo(BeNil())
		Expect(kube.Get(ctx, client.ObjectKeyFromObject(second), second)).To(Succeed())
		Expect(second.DeletionTimestamp).To(BeNil())
	})

	It("rejects oversized batches before changing any experiment", func() {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/api/experiments?uids="+strings.Repeat("one,", 100)+"two", nil))
		Expect(response.Code).To(Equal(http.StatusInternalServerError))
		Expect(response.Body.String()).To(ContainSubstring("Too many uids"))
		Expect(kube.Get(ctx, client.ObjectKeyFromObject(first), first)).To(Succeed())
		Expect(first.DeletionTimestamp).To(BeNil())
		Expect(kube.Get(ctx, client.ObjectKeyFromObject(second), second)).To(Succeed())
	})
})

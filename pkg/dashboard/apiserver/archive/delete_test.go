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

package archive

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	authv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	authclient "k8s.io/client-go/kubernetes/typed/authorization/v1"
	clienttesting "k8s.io/client-go/testing"

	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type archiveAuthPool struct {
	clientpool.Clients
	auth authclient.AuthorizationV1Interface
}

func (p archiveAuthPool) AuthClient(string) (authclient.AuthorizationV1Interface, error) {
	return p.auth, nil
}

type archiveRemovalStore struct {
	core.ExperimentStore
	calls *[]string
	err   error
}

func (s archiveRemovalStore) FindByUID(_ context.Context, uid string) (*core.Experiment, error) {
	if uid == "missing" {
		return nil, gorm.ErrRecordNotFound
	}
	return &core.Experiment{ExperimentMeta: core.ExperimentMeta{UID: uid}}, nil
}

func (s archiveRemovalStore) Delete(_ context.Context, experiment *core.Experiment) error {
	*s.calls = append(*s.calls, "archive:"+experiment.UID)
	return s.err
}

func (s archiveRemovalStore) DeleteByUIDs(_ context.Context, uids []string) error {
	for _, uid := range uids {
		*s.calls = append(*s.calls, "archive:"+uid)
	}
	return s.err
}

type archiveEventRemovalStore struct {
	core.EventStore
	calls *[]string
	err   error
}

func (s archiveEventRemovalStore) DeleteByUID(_ context.Context, uid string) error {
	*s.calls = append(*s.calls, "events:"+uid)
	return s.err
}

func (s archiveEventRemovalStore) DeleteByUIDList(_ context.Context, uids []string) error {
	for _, uid := range uids {
		*s.calls = append(*s.calls, "events:"+uid)
	}
	return s.err
}

var _ = Describe("Removing archived experiments", func() {
	BeforeEach(func() {
		kube := fake.NewClientset()
		kube.PrependReactor("create", "selfsubjectaccessreviews", func(action clienttesting.Action) (bool, runtime.Object, error) {
			review := action.(clienttesting.CreateAction).GetObject().(*authv1.SelfSubjectAccessReview)
			Expect(review.Spec.ResourceAttributes.Verb).To(Equal("patch"))
			review.Status.Allowed = true
			return true, review, nil
		})
		previous := clientpool.K8sClients
		DeferCleanup(func() { clientpool.K8sClients = previous })
		clientpool.K8sClients = archiveAuthPool{auth: kube.AuthorizationV1()}
	})

	DescribeTable("removes events before archive rows",
		func(path string, status int, eventFailure, archiveFailure error, wantCalls []string) {
			calls := []string{}
			experiments := archiveRemovalStore{calls: &calls, err: archiveFailure}
			events := archiveEventRemovalStore{calls: &calls, err: eventFailure}
			router := gin.New()
			Register(router.Group("/api"), NewService(experiments, nil, events, nil, &config.ChaosDashboardConfig{}))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/api/archives"+path, nil))
			Expect(response.Code).To(Equal(status))
			Expect(calls).To(Equal(wantCalls))
		},
		Entry("single archive", "/first", http.StatusOK, nil, nil, []string{"events:first", "archive:first"}),
		Entry("batch skipping missing rows", "?uids=first,missing,second", http.StatusOK, nil, nil,
			[]string{"events:first", "events:second", "archive:first", "archive:second"}),
		Entry("event removal fails", "/first", http.StatusInternalServerError, errors.New("event removal unavailable"), nil,
			[]string{"events:first"}),
		Entry("archive removal fails", "?uids=first", http.StatusInternalServerError, nil, errors.New("archive removal unavailable"),
			[]string{"events:first", "archive:first"}),
	)
})

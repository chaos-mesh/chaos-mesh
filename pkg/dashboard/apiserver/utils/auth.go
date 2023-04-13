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

package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

func AuthMiddleware(c *gin.Context, config *config.ChaosDashboardConfig) {
	if mockResult := mock.On("AuthMiddleware"); mockResult != nil {
		c.Next()

		return
	}

	kubeCli, err := clientpool.ExtractTokenAndGetAuthClient(c.Request.Header)
	if err != nil {
		SetAPIError(c, ErrBadRequest.WrapWithNoMessage(err))

		return
	}

	ns := c.Query("namespace")

	if ns == "" && !config.ClusterScoped && config.TargetNamespace != "" {
		ns = config.TargetNamespace

		log.L().WithName("auth middleware").V(1).Info("Replace query namespace with", ns)
	}

	verb := "list"
	if c.Request.Method != http.MethodGet {
		// patch is used to indicate create, patch, finalizers and other write operations
		verb = "patch"
	}

	sar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: ns,
				Verb:      verb,
				Group:     "chaos-mesh.org",
				Resource:  "*",
			},
		},
	}

	result, err := kubeCli.SelfSubjectAccessReviews().Create(c.Request.Context(), sar, metav1.CreateOptions{})
	if err != nil {
		SetAPImachineryError(c, ErrInternalServer.WrapWithNoMessage(err))

		return
	}

	if !result.Status.Allowed {
		if len(ns) == 0 {
			SetAPIError(c, ErrNoClusterPrivilege.New("can't %s resource in the cluster", verb))
		} else {
			SetAPIError(c, ErrNoNamespacePrivilege.New("can't %s resource in namespace %s", verb, ns))
		}

		return
	}

	c.Next()
}

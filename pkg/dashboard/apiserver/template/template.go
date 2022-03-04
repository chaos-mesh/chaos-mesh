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
//

package template

import (
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/gin-gonic/gin"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("template api")

// Service defines a handler service for cluster common objects.
type Service struct {
	conf *config.ChaosDashboardConfig
}

func NewService(conf *config.ChaosDashboardConfig) *Service {
	return &Service{conf: conf}
}

func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/templates")

	statusCheckEndpoint := endpoint.Group("/status-check")
	statusCheckEndpoint.GET("", s.listStatusCheck)
	statusCheckEndpoint.POST("", s.createStatusCheck)
	statusCheckEndpoint.GET("/:uid", s.getStatusCheckDetailByUID)
	statusCheckEndpoint.PUT("/:uid", s.updateStatusCheck)
	statusCheckEndpoint.DELETE("/:uid", s.deleteStatusCheck)
}

func (it *Service) listStatusCheck(c *gin.Context) {

}

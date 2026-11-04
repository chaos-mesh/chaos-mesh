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

package swaggerserver

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/swaggerdocs"
)

func Handler(c *gin.Context) {
	if path.Base(c.Param("any")) == "doc.json" {
		spec := *swaggerdocs.SwaggerInfo
		spec.Host = c.Request.Host
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(spec.ReadDoc()))
		return
	}
	files := *swaggerFiles.Handler
	ginSwagger.CustomWrapHandler(
		&ginSwagger.Config{URL: "/api/swagger/doc.json"},
		&files,
	)(c)
}

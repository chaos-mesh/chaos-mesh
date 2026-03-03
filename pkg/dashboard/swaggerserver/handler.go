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
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/swaggerdocs"
)

func Handler(c *gin.Context) {
	swaggerdocs.SwaggerInfo.Host = c.Request.Host

	ginSwagger.CustomWrapHandler(
		&ginSwagger.Config{URL: "/api/swagger/doc.json"},
		swaggerFiles.Handler,
	)(c)
}

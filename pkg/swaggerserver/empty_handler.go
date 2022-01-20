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

//go:build !swagger_server
// +build !swagger_server

package swaggerserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler returns an empty `http.Handler`.
func Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "Swagger UI is not built. Please run `SWAGGER=1 make`.")
	}
}

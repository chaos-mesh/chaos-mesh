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

package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	Log             = ctrl.Log.WithName("apiserver")
	ResponseSuccess = Response{Status: "success"}
)

// Response defines a common status struct.
type Response struct {
	Status string `json:"status"`
}

func ShouldBindBodyWithJSON(c *gin.Context, obj interface{}) {
	if err := c.ShouldBindBodyWith(obj, binding.JSON); err != nil {
		c.Status(http.StatusBadRequest)
		c.Error(ErrBadRequest.WrapWithNoMessage(err))
	}
}

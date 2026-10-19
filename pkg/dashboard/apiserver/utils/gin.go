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
	"bytes"
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// MapStringSliceResponse maps response keys to lists of string values.
type MapStringSliceResponse map[string][]string

// Response defines a common status struct.
type Response struct {
	Status string `json:"status"`
}

var (
	ResponseSuccess = Response{Status: "success"}
)

func ShouldBindBodyWithJSON(c *gin.Context, obj interface{}) (err error) {
	err = c.ShouldBindBodyWith(obj, binding.JSON)
	if err == nil {
		body := c.MustGet(gin.BodyBytesKey).([]byte)
		if !json.Valid(body) {
			err = errors.New("request body must contain one complete JSON document")
		} else if bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
			err = errors.New("request body must not be null")
		}
	}
	if err != nil {
		SetAPIError(c, ErrBadRequest.WrapWithNoMessage(err))
	}

	return
}

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
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joomcode/errorx"
)

var (
	ErrNS             = errorx.NewNamespace("error.api")
	ErrOther          = ErrNS.NewType("other")
	ErrInvalidRequest = ErrNS.NewType("invalid_request")
	ErrInternalServer = ErrNS.NewType("internal_server_error")
	ErrNotFound       = ErrNS.NewType("resource_not_found")
)

type APIError struct {
	Error    bool   `json:"error"`
	Message  string `json:"message"`
	Code     string `json:"code"`
	FullText string `json:"full_text"`
}

// MWHandleErrors creates a middleware that turns (last) error in the context into an APIError json response.
// In handlers, `c.Error(err)` can be used to attach the error to the context.
// When error is attached in the context:
// - The handler can optionally assign the HTTP status code.
// - The handler must not self-generate a response body.
func MWHandleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		err := c.Errors.Last()
		if err == nil {
			return
		}

		statusCode := c.Writer.Status()
		if statusCode == http.StatusOK {
			statusCode = http.StatusInternalServerError
		}

		innerErr := errorx.Cast(err.Err)
		if innerErr == nil {
			innerErr = ErrOther.WrapWithNoMessage(err.Err)
		}

		c.AbortWithStatusJSON(statusCode, APIError{
			Error:    true,
			Message:  innerErr.Error(),
			Code:     errorx.GetTypeName(innerErr),
			FullText: fmt.Sprintf("%+v", innerErr),
		})
	}
}

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
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joomcode/errorx"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	ErrNS                   = errorx.NewNamespace("error.api")
	ErrUnknown              = ErrNS.NewType("unknown")
	ErrInvalidRequest       = ErrNS.NewType("invalid_request")
	ErrInternalServer       = ErrNS.NewType("internal_server_error")
	ErrNotFound             = ErrNS.NewType("resource_not_found")
	ErrNoClusterPrivilege   = ErrNS.NewType("no_cluster_privilege")
	ErrNoNamespacePrivilege = ErrNS.NewType("no_namespace_privilege")
)

type APIError struct {
	Message  string `json:"message"`
	Code     string `json:"code"`
	FullText string `json:"full_text"`
	Status   string `json:"status"`
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
			innerErr = ErrUnknown.WrapWithNoMessage(err.Err)
		}

		c.AbortWithStatusJSON(statusCode, APIError{
			Status:   "error",
			Message:  innerErr.Error(),
			Code:     errorx.GetTypeName(innerErr),
			FullText: fmt.Sprintf("%+v", innerErr),
		})
	}
}

func SetErrorForGinCtx(c *gin.Context, err error) {
	if apierrors.IsForbidden(err) && strings.Contains(err.Error(), "at the cluster scope") {
		_ = c.Error(ErrNoClusterPrivilege.WrapWithNoMessage(err))
		return
	} else if apierrors.IsForbidden(err) && strings.Contains(err.Error(), "in the namespace") {
		_ = c.Error(ErrNoNamespacePrivilege.WrapWithNoMessage(err))
		return
	} else if apierrors.IsNotFound(err) {
		_ = c.Error(ErrNotFound.WrapWithNoMessage(err))
		return
	}

	_ = c.Error(ErrInternalServer.WrapWithNoMessage(err))
}

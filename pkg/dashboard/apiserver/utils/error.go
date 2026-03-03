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
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joomcode/errorx"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/log"
)

var (
	ErrNS             = errorx.NewNamespace("error.api")
	ErrUnknown        = ErrNS.NewType("unknown")               // 500
	ErrBadRequest     = ErrNS.NewType("bad_request")           // 400
	ErrNotFound       = ErrNS.NewType("resource_not_found")    // 404
	ErrInternalServer = ErrNS.NewType("internal_server_error") // 500
	// Custom
	ErrNoClusterPrivilege   = ErrNS.NewType("no_cluster_privilege")   // 401
	ErrNoNamespacePrivilege = ErrNS.NewType("no_namespace_privilege") // 401
)

type APIError struct {
	Code     int    `json:"code"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	FullText string `json:"full_text"`
}

func SetAPIError(c *gin.Context, err *errorx.Error) {
	typeName := errorx.GetTypeName(err)

	var code int
	switch typeName {
	case ErrBadRequest.FullName():
		code = http.StatusBadRequest
	case ErrNoClusterPrivilege.FullName(), ErrNoNamespacePrivilege.FullName():
		code = http.StatusUnauthorized
	case ErrNotFound.FullName():
		code = http.StatusNotFound
	case ErrUnknown.FullName(), ErrInternalServer.FullName():
		code = http.StatusInternalServerError
	default:
		code = http.StatusInternalServerError
	}

	apiError := APIError{
		Code:     code,
		Type:     typeName,
		Message:  err.Error(),
		FullText: fmt.Sprintf("%+v", err),
	}

	log.L().WithName("auth middleware").Error(err.Cause(), typeName)
	c.AbortWithStatusJSON(code, &apiError)
}

func SetAPImachineryError(c *gin.Context, err error) {
	if apierrors.IsForbidden(err) && strings.Contains(err.Error(), "at the cluster scope") {
		SetAPIError(c, ErrNoClusterPrivilege.WrapWithNoMessage(err))

		return
	} else if apierrors.IsForbidden(err) && strings.Contains(err.Error(), "in the namespace") {
		SetAPIError(c, ErrNoNamespacePrivilege.WrapWithNoMessage(err))

		return
	} else if apierrors.IsNotFound(err) {
		SetAPIError(c, ErrNotFound.WrapWithNoMessage(err))

		return
	}

	SetAPIError(c, ErrInternalServer.WrapWithNoMessage(err))
}

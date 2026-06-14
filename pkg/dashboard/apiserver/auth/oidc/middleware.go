// Copyright 2024 Chaos Mesh Authors.
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

package oidc

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
)

func (s *Service) Middleware(c *gin.Context) {
	ctx := c.Request.Context()

	s.logger.V(1).Info("handling oidc middleware")
	// The frontend reuses the GCP cookie path for OIDC sessions and sends
	// X-Authorization-Method: "gcp", so this must match "gcp" to stay consistent
	// with the frontend. Unifying this on "oidc" requires a coordinated
	// frontend change and is deferred to a follow-up auth refactor.
	if c.Request.Header.Get("X-Authorization-Method") != "gcp" {
		c.Next()
		return
	}

	expiry, err := time.Parse(time.RFC3339, c.Request.Header.Get("X-Authorization-Expiry"))
	if err != nil {
		utils.SetAPIError(c, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	oauth, err := s.getOauthConfig(c)
	if err != nil {
		utils.SetAPIError(c, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	token, err := oauth.TokenSource(ctx, &oauth2.Token{
		AccessToken:  c.Request.Header.Get("X-Authorization-AccessToken"),
		RefreshToken: c.Request.Header.Get("X-Authorization-RefreshToken"),
		Expiry:       expiry,
	}).Token()

	if err != nil {
		utils.SetAPIError(c, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	token.SetAuthHeader(c.Request)
	setCookie(c, token)

	c.Next()
}

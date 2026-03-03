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

package gcp

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
)

func (s *Service) Middleware(c *gin.Context) {
	ctx := c.Request.Context()

	s.logger.Info("handling gcp middleware")
	if c.Request.Header.Get("X-Authorization-Method") != "gcp" {
		c.Next()
		return
	}

	expiry, err := time.Parse(time.RFC3339, c.Request.Header.Get("X-Authorization-Expiry"))
	if err != nil {
		utils.SetAPIError(c, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	oauth := s.getOauthConfig(c)
	token, err := oauth.TokenSource(ctx, &oauth2.Token{
		AccessToken:  c.Request.Header.Get("X-Authorization-AccessToken"),
		RefreshToken: c.Request.Header.Get("X-Authorization-RefreshToken"),
		Expiry:       expiry,
	}).Token()

	if err != nil {
		utils.SetAPIError(c, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	s.logger.Info("setting request header")
	token.SetAuthHeader(c.Request)
	s.logger.Info("setting request header", "header", c.Request.Header)
	setCookie(c, token)

	c.Next()
}

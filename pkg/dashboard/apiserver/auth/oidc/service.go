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
	"net/http"
	"net/url"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"golang.org/x/oauth2"

	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
)

type Service struct {
	issuerUrl    string
	clientId     string
	clientSecret string
	rootUrl      *url.URL
	logger       logr.Logger
}

// NewService returns an experiment service instance.
func NewService(
	conf *config.ChaosDashboardConfig,
	logger logr.Logger,
) (*Service, error) {
	rootUrl, err := url.Parse(conf.RootUrl)
	if err != nil {
		return nil, err
	}
	if rootUrl.Path == "" {
		rootUrl.Path = "/"
	}

	return &Service{
		clientId:     conf.OidcClientId,
		clientSecret: conf.OidcClientSecret,
		issuerUrl:    conf.OidcIssuerURL,
		rootUrl:      rootUrl,
		logger:       logger.WithName("oidc auth api"),
	}, nil
}

// Register mounts HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service, conf *config.ChaosDashboardConfig) {
	// If the oidc security mode is not set, just skip the registration
	if !conf.OidcSecurityMode {
		return
	}

	r.Use(s.Middleware)

	endpoint := r.Group("/auth/oidc")
	endpoint.GET("/redirect", s.handleRedirect)
	endpoint.GET("/callback", s.authCallback)
}

func (s *Service) getOauthConfig(c *gin.Context) oauth2.Config {
	url := *s.rootUrl
	url.Path = path.Join(s.rootUrl.Path, "./api/auth/oidc/callback")

	var Endpoint = oauth2.Endpoint{
		AuthURL:  s.issuerUrl + "/oauth2/authorize",
		TokenURL: s.issuerUrl + "/oauth2/token",
	}

	return oauth2.Config{
		ClientID:     s.clientId,
		ClientSecret: s.clientSecret,
		RedirectURL:  url.String(),
		Scopes: []string{
			"openid",
		},
		Endpoint: Endpoint,
	}
}

func (s *Service) handleRedirect(c *gin.Context) {
	oauth := s.getOauthConfig(c)
	uri := oauth.AuthCodeURL("")

	s.logger.Info("Redirecting to: ", "URI", uri) // This will log the URL using your service's logger

	c.Redirect(http.StatusFound, uri)
}

func (s *Service) authCallback(c *gin.Context) {
	ctx := c.Request.Context()

	oauth := s.getOauthConfig(c)
	oauth2Token, err := oauth.Exchange(ctx, c.Request.URL.Query().Get("code"), oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	if err != nil {
		utils.SetAPIError(c, utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	oauth2Token.AccessToken = oauth2Token.Extra("id_token").(string)
	setCookie(c, oauth2Token)
	target := url.URL{
		Path: "/",
	}
	c.Redirect(http.StatusFound, target.RequestURI())
}

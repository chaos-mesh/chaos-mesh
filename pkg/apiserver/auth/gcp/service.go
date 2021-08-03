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

package gcp

import (
	"context"
	"net/http"
	"net/url"

	container "cloud.google.com/go/container/apiv1"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("gcp auth api")

type Service struct {
	clientId     string
	clientSecret string

	project  string
	location string
	cluster  string
}

// NewService returns an experiment service instance.
func NewService(
	conf *config.ChaosDashboardConfig,
) *Service {
	return &Service{
		clientId:     conf.GcpClientId,
		clientSecret: conf.GcpClientSecret,

		project:  conf.GcpProject,
		location: conf.GcpLocation,
		cluster:  conf.GcpCluster,
	}
}

// Register mounts HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/auth/gcp")

	endpoint.GET("/redirect", s.handleRedirect)
	endpoint.GET("/refresh", s.handleRedirect)
	endpoint.GET("/callback", s.authCallback)
}

func (s *Service) getOauthConfig(c *gin.Context) oauth2.Config {
	return oauth2.Config{
		ClientID:     s.clientId,
		ClientSecret: s.clientSecret,
		// TODO: use a better way to construct the url
		// TODO: support https
		RedirectURL: "http://" + c.Request.Host + "/api/auth/gcp/callback",
		Scopes: []string{
			"email", "profile",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/compute",
			"https://www.googleapis.com/auth/cloud-platform",
		},
		Endpoint: google.Endpoint,
	}
}

func (s *Service) handleRedirect(c *gin.Context) {
	oauth := s.getOauthConfig(c)
	uri := oauth.AuthCodeURL("")

	c.Redirect(http.StatusFound, uri)
}

func (s *Service) authCallback(c *gin.Context) {
	ctx := c.Request.Context()

	oauth := s.getOauthConfig(c)
	oauth2Token, err := oauth.Exchange(context.TODO(), c.Request.URL.Query().Get("code"))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	cc, err := container.NewClusterManagerClient(ctx, option.WithTokenSource(oauth.TokenSource(ctx, oauth2Token)))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}
	req := &containerpb.GetClusterRequest{
		Name: "projects/" + s.project + "/locations/" + s.location + "/clusters/" + s.cluster,
	}
	resp, err := cc.GetCluster(ctx, req)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	// TODO: handle different kinds of token
	c.SetCookie("access_token", oauth2Token.AccessToken, 0, "", "", false, false)
	c.SetCookie("expiry", oauth2Token.Expiry.String(), 0, "", "", false, false)
	c.SetCookie("ca", resp.MasterAuth.ClusterCaCertificate, 0, "", "", false, false)
	target := url.URL{
		Path: "/",
	}
	c.Redirect(http.StatusFound, target.RequestURI())
}

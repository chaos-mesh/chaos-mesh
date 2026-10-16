// Copyright 2026 Chaos Mesh Authors.
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

package oidc

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func TestOIDCProviderFailureRecovery(t *testing.T) {
	const validToken = `{"access_token":"oauth-access","id_token":"recovered-identity","refresh_token":"refresh","expires_in":3600,"token_type":"Bearer"}`
	for _, endpoint := range []string{"callback", "refresh"} {
		for _, tc := range []struct {
			name, path, body string
			status           int
			transportError   bool
			stop             string
			missingRefresh   bool
		}{
			{name: "discovery transport", path: "/.well-known/openid-configuration", transportError: true},
			{name: "discovery unavailable", path: "/.well-known/openid-configuration", body: `{"error":"unavailable"}`, status: 503},
			{name: "issuer mismatch", path: "/.well-known/openid-configuration", body: `{"issuer":"https://other.test","token_endpoint":"https://issuer.test/token"}`, status: 200},
			{name: "malformed discovery", path: "/.well-known/openid-configuration", body: `{`, status: 200},
			{name: "token transport", path: "/token", transportError: true},
			{name: "invalid grant", path: "/token", body: `{"error":"invalid_grant"}`, status: 400},
			{name: "malformed token", path: "/token", body: `{`, status: 200},
			{name: "numeric identity", path: "/token", body: `{"access_token":"access","id_token":42}`, status: 200},
			{name: "null identity", path: "/token", body: `{"access_token":"access","id_token":null}`, status: 200},
			{name: "missing access token", path: "/token", body: `{"id_token":"identity"}`, status: 200},
			{name: "numeric access token", path: "/token", body: `{"access_token":42,"id_token":"identity"}`, status: 200},
			{name: "canceled discovery", path: "/.well-known/openid-configuration", stop: "cancel"},
			{name: "deadline discovery", path: "/.well-known/openid-configuration", stop: "deadline"},
			{name: "canceled token", path: "/token", stop: "cancel"},
			{name: "deadline token", path: "/token", stop: "deadline"},
			{name: "missing refresh token", missingRefresh: true},
		} {
			if endpoint == "callback" && tc.missingRefresh {
				continue
			}
			t.Run(endpoint+"/"+tc.name, func(t *testing.T) {
				service := newTestService(t)
				router := gin.New()
				router.Use(service.Middleware)
				router.GET("/callback", service.authCallback)
				downstream := 0
				router.GET("/", func(c *gin.Context) {
					downstream++
					if got := c.GetHeader("Authorization"); got != "Bearer recovered-identity" {
						t.Errorf("Authorization %q", got)
					}
					c.Status(http.StatusNoContent)
				})
				target := "/"
				successStatus := http.StatusNoContent
				if endpoint == "callback" {
					target = "/callback?code=test-code"
					successStatus = http.StatusFound
				}
				faultActive := true
				faultCalls, tokenCalls := 0, 0
				request := func() (*httptest.ResponseRecorder, *http.Request) {
					req := oidcRequest(t, target, validToken, http.StatusOK)
					delegate := req.Context().Value(oauth2.HTTPClient).(*http.Client).Transport
					ctx := req.Context()
					cancel := func() {}
					var wantErr error
					if faultActive && tc.stop == "cancel" {
						ctx, cancel = context.WithCancel(ctx)
						wantErr = context.Canceled
					}
					if faultActive && tc.stop == "deadline" {
						ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Second))
						wantErr = context.DeadlineExceeded
					}
					defer cancel()
					transport := providerTransport(func(r *http.Request) (*http.Response, error) {
						if r.URL.Path == "/token" {
							tokenCalls++
						}
						if faultActive && r.URL.Path == tc.path {
							faultCalls++
							if tc.stop != "" {
								cancel()
								if !errors.Is(r.Context().Err(), wantErr) {
									t.Errorf("provider context error %v, want %v", r.Context().Err(), wantErr)
								}
								return nil, r.Context().Err()
							}
							if tc.transportError {
								return nil, errors.New("provider connection reset")
							}
							return &http.Response{StatusCode: tc.status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(tc.body)), Request: r}, nil
						}
						return delegate.RoundTrip(r)
					})
					req = req.WithContext(context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: transport}))
					if endpoint == "refresh" {
						req.Header.Set("X-Authorization-Method", "gcp")
						req.Header.Set("X-Authorization-AccessToken", "expired-identity")
						req.Header.Set("X-Authorization-Expiry", time.Now().Add(-time.Hour).Format(time.RFC3339))
						if !faultActive || !tc.missingRefresh {
							req.Header.Set("X-Authorization-RefreshToken", "refresh")
						}
					}
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
					return w, req
				}
				failed, req := request()
				if failed.Code != http.StatusInternalServerError || len(failed.Result().Cookies()) != 0 || failed.Header().Get("Location") != "" || req.Header.Get("Authorization") != "" || downstream != 0 {
					t.Fatalf("failure status %d, cookies %v, location %q, authorization %q, downstream %d, body %s", failed.Code, responseCookies(failed), failed.Header().Get("Location"), req.Header.Get("Authorization"), downstream, failed.Body)
				}
				if tc.missingRefresh && tokenCalls != 0 {
					t.Fatalf("missing refresh token sent %d token requests", tokenCalls)
				}
				if !tc.missingRefresh && faultCalls == 0 {
					t.Fatal("provider fault was not exercised")
				}
				faultActive = false
				recovered, _ := request()
				if recovered.Code != successStatus || responseCookies(recovered)["access_token"] != "recovered-identity" || responseCookies(recovered)["refresh_token"] != "refresh" {
					t.Fatalf("recovery status %d, cookies %v, body %s", recovered.Code, responseCookies(recovered), recovered.Body)
				}
				if endpoint == "refresh" && downstream != 1 {
					t.Fatalf("recovery downstream %d", downstream)
				}
			})
		}
	}
}

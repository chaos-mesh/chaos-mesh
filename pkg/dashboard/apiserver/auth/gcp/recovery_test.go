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

package gcp

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

func TestGCPProviderFailureRecovery(t *testing.T) {
	const validToken = `{"access_token":"recovered-access","refresh_token":"refresh","expires_in":3600,"token_type":"Bearer"}`
	for _, endpoint := range []string{"callback", "refresh"} {
		for _, tc := range []struct {
			name, body, stop               string
			status                         int
			transportError, missingRefresh bool
		}{
			{name: "token transport", transportError: true},
			{name: "invalid grant", body: `{"error":"invalid_grant"}`, status: 400},
			{name: "malformed token", body: `{`, status: 200},
			{name: "missing access token", body: `{"expires_in":3600}`, status: 200},
			{name: "numeric access token", body: `{"access_token":42}`, status: 200},
			{name: "canceled token", stop: "cancel"},
			{name: "deadline token", stop: "deadline"},
			{name: "missing refresh token", missingRefresh: true},
		} {
			if endpoint == "callback" && tc.missingRefresh {
				continue
			}
			t.Run(endpoint+"/"+tc.name, func(t *testing.T) {
				service := testService(t)
				router := gin.New()
				router.Use(service.Middleware)
				router.GET("/callback", service.authCallback)
				downstream := 0
				router.GET("/", func(c *gin.Context) {
					downstream++
					if got := c.GetHeader("Authorization"); got != "Bearer recovered-access" {
						t.Errorf("Authorization %q", got)
					}
					c.Status(http.StatusNoContent)
				})
				target, grant := "/", "refresh_token"
				successStatus := http.StatusNoContent
				if endpoint == "callback" {
					target, grant = "/callback?code=code", "authorization_code"
					successStatus = http.StatusFound
				}
				faultActive := true
				faultCalls := 0
				request := func() (*httptest.ResponseRecorder, *http.Request) {
					req := gcpRequest(t, target, grant, validToken, http.StatusOK)
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
					transport := tokenTransport(func(r *http.Request) (*http.Response, error) {
						if faultActive {
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
						req.Header.Set("X-Authorization-AccessToken", "expired-access")
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
					t.Fatalf("failure status %d, cookies %v, location %q, authorization %q, downstream %d, body %s", failed.Code, cookies(failed), failed.Header().Get("Location"), req.Header.Get("Authorization"), downstream, failed.Body)
				}
				if tc.missingRefresh && faultCalls != 0 {
					t.Fatalf("missing refresh token sent %d token requests", faultCalls)
				}
				if !tc.missingRefresh && faultCalls == 0 {
					t.Fatal("provider fault was not exercised")
				}
				faultActive = false
				recovered, _ := request()
				if recovered.Code != successStatus || cookies(recovered)["access_token"] != "recovered-access" || cookies(recovered)["refresh_token"] != "refresh" {
					t.Fatalf("recovery status %d, cookies %v, body %s", recovered.Code, cookies(recovered), recovered.Body)
				}
				if endpoint == "refresh" && downstream != 1 {
					t.Fatalf("recovery downstream %d", downstream)
				}
			})
		}
	}
}

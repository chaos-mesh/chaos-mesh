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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"golang.org/x/oauth2"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
)

type tokenTransport func(*http.Request) (*http.Response, error)

func (f tokenTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func gcpRequest(t *testing.T, target, grant, response string, status int) *http.Request {
	t.Helper()
	transport := tokenTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://oauth2.googleapis.com/token" {
			return nil, fmt.Errorf("unexpected provider URL %s", r.URL)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("grant_type") != grant {
			t.Errorf("grant %q, want %q", r.Form.Get("grant_type"), grant)
		}
		if grant == "authorization_code" && (r.Form.Get("code") != "code" || r.Form.Get("redirect_uri") != "https://dashboard.test/mesh/api/auth/gcp/callback") {
			t.Errorf("token form %v", r.Form)
		}
		if grant == "refresh_token" && r.Form.Get("refresh_token") != "refresh" {
			t.Errorf("token form %v", r.Form)
		}
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(response)), Request: r}, nil
	})
	r := httptest.NewRequest("GET", target, nil)
	return r.WithContext(context.WithValue(r.Context(), oauth2.HTTPClient, &http.Client{Transport: transport}))
}

func testService(t *testing.T) *Service {
	t.Helper()
	s, err := NewService(&config.ChaosDashboardConfig{RootURL: "https://dashboard.test/mesh", GcpClientID: "client", GcpClientSecret: "secret"}, logr.Discard())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func cookies(w *httptest.ResponseRecorder) map[string]string {
	values := make(map[string]string)
	for _, c := range w.Result().Cookies() {
		values[c.Name], _ = url.QueryUnescape(c.Value)
	}
	return values
}

func TestGCPRedirect(t *testing.T) {
	r := gin.New()
	Register(r.Group("/api"), testService(t), &config.ChaosDashboardConfig{GcpSecurityMode: true})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/auth/gcp/redirect", nil))
	u, err := url.Parse(w.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if w.Code != 302 || u.Host != "accounts.google.com" || u.Query().Get("client_id") != "client" || u.Query().Get("access_type") != "offline" || u.Query().Get("redirect_uri") != "https://dashboard.test/mesh/api/auth/gcp/callback" {
		t.Fatalf("status %d redirect %s", w.Code, u)
	}
}

func TestGCPCallback(t *testing.T) {
	for _, tc := range []struct {
		name, body             string
		providerStatus, status int
	}{
		{"success", `{"access_token":"access","refresh_token":"refresh","expires_in":3600,"token_type":"Bearer"}`, 200, 302},
		{"provider error", `{"error":"invalid_grant"}`, 400, 500},
		{"malformed token", `{`, 200, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			Register(r.Group("/api"), testService(t), &config.ChaosDashboardConfig{GcpSecurityMode: true})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, gcpRequest(t, "/api/auth/gcp/callback?code=code", "authorization_code", tc.body, tc.providerStatus))
			got := cookies(w)
			if w.Code != tc.status {
				t.Fatalf("status %d body %s", w.Code, w.Body)
			}
			if tc.status == 302 {
				if got["access_token"] != "access" || got["refresh_token"] != "refresh" || w.Header().Get("Location") != "/" {
					t.Fatalf("cookies %v location %q", got, w.Header().Get("Location"))
				}
				if expiry, err := time.Parse(time.RFC3339, got["expiry"]); err != nil || !expiry.After(time.Now()) {
					t.Fatalf("expiry %q error %v", got["expiry"], err)
				}
			} else if len(got) != 0 {
				t.Fatalf("failed authentication set cookies %v", got)
			}
		})
	}
}

func TestGCPMiddleware(t *testing.T) {
	for _, tc := range []struct {
		name, method, expiry, body, token string
		providerStatus, status            int
	}{
		{"other authentication", "", "", "", "", 200, 204},
		{"invalid expiry", "gcp", "invalid", "", "", 200, 500},
		{"valid token", "gcp", time.Now().Add(time.Hour).Format(time.RFC3339), "", "Bearer existing", 200, 204},
		{"refresh", "gcp", time.Now().Add(-time.Hour).Format(time.RFC3339), `{"access_token":"new-access","expires_in":3600,"token_type":"Bearer"}`, "Bearer new-access", 200, 204},
		{"refresh error", "gcp", time.Now().Add(-time.Hour).Format(time.RFC3339), `{"error":"invalid_grant"}`, "", 400, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(testService(t).Middleware)
			continued := false
			r.GET("/", func(c *gin.Context) {
				continued = true
				if got := c.GetHeader("Authorization"); got != tc.token {
					t.Errorf("token %q, want %q", got, tc.token)
				}
				c.Status(204)
			})
			req := gcpRequest(t, "/", "refresh_token", tc.body, tc.providerStatus)
			req.Header.Set("X-Authorization-Method", tc.method)
			req.Header.Set("X-Authorization-Expiry", tc.expiry)
			req.Header.Set("X-Authorization-AccessToken", "existing")
			req.Header.Set("X-Authorization-RefreshToken", "refresh")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status || continued != (tc.status == 204) {
				t.Fatalf("status %d continued %v body %s", w.Code, continued, w.Body)
			}
			if tc.token != "" && (cookies(w)["access_token"] != strings.TrimPrefix(tc.token, "Bearer ") || cookies(w)["refresh_token"] != "refresh") {
				t.Fatalf("cookies %v", cookies(w))
			}
			if tc.status == 500 && len(cookies(w)) != 0 {
				t.Fatalf("failed authentication set cookies %v", cookies(w))
			}
		})
	}
}

func TestGCPConfiguration(t *testing.T) {
	if _, err := NewService(&config.ChaosDashboardConfig{RootURL: "://invalid"}, logr.Discard()); err == nil {
		t.Fatal("accepted invalid root URL")
	}
	r := gin.New()
	Register(r.Group("/api"), testService(t), &config.ChaosDashboardConfig{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/auth/gcp/redirect", nil))
	if w.Code != 404 {
		t.Fatalf("disabled GCP route status %d", w.Code)
	}
}

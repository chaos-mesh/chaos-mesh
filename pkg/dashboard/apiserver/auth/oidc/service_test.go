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
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"golang.org/x/oauth2"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
)

type providerTransport func(*http.Request) (*http.Response, error)

func (f providerTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func oidcRequest(t *testing.T, target, tokenResponse string, tokenStatus int) *http.Request {
	t.Helper()
	transport := providerTransport(func(r *http.Request) (*http.Response, error) {
		body, status := "", http.StatusOK
		switch r.URL.String() {
		case "https://issuer.test/.well-known/openid-configuration":
			body = `{"issuer":"https://issuer.test","authorization_endpoint":"https://issuer.test/authorize","token_endpoint":"https://issuer.test/token","jwks_uri":"https://issuer.test/keys"}`
		case "https://issuer.test/token":
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			switch r.Form.Get("grant_type") {
			case "authorization_code":
				if r.Form.Get("code") != "test-code" || r.Form.Get("redirect_uri") != "https://dashboard.test/mesh/api/auth/oidc/callback" {
					t.Errorf("token form %v", r.Form)
				}
			case "refresh_token":
				if r.Form.Get("refresh_token") != "refresh" {
					t.Errorf("refresh form %v", r.Form)
				}
			default:
				t.Errorf("unexpected token grant %v", r.Form)
			}
			body, status = tokenResponse, tokenStatus
		default:
			return nil, fmt.Errorf("unexpected provider URL %s", r.URL)
		}
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	request := httptest.NewRequest(http.MethodGet, target, nil)
	return request.WithContext(context.WithValue(request.Context(), oauth2.HTTPClient, &http.Client{Transport: transport}))
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	s, err := NewService(&config.ChaosDashboardConfig{RootURL: "https://dashboard.test/mesh", OidcIssuerURL: "https://issuer.test", OidcClientID: "client", OidcClientSecret: "secret"}, logr.Discard())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func responseCookies(w *httptest.ResponseRecorder) map[string]string {
	values := make(map[string]string)
	for _, c := range w.Result().Cookies() {
		values[c.Name], _ = url.QueryUnescape(c.Value)
	}
	return values
}

func TestOIDCRedirect(t *testing.T) {
	s := newTestService(t)
	r := gin.New()
	Register(r.Group("/api"), s, &config.ChaosDashboardConfig{OidcSecurityMode: true})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, oidcRequest(t, "/api/auth/oidc/redirect", "", 200))
	u, err := url.Parse(w.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if w.Code != 302 || u.Host != "issuer.test" || u.Path != "/authorize" || u.Query().Get("client_id") != "client" || u.Query().Get("scope") != "openid" || u.Query().Get("redirect_uri") != "https://dashboard.test/mesh/api/auth/oidc/callback" {
		t.Fatalf("status %d redirect %s", w.Code, u)
	}
}

func TestOIDCCallback(t *testing.T) {
	for _, tc := range []struct {
		name, token            string
		providerStatus, status int
	}{
		{"identity token", `{"access_token":"oauth-access","id_token":"identity","refresh_token":"refresh","expires_in":3600,"token_type":"Bearer"}`, 200, 302},
		{"missing identity token", `{"access_token":"oauth-access","expires_in":3600}`, 200, 500},
		{"empty identity token", `{"access_token":"oauth-access","id_token":"","expires_in":3600}`, 200, 500},
		{"provider error", `{"error":"invalid_grant"}`, 400, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestService(t)
			r := gin.New()
			Register(r.Group("/api"), s, &config.ChaosDashboardConfig{OidcSecurityMode: true})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, oidcRequest(t, "/api/auth/oidc/callback?code=test-code", tc.token, tc.providerStatus))
			cookies := responseCookies(w)
			if w.Code != tc.status {
				t.Fatalf("status %d, body %s", w.Code, w.Body)
			}
			if tc.status == 302 {
				if cookies["access_token"] != "identity" || cookies["refresh_token"] != "refresh" || w.Header().Get("Location") != "/" {
					t.Fatalf("cookies %v, location %q", cookies, w.Header().Get("Location"))
				}
				if expiry, err := time.Parse(time.RFC3339, cookies["expiry"]); err != nil || !expiry.After(time.Now()) {
					t.Fatalf("expiry %q error %v", cookies["expiry"], err)
				}
			} else if len(cookies) != 0 {
				t.Fatalf("failed authentication set cookies %v", cookies)
			}
		})
	}
}

func TestOIDCMiddleware(t *testing.T) {
	for _, tc := range []struct {
		name, method, expiry, response, wantToken, wantRefresh string
		status                                                 int
	}{
		{"other authentication", "", "", "", "", "", 204},
		{"invalid expiry", "gcp", "invalid", "", "", "", 500},
		{"valid identity", "gcp", time.Now().Add(time.Hour).Format(time.RFC3339), "", "Bearer existing-identity", "refresh", 204},
		{"refresh identity", "gcp", time.Now().Add(-time.Hour).Format(time.RFC3339), `{"access_token":"oauth-access","id_token":"new-identity","refresh_token":"new-refresh","expires_in":3600,"token_type":"Bearer"}`, "Bearer new-identity", "new-refresh", 204},
		{"refresh without rotation", "gcp", time.Now().Add(-time.Hour).Format(time.RFC3339), `{"access_token":"oauth-access","id_token":"new-identity","expires_in":3600,"token_type":"Bearer"}`, "Bearer new-identity", "refresh", 204},
		{"refresh missing identity", "gcp", time.Now().Add(-time.Hour).Format(time.RFC3339), `{"access_token":"oauth-access","expires_in":3600,"token_type":"Bearer"}`, "", "", 500},
		{"refresh empty identity", "gcp", time.Now().Add(-time.Hour).Format(time.RFC3339), `{"access_token":"oauth-access","id_token":"","expires_in":3600,"token_type":"Bearer"}`, "", "", 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestService(t)
			r := gin.New()
			r.Use(s.Middleware)
			continued := false
			r.GET("/", func(c *gin.Context) {
				continued = true
				if got := c.GetHeader("Authorization"); got != tc.wantToken {
					t.Errorf("Authorization %q, want %q", got, tc.wantToken)
				}
				c.Status(204)
			})
			req := oidcRequest(t, "/", tc.response, 200)
			req.Header.Set("X-Authorization-Method", tc.method)
			req.Header.Set("X-Authorization-Expiry", tc.expiry)
			req.Header.Set("X-Authorization-AccessToken", "existing-identity")
			req.Header.Set("X-Authorization-RefreshToken", "refresh")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.status || continued != (tc.status == 204) {
				t.Fatalf("status %d, continued %v, body %s", w.Code, continued, w.Body)
			}
			if tc.wantToken != "" && responseCookies(w)["access_token"] != strings.TrimPrefix(tc.wantToken, "Bearer ") {
				t.Fatalf("cookies %v", responseCookies(w))
			}
			if tc.wantRefresh != "" {
				got := responseCookies(w)
				if got["refresh_token"] != tc.wantRefresh {
					t.Errorf("refresh cookie %q, want %q", got["refresh_token"], tc.wantRefresh)
				}
				expiry, err := time.Parse(time.RFC3339, got["expiry"])
				if err != nil || !expiry.After(time.Now().Add(30*time.Minute)) || expiry.After(time.Now().Add(2*time.Hour)) {
					t.Errorf("expiry cookie %q, error %v", got["expiry"], err)
				}
			}
			if tc.status == 500 && len(responseCookies(w)) != 0 {
				t.Fatalf("failed authentication set cookies %v", responseCookies(w))
			}
		})
	}
}

func TestOIDCConfigurationErrors(t *testing.T) {
	for _, conf := range []*config.ChaosDashboardConfig{
		{RootURL: "://invalid"}, {RootURL: "https://dashboard.test", OidcCaFile: "/nonexistent/dashboard-test-ca"},
	} {
		if _, err := NewService(conf, logr.Discard()); err == nil {
			t.Fatalf("accepted invalid configuration %+v", conf)
		}
	}
	r := gin.New()
	Register(r.Group("/api"), newTestService(t), &config.ChaosDashboardConfig{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/auth/oidc/redirect", nil))
	if w.Code != 404 {
		t.Fatalf("disabled OIDC route status %d", w.Code)
	}
}

func TestOIDCCustomCARefresh(t *testing.T) {
	var issuer string
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			fmt.Fprintf(w, `{"issuer":%q,"authorization_endpoint":%q,"token_endpoint":%q,"jwks_uri":%q}`, issuer, issuer+"/authorize", issuer+"/token", issuer+"/keys")
		case "/token":
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			switch r.Form.Get("grant_type") {
			case "refresh_token":
				if r.Form.Get("refresh_token") != "refresh" {
					t.Errorf("refresh form %v", r.Form)
				}
			case "authorization_code":
				if r.Form.Get("code") != "test-code" {
					t.Errorf("callback form %v", r.Form)
				}
			default:
				t.Errorf("unexpected token form %v", r.Form)
			}
			fmt.Fprint(w, `{"access_token":"oauth-access","id_token":"trusted-identity","expires_in":3600,"token_type":"Bearer"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(provider.Close)
	issuer = provider.URL
	caFile := filepath.Join(t.TempDir(), "issuer-ca.pem")
	if err := os.WriteFile(caFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: provider.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := NewService(&config.ChaosDashboardConfig{RootURL: "https://dashboard.test", OidcIssuerURL: issuer, OidcCaFile: caFile}, logr.Discard())
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(s.Middleware)
	router.GET("/callback", s.authCallback)
	router.GET("/", func(c *gin.Context) {
		if got := c.GetHeader("Authorization"); got != "Bearer trusted-identity" {
			t.Errorf("Authorization %q", got)
		}
		c.Status(204)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Authorization-Method", "gcp")
	req.Header.Set("X-Authorization-AccessToken", "expired-identity")
	req.Header.Set("X-Authorization-RefreshToken", "refresh")
	req.Header.Set("X-Authorization-Expiry", time.Now().Add(-time.Hour).Format(time.RFC3339))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 204 || responseCookies(w)["access_token"] != "trusted-identity" {
		t.Fatalf("status %d, body %s, cookies %v", w.Code, w.Body, responseCookies(w))
	}
	untrusted, err := NewService(&config.ChaosDashboardConfig{RootURL: "https://dashboard.test", OidcIssuerURL: issuer}, logr.Discard())
	if err != nil {
		t.Fatal(err)
	}
	untrustedRouter := gin.New()
	untrustedRouter.GET("/callback", untrusted.authCallback)
	untrustedResponse := httptest.NewRecorder()
	untrustedRouter.ServeHTTP(untrustedResponse, httptest.NewRequest(http.MethodGet, "/callback?code=test-code", nil))
	if untrustedResponse.Code != http.StatusInternalServerError || len(untrustedResponse.Result().Cookies()) != 0 {
		t.Fatalf("service without the custom CA returned status %d, cookies %v", untrustedResponse.Code, responseCookies(untrustedResponse))
	}
	for _, stopped := range []string{"canceled", "deadline"} {
		t.Run(stopped, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			if stopped == "deadline" {
				cancel()
				ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			}
			cancel()
			request := httptest.NewRequest(http.MethodGet, "/callback?code=test-code", nil).WithContext(ctx)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusInternalServerError || len(response.Result().Cookies()) != 0 || response.Header().Get("Location") != "" {
				t.Fatalf("stopped request status %d, cookies %v, location %q", response.Code, responseCookies(response), response.Header().Get("Location"))
			}
		})
	}
	callback := httptest.NewRecorder()
	router.ServeHTTP(callback, httptest.NewRequest(http.MethodGet, "/callback?code=test-code", nil))
	if callback.Code != http.StatusFound || responseCookies(callback)["access_token"] != "trusted-identity" {
		t.Fatalf("callback status %d, body %s, cookies %v", callback.Code, callback.Body, responseCookies(callback))
	}
}

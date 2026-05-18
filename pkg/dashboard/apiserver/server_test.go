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
//

package apiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeBasePath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"/", ""},
		{"chaos-mesh", "/chaos-mesh"},
		{"/chaos-mesh", "/chaos-mesh"},
		{"/chaos-mesh/", "/chaos-mesh"},
		{"chaos-mesh/", "/chaos-mesh"},
		{"/a/b/c/", "/a/b/c"},
	}
	for _, tc := range cases {
		if got := normalizeBasePath(tc.in); got != tc.want {
			t.Errorf("normalizeBasePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestStripBasePath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name    string
		prefix  string
		reqPath string
		want    string
	}{
		{"exact prefix becomes root", "/chaos-mesh", "/chaos-mesh", "/"},
		{"prefix with trailing slash becomes root", "/chaos-mesh", "/chaos-mesh/", "/"},
		{"asset under prefix is stripped", "/chaos-mesh", "/chaos-mesh/assets/index.js", "/assets/index.js"},
		{"api under prefix is stripped", "/chaos-mesh", "/chaos-mesh/api/experiments", "/api/experiments"},
		{"unrelated path is untouched", "/chaos-mesh", "/assets/index.js", "/assets/index.js"},
		{"sibling sharing prefix substring is untouched", "/chaos-mesh", "/chaos-mesh-other/foo", "/chaos-mesh-other/foo"},
		{"root is untouched", "/chaos-mesh", "/", "/"},
		{"nested prefix is stripped", "/a/b", "/a/b/c/d", "/c/d"},
		{"nested prefix sibling is untouched", "/a/b", "/a/bc", "/a/bc"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var seen string
			r := gin.New()
			r.Use(stripBasePath(tc.prefix))
			r.NoRoute(func(c *gin.Context) {
				seen = c.Request.URL.Path
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, tc.reqPath, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if seen != tc.want {
				t.Errorf("path after middleware = %q, want %q", seen, tc.want)
			}
		})
	}
}

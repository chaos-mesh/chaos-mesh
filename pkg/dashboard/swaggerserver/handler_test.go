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

package swaggerserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandlerUsesEachRequestHost(t *testing.T) {
	router := gin.New()
	router.GET("/api/swagger/*any", Handler)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := range 16 {
		workers.Go(func() {
			<-start
			host := fmt.Sprintf("dashboard-%d.example:2333", i)
			docPath := "/api/swagger/doc.json"
			if i%2 == 0 {
				docPath = "/api/swagger/nested/doc.json"
			}
			request := httptest.NewRequest(http.MethodGet, "http://"+host+docPath, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Errorf("status = %d, body = %s", response.Code, response.Body)
				return
			}
			var doc struct {
				Host    string `json:"host"`
				Swagger string `json:"swagger"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &doc); err != nil {
				t.Error(err)
				return
			}
			if doc.Host != host || doc.Swagger != "2.0" {
				t.Errorf("document host = %q, want %q; version = %q", doc.Host, host, doc.Swagger)
			}
		})
	}
	close(start)
	workers.Wait()
}

func TestHandlerServesSwaggerUI(t *testing.T) {
	router := gin.New()
	router.GET("/api/swagger/*any", Handler)
	for _, tc := range []struct {
		path, contentType, contains string
		status                      int
	}{
		{"index.html", "text/html", "doc.json", http.StatusOK},
		{"swagger-ui.css", "text/css", ".swagger-ui", http.StatusOK},
		{"missing", "", "", http.StatusNotFound},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/swagger/"+tc.path, nil))
			if response.Code != tc.status || !strings.Contains(response.Header().Get("Content-Type"), tc.contentType) || !strings.Contains(response.Body.String(), tc.contains) {
				t.Fatalf("status = %d, content type = %q, body = %.200s", response.Code, response.Header().Get("Content-Type"), response.Body.String())
			}
		})
	}
}

func TestHandlerAssetsDoNotFixTheNextRequestPrefix(t *testing.T) {
	router := gin.New()
	router.GET("/api/swagger/*any", Handler)
	for _, asset := range []string{"nested/swagger-ui.css", "swagger-ui.css", "index.html", "other/swagger-ui.css", "swagger-ui.css"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/swagger/"+asset, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("asset %s: status %d, body %.200s", asset, response.Code, response.Body.String())
		}
	}
}

func TestHandlerConcurrentAssetPrefixes(t *testing.T) {
	router := gin.New()
	router.GET("/api/swagger/*any", Handler)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := range 16 {
		workers.Go(func() {
			<-start
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/swagger/prefix-%d/swagger-ui.css", i), nil)
			router.ServeHTTP(response, request)
			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), ".swagger-ui") {
				t.Errorf("asset %s: status %d", request.URL.Path, response.Code)
			}
		})
	}
	close(start)
	workers.Wait()
}

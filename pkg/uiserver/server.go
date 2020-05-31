// Copyright 2020 PingCAP, Inc.
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

package uiserver

import (
	"io"
	"net/http"

	"github.com/shurcooL/httpgzip"
)

// Handler returns a FileServer `http.Handler` to handle http request.
func Handler(root http.FileSystem) http.Handler {
	if root != nil {
		return httpgzip.FileServer(root, httpgzip.FileServerOptions{IndexHTML: true, ServeError: fallback(root)})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "Dashboard UI is not built. Use `UI=1 make`.\n")
	})
}

// AssetFS returns assets.
func AssetFS() http.FileSystem {
	return assets
}

func fallback(fs http.FileSystem) func(w http.ResponseWriter, r *http.Request, _ error) {
	return func(w http.ResponseWriter, r *http.Request, _ error) {
		localRedirect(w, r, "index.html")
	}
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like http.Redirect does.
func localRedirect(w http.ResponseWriter, req *http.Request, newPath string) {
	if req.URL.RawQuery != "" {
		newPath += "?" + req.URL.RawQuery
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

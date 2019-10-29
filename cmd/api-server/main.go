// Copyright 2019 PingCAP, Inc.
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

package main

import (
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ngaut/log"
	"github.com/pingcap/chaos-operator/pkg/apiserver"
)

func main() {
	server, err := apiserver.NewServer()
	if err != nil {
		log.Errorf("Error while creating server: %s", err)
		return
	}

	http.ListenAndServe("0.0.0.0:80", server.CreateRouter())
}

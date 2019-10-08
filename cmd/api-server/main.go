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
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ngaut/log"
	"github.com/pingcap/chaos-operator/pkg/apiserver"
)

func main() {
	databaseHost := os.Getenv("CHAOS_API_SERVER_DATABASE_SERVICE_HOST")
	databasePort := os.Getenv("CHAOS_API_SERVER_DATABASE_SERVICE_PORT")

	server, err := apiserver.NewServer(fmt.Sprintf("root:@(%s:%s)/chaos_operator", databaseHost, databasePort))
	if err != nil {
		log.Errorf("Error while creating server: %s", err)
		return
	}

	http.ListenAndServe("0.0.0.0:80", server.CreateRouter())
}

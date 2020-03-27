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

package utils

import (
	"fmt"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	InitLog            = ctrl.Log.WithName("setup")
	DashboardNamespace string
	DataSource         string
)

func init() {
	var ok bool

	DashboardNamespace, ok = os.LookupEnv("NAMESPACE")
	if !ok {
		InitLog.Error(nil, "cannot find NAMESPACE")
		DashboardNamespace = "chaos"
	}

	DataSource = fmt.Sprintf("root:@tcp(chaos-collector-database.%s:3306)/chaos_operator", DashboardNamespace)
}

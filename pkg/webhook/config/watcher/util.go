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

package watcher

<<<<<<< HEAD:pkg/apiserver/status_code/status_code.go
// StatusCode represents the code of api status
type StatusCode int

const (
	// Success indicates the successful return of this API interface.
	Success StatusCode = 0
	// GetResourcesWrong indicates an error when getting resources
	GetResourcesWrong StatusCode = 1001
	// GetResourcesFromDBWrong indicates an error when getting resources from DB
	GetResourcesFromDBWrong StatusCode = 1006
	// IncompleteField indicates that some fields are missing.
	IncompleteField StatusCode = 1007
=======
import (
	"bytes"
	"html/template"
>>>>>>> upstream/master:pkg/webhook/config/watcher/util.go
)

func renderTemplateWithArgs(tpl *template.Template, args map[string]string) ([]byte, error) {
	model := make(map[string]interface{}, len(args))
	for k, v := range args {
		model[k] = v
	}
	buff := new(bytes.Buffer)
	if err := tpl.Execute(buff, model); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

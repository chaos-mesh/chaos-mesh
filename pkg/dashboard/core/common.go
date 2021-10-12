// Copyright 2021 Chaos Mesh Authors.
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

package core

import (
	"encoding/json"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ObjectBase struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	UID       string `json:"uid"`
	Created   string `json:"created_at"`
}

// KubeObjectDesc defines a simple kube object description which uses in apiserver.
type KubeObjectDesc struct {
	metav1.TypeMeta
	Meta KubeObjectMeta `json:"metadata"`
	Spec interface{}    `json:"spec"`
}

// KubeObjectMetadata extracts the required fields from metav1.ObjectMeta.
type KubeObjectMeta struct {
	Namespace   string            `json:"namespace"`
	Name        string            `json:"name"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type Filter struct {
	ObjectID  string `json:"object_id"`
	Start     string `json:"start"`
	End       string `json:"end"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Limit     string `json:"limit"`
}

func (f *Filter) toMap() map[string]interface{} {
	var fMap map[string]interface{}

	marshal, _ := json.Marshal(f)
	_ = json.Unmarshal(marshal, &fMap)

	return fMap
}

const zeroTime = "0001-01-01 00:00:00"

func (f *Filter) ConstructQueryArgs() (string, []interface{}) {
	fMap, query, args := f.toMap(), make([]string, 0), make([]interface{}, 0)

	for k, v := range fMap {
		if v != "" {
			if k == "start" || k == "end" || k == "limit" {
				continue
			}

			if len(args) > 0 {
				query = append(query, "AND", k, "= ?")
			} else {
				query = append(query, k, "= ?")
			}

			args = append(args, v)
		}
	}

	startEnd := ""
	if f.Start != zeroTime && f.End != zeroTime {
		startEnd = "created_at BETWEEN ? AND ?"
		args = append(args, f.Start, f.End)
	} else if f.Start != zeroTime && f.End == zeroTime {
		startEnd = "created_at >= ?"
		args = append(args, f.Start)
	} else if f.Start == zeroTime && f.End != zeroTime {
		startEnd = "created_at <= ?"
		args = append(args, f.End)
	}

	if startEnd != "" {
		if len(query) > 0 {
			query = append(query, "AND", startEnd)
		} else {
			query = append(query, startEnd)
		}
	}

	return strings.Join(query, " "), args
}

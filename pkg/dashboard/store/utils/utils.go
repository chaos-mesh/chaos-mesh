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

package utils

func ConstructQueryArgs(kind, ns, name, uid string) (string, []interface{}) {
	query := ""
	args := make([]interface{}, 0)

	if kind != "" {
		query += "kind = ?"
		args = append(args, kind)
	}

	if ns != "" {
		if len(args) > 0 {
			query += " AND namespace = ?"
		} else {
			query += "namespace = ?"
		}
		args = append(args, ns)
	}

	if name != "" {
		if len(args) > 0 {
			query += " AND name = ?"
		} else {
			query += "name = ?"
		}
		args = append(args, name)
	}

	if uid != "" {
		if len(args) > 0 {
			query += " AND uid = ?"
		} else {
			query += "uid = ?"
		}
		args = append(args, uid)
	}

	return query, args
}

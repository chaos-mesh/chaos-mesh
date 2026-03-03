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

package label

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Label is the label field in metadata
type Label map[string]string

// String converts label to a string
func (l Label) String() string {
	var arr []string

	for k, v := range l {
		if len(k) == 0 {
			continue
		}

		arr = append(arr, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(arr, ",")
}

func ParseLabel(data string) (Label, error) {
	if len(data) == 0 {
		return Label{}, nil
	}

	labels := make(map[string]string)
	for _, tok := range strings.Split(data, ",") {
		kv := strings.Split(tok, "=")
		if len(kv) != 2 {
			return nil, errors.Errorf("invalid labels: %s", data)
		}
		labels[kv[0]] = kv[1]
	}
	return labels, nil
}

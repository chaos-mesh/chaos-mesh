// Copyright 2023 Chaos Mesh Authors.
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

package k8schaos

import (
	"context"
	"io"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type K8sChaosTarget struct {
	unstructured.Unstructured
}

func (target *K8sChaosTarget) Id() string {
	namespace := target.GetNamespace()
	if namespace == "" {
		namespace = "-"
	}
	return strings.Join([]string{target.GetKind(), namespace, target.GetName()}, "/")
}

type SelectImpl struct{}

func (impl *SelectImpl) Select(ctx context.Context, k8sChaosSelector *v1alpha1.K8SChaosAPIObjects) ([]*K8sChaosTarget, error) {
	decoder := yaml.NewDecoder(strings.NewReader(k8sChaosSelector.Value))

	targets := []*K8sChaosTarget{}

	for {
		target := K8sChaosTarget{}

		err := decoder.Decode(&target.Object)
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		targets = append(targets, &target)
	}

	return targets, nil
}

func New() *SelectImpl {
	return &SelectImpl{}
}

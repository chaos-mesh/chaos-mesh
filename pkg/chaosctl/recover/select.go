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

package recover

import (
	"context"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrl/server/model"
)

func SelectPods(ctx context.Context, client *ctrlclient.CtrlClient, selector v1alpha1.PodSelectorSpec) ([]*PartialPod, error) {
	selectInput := model.PodSelectorInput{
		Pods:                map[string]interface{}{},
		NodeSelectors:       map[string]interface{}{},
		Nodes:               selector.Nodes,
		PodPhaseSelectors:   selector.PodPhaseSelectors,
		Namespaces:          selector.Namespaces,
		FieldSelectors:      map[string]interface{}{},
		LabelSelectors:      map[string]interface{}{},
		AnnotationSelectors: map[string]interface{}{},
	}

	for k, v := range selector.Pods {
		selectInput.Pods[k] = v
	}

	for k, v := range selector.NodeSelectors {
		selectInput.NodeSelectors[k] = v
	}

	for k, v := range selector.FieldSelectors {
		selectInput.FieldSelectors[k] = v
	}

	for k, v := range selector.LabelSelectors {
		selectInput.LabelSelectors[k] = v
	}

	for k, v := range selector.AnnotationSelectors {
		selectInput.AnnotationSelectors[k] = v
	}

	podsQuery := new(struct {
		Pods []*PartialPod `graphql:"pods(selector: $selector)"`
	})

	err := client.QueryClient.Query(ctx, podsQuery, map[string]interface{}{"selector": selectInput})
	if err != nil {
		return nil, errors.Wrapf(err, "select pods with selector: %+v", selector)
	}

	return podsQuery.Pods, nil
}

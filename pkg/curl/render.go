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

package curl

import (
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

const image = "curlimages/curl:7.78.0"

const nameSuffix = "-rendered-http-request"

func renderCommands(request CommandFlags) (Commands, error) {
	// TODO: validation of request
	result := []string{"curl", "-i", "-s"}

	// follow the request
	if request.FollowLocation {
		result = append(result, "-L")
	}

	if request.Method != http.MethodGet {
		result = append(result, "-X", request.Method)

		if len(request.Body) > 0 {
			result = append(result, "-d", request.Body)
		}
	}

	renderedHeaders := http.Header{}
	for k, v := range request.Header {
		renderedHeaders[k] = v
	}
	if request.JsonContent {
		if request.Header == nil {
			request.Header = Header{}
		}
		renderedHeaders[HeaderContentType] = []string{ApplicationJson}
	}

	for key, values := range renderedHeaders {
		for _, value := range values {
			result = append(result, "-H", fmt.Sprintf("%s: %s", key, value))
		}
	}

	result = append(result, request.URL)

	return result, nil
}

func RenderWorkflowTaskTemplate(request RequestForm) (*v1alpha1.Template, error) {
	commands, err := renderCommands(request.CommandFlags)
	if err != nil {
		return nil, err
	}
	containerName := fmt.Sprintf("%s%s", request.Name, nameSuffix)
	return &v1alpha1.Template{
		Name: request.Name,
		Type: v1alpha1.TypeTask,
		Task: &v1alpha1.Task{
			Container: &corev1.Container{
				Name:    containerName,
				Image:   image,
				Command: commands,
			},
		},
		ConditionalBranches: nil,
	}, nil
}

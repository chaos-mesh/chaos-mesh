// Copyright 2020 Chaos Mesh Authors.
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

package core

// TODO: using YAML in name might make confusion, because it is actually transferred by json.
// TODO: how about "raw"?

// KubeObjectYAMLDescription defines the YAML structure of an object stored in kubernetes API.
type KubeObjectYAMLDescription struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   KubeObjectYAMLMetadata `json:"metadata"`
	Spec       interface{}            `json:"spec"`
}

// KubeObjectYAMLMetadata defines the metadata of KubeObjectYAMLDescription.
type KubeObjectYAMLMetadata struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

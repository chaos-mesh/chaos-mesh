// Copyright 2019 Chaos Mesh Authors.
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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true

// HttpFaultChaos is the Schema for the HttpFaultchaos API
type HttpFaultChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func (h HttpFaultChaos) DeepCopyObject() runtime.Object {
	panic("implement me")
}

// +kubebuilder:object:root=true

// HttpFaultChaosList contains a list of HttpFaultChaos
type HttpFaultChaosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HttpFaultChaos `json:"items"`
}

func (h HttpFaultChaosList) DeepCopyObject() runtime.Object {
	panic("implement me")
}

func init() {
	SchemeBuilder.Register(&HttpFaultChaos{}, &HttpFaultChaosList{})
}

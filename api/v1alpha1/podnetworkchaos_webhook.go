// Copyright 2021 Chaos Mesh Authors.
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
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var _ webhook.Defaulter = &PodNetworkChaos{}

func (in *PodNetworkChaos) Default() {
}

var _ webhook.Validator = &PodNetworkChaos{}

func (in *PodNetworkChaos) ValidateCreate() error {
	return nil
}

func (in *PodNetworkChaos) ValidateUpdate(old runtime.Object) error {
	return nil
}

func (in *PodNetworkChaos) ValidateDelete() error {
	return nil
}

// Copyright 2025 Chaos Mesh Authors.
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

package v1alpha1

import "testing"

// TestKernelChaosGetSelectorSpecsIsContainerScoped guards against a regression of
// the "too few parts of namespacedname" bug (#4059). KernelChaos.Apply resolves
// each record with controller.ParseNamespacedNameContainer, which expects a
// 3-part namespace/pod/container id and uses the container id to fetch the
// target PID. GetSelectorSpecs must therefore expose the ContainerSelector
// (container-scoped, like IOChaos and StressChaos) rather than the embedded
// pod-scoped PodSelector, which yields 2-part ids that fail to parse and leave
// the experiment stuck NotInjected.
func TestKernelChaosGetSelectorSpecsIsContainerScoped(t *testing.T) {
	kc := &KernelChaos{}
	specs := kc.GetSelectorSpecs()

	sel, ok := specs["."]
	if !ok {
		t.Fatalf(`GetSelectorSpecs is missing the "." selector entry`)
	}
	if _, ok := sel.(*ContainerSelector); !ok {
		t.Fatalf("KernelChaos selector spec = %T; want *ContainerSelector so records carry a container id for ParseNamespacedNameContainer", sel)
	}
}

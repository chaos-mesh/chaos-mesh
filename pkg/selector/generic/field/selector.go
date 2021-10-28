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

package field

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

const Name = "field"

type fieldSelector struct {
	fields.Selector
}

var _ generic.Selector = &fieldSelector{}

func (s *fieldSelector) ListOption() client.ListOption {
	if !s.Empty() {
		return client.MatchingFieldsSelector{Selector: s}
	}
	return nil
}

func (s *fieldSelector) ListFunc(r client.Reader) generic.ListFunc {
	// Since FieldSelectors need to implement index creation, Reader.List is used to get the pod list.
	// Otherwise, just call Client.List directly, which can be obtained through cache.
	if !s.Empty() && r != nil {
		return r.List
	}
	return nil
}

func (s *fieldSelector) Match(obj client.Object) bool {
	var objFields fields.Set
	switch obj := obj.(type) {
	case *v1.Pod:
		objFields = toPodSelectableFields(obj)
	case *v1alpha1.PhysicalMachine:
		objFields = toPhysicalMachineSelectableFields(obj)
	default:
		// not support
		return false
	}
	return s.Matches(objFields)
}

func New(spec v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
	return &fieldSelector{
		Selector: fields.SelectorFromSet(spec.FieldSelectors),
	}, nil
}

// toPodSelectableFields returns a field set that represents the object
// https://github.com/kubernetes/kubernetes/blob/v1.22.2/pkg/registry/core/pod/strategy.go#L306
func toPodSelectableFields(pod *v1.Pod) fields.Set {
	// The purpose of allocation with a given number of elements is to reduce
	// amount of allocations needed to create the fields.Set. If you add any
	// field here or the number of object-meta related fields changes, this should
	// be adjusted.
	podSpecificFieldsSet := make(fields.Set, 9)
	podSpecificFieldsSet["spec.nodeName"] = pod.Spec.NodeName
	podSpecificFieldsSet["spec.restartPolicy"] = string(pod.Spec.RestartPolicy)
	podSpecificFieldsSet["spec.schedulerName"] = pod.Spec.SchedulerName
	podSpecificFieldsSet["spec.serviceAccountName"] = pod.Spec.ServiceAccountName
	podSpecificFieldsSet["status.phase"] = string(pod.Status.Phase)
	podIP := ""
	if len(pod.Status.PodIPs) > 0 {
		podIP = string(pod.Status.PodIPs[0].IP)
	}
	podSpecificFieldsSet["status.podIP"] = podIP
	podSpecificFieldsSet["status.nominatedNodeName"] = pod.Status.NominatedNodeName
	return addObjectMetaFieldsSet(podSpecificFieldsSet, &pod.ObjectMeta, true)
}

// toPhysicalMachineSelectableFields returns a field set that represents the object
func toPhysicalMachineSelectableFields(physicalMachine *v1alpha1.PhysicalMachine) fields.Set {
	pmSpecificFieldsSet := make(fields.Set, 3)
	pmSpecificFieldsSet["spec.address"] = physicalMachine.Spec.Address
	return addObjectMetaFieldsSet(pmSpecificFieldsSet, &physicalMachine.ObjectMeta, true)
}

// addObjectMetaFieldsSet add fields that represent the ObjectMeta to source.
func addObjectMetaFieldsSet(source fields.Set, objectMeta *metav1.ObjectMeta, hasNamespaceField bool) fields.Set {
	source["metadata.name"] = objectMeta.Name
	if hasNamespaceField {
		source["metadata.namespace"] = objectMeta.Namespace
	}
	return source
}

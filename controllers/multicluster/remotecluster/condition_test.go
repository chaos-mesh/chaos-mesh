// Copyright 2022 Chaos Mesh Authors.
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

package remotecluster

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestSetRemoteClusterCondition(t *testing.T) {
	RegisterTestingT(t)

	obj := &v1alpha1.RemoteCluster{}
	setRemoteClusterCondition(obj, v1alpha1.RemoteClusterConditionInstalled, corev1.ConditionTrue, "test")

	Expect(len(obj.Status.Conditions)).To(Equal(2))
	haveConditionInstalled := false
	for _, condition := range obj.Status.Conditions {
		if condition.Type == v1alpha1.RemoteClusterConditionInstalled {
			Expect(condition.Status).To(Equal(corev1.ConditionTrue))
			Expect(condition.Reason).To(Equal("test"))

			haveConditionInstalled = true
		} else {
			Expect(condition.Status).To(Equal(corev1.ConditionFalse))
		}
	}

	Expect(haveConditionInstalled).To(Equal(true))
}

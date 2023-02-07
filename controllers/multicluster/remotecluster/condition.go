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
	corev1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func setRemoteClusterCondition(obj *v1alpha1.RemoteCluster, typ v1alpha1.RemoteClusterConditionType, status corev1.ConditionStatus, reason string) {
	conditionMap := map[v1alpha1.RemoteClusterConditionType]v1alpha1.RemoteClusterCondition{
		v1alpha1.RemoteClusterConditionInstalled: {Type: v1alpha1.RemoteClusterConditionInstalled, Status: corev1.ConditionFalse},
		v1alpha1.RemoteClusterConditionReady:     {Type: v1alpha1.RemoteClusterConditionReady, Status: corev1.ConditionFalse},
	}

	for _, condition := range obj.Status.Conditions {
		conditionMap[condition.Type] = condition
	}

	conditionMap[typ] = v1alpha1.RemoteClusterCondition{Type: typ, Status: status, Reason: reason}

	conditions := make([]v1alpha1.RemoteClusterCondition, 0, len(conditionMap))
	for _, condition := range conditionMap {
		conditions = append(conditions, condition)
	}

	obj.Status.Conditions = conditions
}

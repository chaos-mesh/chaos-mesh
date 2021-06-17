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

package task

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/expr"
)

type Evaluator struct {
	logger     logr.Logger
	kubeclient client.Client
}

func NewEvaluator(logger logr.Logger, kubeclient client.Client) *Evaluator {
	return &Evaluator{logger: logger, kubeclient: kubeclient}
}

func (it *Evaluator) EvaluateConditionBranches(tasks []v1alpha1.ConditionalBranch, resultEnv map[string]interface{}) (branches []v1alpha1.ConditionalBranchStatus, err error) {

	var result []v1alpha1.ConditionalBranchStatus
	for _, task := range tasks {
		it.logger.V(4).Info("evaluate for expression", "expression", task.Expression, "env", resultEnv)
		var evalResult corev1.ConditionStatus
		eval, err := expr.EvalBool(task.Expression, resultEnv)

		if err != nil {
			it.logger.Error(err, "failed to evaluate expression", "expression", task.Expression, "env", resultEnv)
			evalResult = corev1.ConditionUnknown
		} else {
			if eval {
				evalResult = corev1.ConditionTrue
			} else {
				evalResult = corev1.ConditionFalse
			}
		}

		result = append(result, v1alpha1.ConditionalBranchStatus{
			Target:           task.Target,
			EvaluationResult: evalResult,
		})
	}
	return result, nil
}

// Copyright 2020 PingCAP, Inc.
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
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	invalidConfigurationMsg = "invalid configuration"
)

var log = ctrl.Log.WithName("validate-webhook")

// +kubebuilder:object:generate=false
// Admission describe the interface should be implemented in admission webhook
type Admission interface {
	// Validate describe the interface should be implemented in validation webhook
	Validate() (bool, string, error)
}

// Validate describe the podchaos validation logic
func (chaos *PodChaos) Validate() (bool, string, error) {
	switch chaos.Spec.Action {
	case PodFailureAction:
		if chaos.Spec.Duration != nil && chaos.Spec.Scheduler != nil {
			return true, "", nil
		} else if chaos.Spec.Duration == nil && chaos.Spec.Scheduler == nil {
			return true, "", nil
		} else {
			return false, invalidConfigurationMsg, nil
		}
	case PodKillAction:
		// We choose to ignore the Duration property even user define it
		if chaos.Spec.Scheduler == nil {
			return false, invalidConfigurationMsg, nil
		}
		return true, "", nil
	case ContainerKillAction:
		// We choose to ignore the Duration property even user define it
		if chaos.Spec.Scheduler == nil {
			return false, invalidConfigurationMsg, nil
		}
		return true, "", nil
	default:
		err := fmt.Errorf("podchaos[%s/%s] have unknown action type", chaos.Namespace, chaos.Name)
		log.Error(err, "Wrong PodChaos Action type")
		return false, err.Error(), err
	}
}

// Validate describe the iochaos validation logic
func (chaos *IoChaos) Validate() (bool, string, error) {
	if chaos.Spec.Duration != nil && chaos.Spec.Scheduler != nil {
		return true, "", nil
	} else if chaos.Spec.Duration == nil && chaos.Spec.Scheduler == nil {
		return true, "", nil
	}
	return false, invalidConfigurationMsg, nil
}

// Validate describe the network validation logic
func (chaos *NetworkChaos) Validate() (bool, string, error) {
	if chaos.Spec.Duration != nil && chaos.Spec.Scheduler != nil {
		return true, "", nil
	} else if chaos.Spec.Duration == nil && chaos.Spec.Scheduler == nil {
		return true, "", nil
	}
	return false, invalidConfigurationMsg, nil
}

// Validate describe the timechaos validation logic
func (chaos *TimeChaos) Validate() (bool, string, error) {
	if chaos.Spec.Duration != nil && chaos.Spec.Scheduler != nil {
		return true, "", nil
	} else if chaos.Spec.Duration == nil && chaos.Spec.Scheduler == nil {
		return true, "", nil
	}
	return false, invalidConfigurationMsg, nil
}

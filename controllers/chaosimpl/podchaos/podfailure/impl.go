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

package podfailure

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/pkg/annotation"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

const stateAnnotation = "chaos-mesh.org/pod-failure"

type failureState struct {
	Owners         []types.UID       `json:"owners"`
	Containers     map[string]string `json:"containers"`
	InitContainers map[string]string `json:"initContainers,omitempty"`
}

type Impl struct {
	client.Client
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	chaos := obj.(*v1alpha1.PodChaos)
	name, err := controller.ParseNamespacedName(records[index].Id)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	err = impl.updatePod(ctx, name, func(pod *v1.Pod) (bool, error) {
		state, err := readState(pod)
		if err != nil {
			return false, err
		}
		if state == nil {
			ownLegacy, anyLegacy := legacyInjection(pod, chaos)
			if ownLegacy {
				// Preserve retries of an injection made before shared ownership was introduced.
				return false, nil
			}
			if anyLegacy {
				return false, errors.New("waiting for an existing legacy pod-failure experiment to recover")
			}
			state = &failureState{
				Containers:     make(map[string]string),
				InitContainers: make(map[string]string),
			}
			for _, container := range pod.Spec.Containers {
				state.Containers[container.Name] = container.Image
			}
			for _, container := range pod.Spec.InitContainers {
				state.InitContainers[container.Name] = container.Image
			}
		}
		if slices.Contains(state.Owners, chaos.UID) {
			return false, nil
		}
		for i := range pod.Spec.Containers {
			pod.Spec.Containers[i].Image = config.ControllerCfg.PodFailurePauseImage
		}
		for i := range pod.Spec.InitContainers {
			pod.Spec.InitContainers[i].Image = config.ControllerCfg.PodFailurePauseImage
		}
		state.Owners = append(state.Owners, chaos.UID)
		return true, writeState(pod, state)
	})
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	chaos := obj.(*v1alpha1.PodChaos)
	name, err := controller.ParseNamespacedName(records[index].Id)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	err = impl.updatePod(ctx, name, func(pod *v1.Pod) (bool, error) {
		state, err := readState(pod)
		if err != nil {
			return false, err
		}
		if state == nil {
			return recoverLegacyImages(pod, chaos), nil
		}
		owner := slices.Index(state.Owners, chaos.UID)
		if owner < 0 {
			return false, nil
		}
		state.Owners = slices.Delete(state.Owners, owner, owner+1)
		if len(state.Owners) != 0 {
			return true, writeState(pod, state)
		}
		for i, container := range pod.Spec.Containers {
			pod.Spec.Containers[i].Image = state.Containers[container.Name]
		}
		for i, container := range pod.Spec.InitContainers {
			pod.Spec.InitContainers[i].Image = state.InitContainers[container.Name]
		}
		delete(pod.Annotations, stateAnnotation)
		return true, nil
	})
	if err != nil && !k8sError.IsNotFound(err) {
		return v1alpha1.Injected, err
	}
	return v1alpha1.NotInjected, nil
}

func (impl *Impl) updatePod(ctx context.Context, name types.NamespacedName, mutate func(*v1.Pod) (bool, error)) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var original v1.Pod
		if err := impl.Get(ctx, name, &original); err != nil {
			return err
		}
		pod := original.DeepCopy()
		changed, err := mutate(pod)
		if err != nil || !changed {
			return err
		}
		return impl.Patch(ctx, pod, client.MergeFromWithOptions(&original, client.MergeFromWithOptimisticLock{}))
	})
}

func readState(pod *v1.Pod) (*failureState, error) {
	raw, ok := pod.Annotations[stateAnnotation]
	if !ok {
		return nil, nil
	}
	var state failureState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return nil, errors.Wrap(err, "read pod-failure state")
	}
	if len(state.Owners) == 0 {
		return nil, errors.New("pod-failure state has no owners")
	}
	owners := make(map[types.UID]struct{}, len(state.Owners))
	for _, owner := range state.Owners {
		if _, duplicate := owners[owner]; owner == "" || duplicate {
			return nil, errors.New("pod-failure state has an empty or duplicate owner")
		}
		owners[owner] = struct{}{}
	}
	for _, container := range pod.Spec.Containers {
		if image, ok := state.Containers[container.Name]; !ok || image == "" {
			return nil, errors.Errorf("pod-failure state has no original image for container %s", container.Name)
		}
	}
	for _, container := range pod.Spec.InitContainers {
		if image, ok := state.InitContainers[container.Name]; !ok || image == "" {
			return nil, errors.Errorf("pod-failure state has no original image for init container %s", container.Name)
		}
	}
	return &state, nil
}

func writeState(pod *v1.Pod, state *failureState) error {
	raw, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "write pod-failure state")
	}
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	pod.Annotations[stateAnnotation] = string(raw)
	return nil
}

// Legacy annotations lack experiment UIDs, so they cannot safely share the new state.
func legacyInjection(pod *v1.Pod, chaos *v1alpha1.PodChaos) (own, found bool) {
	for index, containers := range [][]v1.Container{pod.Spec.Containers, pod.Spec.InitContainers} {
		for _, container := range containers {
			ownKey := annotation.GenKeyForImage(chaos, container.Name, index == 1)
			suffix := container.Name + "-normal"
			if index == 1 {
				suffix = container.Name + "-init"
			}
			for key := range pod.Annotations {
				if key == suffix || strings.HasPrefix(key, annotation.AnnotationPrefix+"-") && strings.HasSuffix(key, "-pod-failure-"+suffix+"-image") {
					found = true
					own = own || key == ownKey
				}
			}
		}
	}
	return
}

func recoverLegacyImages(pod *v1.Pod, chaos *v1alpha1.PodChaos) bool {
	changed := false
	for i, container := range pod.Spec.Containers {
		key := annotation.GenKeyForImage(chaos, container.Name, false)
		if image, ok := pod.Annotations[key]; ok {
			pod.Spec.Containers[i].Image = image
			delete(pod.Annotations, key)
			changed = true
		}
	}
	for i, container := range pod.Spec.InitContainers {
		key := annotation.GenKeyForImage(chaos, container.Name, true)
		if image, ok := pod.Annotations[key]; ok {
			pod.Spec.InitContainers[i].Image = image
			delete(pod.Annotations, key)
			changed = true
		}
	}
	return changed
}

func NewImpl(c client.Client) *Impl {
	return &Impl{
		Client: c,
	}
}

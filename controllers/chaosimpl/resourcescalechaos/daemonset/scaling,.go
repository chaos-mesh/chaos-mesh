// Copyright 2023 Chaos Mesh Authors.
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

package daemonset

import (
	"context"
	"fmt"

	v1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

func New(daemonSet appsv1.DaemonSetInterface) *DaemonSetScaling {
	return &DaemonSetScaling{
		daemonSet: daemonSet,
	}
}

type DaemonSetScaling struct {
	daemonSet appsv1.DaemonSetInterface
}

func (d *DaemonSetScaling) GetScale(ctx context.Context, resourceName string, options metav1.GetOptions) (*autoscalingv1.Scale, error) {
	ds, err := d.daemonSet.Get(ctx, resourceName, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get scale for daemonset/%s: %w", resourceName, err)
	}

	return d.getAutoscalingV1Scale(ctx, ds), nil
}

func (d *DaemonSetScaling) UpdateScale(ctx context.Context, daemonName string, scale *autoscalingv1.Scale, opts metav1.UpdateOptions) (*autoscalingv1.Scale, error) {
	// Scale down
	if scale.Spec.Replicas == 0 {
		return d.scaleDown(ctx, daemonName)
	} else {
		return d.scaleUp(ctx, daemonName)
	}
}

func (d *DaemonSetScaling) scaleDown(ctx context.Context, daemonName string) (*autoscalingv1.Scale, error) {
	data := []byte(`{"spec": {"nodeSelector": {"non-existing": "true"}}}`)
	ds, err := d.daemonSet.Patch(ctx, daemonName, k8stypes.MergePatchType, data, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed patching node selector spec for daemonset/%s: %w", daemonName, err)
	}

	return d.getAutoscalingV1Scale(ctx, ds), nil
}

func (d *DaemonSetScaling) scaleUp(ctx context.Context, daemonName string) (*autoscalingv1.Scale, error) {
	data := []byte(`[{"op": "remove", "path": "/spec/template/spec/nodeSelector/non-existing"}]`)
	ds, err := d.daemonSet.Patch(ctx, daemonName, k8stypes.JSONPatchType, data, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed patching node selector spec for daemonset/%s: %w", daemonName, err)
	}

	return d.getAutoscalingV1Scale(ctx, ds), nil
}

func (d *DaemonSetScaling) getAutoscalingV1Scale(ctx context.Context, ds *v1.DaemonSet) *autoscalingv1.Scale {
	scale := new(autoscalingv1.Scale)
	scale.ObjectMeta = *ds.ObjectMeta.DeepCopy()
	scale.TypeMeta = ds.TypeMeta

	// Desired replicas
	var desiredReplicas int32 = 0
	if ds.Status.DesiredNumberScheduled > 0 {
		desiredReplicas = 1
	}
	scale.Spec = autoscalingv1.ScaleSpec{
		Replicas: desiredReplicas,
	}

	// Current replicas
	var readyReplicas int32 = 0
	if ds.Status.NumberReady > 0 {
		readyReplicas = 1
	}
	scale.Status = autoscalingv1.ScaleStatus{
		Replicas: readyReplicas,
	}

	return scale
}

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

package nodevolumepath

import (
	"context"

	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

type SelectImpl struct {
	c client.Client
	r client.Reader

	generic.Option
}

type NodeVolumePath struct {
	nodeName string
	// volumePath is a volumePath to the block device or directory on a node
	volumePath string
}

func (n *NodeVolumePath) Id() string {
	// The path may contain "/", but it doesn't matter
	return n.nodeName + "/" + n.volumePath
}

func (impl *SelectImpl) Select(ctx context.Context, selector *v1alpha1.NodeVolumePathSelector) ([]*NodeVolumePath, error) {
	pods, err := pod.SelectAndFilterPods(ctx, impl.c, impl.r, &selector.PodSelector, impl.ClusterScoped, impl.TargetNamespace, impl.EnableFilterNamespace)
	if err != nil {
		return nil, err
	}

	var result []*NodeVolumePath
	for _, pod := range pods {
		for _, volume := range pod.Spec.Volumes {
			if volume.Name == selector.VolumeName {
				// Find the path of the volume
				// If the volume is a `HostPath`, the path is the `Path` field
				// If the volume is a `PersistentVolumeClaim`, we can get the path from the related `PersistentVolume`
				if volume.HostPath != nil {
					result = append(result, &NodeVolumePath{
						nodeName:   pod.Spec.NodeName,
						volumePath: volume.HostPath.Path,
					})
				} else if volume.PersistentVolumeClaim != nil {
					var pvc v1.PersistentVolumeClaim
					impl.c.Get(ctx, types.NamespacedName{
						Namespace: pod.Namespace,
						Name:      volume.PersistentVolumeClaim.ClaimName,
					}, &pvc)
					if pvc.Status.Phase == v1.ClaimBound {
						var pv v1.PersistentVolume
						impl.c.Get(ctx, types.NamespacedName{
							Name: pvc.Spec.VolumeName,
						}, &pv)

						// Only `HostPath` and `LocalVolume` are supported
						// TODO: Possibly support more PersistentVolume source.
						if pv.Spec.HostPath != nil {
							result = append(result, &NodeVolumePath{
								nodeName:   pod.Spec.NodeName,
								volumePath: pv.Spec.HostPath.Path,
							})
						} else if pv.Spec.Local != nil {
							result = append(result, &NodeVolumePath{
								nodeName:   pod.Spec.NodeName,
								volumePath: pv.Spec.Local.Path,
							})
						} else {
							// TODO: handle the case when the PV source is not supported
							continue
						}
					} else {
						// TODO: handle the case that pvc is not bound
						continue
					}
				} else {
					// TODO: handle the case (at least reutrn an error) that the volume source is not supported
					continue
				}
			}
		}
	}

	return result, nil
}

type Params struct {
	fx.In

	Client client.Client
	Reader client.Reader `name:"no-cache"`
}

func New(params Params) *SelectImpl {
	return &SelectImpl{
		params.Client,
		params.Reader,
		generic.Option{
			ClusterScoped:         config.ControllerCfg.ClusterScoped,
			TargetNamespace:       config.ControllerCfg.TargetNamespace,
			EnableFilterNamespace: config.ControllerCfg.EnableFilterNamespace,
		},
	}
}

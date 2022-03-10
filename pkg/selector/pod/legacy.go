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

package pod

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	genericnamespace "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/namespace"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/registry"
)

// LegacyPodSelector is the temporary place to hold the legacy logic for selecting then filtering the Pod.
type LegacyPodSelector struct {
	// c is the kubernetes client which we usually use.
	c client.Client
	// r is actually the kubernetes client without indexing and caching, only used/required with fieldSelector.
	r client.Reader
	// clusterScoped means selector from all namespace.
	clusterScoped bool
	// targetNamespace means select from this namespace, only work with clusterScoped is true.
	targetNamespace string
	// enableFilterNamespace means select only from namespace annotated with generic.InjectAnnotationKey.
	enableFilterNamespace bool
}

func NewLegacyPodSelector(c client.Client, r client.Reader, clusterScoped bool, targetNamespace string, enableFilterNamespace bool) *LegacyPodSelector {
	return &LegacyPodSelector{c: c, r: r, clusterScoped: clusterScoped, targetNamespace: targetNamespace, enableFilterNamespace: enableFilterNamespace}
}

// SelectAndFilterPods returns the list of pods that filtered by selector and SelectorMode
// Deprecated: use pod.SelectImpl as instead
func (lps *LegacyPodSelector) SelectAndFilterPods(ctx context.Context, spec *v1alpha1.PodSelector) ([]v1.Pod, error) {
	if pods := mock.On("MockSelectAndFilterPods"); pods != nil {
		return pods.(func() []v1.Pod)(), nil
	}
	if err := mock.On("MockSelectedAndFilterPodsError"); err != nil {
		return nil, err.(error)
	}

	selector := spec.Selector
	mode := spec.Mode
	value := spec.Value

	pods, err := lps.SelectPods(ctx, selector)
	if err != nil {
		return nil, err
	}

	if len(pods) == 0 {
		err = errors.New("no pod is selected")
		return nil, err
	}

	filteredPod, err := lps.filterPodsByMode(pods, mode, value)
	if err != nil {
		return nil, err
	}

	return filteredPod, nil
}

//revive:disable:flag-parameter

// SelectPods returns the list of pods that are available for pod chaos action.
// It returns all pods that match the configured label, annotation and namespace selectors.
// If pods are specifically specified by `selector.Pods`, it just returns the selector.Pods.
//
// Deprecated: try to avoid using this method. Only use it until you REALLY need the all possible pods before filtering.
// For example, frontend hint for previewing/selecting pods.
func (lps *LegacyPodSelector) SelectPods(ctx context.Context, selector v1alpha1.PodSelectorSpec) ([]v1.Pod, error) {
	// pods are specifically specified
	if len(selector.Pods) > 0 {
		return lps.selectSpecifiedPods(ctx, selector)
	}

	selectorRegistry := newSelectorRegistry(ctx, lps.c, selector)
	selectorChain, err := registry.Parse(selectorRegistry, selector.GenericSelectorSpec, generic.Option{
		ClusterScoped:         lps.clusterScoped,
		TargetNamespace:       lps.targetNamespace,
		EnableFilterNamespace: lps.enableFilterNamespace,
	})
	if err != nil {
		return nil, err
	}

	return lps.listPods(ctx, selector, selectorChain)
}

// CheckPodMeetSelector checks if this pod meets the selection criteria.
func (lps *LegacyPodSelector) CheckPodMeetSelector(ctx context.Context, pod v1.Pod, selector v1alpha1.PodSelectorSpec) (bool, error) {
	if len(selector.Pods) > 0 {
		meet := false
		for ns, names := range selector.Pods {
			if pod.Namespace != ns {
				continue
			}

			for _, name := range names {
				if pod.Name == name {
					meet = true
				}
			}

			if !meet {
				return false, nil
			}
		}
	}

	selectorRegistry := newSelectorRegistry(ctx, lps.c, selector)
	selectorChain, err := registry.Parse(selectorRegistry, selector.GenericSelectorSpec, generic.Option{
		ClusterScoped:         lps.clusterScoped,
		TargetNamespace:       lps.targetNamespace,
		EnableFilterNamespace: lps.enableFilterNamespace,
	})
	if err != nil {
		return false, err
	}

	return selectorChain.Match(&pod), nil
}

func (lps *LegacyPodSelector) selectSpecifiedPods(ctx context.Context, spec v1alpha1.PodSelectorSpec) ([]v1.Pod, error) {
	var pods []v1.Pod
	namespaceCheck := make(map[string]bool)

	for ns, names := range spec.Pods {
		if !lps.clusterScoped {
			if lps.targetNamespace != ns {
				log.Info("skip namespace because ns is out of scope within namespace scoped mode", "namespace", ns)
				continue
			}
		}

		if lps.enableFilterNamespace {
			allow, ok := namespaceCheck[ns]
			if !ok {
				allow = genericnamespace.CheckNamespace(ctx, lps.c, ns, log)
				namespaceCheck[ns] = allow
			}
			if !allow {
				continue
			}
		}
		for _, name := range names {
			var pod v1.Pod
			err := lps.c.Get(ctx, types.NamespacedName{
				Namespace: ns,
				Name:      name,
			}, &pod)
			if err == nil {
				pods = append(pods, pod)
				continue
			}

			if apierrors.IsNotFound(err) {
				log.Error(err, "Pod is not found", "namespace", ns, "pod name", name)
				continue
			}

			return nil, err
		}
	}
	return pods, nil
}

func (lps *LegacyPodSelector) listPods(ctx context.Context, spec v1alpha1.PodSelectorSpec,
	selectorChain generic.SelectorChain) ([]v1.Pod, error) {
	var pods []v1.Pod
	namespaceCheck := make(map[string]bool)

	if err := selectorChain.ListObjects(lps.c, lps.r,
		func(listFunc generic.ListFunc, opts client.ListOptions) error {
			var podList v1.PodList
			if len(spec.Namespaces) > 0 {
				for _, namespace := range spec.Namespaces {
					if lps.enableFilterNamespace {
						allow, ok := namespaceCheck[namespace]
						if !ok {
							allow = genericnamespace.CheckNamespace(ctx, lps.c, namespace, log)
							namespaceCheck[namespace] = allow
						}
						if !allow {
							continue
						}
					}

					opts.Namespace = namespace
					if err := listFunc(ctx, &podList, &opts); err != nil {
						return err
					}
					pods = append(pods, podList.Items...)
				}
			} else {
				// in fact, this will never happen
				if err := listFunc(ctx, &podList, &opts); err != nil {
					return err
				}
				pods = append(pods, podList.Items...)
			}
			return nil
		}); err != nil {
		return nil, err
	}

	filterPods := make([]v1.Pod, 0, len(pods))
	for _, pod := range pods {
		pod := pod
		if selectorChain.Match(&pod) {
			filterPods = append(filterPods, pod)
		}
	}
	return filterPods, nil
}

// filterPodsByMode filters pods by mode from pod list
func (lps *LegacyPodSelector) filterPodsByMode(pods []v1.Pod, mode v1alpha1.SelectorMode, value string) ([]v1.Pod, error) {
	indexes, err := generic.FilterObjectsByMode(mode, value, len(pods))
	if err != nil {
		return nil, err
	}

	var filteredPods []v1.Pod

	for _, index := range indexes {
		index := index
		filteredPods = append(filteredPods, pods[index])
	}
	return filteredPods, nil
}

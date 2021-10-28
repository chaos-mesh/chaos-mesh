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

package pod

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/registry"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	generic_annotation "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/annotation"
	generic_field "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/field"
	generic_label "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/label"
	generic_namespace "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/namespace"
)

var log = ctrl.Log.WithName("podselector")

type SelectImpl struct {
	c client.Client
	r client.Reader

	generic.Option
}

type Pod struct {
	v1.Pod
}

func (pod *Pod) Id() string {
	return (types.NamespacedName{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	}).String()
}

func (impl *SelectImpl) Select(ctx context.Context, ps *v1alpha1.PodSelector) ([]*Pod, error) {
	if ps == nil {
		return []*Pod{}, nil
	}

	pods, err := SelectAndFilterPods(ctx, impl.c, impl.r, ps, impl.ClusterScoped, impl.TargetNamespace, impl.EnableFilterNamespace)
	if err != nil {
		return nil, err
	}

	var result []*Pod
	for _, pod := range pods {
		result = append(result, &Pod{
			pod,
		})
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

// SelectAndFilterPods returns the list of pods that filtered by selector and PodMode
func SelectAndFilterPods(ctx context.Context, c client.Client, r client.Reader, spec *v1alpha1.PodSelector, clusterScoped bool, targetNamespace string, enableFilterNamespace bool) ([]v1.Pod, error) {
	if pods := mock.On("MockSelectAndFilterPods"); pods != nil {
		return pods.(func() []v1.Pod)(), nil
	}
	if err := mock.On("MockSelectedAndFilterPodsError"); err != nil {
		return nil, err.(error)
	}

	selector := spec.Selector
	mode := spec.Mode
	value := spec.Value

	pods, err := SelectPods(ctx, c, r, selector, clusterScoped, targetNamespace, enableFilterNamespace)
	if err != nil {
		return nil, err
	}

	if len(pods) == 0 {
		err = errors.New("no pod is selected")
		return nil, err
	}

	filteredPod, err := filterPodsByMode(pods, mode, value)
	if err != nil {
		return nil, err
	}

	return filteredPod, nil
}

//revive:disable:flag-parameter

// SelectPods returns the list of pods that are available for pod chaos action.
// It returns all pods that match the configured label, annotation and namespace selectors.
// If pods are specifically specified by `selector.Pods`, it just returns the selector.Pods.
func SelectPods(ctx context.Context, c client.Client, r client.Reader, selector v1alpha1.PodSelectorSpec, clusterScoped bool, targetNamespace string, enableFilterNamespace bool) ([]v1.Pod, error) {
	// pods are specifically specified
	if len(selector.Pods) > 0 {
		return selectSpecifiedPods(ctx, c, selector, clusterScoped, targetNamespace, enableFilterNamespace)
	}

	selectorRegistry := newSelectorRegistry(ctx, c, selector)
	selectorChain, err := registry.Parse(selectorRegistry, selector.GenericSelectorSpec, generic.Option{
		ClusterScoped:         clusterScoped,
		TargetNamespace:       targetNamespace,
		EnableFilterNamespace: enableFilterNamespace,
	})
	if err != nil {
		return nil, err
	}

	var pods []v1.Pod
	pods, err = listPods(ctx, c, r, selector, selectorChain, enableFilterNamespace)
	if err != nil {
		return nil, err
	}

	filterPods := make([]v1.Pod, 0, len(pods))
	for _, pod := range pods {
		if selectorChain.Match(&pod) {
			filterPods = append(filterPods, pod)
		}
	}
	return filterPods, nil
}

func selectSpecifiedPods(ctx context.Context, c client.Client, spec v1alpha1.PodSelectorSpec,
	clusterScoped bool, targetNamespace string, enableFilterNamespace bool) ([]v1.Pod, error) {
	var pods []v1.Pod
	namespaceCheck := make(map[string]bool)

	for ns, names := range spec.Pods {
		if !clusterScoped {
			if targetNamespace != ns {
				log.Info("skip namespace because ns is out of scope within namespace scoped mode", "namespace", ns)
				continue
			}
		}

		if enableFilterNamespace {
			allow, ok := namespaceCheck[ns]
			if !ok {
				allow = generic_namespace.CheckNamespace(ctx, c, ns)
				namespaceCheck[ns] = allow
			}
			if !allow {
				continue
			}
		}
		for _, name := range names {
			var pod v1.Pod
			err := c.Get(ctx, types.NamespacedName{
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

//revive:enable:flag-parameter

// GetService get k8s service by service name
func GetService(ctx context.Context, c client.Client, namespace, controllerNamespace string, serviceName string) (*v1.Service, error) {
	// use the environment value if namespace is empty
	if len(namespace) == 0 {
		namespace = controllerNamespace
	}

	service := &v1.Service{}
	err := c.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      serviceName,
	}, service)
	if err != nil {
		return nil, err
	}

	return service, nil
}

// CheckPodMeetSelector checks if this pod meets the selection criteria.
func CheckPodMeetSelector(ctx context.Context, c client.Client, pod v1.Pod, selector v1alpha1.PodSelectorSpec, clusterScoped bool, targetNamespace string, enableFilterNamespace bool) (bool, error) {
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

	selectorRegistry := newSelectorRegistry(ctx, c, selector)
	selectorChain, err := registry.Parse(selectorRegistry, selector.GenericSelectorSpec, generic.Option{
		ClusterScoped:         clusterScoped,
		TargetNamespace:       targetNamespace,
		EnableFilterNamespace: enableFilterNamespace,
	})
	if err != nil {
		return false, err
	}

	return selectorChain.Match(&pod), nil
}

func newSelectorRegistry(ctx context.Context, c client.Client, spec v1alpha1.PodSelectorSpec) registry.Registry {
	return map[string]registry.SelectorFactory{
		generic_label.Name:      generic_label.New,
		generic_namespace.Name:  generic_namespace.New,
		generic_field.Name:      generic_field.New,
		generic_annotation.Name: generic_annotation.New,
		nodeSelectorName: func(selector v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
			return newNodeSelector(ctx, c, spec)
		},
		phaseSelectorName: func(selector v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
			return newPhaseSelector(spec)
		},
	}
}

func listPods(ctx context.Context, c client.Client, r client.Reader, spec v1alpha1.PodSelectorSpec,
	selectorChain generic.SelectorChain, enableFilterNamespace bool) ([]v1.Pod, error) {
	var pods []v1.Pod
	namespaceCheck := make(map[string]bool)

	if err := selectorChain.ListObjects(c, r,
		func(listFunc generic.ListFunc, opts client.ListOptions) error {
			var podList v1.PodList
			if len(spec.Namespaces) > 0 {
				for _, namespace := range spec.Namespaces {
					if enableFilterNamespace {
						allow, ok := namespaceCheck[namespace]
						if !ok {
							allow = generic_namespace.CheckNamespace(ctx, c, namespace)
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

	return pods, nil
}

// filterPodsByMode filters pods by mode from pod list
func filterPodsByMode(pods []v1.Pod, mode v1alpha1.PodMode, value string) ([]v1.Pod, error) {
	if len(pods) == 0 {
		return nil, errors.New("cannot generate pods from empty list")
	}

	switch mode {
	case v1alpha1.OnePodMode:
		index := getRandomNumber(len(pods))
		pod := pods[index]

		return []v1.Pod{pod}, nil
	case v1alpha1.AllPodMode:
		return pods, nil
	case v1alpha1.FixedPodMode:
		num, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		if len(pods) < num {
			num = len(pods)
		}

		if num <= 0 {
			return nil, errors.New("cannot select any pod as value below or equal 0")
		}

		return getFixedSubListFromPodList(pods, num), nil
	case v1alpha1.FixedPercentPodMode:
		percentage, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		if percentage == 0 {
			return nil, errors.New("cannot select any pod as value below or equal 0")
		}

		if percentage < 0 || percentage > 100 {
			return nil, fmt.Errorf("fixed percentage value of %d is invalid, Must be (0,100]", percentage)
		}

		num := int(math.Floor(float64(len(pods)) * float64(percentage) / 100))

		return getFixedSubListFromPodList(pods, num), nil
	case v1alpha1.RandomMaxPercentPodMode:
		maxPercentage, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		if maxPercentage == 0 {
			return nil, errors.New("cannot select any pod as value below or equal 0")
		}

		if maxPercentage < 0 || maxPercentage > 100 {
			return nil, fmt.Errorf("fixed percentage value of %d is invalid, Must be [0-100]", maxPercentage)
		}

		percentage := getRandomNumber(maxPercentage + 1) // + 1 because Intn works with half open interval [0,n) and we want [0,n]
		num := int(math.Floor(float64(len(pods)) * float64(percentage) / 100))

		return getFixedSubListFromPodList(pods, num), nil
	default:
		return nil, fmt.Errorf("mode %s not supported", mode)
	}
}

func getFixedSubListFromPodList(pods []v1.Pod, num int) []v1.Pod {
	indexes := RandomFixedIndexes(0, uint(len(pods)), uint(num))

	var filteredPods []v1.Pod

	for _, index := range indexes {
		index := index
		filteredPods = append(filteredPods, pods[index])
	}

	return filteredPods
}

// RandomFixedIndexes returns the `count` random indexes between `start` and `end`.
// [start, end)
func RandomFixedIndexes(start, end, count uint) []uint {
	var indexes []uint
	m := make(map[uint]uint, count)

	if end < start {
		return indexes
	}

	if count > end-start {
		for i := start; i < end; i++ {
			indexes = append(indexes, i)
		}

		return indexes
	}

	for i := 0; i < int(count); {
		index := uint(getRandomNumber(int(end-start))) + start

		_, exist := m[index]
		if exist {
			continue
		}

		m[index] = index
		indexes = append(indexes, index)
		i++
	}

	return indexes
}

func getRandomNumber(max int) uint64 {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return num.Uint64()
}

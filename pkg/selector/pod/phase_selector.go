package pod

import (
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type phaseSelector struct {
	reqIncl []labels.Requirement
	reqExcl []labels.Requirement
}

var _ generic.Selector = &phaseSelector{}

func (s *phaseSelector) AddListOption(opts client.ListOptions) client.ListOptions {
	return opts
}

func (s *phaseSelector) SetListFunc(f generic.ListFunc) generic.ListFunc {
	return f
}

func (s *phaseSelector) Match(obj client.Object) bool {
	included := len(s.reqIncl) == 0
	pod := obj.(*v1.Pod)
	selector := labels.Set{string(pod.Status.Phase): ""}

	// include pod if one including requirement matches
	for _, req := range s.reqIncl {
		if req.Matches(selector) {
			included = true
			break
		}
	}

	// exclude pod if it is filtered out by at least one excluding requirement
	for _, req := range s.reqExcl {
		if !req.Matches(selector) {
			included = false
			break
		}
	}

	return included
}

func NewPhaseSelector(spec v1alpha1.PodSelectorSpec) (generic.Selector, error) {
	selectorStr := strings.Join(spec.PodPhaseSelectors, ",")
	selector, err := labels.Parse(selectorStr)
	if err != nil {
		return nil, err
	}

	// TODO
	//if selector.Empty() {
	//	return pods, nil
	//}

	reqs, _ := selector.Requirements()
	var (
		reqIncl []labels.Requirement
		reqExcl []labels.Requirement
	)

	for _, req := range reqs {
		switch req.Operator() {
		case selection.Exists:
			reqIncl = append(reqIncl, req)
		case selection.DoesNotExist:
			reqExcl = append(reqExcl, req)
		default:
			return nil, fmt.Errorf("unsupported operator: %s", req.Operator())
		}
	}

	return &phaseSelector{
		reqIncl: reqIncl,
		reqExcl: reqExcl,
	}, nil
}

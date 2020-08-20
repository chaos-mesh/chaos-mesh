package injector

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/filter"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/injector/config"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type PodItem struct {
	Pod *corev1.Pod
	Config *config.Config
}

func (PodItem) TypeName() string {
	return "POD"
}

var PodFilter = filter.ItemFilter{
	MustOkRuleList:[]filter.Rule{
		podValid{},
		podInjectValid{},
		podPolicyValid{},
		podInNeverInjectSelector{},
	},
	MustOkRuleListUseCode: []bool{true,true,true,true},
	OneOkRuleList:[]filter.Rule{
		podInAlwaysInjectSelector{},
		podAnnotationOk{},
		podPolicyEnabled{},
	},
	OneOkRuleListUseCode: []bool{true,true,true},
}

type podValid struct {}

func (podValid)Check(item filter.Item) (bool,error) {
	if item.TypeName() != "POD" {
		return false,errors.Wrap(errors.New("Type Error: Not Pod;"),
			"podValid check failed;")
	}
	pod := item.(PodItem).Pod
	if len(pod.Spec.Containers) < 1{
		return false,errors.Wrap(errors.New("Pod Error: Pod Container not exist;"),
			"podValid check failed;")
	}
	return true,nil
}

type podInjectValid struct {}

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

func (podInjectValid) Check(item filter.Item) (bool, error) {
	pod := item.(PodItem).Pod
	if pod.Spec.HostNetwork {
		return false,nil
	}
	for _, namespace := range ignoredNamespaces {
		if pod.ObjectMeta.Namespace == namespace {
			return false,nil
		}
	}
	return true,nil
}

type podPolicyValid struct {}

func (podPolicyValid) Check(item filter.Item) (bool, error) {
	// Todo: write config change struct
	c := *item.(PodItem).Config
	if c.Policy != config.InjectionPolicyDisabled &&
		c.Policy != config.InjectionPolicyEnabled {
		return false, errors.Errorf("Illegal value for autoInject:%s, must be one of [%s,%s]. Auto injection disabled!",
			c.Policy, config.InjectionPolicyDisabled, config.InjectionPolicyEnabled)
	}
	return true,nil
}

type podInNeverInjectSelector struct {}

func (podInNeverInjectSelector)Check(item filter.Item) (bool,error) {
	pod := item.(PodItem).Pod
	c := *item.(PodItem).Config
	for _, neverSelector := range c.NeverInjectSelector {
		selector, err := metav1.LabelSelectorAsSelector(&neverSelector)
		if err != nil {
			return false, errors.Errorf("Invalid selector for NeverInjectSelector: %v (%v)",
				neverSelector, err)
		} else if !selector.Empty() && selector.Matches(labels.Set(pod.ObjectMeta.Labels)) {
			log.Info("Explicitly disabling injection for Pod %s due to Pod labels matching NeverInjectSelector config map entry.",
				pod.ObjectMeta.Namespace)
			return false,nil
		}
	}
	return true,nil
}

type podAnnotationOk struct {}

func (podAnnotationOk)Check(item filter.Item) (bool,error) {
	//todo:
	return false,nil
}

type podInAlwaysInjectSelector struct {}

func (podInAlwaysInjectSelector)Check(item filter.Item) (bool,error) {
	pod := item.(PodItem).Pod
	c := *item.(PodItem).Config
	for _, alwaysSelector := range c.AlwaysInjectSelector {
		selector, err := metav1.LabelSelectorAsSelector(&alwaysSelector)
		if err != nil {
			return false, errors.Errorf("Invalid selector for AlwaysInjectSelector: %v (%v)", alwaysSelector, err)
		} else if !selector.Empty() && selector.Matches(labels.Set(pod.ObjectMeta.Labels)) {
			log.Info("Explicitly enabling injection for Pod %s due to Pod labels matching AlwaysInjectSelector config map entry.",
				pod.ObjectMeta.Namespace)
			return true,nil
		}
	}
	return false,nil
}

type podPolicyEnabled struct {}

func (podPolicyEnabled)Check(item filter.Item) (bool,error) {
	c := *item.(PodItem).Config
	if c.Policy == config.InjectionPolicyEnabled {
		return true, nil
	}
	return false,nil
}


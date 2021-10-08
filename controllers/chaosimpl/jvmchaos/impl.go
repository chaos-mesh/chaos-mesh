// Copyright 2020 Chaos Mesh Authors.
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

package jvmchaos

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/go-logr/logr"
	"github.com/pingcap/errors"
	"go.uber.org/fx"

	//v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	//"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	//"github.com/chaos-mesh/chaos-mesh/pkg/jvm"
)

const CommonRuleTemplate = `
RULE {{.Name}}
CLASS {{.Class}}
METHOD {{.Method}}
AT ENTRY
IF true
DO 
	{{.Do}};
ENDRULE
`

const StressRuleTemplate = `
RULE {{.Name}}
STRESS {{.StressType}}
{{.StressValueName}} {{.StressValue}}
ENDRULE
`

const GcRuleTemplate = `
RULE {{.Name}}
GC
ENDRULE
`

type Impl struct {
	client.Client
	Log logr.Logger

	decoder *utils.ContianerRecordDecoder
}

var _ common.ChaosImpl = (*Impl)(nil)

// Apply applies jvm-chaos
func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("jvm chaos apply", "record", records[index])
	if impl.decoder == nil {
		return v1alpha1.NotInjected, fmt.Errorf("impl decoder is nil")
	}
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index])
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	if decodedContainer.PbClient != nil {
		defer decodedContainer.PbClient.Close()
	}

	jvmChaos := obj.(*v1alpha1.JVMChaos)
	jvmChaos.Spec.Name = jvmChaos.Name
	err = generateRuleData(&jvmChaos.Spec)
	if err != nil {
		return v1alpha1.Injected, err
	}

	_, err = decodedContainer.PbClient.InstallJVMRules(ctx, &pb.InstallJVMRulesRequest{
		ContainerId: decodedContainer.ContainerId,
		Rule:        jvmChaos.Spec.RuleData,
		Enable:      true,
		EnterNS:     true,
	})
	if err != nil {
		impl.Log.Error(err, "install jvm rules")
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil

	/*
		jvmchaos := obj.(*v1alpha1.JVMChaos)

		var pod v1.Pod
		namespacedName, err := controller.ParseNamespacedName(records[index].Id)
		if err != nil {
			return v1alpha1.NotInjected, err
		}
		err = impl.Client.Get(ctx, namespacedName, &pod)
		if err != nil {
			// TODO: handle this error
			return v1alpha1.NotInjected, err
		}

		impl.Log.Info("Try to apply jvm chaos", "namespace",
			pod.Namespace, "name", pod.Name)

		// TODO: Custom port may be required
		err = jvm.ActiveSandbox(pod.Status.PodIP, sandboxPort)
		if err != nil {
			return v1alpha1.NotInjected, err
		}

		impl.Log.Info("active sandbox", "pod", pod.Name)

		suid := genSUID(&pod, jvmchaos)
		jsonBytes, err := jvm.ToSandboxAction(suid, jvmchaos)

		if err != nil {
			return v1alpha1.NotInjected, err
		}
		// TODO: Custom port may be required
		err = jvm.InjectChaos(pod.Status.PodIP, sandboxPort, jsonBytes)
		if err != nil {
			return v1alpha1.NotInjected, err
		}
		impl.Log.Info("Inject JVM Chaos", "pod", pod.Name, "action", jvmchaos.Spec.Action)

		return v1alpha1.Injected, nil
	*/
}

// Recover means the reconciler recovers the chaos action
func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	if impl.decoder == nil {
		return v1alpha1.Injected, fmt.Errorf("impl decoder is nil")
	}
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index])
	if decodedContainer.PbClient != nil {
		defer decodedContainer.PbClient.Close()
	}
	if err != nil {
		return v1alpha1.Injected, err
	}

	jvmChaos := obj.(*v1alpha1.JVMChaos)
	err = generateRuleData(&jvmChaos.Spec)
	if err != nil {
		return v1alpha1.Injected, err
	}

	_, err = decodedContainer.PbClient.InstallJVMRules(ctx, &pb.InstallJVMRulesRequest{
		ContainerId: decodedContainer.ContainerId,
		Rule:        jvmChaos.Spec.RuleData,
		Enable:      false,
		EnterNS:     true,
	})
	if err != nil {
		impl.Log.Error(err, "uninstall jvm rules")
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil

	/*
		jvmchaos := obj.(*v1alpha1.JVMChaos)

		var pod v1.Pod
		namespacedName, err := controller.ParseNamespacedName(records[index].Id)
		if err != nil {
			// This error is not expected to exist
			return v1alpha1.NotInjected, err
		}
		err = impl.Client.Get(ctx, namespacedName, &pod)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				return v1alpha1.Injected, err
			}

			impl.Log.Info("Target pod has been deleted", "namespace", pod.Namespace, "name", pod.Name)
			return v1alpha1.NotInjected, nil

		}

		impl.Log.Info("Try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

		suid := genSUID(&pod, jvmchaos)
		jsonBytes, err := jvm.ToSandboxAction(suid, jvmchaos)
		if err != nil {
			return v1alpha1.Injected, err
		}

		// TODO: Custom port may be required
		err = jvm.RecoverChaos(pod.Status.PodIP, sandboxPort, jsonBytes)

		if err != nil {
			return v1alpha1.Injected, err
		}

		return v1alpha1.NotInjected, nil
	*/
}

// JVMRuleParameter is only used to generate rule data
type JVMRuleParameter struct {
	v1alpha1.JVMParameter

	StressType      string
	StressValue     string
	StressValueName string
	Do              string
}

func generateRuleData(spec *v1alpha1.JVMChaosSpec) error {
	if len(spec.RuleData) != 0 {
		return nil
	}

	ruleParameter := &JVMRuleParameter{
		JVMParameter: spec.JVMParameter,
	}

	switch spec.Action {
	case v1alpha1.JVMLatencyAction:
		ruleParameter.Do = fmt.Sprintf("Thread.sleep(%d)", ruleParameter.LatencyDuration)
	case v1alpha1.JVMExceptionAction:
		ruleParameter.Do = fmt.Sprintf("throw new %s", ruleParameter.ThrowException)
	case v1alpha1.JVMReturnAction:
		ruleParameter.Do = fmt.Sprintf("return %s", ruleParameter.ReturnValue)
	case v1alpha1.JVMStressAction:
		if ruleParameter.CPUCount > 0 {
			ruleParameter.StressType = "CPU"
			ruleParameter.StressValueName = "CPUCOUNT"
			ruleParameter.StressValue = fmt.Sprintf("%d", ruleParameter.CPUCount)
		} else {
			ruleParameter.StressType = "MEMORY"
			ruleParameter.StressValueName = "MEMORYTYPE"
			ruleParameter.StressValue = ruleParameter.MemoryType
		}
	}

	buf := new(bytes.Buffer)
	var t *template.Template
	switch spec.Action {
	case v1alpha1.JVMStressAction:
		t = template.Must(template.New("byteman rule").Parse(StressRuleTemplate))
	case v1alpha1.JVMExceptionAction, v1alpha1.JVMLatencyAction, v1alpha1.JVMReturnAction:
		t = template.Must(template.New("byteman rule").Parse(CommonRuleTemplate))
	case v1alpha1.JVMGCAction:
		t = template.Must(template.New("byteman rule").Parse(GcRuleTemplate))
	default:
		return errors.Errorf("jvm action %s not supported", spec.Action)
	}
	if t == nil {
		return errors.Errorf("parse byeman rule template failed")
	}
	err := t.Execute(buf, ruleParameter)
	if err != nil {
		return err
	}

	spec.RuleData = buf.String()
	return nil
}

// Object would return the instance of chaos

func NewImpl(c client.Client, log logr.Logger, decoder *utils.ContianerRecordDecoder) *common.ChaosImplPair {
	return &common.ChaosImplPair{
		Name:   "jvmchaos",
		Object: &v1alpha1.JVMChaos{},
		Impl: &Impl{
			Client:  c,
			Log:     log.WithName("jvmchaos"),
			decoder: decoder,
		},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)

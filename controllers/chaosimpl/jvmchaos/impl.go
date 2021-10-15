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

package jvmchaos

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/go-logr/logr"
	"github.com/pingcap/errors"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
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
	err = generateRuleData(&jvmChaos.Spec)
	if err != nil {
		return v1alpha1.Injected, err
	}

	_, err = decodedContainer.PbClient.InstallJVMRules(ctx, &pb.InstallJVMRulesRequest{
		ContainerId: decodedContainer.ContainerId,
		Rule:        jvmChaos.Spec.RuleData,
		Port:        jvmChaos.Spec.Port,
		Enable:      true,
		EnterNS:     true,
	})
	if err != nil {
		impl.Log.Error(err, "install jvm rules")
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
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
		Port:        jvmChaos.Spec.Port,
		Enable:      false,
		EnterNS:     true,
	})
	if err != nil {
		impl.Log.Error(err, "uninstall jvm rules")
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil
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

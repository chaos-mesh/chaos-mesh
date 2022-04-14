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
	"strings"
	"text/template"

	"github.com/go-logr/logr"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

var (
	errNilDecoder error = errors.New("impl decoder is nil")
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

const (
	// byteman rule template
	SimpleRuleTemplate = `
RULE {{.Name}}
CLASS {{.Class}}
METHOD {{.Method}}
AT ENTRY
IF true
DO
	{{.Do}};
ENDRULE
`

	CompleteRuleTemplate = `
RULE {{.Name}}
CLASS {{.Class}}
METHOD {{.Method}}
HELPER {{.Helper}}
AT ENTRY
BIND {{.Bind}};
IF {{.Condition}}
DO
	{{.Do}};
ENDRULE
`

	// for action 'gc' and 'stress'
	GCHelper     = "org.chaos_mesh.byteman.helper.GCHelper"
	StressHelper = "org.chaos_mesh.byteman.helper.StressHelper"

	// the trigger point for 'gc' and 'stress'
	TriggerClass  = "org.chaos_mesh.chaos_agent.TriggerThread"
	TriggerMethod = "triggerFunc"
)

// BytemanTemplateSpec is the template spec for byteman rule
type BytemanTemplateSpec struct {
	Name      string
	Class     string
	Method    string
	Helper    string
	Bind      string
	Condition string
	Do        string

	// below is only used for stress template
	StressType      string
	StressValueName string
	StressValue     string
}

type Impl struct {
	client.Client
	Log logr.Logger

	decoder *utils.ContainerRecordDecoder
}

// Apply applies jvm-chaos
func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("jvm chaos apply", "record", records[index])
	if impl.decoder == nil {
		return v1alpha1.NotInjected, errors.WithStack(errNilDecoder)
	}
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index], obj)
	if decodedContainer.PbClient != nil {
		defer func() {
			err := decodedContainer.PbClient.Close()
			if err != nil {
				impl.Log.Error(err, "fail to close pb client")
			}
		}()
	}
	if err != nil {
		return v1alpha1.NotInjected, err
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
		return v1alpha1.Injected, errors.WithStack(errNilDecoder)
	}
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index], obj)
	if decodedContainer.PbClient != nil {
		defer func() {
			err := decodedContainer.PbClient.Close()
			if err != nil {
				impl.Log.Error(err, "fail to close pb client")
			}
		}()
	}
	if err != nil && strings.Contains(err.Error(), "container not found") {
		// Unable to find the container, so we are unable to remove the experiment from the jvm as it has gone
		impl.Log.Error(err, "finding container")
		return v1alpha1.Recovered, nil
	}
	if err != nil {
		return v1alpha1.Injected, err
	}

	jvmChaos := obj.(*v1alpha1.JVMChaos)
	err = generateRuleData(&jvmChaos.Spec)
	if err != nil {
		return v1alpha1.Injected, err
	}

	_, err = decodedContainer.PbClient.UninstallJVMRules(ctx, &pb.UninstallJVMRulesRequest{
		ContainerId: decodedContainer.ContainerId,
		Rule:        jvmChaos.Spec.RuleData,
		Port:        jvmChaos.Spec.Port,
		EnterNS:     true,
	})
	if err != nil && strings.Contains(err.Error(), "Connection refused") {
		// Unable to connect to the jvm - meaning that there is no agent running on the jvm, most likely because the jvm process has been restarted
		impl.Log.Error(err, "uninstall jvm rules (possible restart of jvm process)")
		return v1alpha1.Recovered, nil
	}
	if err != nil {
		impl.Log.Error(err, "uninstall jvm rules")
		return v1alpha1.Injected, err
	}

	return v1alpha1.Recovered, nil
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

	bytemanTemplateSpec := BytemanTemplateSpec{
		Name:   spec.Name,
		Class:  spec.Class,
		Method: spec.Method,
	}

	switch spec.Action {
	case v1alpha1.JVMLatencyAction:
		bytemanTemplateSpec.Do = fmt.Sprintf("Thread.sleep(%d)", spec.LatencyDuration)
	case v1alpha1.JVMExceptionAction:
		bytemanTemplateSpec.Do = fmt.Sprintf("throw new %s", spec.ThrowException)
	case v1alpha1.JVMReturnAction:
		bytemanTemplateSpec.Do = fmt.Sprintf("return %s", spec.ReturnValue)
	case v1alpha1.JVMStressAction:
		bytemanTemplateSpec.Helper = StressHelper
		bytemanTemplateSpec.Class = TriggerClass
		bytemanTemplateSpec.Method = TriggerMethod
		// the bind and condition is useless, only used for fill the template
		bytemanTemplateSpec.Bind = "flag:boolean=true"
		bytemanTemplateSpec.Condition = "true"
		if spec.CPUCount > 0 {
			bytemanTemplateSpec.Do = fmt.Sprintf("injectCPUStress(\"%s\", %d)", spec.Name, spec.CPUCount)
		} else {
			bytemanTemplateSpec.Do = fmt.Sprintf("injectMemStress(\"%s\", \"%s\")", spec.Name, spec.MemoryType)
		}
	case v1alpha1.JVMGCAction:
		bytemanTemplateSpec.Helper = GCHelper
		bytemanTemplateSpec.Class = TriggerClass
		bytemanTemplateSpec.Method = TriggerMethod
		// the bind and condition is useless, only used for fill the template
		bytemanTemplateSpec.Bind = "flag:boolean=true"
		bytemanTemplateSpec.Condition = "true"
		bytemanTemplateSpec.Do = "gc()"
	}

	buf := new(bytes.Buffer)
	var t *template.Template
	switch spec.Action {
	case v1alpha1.JVMStressAction, v1alpha1.JVMGCAction:
		t = template.Must(template.New("byteman rule").Parse(CompleteRuleTemplate))
	case v1alpha1.JVMExceptionAction, v1alpha1.JVMLatencyAction, v1alpha1.JVMReturnAction:
		t = template.Must(template.New("byteman rule").Parse(SimpleRuleTemplate))
	default:
		return errors.Errorf("jvm action %s not supported", spec.Action)
	}
	if t == nil {
		return errors.Errorf("parse byeman rule template failed")
	}
	err := t.Execute(buf, bytemanTemplateSpec)
	if err != nil {
		log.Error("executing template", zap.Error(err))
		return err
	}

	spec.RuleData = buf.String()
	return nil
}

// Object would return the instance of chaos
func NewImpl(c client.Client, log logr.Logger, decoder *utils.ContainerRecordDecoder) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
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

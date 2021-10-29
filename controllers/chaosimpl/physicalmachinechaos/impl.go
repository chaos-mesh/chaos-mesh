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

package physicalmachinechaos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
)

type Impl struct {
	client.Client
	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("apply physical machine chaos")

	var physicalMachine v1alpha1.PhysicalMachine
	physicalMachineChaos := obj.(*v1alpha1.PhysicalMachineChaos)
	namespacedName, err := controller.ParseNamespacedName(records[index].Id)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	err = impl.Get(ctx, namespacedName, &physicalMachine)
	if err != nil {
		// TODO: handle this error
		return v1alpha1.NotInjected, err
	}

	// for example, physicalMachinechaos.Spec.Action is 'network-delay', action is 'network', subAction is 'delay'
	// notice: 'process' and 'clock' action has no subAction, set subAction to ""
	actions := strings.SplitN(string(physicalMachineChaos.Spec.Action), "-", 2)
	if len(actions) == 1 {
		actions = append(actions, "")
	} else if len(actions) != 2 {
		err := errors.New("action invalid")
		return v1alpha1.NotInjected, err
	}
	action, subAction := actions[0], actions[1]
	physicalMachineChaos.Spec.ExpInfo.Action = subAction

	/*
		transform ExpInfo in PhysicalMachineChaos to json data required by chaosd
		for example:
		    ExpInfo: &ExpInfo {
			    UID: "123",
				Action: "cpu",
				StressCPU: &StressCPU {
					Load: 1,
					Workers: 1,
				}
			}

			transform to json data: "{\"uid\":\"123\",\"action\":\"cpu\",\"load\":1,\"workers\":1}
	*/
	var expInfoMap map[string]interface{}
	expInfoBytes, _ := json.Marshal(physicalMachineChaos.Spec.ExpInfo)
	err = json.Unmarshal(expInfoBytes, &expInfoMap)
	if err != nil {
		impl.Log.Error(err, "fail to unmarshal experiment info")
		return v1alpha1.NotInjected, err
	}
	configKV, ok := expInfoMap[string(physicalMachineChaos.Spec.Action)].(map[string]interface{})
	if !ok {
		err = errors.New("transform action config to map failed")
		impl.Log.Error(err, "")
		return v1alpha1.NotInjected, err
	}
	for k, v := range configKV {
		expInfoMap[k] = v
	}
	delete(expInfoMap, string(physicalMachineChaos.Spec.Action))

	expInfoBytes, err = json.Marshal(expInfoMap)
	if err != nil {
		impl.Log.Error(err, "fail to marshal experiment info")
		return v1alpha1.NotInjected, err
	}

	url := fmt.Sprintf("%s/api/attack/%s", physicalMachine.Spec.Address, action)
	impl.Log.Info("HTTP request", "address", physicalMachine.Spec.Address, "data", string(expInfoBytes))

	statusCode, err := impl.doHttpRequest("POST", url, bytes.NewBuffer(expInfoBytes))
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	if statusCode != http.StatusOK {
		err = errors.New("HTTP status is not OK")
		impl.Log.Error(err, "")
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("recover physical machine chaos")

	var physicalMachine v1alpha1.PhysicalMachine
	physicalMachineChaos := obj.(*v1alpha1.PhysicalMachineChaos)
	namespacedName, err := controller.ParseNamespacedName(records[index].Id)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	err = impl.Get(ctx, namespacedName, &physicalMachine)
	if err != nil {
		// TODO: handle this error
		return v1alpha1.NotInjected, err
	}

	url := fmt.Sprintf("%s/api/attack/%s", physicalMachine.Spec.Address, physicalMachineChaos.Spec.ExpInfo.UID)
	statusCode, err := impl.doHttpRequest("DELETE", url, nil)
	if err != nil {
		return v1alpha1.Injected, err
	}

	if statusCode == http.StatusNotFound {
		impl.Log.Info("experiment not found", "uid", physicalMachineChaos.Spec.ExpInfo.UID)
	} else if statusCode != http.StatusOK {
		err = errors.New("HTTP status is not OK")
		impl.Log.Error(err, "")
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil
}

func (impl *Impl) doHttpRequest(method, url string, data io.Reader) (int, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		impl.Log.Error(err, "fail to generate HTTP request")
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		impl.Log.Error(err, "do HTTP request")
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	impl.Log.Info("HTTP response", "url", url, "status", resp.Status, "body", string(body))

	return resp.StatusCode, nil
}

func NewImpl(c client.Client, log logr.Logger) *common.ChaosImplPair {
	return &common.ChaosImplPair{
		Name:   "physicalmachinechaos",
		Object: &v1alpha1.PhysicalMachineChaos{},
		Impl: &Impl{
			Client: c,
			Log:    log.WithName("physicalmachinechaos"),
		},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)

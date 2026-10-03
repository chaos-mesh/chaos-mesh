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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client
	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("apply physical machine chaos")

	physicalMachineChaos := obj.(*v1alpha1.PhysicalMachineChaos)
	var address string
	// For backwards compatibility, physical machines can be selected in two ways.
	// Consequently, a record ID can have either of the following forms:
	//
	// 1. When using spec.address, it is the physical machine address.
	// 2. When using a selector, it is a namespaced name.
	if len(physicalMachineChaos.Spec.Address) > 0 {
		address = records[index].Id
	} else {
		var physicalMachine v1alpha1.PhysicalMachine
		namespacedName, err := controller.ParseNamespacedName(records[index].Id)
		if err != nil {
			return v1alpha1.NotInjected, err
		}
		err = impl.Get(ctx, namespacedName, &physicalMachine)
		if err != nil {
			// TODO: Handle this error.
			return v1alpha1.NotInjected, err
		}
		address = physicalMachine.Spec.Address
	}

	// Split an action such as "network-delay" into "network" and "delay".
	// The "process", "vm", "clock", and "user_defined" actions have no sub-action.
	actions := strings.SplitN(string(physicalMachineChaos.Spec.Action), "-", 2)
	if len(actions) == 1 {
		actions = append(actions, "")
	} else if len(actions) != 2 {
		err := errors.New("action invalid")
		return v1alpha1.NotInjected, err
	}
	action, subAction := actions[0], actions[1]
	physicalMachineChaos.Spec.ExpInfo.Action = subAction

	// Chaosd expects action configuration at the top level, while ExpInfo stores it
	// beneath the action name. Flatten the action configuration for the request.
	var expInfoMap map[string]any
	expInfoBytes, err := json.Marshal(physicalMachineChaos.Spec.ExpInfo)
	if err != nil {
		impl.Log.Error(err, "fail to marshal experiment info")
		return v1alpha1.NotInjected, err
	}
	err = json.Unmarshal(expInfoBytes, &expInfoMap)
	if err != nil {
		impl.Log.Error(err, "fail to unmarshal experiment info")
		return v1alpha1.NotInjected, err
	}
	configKV, ok := expInfoMap[string(physicalMachineChaos.Spec.Action)].(map[string]any)
	if !ok {
		err = errors.New("transform action config to map failed")
		impl.Log.Error(err, "")
		return v1alpha1.NotInjected, err
	}
	delete(expInfoMap, string(physicalMachineChaos.Spec.Action))
	maps.Copy(expInfoMap, configKV)

	expInfoBytes, err = json.Marshal(expInfoMap)
	if err != nil {
		impl.Log.Error(err, "fail to marshal experiment info")
		return v1alpha1.NotInjected, err
	}

	url := fmt.Sprintf("%s/api/attack/%s", address, action)
	impl.Log.Info("HTTP request", "address", address, "data", string(expInfoBytes))

	statusCode, body, err := impl.doHttpRequest("POST", url, bytes.NewBuffer(expInfoBytes))
	if err != nil {
		return v1alpha1.NotInjected, errors.Wrap(err, body)
	}

	if statusCode != http.StatusOK {
		err = errors.New("HTTP status is not OK")
		impl.Log.Error(err, body)
		return v1alpha1.NotInjected, errors.Wrap(err, body)
	}

	var response struct {
		UID string `json:"uid"`
	}
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return v1alpha1.NotInjected, errors.Wrap(err, "unmarshal chaosd response")
	}
	if response.UID == "" {
		return v1alpha1.NotInjected, errors.Errorf("chaosd response does not contain a uid for target %s", records[index].Id)
	}
	// Store the UID for this target because one experiment can apply to multiple chaosd instances.
	if physicalMachineChaos.Status.ChaosdUIDs == nil {
		physicalMachineChaos.Status.ChaosdUIDs = make(map[string]string)
	}
	physicalMachineChaos.Status.ChaosdUIDs[records[index].Id] = response.UID

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("recover physical machine chaos")

	physicalMachineChaos := obj.(*v1alpha1.PhysicalMachineChaos)
	var address string
	if len(physicalMachineChaos.Spec.Address) > 0 {
		address = records[index].Id
	} else {
		var physicalMachine v1alpha1.PhysicalMachine
		namespacedName, err := controller.ParseNamespacedName(records[index].Id)
		if err != nil {
			return v1alpha1.Injected, err
		}
		err = impl.Get(ctx, namespacedName, &physicalMachine)
		if err != nil {
			// TODO: Handle this error.
			return v1alpha1.Injected, err
		}
		address = physicalMachine.Spec.Address
	}

	uid := physicalMachineChaos.Status.ChaosdUIDs[records[index].Id]
	if uid == "" {
		// Preserve recovery for existing experiments that specify a UID in their spec.
		uid = physicalMachineChaos.Spec.ExpInfo.UID
	}
	if uid == "" {
		return v1alpha1.Injected, errors.Errorf("chaosd uid not found for target %s", records[index].Id)
	}

	url := fmt.Sprintf("%s/api/attack/%s", address, uid)
	statusCode, body, err := impl.doHttpRequest("DELETE", url, nil)
	if err != nil {
		return v1alpha1.Injected, errors.Wrap(err, body)
	}

	if statusCode == http.StatusNotFound {
		impl.Log.Info("experiment not found", "uid", uid)
	} else if statusCode != http.StatusOK {
		err = errors.New("HTTP status is not OK")
		impl.Log.Error(err, body)
		return v1alpha1.Injected, errors.Wrap(err, body)
	}
	// A successful or idempotent recovery no longer needs a persisted chaosd UID.
	delete(physicalMachineChaos.Status.ChaosdUIDs, records[index].Id)

	return v1alpha1.NotInjected, nil
}

func (impl *Impl) doHttpRequest(method, url string, data io.Reader) (int, string, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		impl.Log.Error(err, "fail to generate HTTP request")
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	var httpClient *http.Client
	if config.ControllerCfg.ChaosdSecurityMode {
		httpClient, err = securityHTTPClient(url)
		if err != nil {
			impl.Log.Error(err, "generate HTTPS client")
			return 0, "", err
		}
	} else {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		impl.Log.Error(err, "do HTTP request")
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}
	impl.Log.Info("HTTP response", "url", url, "status", resp.Status, "body", string(body))

	return resp.StatusCode, string(body), nil
}

func securityHTTPClient(url string) (*http.Client, error) {
	if !strings.Contains(url, "https") {
		return nil, errors.Errorf("a secure url should begin with `https` rather than `http`, url: %s", url)
	}

	pair, err := tls.LoadX509KeyPair(config.ControllerCfg.ChaosdClientCert, config.ControllerCfg.ChaosdClientKey)
	if err != nil {
		return nil, errors.Wrap(err, "load x509 key pair failed")
	}

	pool := x509.NewCertPool()
	ca, err := os.ReadFile(config.ControllerCfg.ChaosdCACert)
	if err != nil {
		return nil, errors.Wrap(err, "read ChaosdCACert file failed")
	}
	pool.AppendCertsFromPEM(ca)

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      pool,
				Certificates: []tls.Certificate{pair},
				ServerName:   "chaosd.chaos-mesh.org",
			},
		},
		Timeout: 5 * time.Second,
	}, nil
}

func NewImpl(c client.Client, log logr.Logger) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
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

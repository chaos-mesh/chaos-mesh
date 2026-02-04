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

package common

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func GetYandexCloudSDK(ctx context.Context, client client.Client, ycchaos *v1alpha1.YCChaos) (*ycsdk.SDK, error) {
	if ycchaos.Spec.SecretName == nil {
		return nil, errors.New("secret name is required for Yandex Cloud authentication")
	}

	secret := &v1.Secret{}
	err := client.Get(ctx, types.NamespacedName{
		Name:      *ycchaos.Spec.SecretName,
		Namespace: ycchaos.Namespace,
	}, secret)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get cloud secret")
	}

	saKeyData, exists := secret.Data["sa-key.json"]
	if !exists {
		return nil, errors.New("sa-key.json not found in secret")
	}

	key, err := iamkey.ReadFromJSONBytes(saKeyData)
	if err != nil {
		return nil, errors.Wrap(err, "fail to parse service account key")
	}

	creds, err := ycsdk.ServiceAccountKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create credentials from service account key")
	}

	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: creds,
		Endpoint:    "api.cloud.yandex.net:443",
	})
	if err != nil {
		return nil, errors.Wrap(err, "fail to build Yandex Cloud SDK")
	}

	return sdk, nil
}

func ParseYCChaosAndSelector(obj v1alpha1.InnerObject, records []*v1alpha1.Record, index int, log logr.Logger) (*v1alpha1.YCChaos, *v1alpha1.YCSelector, error) {
	ycchaos, ok := obj.(*v1alpha1.YCChaos)
	if !ok {
		err := errors.New("chaos is not ycchaos")
		log.Error(err, "chaos is not YCChaos", "chaos", obj)
		return nil, nil, err
	}

	var selected v1alpha1.YCSelector
	err := json.Unmarshal([]byte(records[index].Id), &selected)
	if err != nil {
		log.Error(err, "fail to unmarshal the selector")
		return nil, nil, err
	}

	return ycchaos, &selected, nil
}

func WaitForOperation(ctx context.Context, sdk *ycsdk.SDK, op *operation.Operation, log logr.Logger, operationName string) error {
	wrappedOp, err := sdk.WrapOperation(op, nil)
	if err != nil {
		log.Error(err, "fail to wrap operation", "operation", operationName)
		return errors.Wrapf(err, "fail to wrap %s operation", operationName)
	}

	err = wrappedOp.Wait(ctx)
	if err != nil {
		log.Error(err, "fail to wait for operation to complete", "operation", operationName)
		return errors.Wrapf(err, "fail to wait for %s operation to complete", operationName)
	}

	return nil
}

func LogOperationSuccess(log logr.Logger, operationName string, instanceId string) {
	log.Info("compute instance operation completed successfully",
		"operation", operationName,
		"instanceId", instanceId)
}

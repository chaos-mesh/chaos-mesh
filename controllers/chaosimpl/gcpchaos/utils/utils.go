// Copyright 2021 Chaos Mesh Authors.
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

package utils

import (
	"context"
	"encoding/base64"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// GetComputeService is used to get the GCP compute Service.
func GetComputeService(ctx context.Context, cli client.Client, gcpchaos *v1alpha1.GcpChaos) (*compute.Service, error) {
	if gcpchaos.Spec.SecretName != nil {
		secret := &v1.Secret{}
		err := cli.Get(ctx, types.NamespacedName{
			Name:      *gcpchaos.Spec.SecretName,
			Namespace: gcpchaos.Namespace,
		}, secret)
		if err != nil {
			return nil, err
		}

		decodeBytes, err := base64.StdEncoding.DecodeString(string(secret.Data["service_account"]))
		if err != nil {
			return nil, err
		}
		computeService, err := compute.NewService(ctx, option.WithCredentialsJSON(decodeBytes))
		if err != nil {
			return nil, err
		}
		return computeService, nil
	}

	computeService, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return computeService, nil
}

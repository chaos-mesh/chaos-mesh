// Copyright 2022 Chaos Mesh Authors.
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

package utils

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// GetVMClient is used to get the azure VM Client
func GetVMClient(ctx context.Context, cli client.Client, azurechaos *v1alpha1.AzureChaos) (*compute.VirtualMachinesClient, error) {
	authorizer, err := GetAuthorizer(ctx, cli, azurechaos)
	if err != nil {
		return nil, err
	}

	vmClient := compute.NewVirtualMachinesClient(azurechaos.Spec.SubscriptionID)
	vmClient.Authorizer = authorizer

	return &vmClient, nil
}

// GetDiskClient is used to get the azure disk Client
func GetDiskClient(ctx context.Context, cli client.Client, azurechaos *v1alpha1.AzureChaos) (*compute.DisksClient, error) {
	authorizer, err := GetAuthorizer(ctx, cli, azurechaos)
	if err != nil {
		return nil, err
	}
	disksClient := compute.NewDisksClient(azurechaos.Spec.SubscriptionID)
	disksClient.Authorizer = authorizer

	return &disksClient, nil
}

// GetAuthorizer is used to get the azure authorizer
func GetAuthorizer(ctx context.Context, cli client.Client, azurechaos *v1alpha1.AzureChaos) (autorest.Authorizer, error) {
	secret := &v1.Secret{}
	err := cli.Get(ctx, types.NamespacedName{
		Name:      *azurechaos.Spec.SecretName,
		Namespace: azurechaos.Namespace,
	}, secret)
	if err != nil {
		return nil, err
	}

	clientCredentialConfig := auth.NewClientCredentialsConfig(
		string(secret.Data["client_id"]),
		string(secret.Data["client_secret"]),
		string(secret.Data["tenant_id"]))

	authorizer, err := clientCredentialConfig.Authorizer()
	if err != nil {
		return nil, err
	}

	return authorizer, nil
}

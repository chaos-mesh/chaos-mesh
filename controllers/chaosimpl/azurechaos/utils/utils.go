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
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// GetVMClient is used to get the azure VM Client
func GetVMClient(azurechaos *v1alpha1.AzureChaos) (*compute.VirtualMachinesClient, error) {
	authorizer, err := auth.NewAuthorizerFromEnvironment()

	if err != nil {
		return nil, err
	}

	vmClient := compute.NewVirtualMachinesClient(azurechaos.Spec.SubscriptionID)
	vmClient.Authorizer = authorizer

	return &vmClient, nil
}

// GetDiskClient is used to get the azure disk Client
func GetDiskClient(azurechaos *v1alpha1.AzureChaos) (*compute.DisksClient, error) {
	authorizer, err := auth.NewAuthorizerFromEnvironment()

	if err != nil {
		return nil, err
	}

	disksClient := compute.NewDisksClient(azurechaos.Spec.SubscriptionID)
	disksClient.Authorizer = authorizer

	return &disksClient, nil
}
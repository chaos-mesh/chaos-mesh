// Copyright 2023 Chaos Mesh Authors.
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
	"fmt"

	"github.com/apache/cloudstack-go/v2/cloudstack"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// GetCloudStackClient is used to get a CloudStack client.
func GetCloudStackClient(ctx context.Context, cli client.Client, cloudstackchaos *v1alpha1.CloudStackVMChaos) (*cloudstack.CloudStackClient, error) {
	apiConfig := cloudstackchaos.Spec.APIConfig

	var secret v1.Secret
	if err := cli.Get(ctx, types.NamespacedName{Namespace: cloudstackchaos.Namespace, Name: apiConfig.SecretName}, &secret); err != nil {
		return nil, fmt.Errorf("retrieving secret for cloudstack api client: %w", err)
	}

	apiKey, ok := secret.Data[apiConfig.APIKeyField]
	if !ok {
		return nil, fmt.Errorf("field %s not found in secret %s", apiConfig.APIKeyField, apiConfig.SecretName)
	}

	apiSecret, ok := secret.Data[apiConfig.APISecretField]
	if !ok {
		return nil, fmt.Errorf("field %s not found in secret %s", apiConfig.APIKeyField, apiConfig.SecretName)
	}

	return cloudstack.NewAsyncClient(
		cloudstackchaos.Spec.APIConfig.Address,
		string(apiKey),
		string(apiSecret),
		apiConfig.VerifySSL,
	), nil
}

func SelectorToListParams(s *v1alpha1.CloudStackVMChaosSelector) *cloudstack.ListVirtualMachinesParams {
	params := &cloudstack.ListVirtualMachinesParams{}

	if s.Account != nil {
		params.SetAccount(*s.Account)
	}

	if s.AffinityGroupID != nil {
		params.SetAffinitygroupid(*s.AffinityGroupID)
	}

	if s.DisplayVM {
		params.SetDisplayvm(s.DisplayVM)
	}

	if s.DomainID != nil {
		params.SetDomainid(*s.DomainID)
	}

	if s.GroupID != nil {
		params.SetGroupid(*s.GroupID)
	}

	if s.HostID != nil {
		params.SetHostid(*s.HostID)
	}

	if s.Hypervisor != nil {
		params.SetHypervisor(*s.Hypervisor)
	}

	if s.ID != nil {
		params.SetId(*s.ID)
	}

	if len(s.IDs) > 0 {
		params.SetIds(s.IDs)
	}

	if s.ISOID != nil {
		params.SetIsoid(*s.ISOID)
	}

	if s.IsRecursive {
		params.SetIsrecursive(s.IsRecursive)
	}

	if s.KeyPair != nil {
		params.SetKeypair(*s.KeyPair)
	}

	if s.Keyword != nil {
		params.SetKeyword(*s.Keyword)
	}

	if s.ListAll {
		params.SetListall(s.ListAll)
	}

	if s.Name != nil {
		params.SetName(*s.Name)
	}

	if s.NetworkID != nil {
		params.SetNetworkid(*s.NetworkID)
	}

	if s.ProjectID != nil {
		params.SetProjectid(*s.ProjectID)
	}

	if s.ServiceOffering != nil {
		params.SetServiceofferingid(*s.ServiceOffering)
	}

	if s.State != nil {
		params.SetState(*s.State)
	}

	if s.StorageID != nil {
		params.SetStorageid(*s.StorageID)
	}

	if len(s.Tags) > 0 {
		params.SetTags(s.Tags)
	}

	if s.TempalteID != nil {
		params.SetTemplateid(*s.TempalteID)
	}

	if s.UserID != nil {
		params.SetUserid(*s.UserID)
	}

	if s.VPCID != nil {
		params.SetVpcid(*s.VPCID)
	}

	if s.ZoneID != nil {
		params.SetZoneid(*s.ZoneID)
	}

	return params
}

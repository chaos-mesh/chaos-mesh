// Copyright 2026 Chaos Mesh Authors.
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

package selector_test

import (
	"encoding/json"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/aws"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/azure"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/gcp"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/internal/cloud"
)

func TestSelector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Selector Suite")
}

var _ = Describe("Cloud target selection", func() {
	DescribeTable("returns the selector JSON as the record ID",
		func(ctx SpecContext, obj v1alpha1.InnerObjectWithSelector, expectedID string) {
			dispatcher := selector.New(selector.Params{
				AWSSelector:   aws.New(),
				AzureSelector: azure.New(),
				GCPSelector:   gcp.New(),
			})
			targets, err := dispatcher.Select(ctx, obj.GetSelectorSpecs()["."])
			Expect(err).NotTo(HaveOccurred())
			Expect(targets).To(HaveLen(1))
			Expect(targets[0].Id()).To(Equal(expectedID))
		},
		Entry("AWS", &v1alpha1.AWSChaos{Spec: v1alpha1.AWSChaosSpec{
			AWSSelector: v1alpha1.AWSSelector{AWSRegion: "us-west-2", Ec2Instance: "i-123"},
		}}, `{"awsRegion":"us-west-2","ec2Instance":"i-123"}`),
		Entry("Azure", &v1alpha1.AzureChaos{Spec: v1alpha1.AzureChaosSpec{
			AzureSelector: v1alpha1.AzureSelector{SubscriptionID: "sub", ResourceGroupName: "rg", VMName: "vm", RemoteCluster: "remote"},
		}}, `{"subscriptionID":"sub","resourceGroupName":"rg","vmName":"vm","remoteCluster":"remote"}`),
		Entry("GCP", &v1alpha1.GCPChaos{Spec: v1alpha1.GCPChaosSpec{
			GCPSelector: v1alpha1.GCPSelector{Project: "project", Zone: "us-west1-a", Instance: "vm", DeviceNames: []string{"disk-a", "disk-b"}},
		}}, `{"project":"project","zone":"us-west1-a","instance":"vm","deviceNames":["disk-a","disk-b"]}`),
	)

	It("returns an encoding error and no targets for an unsupported selector", func() {
		targets, err := cloud.Select(make(chan int))

		Expect(targets).To(BeNil())
		Expect(err).To(HaveOccurred())
		var unsupported *json.UnsupportedTypeError
		Expect(errors.As(err, &unsupported)).To(BeTrue())
	})
})

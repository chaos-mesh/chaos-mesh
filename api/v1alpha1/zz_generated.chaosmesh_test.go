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

package v1alpha1

import (
	"reflect"
	"testing"

	"github.com/bxcodec/faker"
	. "github.com/onsi/gomega"
)

func TestAwsChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestAwsChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestAwsChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestAwsChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestAwsChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestAwsChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestAwsChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestDNSChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestDNSChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestDNSChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestDNSChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestDNSChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestDNSChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestDNSChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestGcpChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &GcpChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestGcpChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &GcpChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestGcpChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &GcpChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestGcpChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &GcpChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestGcpChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &GcpChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestGcpChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &GcpChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestGcpChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &GcpChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestHTTPChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestHTTPChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestHTTPChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestHTTPChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestHTTPChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestHTTPChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestHTTPChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestIOChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IOChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestIOChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IOChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestIOChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IOChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestIOChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IOChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestIOChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IOChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestIOChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &IOChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestIOChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IOChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestJVMChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestJVMChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestJVMChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestJVMChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestJVMChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestJVMChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestJVMChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestKernelChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestKernelChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestKernelChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestKernelChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestKernelChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestKernelChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestKernelChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestNetworkChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestNetworkChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestNetworkChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestNetworkChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestNetworkChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestNetworkChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestNetworkChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestPodChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestPodChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestPodChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestPodChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestPodChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestPodChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &PodChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestPodChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestStressChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestStressChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestStressChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestStressChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestStressChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestStressChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &StressChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestStressChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestTimeChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestTimeChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestTimeChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.Spec.GetDuration()
}

func TestTimeChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestTimeChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestTimeChaosGetSpecAndMetaString(t *testing.T) {
	g := NewGomegaWithT(t)
	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)
	g.Expect(err).To(BeNil())
	chaos.GetSpecAndMetaString()
}

func TestTimeChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func init() {
	faker.AddProvider("ioMethods", func(v reflect.Value) (interface{}, error) {
		return []IoMethod{LookUp}, nil
	})
}

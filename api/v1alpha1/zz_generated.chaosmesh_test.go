// Copyright 2020 Chaos Mesh Authors.
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
	"time"

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

	chaos.GetDuration()
}

func TestAwsChaosGetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextStart()
}

func TestAwsChaosSetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextStart(time.Now())
}

func TestAwsChaosGetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextRecover()
}

func TestAwsChaosSetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextRecover(time.Now())
}

func TestAwsChaosGetScheduler(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &AwsChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetScheduler()
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

	chaos.GetDuration()
}

func TestDNSChaosGetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextStart()
}

func TestDNSChaosSetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextStart(time.Now())
}

func TestDNSChaosGetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextRecover()
}

func TestDNSChaosSetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextRecover(time.Now())
}

func TestDNSChaosGetScheduler(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetScheduler()
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

func TestDNSChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &DNSChaosList{}
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

	chaos.GetDuration()
}

func TestHTTPChaosGetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextStart()
}

func TestHTTPChaosSetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextStart(time.Now())
}

func TestHTTPChaosGetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextRecover()
}

func TestHTTPChaosSetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextRecover(time.Now())
}

func TestHTTPChaosGetScheduler(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetScheduler()
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

func TestHTTPChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &HTTPChaosList{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.ListChaos()
}

func TestIoChaosIsDeleted(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsDeleted()
}

func TestIoChaosIsIsPaused(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.IsPaused()
}

func TestIoChaosGetDuration(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetDuration()
}

func TestIoChaosGetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextStart()
}

func TestIoChaosSetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextStart(time.Now())
}

func TestIoChaosGetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextRecover()
}

func TestIoChaosSetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextRecover(time.Now())
}

func TestIoChaosGetScheduler(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetScheduler()
}

func TestIoChaosGetChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetChaos()
}

func TestIoChaosGetStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetStatus()
}

func TestIoChaosListChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &IoChaosList{}
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

	chaos.GetDuration()
}

func TestJVMChaosGetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextStart()
}

func TestJVMChaosSetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextStart(time.Now())
}

func TestJVMChaosGetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextRecover()
}

func TestJVMChaosSetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextRecover(time.Now())
}

func TestJVMChaosGetScheduler(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &JVMChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetScheduler()
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

	chaos.GetDuration()
}

func TestKernelChaosGetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextStart()
}

func TestKernelChaosSetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextStart(time.Now())
}

func TestKernelChaosGetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextRecover()
}

func TestKernelChaosSetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextRecover(time.Now())
}

func TestKernelChaosGetScheduler(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &KernelChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetScheduler()
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

	chaos.GetDuration()
}

func TestNetworkChaosGetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextStart()
}

func TestNetworkChaosSetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextStart(time.Now())
}

func TestNetworkChaosGetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextRecover()
}

func TestNetworkChaosSetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextRecover(time.Now())
}

func TestNetworkChaosGetScheduler(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &NetworkChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetScheduler()
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

	chaos.GetDuration()
}

func TestPodChaosGetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextStart()
}

func TestPodChaosSetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextStart(time.Now())
}

func TestPodChaosGetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextRecover()
}

func TestPodChaosSetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextRecover(time.Now())
}

func TestPodChaosGetScheduler(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &PodChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetScheduler()
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

	chaos.GetDuration()
}

func TestStressChaosGetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextStart()
}

func TestStressChaosSetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextStart(time.Now())
}

func TestStressChaosGetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextRecover()
}

func TestStressChaosSetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextRecover(time.Now())
}

func TestStressChaosGetScheduler(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &StressChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetScheduler()
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

	chaos.GetDuration()
}

func TestTimeChaosGetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextStart()
}

func TestTimeChaosSetNextStart(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextStart(time.Now())
}

func TestTimeChaosGetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetNextRecover()
}

func TestTimeChaosSetNextRecover(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.SetNextRecover(time.Now())
}

func TestTimeChaosGetScheduler(t *testing.T) {
	g := NewGomegaWithT(t)

	chaos := &TimeChaos{}
	err := faker.FakeData(chaos)

	g.Expect(err).To(BeNil())

	chaos.GetScheduler()
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

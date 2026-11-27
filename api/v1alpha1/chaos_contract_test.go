// Copyright Chaos Mesh Authors.
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

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Chaos API contracts", func() {
	DescribeTable("preserves the experiment spec through Workflow and Schedule templates",
		func(ctx SpecContext, kind TemplateType, field, actionSpec string) {
			scheme := runtime.NewScheme()
			Expect(AddToScheme(scheme)).To(Succeed())
			obj, err := scheme.New(SchemeGroupVersion.WithKind(string(kind)))
			Expect(err).NotTo(HaveOccurred())
			expected, ok := obj.(InnerObject)
			Expect(ok).To(BeTrue())
			fixture := json.NewDecoder(strings.NewReader(actionSpec))
			fixture.DisallowUnknownFields()
			Expect(fixture.Decode(expected)).To(Succeed())
			Expect(json.Unmarshal([]byte(`{"spec":{"duration":"2m","remoteCluster":"edge"}}`), expected)).To(Succeed())

			originalTemplate := []byte(fmt.Sprintf(`{%q:{"duration":"1m"}}`, field))
			workflow := &EmbedChaos{}
			Expect(json.Unmarshal(originalTemplate, workflow)).To(Succeed())
			Expect(workflow.RestoreChaosSpec(expected)).To(Succeed())
			workflowChaos, err := workflow.SpawnNewObject(kind)
			Expect(err).NotTo(HaveOccurred())
			Expect(workflowChaos).To(Equal(expected))

			schedule := &ScheduleItem{}
			Expect(json.Unmarshal(originalTemplate, schedule)).To(Succeed())
			Expect(schedule.RestoreChaosSpec(expected)).To(Succeed())
			scheduleChaos, err := schedule.SpawnNewObject(ScheduleTemplateType(kind))
			Expect(err).NotTo(HaveOccurred())
			Expect(scheduleChaos).To(Equal(expected))

			list, err := workflow.SpawnNewList(kind)
			Expect(err).NotTo(HaveOccurred())
			Expect(json.Unmarshal([]byte(`{"items":[{"metadata":{"name":"first"}},{"metadata":{"name":"second"}}]}`), list)).To(Succeed())
			items := list.GetItems()
			Expect(items).To(HaveLen(2))
			Expect(items[0]).To(BeAssignableToTypeOf(expected))
			Expect(items[1]).To(BeAssignableToTypeOf(expected))
			Expect(items[0].GetName()).To(Equal("first"))
			Expect(items[1].GetName()).To(Equal("second"))

			By("admitting metadata updates while protecting the experiment spec")
			expected.SetName("experiment")
			Expect(expected.Default(ctx, expected)).To(Succeed())
			Expect(expected.ValidateCreate(ctx, expected)).Error().NotTo(HaveOccurred())
			updated := expected.DeepCopyObject().(InnerObject)
			updated.SetAnnotations(map[string]string{"test": "updated"})
			Expect(expected.ValidateUpdate(ctx, expected, updated)).Error().NotTo(HaveOccurred())
			Expect(expected.ValidateDelete(ctx, expected)).Error().NotTo(HaveOccurred())

			Expect(json.Unmarshal([]byte(`{"spec":{"duration":"3m"}}`), updated)).To(Succeed())
			Expect(expected.ValidateUpdate(ctx, expected, updated)).Error().To(MatchError(ErrCanNotUpdateChaos))

			unrelated := &corev1.Pod{}
			Expect(expected.ValidateCreate(ctx, unrelated)).Error().To(MatchError(ContainSubstring("expected type")))
			Expect(expected.ValidateUpdate(ctx, unrelated, expected)).Error().To(MatchError(ContainSubstring("expected type")))
			Expect(expected.ValidateUpdate(ctx, expected, unrelated)).Error().To(MatchError(ContainSubstring("expected type")))
			Expect(expected.ValidateDelete(ctx, unrelated)).Error().To(MatchError(ContainSubstring("expected type")))

			By("measuring the duration from the creation timestamp")
			start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
			expected.SetCreationTimestamp(metav1.NewTime(start))
			checkDuration := func(elapsed, wantRemaining time.Duration, wantExpired bool) {
				GinkgoHelper()
				expired, remaining, err := expected.DurationExceeded(start.Add(elapsed))
				Expect(err).NotTo(HaveOccurred())
				Expect(expired).To(Equal(wantExpired))
				Expect(remaining).To(Equal(wantRemaining))
			}
			checkDuration(time.Minute, time.Minute, false)
			checkDuration(2*time.Minute, 0, false)
			checkDuration(3*time.Minute, 0, true)
			Expect(json.Unmarshal([]byte(`{"spec":{"duration":null}}`), expected)).To(Succeed())
			checkDuration(3*time.Minute, 0, false)
			Expect(json.Unmarshal([]byte(`{"spec":{"duration":"invalid"}}`), expected)).To(Succeed())
			expired, remaining, err := expected.DurationExceeded(start)
			Expect(err).To(HaveOccurred())
			Expect(expired).To(BeFalse())
			Expect(remaining).To(BeZero())
		},
		Entry("AWSChaos", TypeAWSChaos, "awsChaos", `{"spec":{"action":"ec2-stop","ec2Instance":"i-example","awsRegion":"test"}}`),
		Entry("AzureChaos", TypeAzureChaos, "azureChaos", `{"spec":{"action":"vm-stop","vmName":"example","resourceGroupName":"test","subscriptionID":"test"}}`),
		Entry("BlockChaos", TypeBlockChaos, "blockChaos", `{"spec":{"action":"delay","mode":"one","volumeName":"test","delay":{"latency":"10ms","correlation":"0"}}}`),
		Entry("DNSChaos", TypeDNSChaos, "dnsChaos", `{"spec":{"action":"error","mode":"one","patterns":["example.test"]}}`),
		Entry("GCPChaos", TypeGCPChaos, "gcpChaos", `{"spec":{"action":"node-stop","instance":"example","project":"test","zone":"test"}}`),
		Entry("HTTPChaos", TypeHTTPChaos, "httpChaos", `{"spec":{"mode":"one","target":"Request","port":8080,"method":"GET","delay":"10ms"}}`),
		Entry("IOChaos", TypeIOChaos, "ioChaos", `{"spec":{"action":"latency","mode":"one","delay":"1ms","percent":50,"volumePath":"/test"}}`),
		Entry("JVMChaos", TypeJVMChaos, "jvmChaos", `{"spec":{"action":"gc","mode":"one","port":9091}}`),
		Entry("KernelChaos", TypeKernelChaos, "kernelChaos", `{"spec":{"mode":"one","failKernRequest":{"failtype":1}}}`),
		Entry("NetworkChaos", TypeNetworkChaos, "networkChaos", `{"spec":{"action":"delay","mode":"one","delay":{"latency":"10ms","correlation":"0"}}}`),
		Entry("PhysicalMachineChaos", TypePhysicalMachineChaos, "physicalmachineChaos", `{"spec":{"action":"network-delay","mode":"one","address":["http://example.test:31767"],"network-delay":{"device":"eth0","latency":"1ms"}}}`),
		Entry("PodChaos", TypePodChaos, "podChaos", `{"spec":{"action":"pod-failure","mode":"one"}}`),
		Entry("StressChaos", TypeStressChaos, "stressChaos", `{"spec":{"mode":"one","stressors":{"cpu":{"workers":1,"load":10}}}}`),
		Entry("TimeChaos", TypeTimeChaos, "timeChaos", `{"spec":{"mode":"one","timeOffset":"1s","clockIds":["CLOCK_REALTIME"]}}`),
	)

	It("schedules a Workflow with its templates intact", func() {
		original := &Workflow{Spec: WorkflowSpec{
			Entry:     "wait",
			Templates: []Template{{Name: "wait", Type: TypeSuspend, Deadline: new("5s")}},
		}}
		item := &ScheduleItem{Workflow: &WorkflowSpec{}}
		Expect(item.RestoreChaosSpec(original)).To(Succeed())
		spawned, err := item.SpawnNewObject(ScheduleTypeWorkflow)
		Expect(err).NotTo(HaveOccurred())
		Expect(spawned).To(Equal(original))
	})

	It("reports missing experiment specs for every supported Schedule kind", func() {
		item := &ScheduleItem{}
		for _, kind := range allScheduleTemplateType {
			Expect(item.SpawnNewObject(kind)).Error().To(MatchError(ContainSubstring(fmt.Sprintf("missing %s spec", kind))))
		}
	})

	It("rejects kinds that cannot be embedded as experiments", func() {
		workflow := &EmbedChaos{}
		schedule := &ScheduleItem{}
		Expect(workflow.SpawnNewObject(TypeTask)).Error().To(MatchError(ContainSubstring("unknown template type Task")))
		Expect(workflow.SpawnNewList(TypeTask)).Error().To(MatchError(ContainSubstring("unknown template type Task")))
		Expect(schedule.SpawnNewObject(ScheduleTemplateType(TypeTask))).Error().To(MatchError(ContainSubstring("unknown template type Task")))
		Expect(workflow.RestoreChaosSpec(&corev1.Pod{})).To(MatchError(ContainSubstring("unknown chaos")))
		Expect(schedule.RestoreChaosSpec(&corev1.Pod{})).To(MatchError(ContainSubstring("unknown chaos")))
	})
})

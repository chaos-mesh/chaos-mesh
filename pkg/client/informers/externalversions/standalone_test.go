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

package externalversions_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	informers "github.com/chaos-mesh/chaos-mesh/pkg/client/informers/externalversions/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/client/informers/externalversions/internalinterfaces"
	"github.com/chaos-mesh/chaos-mesh/pkg/client/versioned"
	clientfake "github.com/chaos-mesh/chaos-mesh/pkg/client/versioned/fake"
)

type indexedChaos interface {
	cache.Object
	metav1.Object
	runtime.Object
}

type informerContract func(SpecContext, string)

func standaloneInformer[T indexedChaos, I cache.TypedSharedIndexInformer[T]](
	prototype T,
	newUntyped func(versioned.Interface, string, time.Duration, cache.Indexers) cache.SharedIndexInformer,
	newFiltered func(versioned.Interface, string, time.Duration, cache.Indexers, internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer,
	withOptions func(versioned.Interface, string, internalinterfaces.InformerOptions) cache.SharedIndexInformer,
	newTyped func(versioned.Interface, string, time.Duration, cache.TypedIndexers[T]) I,
	newTypedFiltered func(versioned.Interface, string, time.Duration, cache.TypedIndexers[T], internalinterfaces.TweakListOptionsFunc) I,
	asTyped func(cache.SharedIndexInformer) I,
) informerContract {
	return func(ctx SpecContext, resource string) {
		const ns = "watched"
		const ownerIndex = "team"
		client := clientfake.NewSimpleClientset()
		gvr := v1alpha1.SchemeGroupVersion.WithResource(resource)
		object := prototype.DeepCopyObject().(T)
		object.SetName("fixture")
		object.SetNamespace(ns)
		object.SetLabels(map[string]string{ownerIndex: "ops"})
		Expect(client.Tracker().Create(gvr, object, ns)).To(Succeed())

		foreign := object.DeepCopyObject().(T)
		foreign.SetNamespace("foreign")
		Expect(client.Tracker().Create(gvr, foreign, foreign.GetNamespace())).To(Succeed())
		excluded := object.DeepCopyObject().(T)
		excluded.SetName("other-team")
		excluded.SetLabels(map[string]string{ownerIndex: "dev"})
		Expect(client.Tracker().Create(gvr, excluded, ns)).To(Succeed())

		typedIndexes := cache.TypedIndexers[T]{ownerIndex: func(obj T) ([]string, error) {
			return []string{obj.GetLabels()[ownerIndex]}, nil
		}}
		indexes := cache.TypedIndexersToIndexers(typedIndexes)
		filter := func(options *metav1.ListOptions) { options.LabelSelector = "team=ops" }

		for _, independent := range []cache.SharedIndexInformer{
			newUntyped(client, ns, 0, indexes),
			newFiltered(client, ns, 0, indexes, filter),
			withOptions(client, ns, internalinterfaces.InformerOptions{Indexers: indexes, TweakListOptions: filter}),
			newTyped(client, ns, 0, typedIndexes),
		} {
			Expect(independent.GetStore().List()).To(BeEmpty())
			Expect(independent.GetStore().Add(object)).To(Succeed())
			Expect(asTyped(independent).GetTypedIndexer().ByTypedIndex(ownerIndex, "ops")).To(ConsistOf(Equal(object)))
		}

		untyped := cache.NewSharedIndexInformer(nil, prototype, 0, indexes)
		Expect(untyped.GetStore().Add(object)).To(Succeed())
		Expect(asTyped(untyped).GetTypedIndexer().ByTypedIndex(ownerIndex, "ops")).To(ConsistOf(Equal(object)))

		typed := newTypedFiltered(client, ns, 0, typedIndexes, filter)
		Expect(asTyped(typed)).To(BeIdenticalTo(typed))
		added := make(chan T, 1)
		_, err := typed.AddTypedEventHandler(cache.TypedResourceEventHandlerFuncs[T]{
			AddFunc: func(obj T) {
				select {
				case added <- obj:
				default:
				}
			},
		})
		Expect(err).NotTo(HaveOccurred())
		runCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		done := make(chan struct{})
		go func() {
			defer close(done)
			typed.RunWithContext(runCtx)
		}()
		DeferCleanup(func() {
			cancel()
			Eventually(done).Should(BeClosed())
		})

		var found T
		Eventually(added).WithContext(runCtx).Should(Receive(&found))
		Expect(found.GetNamespace()).To(Equal(ns))
		Expect(found.GetName()).To(Equal(object.GetName()))
		Expect(cache.WaitForCacheSync(runCtx.Done(), typed.HasSynced)).To(BeTrue())
		Expect(typed.GetStore().List()).To(HaveLen(1))
		Expect(typed.GetTypedIndexer().ByTypedIndex(ownerIndex, "ops")).To(ConsistOf(Equal(found)))
	}
}

var _ = Describe("Standalone informer constructors", func() {
	DescribeTable("preserves typed indexing, filter options and independent caches",
		func(ctx SpecContext, resource string, verify informerContract) {
			verify(ctx, resource)
		},
		Entry("AWSChaos", "awschaos", standaloneInformer(&v1alpha1.AWSChaos{},
			informers.NewAWSChaosInformer, informers.NewFilteredAWSChaosInformer, informers.NewAWSChaosInformerWithOptions,
			informers.NewTypedAWSChaosInformer, informers.NewTypedFilteredAWSChaosInformer, informers.ToAWSChaosIndexInformer)),

		Entry("AzureChaos", "azurechaos", standaloneInformer(&v1alpha1.AzureChaos{},
			informers.NewAzureChaosInformer, informers.NewFilteredAzureChaosInformer, informers.NewAzureChaosInformerWithOptions,
			informers.NewTypedAzureChaosInformer, informers.NewTypedFilteredAzureChaosInformer, informers.ToAzureChaosIndexInformer)),

		Entry("BlockChaos", "blockchaos", standaloneInformer(&v1alpha1.BlockChaos{},
			informers.NewBlockChaosInformer, informers.NewFilteredBlockChaosInformer, informers.NewBlockChaosInformerWithOptions,
			informers.NewTypedBlockChaosInformer, informers.NewTypedFilteredBlockChaosInformer, informers.ToBlockChaosIndexInformer)),

		Entry("GCPChaos", "gcpchaos", standaloneInformer(&v1alpha1.GCPChaos{},
			informers.NewGCPChaosInformer, informers.NewFilteredGCPChaosInformer, informers.NewGCPChaosInformerWithOptions,
			informers.NewTypedGCPChaosInformer, informers.NewTypedFilteredGCPChaosInformer, informers.ToGCPChaosIndexInformer)),

		Entry("HTTPChaos", "httpchaos", standaloneInformer(&v1alpha1.HTTPChaos{},
			informers.NewHTTPChaosInformer, informers.NewFilteredHTTPChaosInformer, informers.NewHTTPChaosInformerWithOptions,
			informers.NewTypedHTTPChaosInformer, informers.NewTypedFilteredHTTPChaosInformer, informers.ToHTTPChaosIndexInformer)),

		Entry("IOChaos", "iochaos", standaloneInformer(&v1alpha1.IOChaos{},
			informers.NewIOChaosInformer, informers.NewFilteredIOChaosInformer, informers.NewIOChaosInformerWithOptions,
			informers.NewTypedIOChaosInformer, informers.NewTypedFilteredIOChaosInformer, informers.ToIOChaosIndexInformer)),

		Entry("JVMChaos", "jvmchaos", standaloneInformer(&v1alpha1.JVMChaos{},
			informers.NewJVMChaosInformer, informers.NewFilteredJVMChaosInformer, informers.NewJVMChaosInformerWithOptions,
			informers.NewTypedJVMChaosInformer, informers.NewTypedFilteredJVMChaosInformer, informers.ToJVMChaosIndexInformer)),

		Entry("KernelChaos", "kernelchaos", standaloneInformer(&v1alpha1.KernelChaos{},
			informers.NewKernelChaosInformer, informers.NewFilteredKernelChaosInformer, informers.NewKernelChaosInformerWithOptions,
			informers.NewTypedKernelChaosInformer, informers.NewTypedFilteredKernelChaosInformer, informers.ToKernelChaosIndexInformer)),

		Entry("NetworkChaos", "networkchaos", standaloneInformer(&v1alpha1.NetworkChaos{},
			informers.NewNetworkChaosInformer, informers.NewFilteredNetworkChaosInformer, informers.NewNetworkChaosInformerWithOptions,
			informers.NewTypedNetworkChaosInformer, informers.NewTypedFilteredNetworkChaosInformer, informers.ToNetworkChaosIndexInformer)),

		Entry("PodChaos", "podchaos", standaloneInformer(&v1alpha1.PodChaos{},
			informers.NewPodChaosInformer, informers.NewFilteredPodChaosInformer, informers.NewPodChaosInformerWithOptions,
			informers.NewTypedPodChaosInformer, informers.NewTypedFilteredPodChaosInformer, informers.ToPodChaosIndexInformer)),

		Entry("PodHttpChaos", "podhttpchaos", standaloneInformer(&v1alpha1.PodHttpChaos{},
			informers.NewPodHttpChaosInformer, informers.NewFilteredPodHttpChaosInformer, informers.NewPodHttpChaosInformerWithOptions,
			informers.NewTypedPodHttpChaosInformer, informers.NewTypedFilteredPodHttpChaosInformer, informers.ToPodHttpChaosIndexInformer)),

		Entry("PodIOChaos", "podiochaos", standaloneInformer(&v1alpha1.PodIOChaos{},
			informers.NewPodIOChaosInformer, informers.NewFilteredPodIOChaosInformer, informers.NewPodIOChaosInformerWithOptions,
			informers.NewTypedPodIOChaosInformer, informers.NewTypedFilteredPodIOChaosInformer, informers.ToPodIOChaosIndexInformer)),

		Entry("PodNetworkChaos", "podnetworkchaos", standaloneInformer(&v1alpha1.PodNetworkChaos{},
			informers.NewPodNetworkChaosInformer, informers.NewFilteredPodNetworkChaosInformer, informers.NewPodNetworkChaosInformerWithOptions,
			informers.NewTypedPodNetworkChaosInformer, informers.NewTypedFilteredPodNetworkChaosInformer, informers.ToPodNetworkChaosIndexInformer)),

		Entry("StressChaos", "stresschaos", standaloneInformer(&v1alpha1.StressChaos{},
			informers.NewStressChaosInformer, informers.NewFilteredStressChaosInformer, informers.NewStressChaosInformerWithOptions,
			informers.NewTypedStressChaosInformer, informers.NewTypedFilteredStressChaosInformer, informers.ToStressChaosIndexInformer)),

		Entry("TimeChaos", "timechaos", standaloneInformer(&v1alpha1.TimeChaos{},
			informers.NewTimeChaosInformer, informers.NewFilteredTimeChaosInformer, informers.NewTimeChaosInformerWithOptions,
			informers.NewTypedTimeChaosInformer, informers.NewTypedFilteredTimeChaosInformer, informers.ToTimeChaosIndexInformer)),
	)
})

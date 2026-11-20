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
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/client/informers/externalversions"
	informers "github.com/chaos-mesh/chaos-mesh/pkg/client/informers/externalversions/api/v1alpha1"
	clientfake "github.com/chaos-mesh/chaos-mesh/pkg/client/versioned/fake"
)

func TestSharedInformers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaos Mesh shared informers")
}

var _ = Describe("Chaos experiment informer factory", func() {
	It("routes each experiment resource to its own namespaced cache and delivers watch updates", func(ctx SpecContext) {
		const namespace = "watched"
		const selection = "team=platform"
		selected := labels.SelectorFromSet(labels.Set{"team": "platform"})
		examples := []struct {
			resource string
			object   runtime.Object
		}{
			{"awschaos", &v1alpha1.AWSChaos{}},
			{"azurechaos", &v1alpha1.AzureChaos{}},
			{"blockchaos", &v1alpha1.BlockChaos{}},
			{"gcpchaos", &v1alpha1.GCPChaos{}},
			{"httpchaos", &v1alpha1.HTTPChaos{}},
			{"iochaos", &v1alpha1.IOChaos{}},
			{"jvmchaos", &v1alpha1.JVMChaos{}},
			{"kernelchaos", &v1alpha1.KernelChaos{}},
			{"networkchaos", &v1alpha1.NetworkChaos{}},
			{"podchaos", &v1alpha1.PodChaos{}},
			{"podhttpchaos", &v1alpha1.PodHttpChaos{}},
			{"podiochaos", &v1alpha1.PodIOChaos{}},
			{"podnetworkchaos", &v1alpha1.PodNetworkChaos{}},
			{"stresschaos", &v1alpha1.StressChaos{}},
			{"timechaos", &v1alpha1.TimeChaos{}},
		}
		client := clientfake.NewSimpleClientset()
		for _, example := range examples {
			gvr := v1alpha1.SchemeGroupVersion.WithResource(example.resource)
			accessor, err := meta.Accessor(example.object)
			Expect(err).NotTo(HaveOccurred())
			accessor.SetName("visible")
			accessor.SetNamespace(namespace)
			accessor.SetLabels(map[string]string{"team": "platform"})
			Expect(client.Tracker().Create(gvr, example.object, namespace)).To(Succeed())

			excluded := example.object.DeepCopyObject()
			excludedMeta, err := meta.Accessor(excluded)
			Expect(err).NotTo(HaveOccurred())
			excludedMeta.SetName("different-team")
			excludedMeta.SetLabels(map[string]string{"team": "other"})
			Expect(client.Tracker().Create(gvr, excluded, namespace)).To(Succeed())

			otherNamespace := example.object.DeepCopyObject()
			otherMeta, err := meta.Accessor(otherNamespace)
			Expect(err).NotTo(HaveOccurred())
			otherMeta.SetNamespace("excluded")
			Expect(client.Tracker().Create(gvr, otherNamespace, "excluded")).To(Succeed())
		}
		watches := make(chan clienttesting.WatchAction, len(examples))
		client.PrependWatchReactor("*", func(action clienttesting.Action) (bool, watch.Interface, error) {
			request := action.(clienttesting.WatchAction)
			stream, err := client.Tracker().Watch(request.GetResource(), request.GetNamespace())
			if err == nil {
				select {
				case watches <- request:
				default:
				}
			}
			return true, stream, err
		})
		name, err := cache.NewInformerName("chaos-mesh-shared-factory-test")
		Expect(err).NotTo(HaveOccurred())
		factory := externalversions.NewSharedInformerFactoryWithOptions(client, 0,
			externalversions.WithNamespace(namespace),
			externalversions.WithInformerName(name),
			externalversions.WithCustomResyncConfig(map[metav1.Object]time.Duration{&v1alpha1.PodChaos{}: time.Hour}),
			externalversions.WithTweakListOptions(func(opts *metav1.ListOptions) { opts.LabelSelector = selection }),
			externalversions.WithTransform(func(obj any) (any, error) {
				accessor, err := meta.Accessor(obj)
				if err != nil {
					return nil, err
				}
				accessor.SetManagedFields(nil)
				return obj, nil
			}),
		)
		runCtx, stop := context.WithCancel(ctx)
		DeferCleanup(func() {
			stop()
			factory.Shutdown()
		})

		var caches []externalversions.GenericInformer
		for _, example := range examples {
			gvr := v1alpha1.SchemeGroupVersion.WithResource(example.resource)
			informer, err := factory.ForResource(gvr)
			Expect(err).NotTo(HaveOccurred(), "%s", gvr)
			caches = append(caches, informer)
		}
		_, err = factory.ForResource(v1alpha1.SchemeGroupVersion.WithResource("unknown"))
		Expect(err).To(MatchError(ContainSubstring("no informer found")))

		factory.StartWithContext(runCtx)
		factory.StartWithContext(runCtx)
		synced := factory.WaitForCacheSyncWithContext(runCtx)
		Expect(synced.Err).NotTo(HaveOccurred())
		Expect(synced.Synced).To(HaveLen(len(examples)))
		for i, example := range examples {
			Expect(synced.Synced[reflect.TypeOf(example.object)]).To(BeTrue(), "%s", example.resource)
			lister := caches[i].Lister().ByNamespace(namespace)
			items, err := lister.List(labels.Everything())
			Expect(err).NotTo(HaveOccurred())
			Expect(items).To(HaveLen(1), "%s", example.resource)
			Expect(items[0]).To(BeAssignableToTypeOf(example.object))
			_, err = lister.Get("different-team")
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		}

		watching := map[string]bool{}
		for range examples {
			var action clienttesting.WatchAction
			Eventually(watches).WithContext(runCtx).WithTimeout(5 * time.Second).Should(Receive(&action))
			Expect(action.GetNamespace()).To(Equal(namespace))
			Expect(action.GetWatchRestrictions().Labels.String()).To(Equal(selection))
			watching[action.GetResource().Resource] = true
		}
		Expect(watching).To(HaveLen(len(examples)))

		for i, example := range examples {
			gvr := v1alpha1.SchemeGroupVersion.WithResource(example.resource)
			update := example.object.DeepCopyObject()
			md, err := meta.Accessor(update)
			Expect(err).NotTo(HaveOccurred())
			md.SetAnnotations(map[string]string{"revision": "watch"})
			md.SetManagedFields([]metav1.ManagedFieldsEntry{{Manager: "fixture"}})
			Expect(client.Tracker().Update(gvr, update, namespace)).To(Succeed())
			lister := caches[i].Lister().ByNamespace(namespace)
			Eventually(func(g Gomega) {
				items, err := lister.List(selected)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(items).To(HaveLen(1))
				md, err := meta.Accessor(items[0])
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(md.GetAnnotations()).To(HaveKeyWithValue("revision", "watch"))
				g.Expect(md.GetManagedFields()).To(BeEmpty())
			}).WithContext(runCtx).WithTimeout(5 * time.Second).Should(Succeed())
		}

		typed := factory.Api().V1alpha1().Podchaos()
		item, err := typed.Lister().Podchaos(namespace).Get("visible")
		Expect(err).NotTo(HaveOccurred())
		Expect(item.Annotations).To(HaveKeyWithValue("revision", "watch"))
		allPods, err := typed.Lister().List(selected)
		Expect(err).NotTo(HaveOccurred())
		Expect(allPods).To(HaveLen(1))
	}, SpecTimeout(20*time.Second))

	It("reports cancellation before cache sync and does not start new watches after shutdown", func() {
		client := clientfake.NewSimpleClientset()
		factory := externalversions.NewSharedInformerFactory(client, 0)
		pods := factory.Api().V1alpha1().Podchaos()
		typed := informers.ToTypedPodChaosInformer(pods)
		Expect(typed.TypedInformer()).NotTo(BeNil())

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		factory.Start(ctx.Done())
		synced := factory.WaitForCacheSyncWithContext(ctx)
		Expect(synced.Err).To(MatchError(context.Canceled))
		Expect(synced.Synced).To(HaveKeyWithValue(reflect.TypeFor[*v1alpha1.PodChaos](), false))
		factory.Shutdown()
		Expect(factory.Api().V1alpha1().Timechaos().Informer().HasSynced()).To(BeFalse())
		factory.StartWithContext(ctx)
		Expect(factory.WaitForCacheSync(ctx.Done())).To(HaveLen(1))
	})
})

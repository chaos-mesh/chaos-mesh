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

package versioned_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/client/versioned"
	typed "github.com/chaos-mesh/chaos-mesh/pkg/client/versioned/typed/api/v1alpha1"
)

const clientNamespace = "experiment-space"

func TestClientset(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chaos Mesh typed REST clients")
}

type resourceAccess struct {
	get  func(context.Context, string) (runtime.Object, error)
	list func(context.Context, metav1.ListOptions) (runtime.Object, error)
}

func bindResource[T runtime.Object, L runtime.Object](resource interface {
	Get(context.Context, string, metav1.GetOptions) (T, error)
	List(context.Context, metav1.ListOptions) (L, error)
}) resourceAccess {
	return resourceAccess{
		get: func(ctx context.Context, name string) (runtime.Object, error) {
			return resource.Get(ctx, name, metav1.GetOptions{})
		},
		list: func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
			return resource.List(ctx, options)
		},
	}
}

var _ = Describe("Chaos Mesh typed clients", func() {
	DescribeTable("addresses and decodes the namespaced resource",
		func(ctx SpecContext, resource, kind string, bind func(typed.ApiV1alpha1Interface) resourceAccess) {
			gv := v1alpha1.SchemeGroupVersion
			prefix := "/apis/" + gv.String() + "/namespaces/" + clientNamespace + "/" + resource
			item := map[string]any{
				"apiVersion": gv.String(), "kind": kind,
				"metadata": map[string]string{"name": "found", "namespace": clientNamespace, "resourceVersion": "51"},
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer fixture" || r.UserAgent() == "" {
					http.Error(w, "incorrect request", http.StatusBadRequest)
					return
				}
				var body any
				switch r.URL.Path {
				case prefix:
					if r.URL.Query().Get("labelSelector") != "team=ops" || r.URL.Query().Get("limit") != "2" {
						http.Error(w, "incorrect list options", http.StatusBadRequest)
						return
					}
					body = map[string]any{
						"apiVersion": gv.String(), "kind": kind + "List", "items": []any{item},
						"metadata": map[string]string{"resourceVersion": "52", "continue": "next"},
					}
				case prefix + "/found":
					body = item
				case prefix + "/gone":
					w.WriteHeader(http.StatusNotFound)
					body = map[string]any{"apiVersion": "v1", "kind": "Status", "status": "Failure", "reason": "NotFound", "code": 404}
				default:
					http.NotFound(w, r)
					return
				}
				_ = json.NewEncoder(w).Encode(body)
			}))
			DeferCleanup(server.Close)

			config := &rest.Config{Host: server.URL, BearerToken: "fixture", QPS: 100, Burst: 5}
			clients, err := versioned.NewForConfig(config)
			Expect(err).NotTo(HaveOccurred())
			Expect(config.RateLimiter).To(BeNil())
			Expect(config.UserAgent).To(BeEmpty())

			access := bind(clients.ApiV1alpha1())
			found, err := access.get(ctx, "found")
			Expect(err).NotTo(HaveOccurred())
			md, err := meta.Accessor(found)
			Expect(err).NotTo(HaveOccurred())
			Expect(md.GetName()).To(Equal("found"))
			Expect(md.GetNamespace()).To(Equal(clientNamespace))
			Expect(md.GetResourceVersion()).To(Equal("51"))

			page, err := access.list(ctx, metav1.ListOptions{LabelSelector: "team=ops", Limit: 2})
			Expect(err).NotTo(HaveOccurred())
			list, err := meta.ListAccessor(page)
			Expect(err).NotTo(HaveOccurred())
			Expect(list.GetResourceVersion()).To(Equal("52"))
			Expect(list.GetContinue()).To(Equal("next"))
			items, err := meta.ExtractList(page)
			Expect(err).NotTo(HaveOccurred())
			Expect(items).To(HaveLen(1))
			Expect(items[0]).To(BeAssignableToTypeOf(found))
			itemMeta, err := meta.Accessor(items[0])
			Expect(err).NotTo(HaveOccurred())
			Expect(itemMeta.GetName()).To(Equal(md.GetName()))
			Expect(itemMeta.GetNamespace()).To(Equal(md.GetNamespace()))
			Expect(itemMeta.GetResourceVersion()).To(Equal(md.GetResourceVersion()))

			_, err = access.get(ctx, "gone")
			Expect(apierrors.IsNotFound(err)).To(BeTrue(), "expected Kubernetes NotFound, got %v", err)
		},
		Entry("AWSChaos", "awschaos", "AWSChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Awschaos(clientNamespace))
		}),
		Entry("AzureChaos", "azurechaos", "AzureChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Azurechaos(clientNamespace))
		}),
		Entry("BlockChaos", "blockchaos", "BlockChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Blockchaos(clientNamespace))
		}),
		Entry("GCPChaos", "gcpchaos", "GCPChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Gcpchaos(clientNamespace))
		}),
		Entry("HTTPChaos", "httpchaos", "HTTPChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Httpchaos(clientNamespace))
		}),
		Entry("IOChaos", "iochaos", "IOChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Iochaos(clientNamespace))
		}),
		Entry("JVMChaos", "jvmchaos", "JVMChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Jvmchaos(clientNamespace))
		}),
		Entry("KernelChaos", "kernelchaos", "KernelChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Kernelchaos(clientNamespace))
		}),
		Entry("NetworkChaos", "networkchaos", "NetworkChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Networkchaos(clientNamespace))
		}),
		Entry("PodChaos", "podchaos", "PodChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Podchaos(clientNamespace))
		}),
		Entry("PodHttpChaos", "podhttpchaos", "PodHttpChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Podhttpchaos(clientNamespace))
		}),
		Entry("PodIOChaos", "podiochaos", "PodIOChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Podiochaos(clientNamespace))
		}),
		Entry("PodNetworkChaos", "podnetworkchaos", "PodNetworkChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Podnetworkchaos(clientNamespace))
		}),
		Entry("StressChaos", "stresschaos", "StressChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Stresschaos(clientNamespace))
		}),
		Entry("TimeChaos", "timechaos", "TimeChaos", func(api typed.ApiV1alpha1Interface) resourceAccess {
			return bindResource(api.Timechaos(clientNamespace))
		}),
	)

	It("uses a supplied transport for discovery and allows the REST client to be reused", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/version" || r.Header.Get("X-From-Transport") != "test" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"major":"1","minor":"36","gitVersion":"v1.36.0"}`))
		}))
		DeferCleanup(server.Close)

		config := &rest.Config{Host: server.URL}
		restClient := typed.NewForConfigOrDie(config).RESTClient()
		Expect(versioned.New(restClient).ApiV1alpha1().RESTClient()).To(BeIdenticalTo(restClient))
		Expect(versioned.NewForConfigOrDie(config).Discovery()).NotTo(BeNil())
		clients, err := versioned.NewForConfigAndClient(config, &http.Client{
			Transport: discoveryTransport{base: server.Client().Transport},
		})
		Expect(err).NotTo(HaveOccurred())
		info, err := clients.Discovery().ServerVersion()
		Expect(err).NotTo(HaveOccurred())
		Expect(info.GitVersion).To(Equal("v1.36.0"))
	})

	It("reports invalid rate limits, hosts and TLS settings to the caller", func() {
		badBurst := &rest.Config{Host: "http://api.test", QPS: 1}
		Expect(versioned.NewForConfig(badBurst)).Error().To(MatchError(ContainSubstring("burst")))
		Expect(func() { versioned.NewForConfigOrDie(badBurst) }).To(Panic())

		badHost := &rest.Config{Host: "://invalid"}
		Expect(versioned.NewForConfigAndClient(badHost, http.DefaultClient)).Error().To(HaveOccurred())
		Expect(func() { typed.NewForConfigOrDie(badHost) }).To(Panic())

		badTLS := &rest.Config{Host: "https://api.test", TLSClientConfig: rest.TLSClientConfig{CAData: []byte("not a certificate")}}
		Expect(versioned.NewForConfig(badTLS)).Error().To(HaveOccurred())
		Expect(typed.NewForConfig(badTLS)).Error().To(HaveOccurred())
	})
})

type discoveryTransport struct {
	base http.RoundTripper
}

func (t discoveryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("X-From-Transport", "test")
	return t.base.RoundTrip(req)
}

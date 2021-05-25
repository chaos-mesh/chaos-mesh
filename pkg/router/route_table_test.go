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

package router

import (
	"context"
	"math/rand"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

func TestRouter(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Router Suit",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	rand.Seed(GinkgoRandomSeed())

	close(done)
})

type testEndpoint struct{}

func (e *testEndpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	return nil
}

func (e *testEndpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	return nil
}

func (e *testEndpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

var _ = Describe("Router", func() {
	It("should register successfully", func() {
		Register("hello", &v1alpha1.NetworkChaos{}, func(obj runtime.Object) bool {
			return true
		}, func(ctx ctx.Context) end.Endpoint {
			return &testEndpoint{}
		})

		Expect(len(routeTable)).To(Equal(1))

		typ := reflect.TypeOf(&v1alpha1.NetworkChaos{})
		Expect(routeTable[typ]).NotTo(BeNil())
	})
})

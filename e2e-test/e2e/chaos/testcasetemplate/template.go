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

package testcasetemplate

import (
	. "github.com/onsi/ginkgo"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestcaseTemplate(
	ns string,
	kubeCli kubernetes.Interface,
	cli client.Client,
	// any other parameters that you need fetch from context
) {
	// describe test steps with By() statement
	// here are some examples.
	By("preparing experiment pods")
	// some logic to create pod which will be injected chaos
	By("create pod failure chaos CRD objects")
	// create chaos CRD
	By("waiting for assertion some pod fall into failure")
	// assert that chaos is effective
	By("delete pod failure chaos CRD objects")
	// delete chaos CRD
	By("waiting for assertion recovering")
	// assert that chaos has gone
}

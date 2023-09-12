// Copyright 2021 Chaos Mesh Authors.
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

package e2e

import (
	"context" // load pprof
	_ "net/http/pprof"
	"os/exec"
	"time"

	"github.com/onsi/ginkgo/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
	_ "k8s.io/kubelet"
	"k8s.io/kubernetes/test/e2e/framework"
	e2edebug "k8s.io/kubernetes/test/e2e/framework/debug"
	e2ekubectl "k8s.io/kubernetes/test/e2e/framework/kubectl"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
	utilnet "k8s.io/utils/net"

	test "github.com/chaos-mesh/chaos-mesh/e2e-test"
	e2econfig "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/config" // ensure auth plugins are loaded
)

const namespaceCleanupTimeout = 15 * time.Minute

// This is modified from framework.SetupSuite().
// setupSuite is the boilerplate that can be used to setup ginkgo test suites, on the SynchronizedBeforeSuite step.
// There are certain operations we only want to run once per overall test invocation
// (such as deleting old namespaces, or verifying that all system pods are running.
// Because of the way Ginkgo runs tests in parallel, we must use SynchronizedBeforeSuite
// to ensure that these operations only run on the first parallel Ginkgo node.
func setupSuite(ctx context.Context) {
	// Run only on Ginkgo node 1

	c, err := framework.LoadClientset()
	if err != nil {
		klog.Fatal("Error loading client: ", err)
	}

	// Delete any namespaces except those created by the system. This ensures no
	// lingering resources are left over from a previous test run.
	if framework.TestContext.CleanStart {
		deleted, err := framework.DeleteNamespaces(ctx, c, nil, /* deleteFilter */
			[]string{
				metav1.NamespaceSystem,
				metav1.NamespaceDefault,
				metav1.NamespacePublic,
				v1.NamespaceNodeLease,
				// kind local path provisioner namespace since 0.7.0
				// https://github.com/kubernetes-sigs/kind/blob/v0.7.0/pkg/build/node/storage.go#L35
				"local-path-storage",
			})
		if err != nil {
			framework.Failf("Error deleting orphaned namespaces: %v", err)
		}
		klog.Infof("Waiting for deletion of the following namespaces: %v", deleted)
		if err := framework.WaitForNamespacesDeleted(ctx, c, deleted, namespaceCleanupTimeout); err != nil {
			framework.Failf("Failed to delete orphaned namespaces %v: %v", deleted, err)
		}
	}

	timeouts := framework.NewTimeoutContext()

	// In large clusters we may get to this point but still have a bunch
	// of nodes without Routes created. Since this would make a node
	// unschedulable, we need to wait until all of them are schedulable.
	framework.ExpectNoError(e2enode.WaitForAllNodesSchedulable(ctx, c, timeouts.NodeSchedulable))

	//// If NumNodes is not specified then auto-detect how many are scheduleable and not tainted
	//if framework.TestContext.CloudConfig.NumNodes == framework.DefaultNumNodes {
	//	framework.TestContext.CloudConfig.NumNodes = len(framework.GetReadySchedulableNodesOrDie(c).Items)
	//}

	// Ensure all pods are running and ready before starting tests (otherwise,
	// cluster infrastructure pods that are being pulled or started can block
	// test pods from running, and tests that ensure all pods are running and
	// ready will fail).
	podStartupTimeout := timeouts.SystemPodsStartup
	// TODO: In large clusters, we often observe a non-starting pods due to
	// #41007. To avoid those pods preventing the whole test runs (and just
	// wasting the whole run), we allow for some not-ready pods (with the
	// number equal to the number of allowed not-ready nodes).
	if err := e2epod.WaitForPodsRunningReady(ctx, c, metav1.NamespaceSystem, int32(framework.TestContext.MinStartupPods), int32(framework.TestContext.AllowedNotReadyNodes), podStartupTimeout); err != nil {
		e2edebug.DumpAllNamespaceInfo(ctx, c, metav1.NamespaceSystem)
		e2ekubectl.LogFailedContainers(ctx, c, metav1.NamespaceSystem, framework.Logf)
		framework.Failf("Error waiting for all pods to be running and ready: %v", err)
	}

	//if err := framework.WaitForDaemonSets(c, metav1.NamespaceSystem, int32(framework.TestContext.AllowedNotReadyNodes), framework.TestContext.SystemDaemonsetStartupTimeout); err != nil {
	//	framework.Logf("WARNING: Waiting for all daemonsets to be ready failed: %v", err)
	//}

	dc := c.DiscoveryClient

	serverVersion, serverErr := dc.ServerVersion()
	if serverErr != nil {
		framework.Logf("Unexpected server error retrieving version: %v", serverErr)
	}
	if serverVersion != nil {
		framework.Logf("kube-apiserver version: %s", serverVersion.GitVersion)
	}
}

var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	if e2econfig.TestConfig.InstallChaosMesh {
		ginkgo.By("Clear all helm releases")
		helmClearCmd := "helm ls --all --short | xargs -n 1 -r helm delete --purge"
		if err := exec.Command("sh", "-c", helmClearCmd).Run(); err != nil {
			framework.Failf("failed to clear helm releases (cmd: %q, error: %v", helmClearCmd, err)
		}
		ginkgo.By("Clear non-kubernetes apiservices")
		clearNonK8SAPIServicesCmd := "kubectl delete apiservices -l kube-aggregator.kubernetes.io/automanaged!=onstart"
		if err := exec.Command("sh", "-c", clearNonK8SAPIServicesCmd).Run(); err != nil {
			framework.Failf("failed to clear non-kubernetes apiservices (cmd: %q, error: %v", clearNonK8SAPIServicesCmd, err)
		}

		setupSuite(context.Background())

		// Get clients
		oa, ocfg, err := test.BuildOperatorActionAndCfg(e2econfig.TestConfig)
		framework.ExpectNoError(err, "failed to create operator action")
		oa.CleanCRDOrDie()
		err = oa.InstallCRD(ocfg)
		framework.ExpectNoError(err, "failed to install crd")
		err = oa.DeployOperator(ocfg)
		framework.ExpectNoError(err, "failed to install chaos-mesh")
	}
	return nil
}, func(data []byte) {
	// Run on all Ginkgo nodes
	setupSuitePerGinkgoNode()
})

func setupSuitePerGinkgoNode() {
	c, err := framework.LoadClientset()
	if err != nil {
		klog.Fatal("Error loading client: ", err)
	}
	framework.TestContext.IPFamily = getDefaultClusterIPFamily(c)
	framework.Logf("Cluster IP family: %s", framework.TestContext.IPFamily)
}

// getDefaultClusterIPFamily obtains the default IP family of the cluster
// using the Cluster IP address of the kubernetes service created in the default namespace
// This unequivocally identifies the default IP family because services are single family
// TODO: dual-stack may support multiple families per service
// but we can detect if a cluster is dual stack because pods have two addresses (one per family)
func getDefaultClusterIPFamily(c kubernetes.Interface) string {
	// Get the ClusterIP of the kubernetes service created in the default namespace
	svc, err := c.CoreV1().Services(metav1.NamespaceDefault).Get(context.TODO(), "kubernetes", metav1.GetOptions{})
	if err != nil {
		framework.Failf("Failed to get kubernetes service ClusterIP: %v", err)
	}

	if utilnet.IsIPv6String(svc.Spec.ClusterIP) {
		return "ipv6"
	}
	return "ipv4"
}

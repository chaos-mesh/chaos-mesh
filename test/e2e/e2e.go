package e2e

import (
	"fmt"
	_ "net/http/pprof"
	"os/exec"

	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	e2elog "k8s.io/kubernetes/test/e2e/framework/log"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"

	// ensure auth plugins are loaded
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// This is modified from framework.SetupSuite().
// setupSuite is the boilerplate that can be used to setup ginkgo test suites, on the SynchronizedBeforeSuite step.
// There are certain operations we only want to run once per overall test invocation
// (such as deleting old namespaces, or verifying that all system pods are running.
// Because of the way Ginkgo runs tests in parallel, we must use SynchronizedBeforeSuite
// to ensure that these operations only run on the first parallel Ginkgo node.
func setupSuite() {
	// Run only on Ginkgo node 1

	c, err := framework.LoadClientset()
	if err != nil {
		klog.Fatal("Error loading client: ", err)
	}

	// Delete any namespaces except those created by the system. This ensures no
	// lingering resources are left over from a previous test run.
	if framework.TestContext.CleanStart {
		deleted, err := framework.DeleteNamespaces(c, nil, /* deleteFilter */
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
			e2elog.Failf("Error deleting orphaned namespaces: %v", err)
		}
		klog.Infof("Waiting for deletion of the following namespaces: %v", deleted)
		if err := framework.WaitForNamespacesDeleted(c, deleted, framework.NamespaceCleanupTimeout); err != nil {
			e2elog.Failf("Failed to delete orphaned namespaces %v: %v", deleted, err)
		}
	}

	// In large clusters we may get to this point but still have a bunch
	// of nodes without Routes created. Since this would make a node
	// unschedulable, we need to wait until all of them are schedulable.
	framework.ExpectNoError(framework.WaitForAllNodesSchedulable(c, framework.TestContext.NodeSchedulableTimeout))

	// If NumNodes is not specified then auto-detect how many are scheduleable and not tainted
	if framework.TestContext.CloudConfig.NumNodes == framework.DefaultNumNodes {
		framework.TestContext.CloudConfig.NumNodes = len(framework.GetReadySchedulableNodesOrDie(c).Items)
	}

	// Ensure all pods are running and ready before starting tests (otherwise,
	// cluster infrastructure pods that are being pulled or started can block
	// test pods from running, and tests that ensure all pods are running and
	// ready will fail).
	podStartupTimeout := framework.TestContext.SystemPodsStartupTimeout
	// TODO: In large clusters, we often observe a non-starting pods due to
	// #41007. To avoid those pods preventing the whole test runs (and just
	// wasting the whole run), we allow for some not-ready pods (with the
	// number equal to the number of allowed not-ready nodes).
	if err := e2epod.WaitForPodsRunningReady(c, metav1.NamespaceSystem, int32(framework.TestContext.MinStartupPods), int32(framework.TestContext.AllowedNotReadyNodes), podStartupTimeout, map[string]string{}); err != nil {
		framework.DumpAllNamespaceInfo(c, metav1.NamespaceSystem)
		framework.LogFailedContainers(c, metav1.NamespaceSystem, e2elog.Logf)
		e2elog.Failf("Error waiting for all pods to be running and ready: %v", err)
	}

	if err := framework.WaitForDaemonSets(c, metav1.NamespaceSystem, int32(framework.TestContext.AllowedNotReadyNodes), framework.TestContext.SystemDaemonsetStartupTimeout); err != nil {
		e2elog.Logf("WARNING: Waiting for all daemonsets to be ready failed: %v", err)
	}

	dc := c.DiscoveryClient

	serverVersion, serverErr := dc.ServerVersion()
	if serverErr != nil {
		e2elog.Logf("Unexpected server error retrieving version: %v", serverErr)
	}
	if serverVersion != nil {
		e2elog.Logf("kube-apiserver version: %s", serverVersion.GitVersion)
	}
}

var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
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

	setupSuite()

	// Get clients
	config, err := framework.LoadConfig()
	framework.ExpectNoError(err, "failed to load config")
	kubeCli, err := kubernetes.NewForConfig(config)
	framework.ExpectNoError(err, "failed to create clientset")
	ginkgo.By("Recycle all local PVs")
	pvList, err := kubeCli.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	framework.ExpectNoError(err, "failed to list pvList")
	for _, pv := range pvList.Items {
		if pv.Spec.PersistentVolumeReclaimPolicy == v1.PersistentVolumeReclaimDelete {
			continue
		}
		ginkgo.By(fmt.Sprintf("Update reclaim policy of PV %s to %s", pv.Name, v1.PersistentVolumeReclaimDelete))
		pv.Spec.PersistentVolumeReclaimPolicy = v1.PersistentVolumeReclaimDelete
		_, err = kubeCli.CoreV1().PersistentVolumes().Update(&pv)
		framework.ExpectNoError(err, fmt.Sprintf("failed to update pv %s", pv.Name))
	}
	return nil
}, func(data []byte) {
	// Run on all Ginkgo nodes
	framework.SetupSuitePerGinkgoNode()
})

var _ = ginkgo.SynchronizedAfterSuite(func() {
	framework.CleanupSuite()
}, func() {
	framework.AfterSuiteActions()
})

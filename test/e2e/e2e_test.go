package e2e_test

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/gomega"
	runtimeutils "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/logs"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/config"
	e2elog "k8s.io/kubernetes/test/e2e/framework/log"
	"k8s.io/kubernetes/test/e2e/framework/testfiles"
	"k8s.io/kubernetes/test/e2e/framework/viperconfig"

	e2econfig "github.com/pingcap/chaos-mesh/test/e2e/config"

	// test sources
	_ "github.com/pingcap/chaos-mesh/test/e2e/chaos"
)

var viperConfig = flag.String("viper-config", "", "The name of a viper config file (https://github.com/spf13/viper#what-is-viper). All e2e command line parameters can also be configured in such a file. May contain a path and may or may not contain the file suffix. The default is to look for an optional file with `e2e` as base name. If a file is specified explicitly, it must be present.")

// handleFlags sets up all flags and parses the command line.
func handleFlags() {
	config.CopyFlags(config.Flags, flag.CommandLine)
	framework.RegisterCommonFlags(flag.CommandLine)
	framework.RegisterClusterFlags(flag.CommandLine)
	e2econfig.RegisterChaosMeshConfig(flag.CommandLine)
	flag.Parse()
}

func TestMain(m *testing.M) {
	// Register test flags, then parse flags.
	handleFlags()

	// Now that we know which Viper config (if any) was chosen,
	// parse it and update those options which weren't already set via command line flags
	// (which have higher priority).
	if err := viperconfig.ViperizeFlags(*viperConfig, "e2e", flag.CommandLine); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	framework.AfterReadingAllFlags(&framework.TestContext)

	if framework.TestContext.RepoRoot != "" {
		testfiles.AddFileSource(testfiles.RootFileSource{Root: framework.TestContext.RepoRoot})
	}

	rand.Seed(time.Now().UnixNano())
	os.Exit(m.Run())
}

func TestE2E(t *testing.T) {
	RunE2ETests(t)
}

func RunE2ETests(t *testing.T) {
	runtimeutils.ReallyCrash = true
	logs.InitLogs()
	defer logs.FlushLogs()

	gomega.RegisterFailHandler(e2elog.Fail)

	// Run tests through the Ginkgo runner with output to console + JUnit for Jenkins
	var r []ginkgo.Reporter
	klog.Infof("Starting e2e run %q on Ginkgo node %d", framework.RunID, ginkgoconfig.GinkgoConfig.ParallelNode)

	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "chaosmesh e2e suit", r)
}

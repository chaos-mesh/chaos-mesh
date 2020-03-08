package chaos

import (
	"context"
	// load pprof
	_ "net/http/pprof"
	"time"

	"github.com/onsi/ginkgo"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
)

var _ = ginkgo.Describe("pod", func() {
	f := framework.NewDefaultFramework("tidb-cluster")
	var ns string
	var fwCancel context.CancelFunc

	ginkgo.BeforeEach(func() {
		ns = f.Namespace.Name
		_, cancel := context.WithCancel(context.Background())
		fwCancel = cancel
	})

	ginkgo.AfterEach(func() {
		if fwCancel != nil {
			fwCancel()
		}
	})

	ginkgo.It("example1", func() {
		klog.Infof("ns=%s,start", ns)
		time.Sleep(10 * time.Second)
		klog.Infof("ns=%s,end", ns)
	})

	ginkgo.It("example2", func() {
		klog.Infof("ns=%s,start", ns)
		time.Sleep(10 * time.Second)
		klog.Infof("ns=%s,end", ns)
	})

})

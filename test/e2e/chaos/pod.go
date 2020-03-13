package chaos

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	chaosmeshv1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/test/pkg/fixture"

	"k8s.io/client-go/kubernetes"
	restClient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"

	// load pprof
	_ "net/http/pprof"
	"time"

	"github.com/onsi/ginkgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
)

var _ = ginkgo.Describe("[chaos-mesh] Basic", func() {
	f := framework.NewDefaultFramework("chaos-mesh")
	var ns string
	var fwCancel context.CancelFunc
	var kubeCli kubernetes.Interface
	var config *restClient.Config
	var cli client.Client

	ginkgo.BeforeEach(func() {
		ns = f.Namespace.Name
		_, cancel := context.WithCancel(context.Background())
		fwCancel = cancel
		kubeCli = f.ClientSet
		var err error
		config, err = framework.LoadConfig()
		framework.ExpectNoError(err, "config error")
		scheme := runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(scheme)
		_ = chaosmeshv1alpha1.AddToScheme(scheme)
		cli, err = client.New(config, client.Options{Scheme: scheme})

	})

	ginkgo.AfterEach(func() {
		if fwCancel != nil {
			fwCancel()
		}
	})

	ginkgo.It("example1", func() {
		ctx, cancel := context.WithCancel(context.Background())
		klog.Infof("ns=%s,start", ns)
		namespace, err := kubeCli.CoreV1().Namespaces().Get("chaos-testing", metav1.GetOptions{})
		framework.ExpectNoError(err, "namespace error")
		klog.Infof(namespace.Name)
		klog.Infof("config = %s", config.Username)
		err = cli.Create(ctx, fixture.NewDefaultPodChaos())
		framework.ExpectNoError(err, "create pod chaos error")
		pc := &v1alpha1.PodChaos{}
		err = cli.Get(ctx, client.ObjectKey{
			Namespace: "chaos-testing",
			Name:      "pod-chaos-example",
		}, pc)
		framework.ExpectNoError(err, "get pod chaos error")
		klog.Infof("podchaos name = %v", pc.GetName())
		cancel()
		klog.Infof("ns=%s,end", ns)
	})

	ginkgo.It("example2", func() {
		klog.Infof("ns=%s,start", ns)
		time.Sleep(10 * time.Second)
		klog.Infof("ns=%s,end", ns)
	})

})

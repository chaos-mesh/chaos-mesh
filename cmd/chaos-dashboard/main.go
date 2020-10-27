package main

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	pkgclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
	//log    = ctrl.Log.WithName("test")
)

func main() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	//_ = chaosmeshv1alpha1.AddToScheme(scheme)

	/*
		options := ctrl.Options{
			Scheme: scheme,
			//MetricsBindAddress: common.ControllerCfg.MetricsAddr,
			//LeaderElection:     common.ControllerCfg.EnableLeaderElection,
			//Port:               9443,
			Namespace: "busybox",
		}
	*/

	config := ctrl.GetConfigOrDie()
	config.BearerToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImRxMHZnTmphR2lKbVFfOUhMTWNNWW5aSWV4ZmwxVHpRMWJlclV2b2tGUjQifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJidXN5Ym94Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6InMtdGVzdDEtdG9rZW4tbmhieDciLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoicy10ZXN0MSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjhiYTg3NzZlLTU1YWEtNGFhNS05MTkzLTM0ZmJkOTQ1MzVjZiIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpidXN5Ym94OnMtdGVzdDEifQ.iEUSvOWnE4_k5OJQGWR0FJFZK1Vc1X2dH_SltTtAYPBGgWHmLl77oqX_o7nFpU96Y_YErAntsCun3LMN-Hvfl_0bkUY9Dhc3qDj5AhrmI7b7Tp0diz6oXQmUQ0ZqlFfKvheZTySNgrkbV0m0ssdp9lIFh39eoLmrTPyHjhQDGFD9xnLHi6HyI-da0BQuo7hEwBlpW4LDWgXyndszyXB5BuXw3YMWrzZzAwdYfmPp1Ew-6vVYx68LyQz0Brz5gRYUHz1Md5TgPr7_5Tb53Syo23JSZb9ktjIQiASkHykTfAU7SN62_6EdslFsmn3HVIQkG2h9EqEmCnWtdtY2RC-MNw"
	config.BearerTokenFile = ""
	//log.Info("config", "cfg", config)
	fmt.Println("config", config)
	/*
		mgr, err := ctrl.NewManager(config, options)
		if err != nil {
			fmt.Println(err, "unable to start manager")
			os.Exit(1)
		}
	*/
	client, err := pkgclient.New(config, pkgclient.Options{})
	if err != nil {
		fmt.Println("new client", err)
	}

	go func() {
		time.Sleep(5 * time.Second)
		//client := mgr.GetClient()
		// Using a typed object.

		pod := &corev1.PodList{}
		var listOptions = pkgclient.ListOptions{}

		fmt.Println("begin list pod in busybox")
		listOptions.Namespace = "busybox"
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err = client.List(ctx, pod, &listOptions)
		fmt.Println("list pod in busybox", pod, "err", err)

		listOptions.Namespace = "chaos-testing"
		ctx2, _ := context.WithTimeout(context.Background(), 3*time.Second)
		err = client.List(ctx2, pod, pkgclient.InNamespace("chaos-testing"))

		fmt.Println("list pod in chaos-testing", pod, "err", err)

	}()

	/*
		stopCh := ctrl.SetupSignalHandler()
		if err := mgr.Start(stopCh); err != nil {
			fmt.Println(err, "unable to start manager")
			os.Exit(1)
		}
	*/

	time.Sleep(3600 * time.Second)
}

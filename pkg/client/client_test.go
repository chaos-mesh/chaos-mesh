package client

import (
	"context"
	"os"
	"testing"

	"github.com/chaos-mesh/chaos-mesh/pkg/client/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestClient(t *testing.T) {
	var config *rest.Config
	var err error

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {

		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}

	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		t.Skipf("Kubeconfig file not found at %s, skipping test", kubeconfig)
		return
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatalf("Failed to load kubeconfig: %v", err)
	}

	clientSet, err := versioned.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create clientSet: %v", err)
	}

	ctx := context.Background()

	podChaosItems, err := clientSet.ApiV1alpha1().Podchaos("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to list PodChaos in namespace %s: %v", "default", err)
	}
	t.Logf("Found %d PodChaos in namespace %s", len(podChaosItems.Items), "default")
}

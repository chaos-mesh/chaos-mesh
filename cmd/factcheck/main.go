package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/factcheck/core"
	reportpkg "github.com/chaos-mesh/chaos-mesh/pkg/factcheck/report"
	"github.com/chaos-mesh/chaos-mesh/pkg/factcheck/verifiers/networkchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/factcheck/verifiers/podchaos"
)

var (
	chaosName   string
	chaosNs     string
	chaosKind   string
	outputFmt   string
	scheme      = runtime.NewScheme()
)

func init() {
	flag.StringVar(&chaosName, "name", "", "Name of the Chaos CR")
	flag.StringVar(&chaosNs, "namespace", "default", "Namespace of the Chaos CR")
	flag.StringVar(&chaosKind, "kind", "PodChaos", "Kind of the Chaos CR (e.g., PodChaos, NetworkChaos)")
	flag.StringVar(&outputFmt, "output", "text", "Output format (text or json)")

	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
}

func main() {
	flag.Parse()

	if chaosName == "" {
		fmt.Println("Error: -name flag is required")
		os.Exit(1)
	}

	// Initialize K8s client
	cfg := ctrl.GetConfigOrDie()
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		fmt.Printf("Failed to create Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("Failed to create Kubernetes clientset: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Fetch target Chaos CR
	var obj client.Object
	if chaosKind == "PodChaos" {
		obj = &v1alpha1.PodChaos{}
	} else if chaosKind == "NetworkChaos" {
		obj = &v1alpha1.NetworkChaos{}
	} else {
		fmt.Printf("Error: Unsupported chaos kind '%s'. Currently only supporting PodChaos and NetworkChaos.\n", chaosKind)
		os.Exit(1)
	}

	err = c.Get(ctx, types.NamespacedName{Namespace: chaosNs, Name: chaosName}, obj)
	if err != nil {
		fmt.Printf("Failed to fetch %s %s/%s: %v\n", chaosKind, chaosNs, chaosName, err)
		os.Exit(1)
	}

	// Register verifiers
	registry := core.NewRegistry()
	registry.Register(podchaos.NewVerifier(c))
	registry.Register(networkchaos.NewVerifier(c, clientset))

	// Dispatch verification via Registry
	verifier, err := registry.Get(chaosKind)
	if err != nil {
		fmt.Printf("Verification failed: %v\n", err)
		os.Exit(1)
	}

	targets, err := verifier.ResolveTargets(ctx, obj)
	if err != nil {
		fmt.Printf("Failed to resolve targets: %v\n", err)
		os.Exit(1)
	}

	results, err := verifier.CollectEvidence(ctx, obj, targets)
	if err != nil {
		fmt.Printf("Failed to collect evidence: %v\n", err)
		os.Exit(1)
	}

	verdict, reason := verifier.Evaluate(results)

	// Flatten evidence for the report
	var allEvidence []core.Evidence
	for _, res := range results {
		allEvidence = append(allEvidence, res.Evidence...)
	}

	report := &core.VerdictReport{
		ChaosName: chaosName,
		ChaosKind: chaosKind,
		Verdict:   verdict,
		Reason:    reason,
		Evidence:  allEvidence,
	}

	// Render report
	renderer := reportpkg.NewRenderer()
	if err := renderer.Render(report, outputFmt); err != nil {
		fmt.Printf("Failed to render report: %v\n", err)
		os.Exit(1)
	}
}

package networkchaos

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/pkg/factcheck/core"
)

type Verifier struct {
	client    client.Client
	clientset *kubernetes.Clientset
}

func NewVerifier(c client.Client, clientset *kubernetes.Clientset) *Verifier {
	return &Verifier{
		client:    c,
		clientset: clientset,
	}
}

func (v *Verifier) Kind() string {
	return "NetworkChaos"
}

func (v *Verifier) ResolveTargets(ctx context.Context, obj client.Object) ([]core.Target, error) {
	networkChaos, ok := obj.(*v1alpha1.NetworkChaos)
	if !ok {
		return nil, fmt.Errorf("expected *v1alpha1.NetworkChaos")
	}

	var targets []core.Target
	for id := range networkChaos.Status.Instances {
		nn, err := controller.ParseNamespacedName(id)
		if err != nil {
			continue
		}
		targets = append(targets, core.Target{
			Namespace: nn.Namespace,
			Name:      nn.Name,
			Kind:      "Pod",
		})
	}
	return targets, nil
}

func (v *Verifier) CollectEvidence(ctx context.Context, obj client.Object, targets []core.Target) ([]core.TargetResult, error) {
	var results []core.TargetResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// TODO: Re-implement active network probes (e.g. ping/curl) using ephemeral pods.
	// Current implementation is a stub.
	var collectors []core.Collector

	for _, target := range targets {
		wg.Add(1)
		go func(t core.Target) {
			defer wg.Done()
			
			var targetEvidence []core.Evidence
			var errorMsg string
			success := false
			
			for _, collector := range collectors {
				ev, err := collector.Collect(ctx, t)
				if err != nil {
					errorMsg = err.Error()
				} else if len(ev) > 0 {
					targetEvidence = append(targetEvidence, ev...)
					success = true
				}
			}
			
			if !success && errorMsg == "" {
				errorMsg = "Not implemented: active probes required for NetworkChaos"
			}
			
			mu.Lock()
			results = append(results, core.TargetResult{
				Target:   t,
				Success:  success,
				ErrorMsg: errorMsg,
				Evidence: targetEvidence,
			})
			mu.Unlock()
			
		}(target)
	}

	wg.Wait()
	return results, nil
}

func (v *Verifier) Evaluate(results []core.TargetResult) (core.Verdict, string) {
	if len(results) == 0 {
		return core.Mismatched, "No valid targets found in NetworkChaos status."
	}

	matchedCount := 0
	var failedReasons []string

	for _, res := range results {
		if res.Success {
			matchedCount++
		} else {
			failedReasons = append(failedReasons, fmt.Sprintf("%s (%s)", res.Target.Name, res.ErrorMsg))
		}
	}

	if matchedCount == len(results) {
		return core.Matched, fmt.Sprintf("Active probes detected artificial network disruption for all %d targets.", len(results))
	} else if matchedCount > 0 {
		return core.Mismatched, fmt.Sprintf("Partial match: Active probes detected disruption for only %d/%d targets. Failures: %s", matchedCount, len(results), strings.Join(failedReasons, ", "))
	}

	return core.Mismatched, fmt.Sprintf("Active probes failed. Failures: %s", strings.Join(failedReasons, ", "))
}

package podchaos

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/pkg/factcheck/collectors/k8s"
	"github.com/chaos-mesh/chaos-mesh/pkg/factcheck/core"
)

type Verifier struct {
	client client.Client
}

func NewVerifier(c client.Client) *Verifier {
	return &Verifier{
		client: c,
	}
}

func (v *Verifier) Kind() string {
	return "PodChaos"
}

func (v *Verifier) ResolveTargets(ctx context.Context, obj client.Object) ([]core.Target, error) {
	podChaos, ok := obj.(*v1alpha1.PodChaos)
	if !ok {
		return nil, fmt.Errorf("expected *v1alpha1.PodChaos")
	}

	var targets []core.Target
	for _, record := range podChaos.Status.Experiment.Records {
		nn, err := controller.ParseNamespacedName(record.Id)
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
	podChaos, ok := obj.(*v1alpha1.PodChaos)
	if !ok {
		return nil, fmt.Errorf("expected *v1alpha1.PodChaos in CollectEvidence")
	}

	var results []core.TargetResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	injectionTime := podChaos.GetCreationTimestamp().Time

	var collectors []core.Collector

	// Assemble collectors based on PodChaos Action.
	switch podChaos.Spec.Action {
	case v1alpha1.PodKillAction:
		collectors = []core.Collector{
			k8s.NewPodStateCollector(v.client, injectionTime),
			k8s.NewEventCollector(v.client, injectionTime, "Killing"),
		}
	case v1alpha1.PodFailureAction:
		// TODO: Implement PodFailureCollector.
		return nil, fmt.Errorf("pod-failure action is not yet supported by collectors")
	case v1alpha1.ContainerKillAction:
		return nil, fmt.Errorf("container-kill action is not yet supported by collectors")
	default:
		return nil, fmt.Errorf("unsupported PodChaos action: %s", podChaos.Spec.Action)
	}

	for _, target := range targets {
		wg.Add(1)
		go func(t core.Target) {
			defer wg.Done()
			
			// Polling loop
			timeout := time.After(30 * time.Second)
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			var targetEvidence []core.Evidence

			for {
				select {
				case <-ctx.Done():
					return
				case <-timeout:
					mu.Lock()
					results = append(results, core.TargetResult{
						Target:   t,
						Success:  false,
						ErrorMsg: "Timed out after 30s waiting for evidence of pod termination",
						Evidence: targetEvidence,
					})
					mu.Unlock()
					return
				case <-ticker.C:
					foundMatch := false
					for _, collector := range collectors {
						ev, err := collector.Collect(ctx, t)
						if err == nil && len(ev) > 0 {
							targetEvidence = append(targetEvidence, ev...)
							foundMatch = true
						}
					}
					
					if foundMatch {
						mu.Lock()
						results = append(results, core.TargetResult{
							Target:   t,
							Success:  true,
							Evidence: targetEvidence,
						})
						mu.Unlock()
						return
					}
				}
			}
		}(target)
	}

	wg.Wait()
	return results, nil
}

func (v *Verifier) Evaluate(results []core.TargetResult) (core.Verdict, string) {
	if len(results) == 0 {
		return core.Mismatched, "No targets could be resolved or no evidence collected."
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
		return core.Matched, fmt.Sprintf("Found conclusive evidence for all %d target pods.", len(results))
	} else if matchedCount > 0 {
		return core.Mismatched, fmt.Sprintf("Partial match: Found conclusive evidence for only %d/%d target pods. Failures: %s", matchedCount, len(results), strings.Join(failedReasons, ", "))
	}

	return core.Mismatched, fmt.Sprintf("Polling window expired. No conclusive evidence of pod termination was found. Failures: %s", strings.Join(failedReasons, ", "))
}

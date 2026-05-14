package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/factcheck/core"
)

// PodStateCollector verifies pod lifecycle states.
type PodStateCollector struct {
	client        client.Client
	injectionTime time.Time
}

func NewPodStateCollector(c client.Client, injectionTime time.Time) *PodStateCollector {
	return &PodStateCollector{
		client:        c,
		injectionTime: injectionTime,
	}
}

func (c *PodStateCollector) Name() string {
	return "PodStateCollector"
}

func (c *PodStateCollector) Collect(ctx context.Context, target core.Target) ([]core.Evidence, error) {
	var evidence []core.Evidence
	var pod corev1.Pod
	
	nn := types.NamespacedName{Namespace: target.Namespace, Name: target.Name}
	err := c.client.Get(ctx, nn, &pod)

	if err != nil {
		if apierrors.IsNotFound(err) {
			desc := fmt.Sprintf("Pod %s is NotFound (successfully deleted)", nn.String())

			// Attempt to resolve replacement pod via ReplicaSet prefix
			lastDash := strings.LastIndex(target.Name, "-")
			if lastDash > 0 {
				rsPrefix := target.Name[:lastDash]
				
				var podList corev1.PodList
				var newPods []string
				if err := c.client.List(ctx, &podList, client.InNamespace(target.Namespace)); err == nil {
					for _, p := range podList.Items {
						if p.Name != target.Name && strings.HasPrefix(p.Name, rsPrefix) && !p.CreationTimestamp.Time.Before(c.injectionTime) {
							newPods = append(newPods, p.Name)
						}
					}
				}
				
				if len(newPods) > 0 {
					desc += fmt.Sprintf("\n      Replacements Spawned: %s", strings.Join(newPods, ", "))
				}
			}

			evidence = append(evidence, core.Evidence{
				Type:        "PodState",
				Source:      c.Name(),
				Target:      nn.String(),
				Timestamp:   time.Now().Format(time.RFC3339),
				Description: desc,
			})
		}
		return evidence, nil
	}

	// Check for recreation (StatefulSet/DaemonSet in-place restart)
	if pod.CreationTimestamp.Time.After(c.injectionTime) {
		evidence = append(evidence, core.Evidence{
			Type:        "PodState",
			Source:      c.Name(),
			Target:      nn.String(),
			Timestamp:   time.Now().Format(time.RFC3339),
			Description: fmt.Sprintf("Pod %s exists but was recreated (CreationTime > Chaos CreationTime)", nn.String()),
		})
	}

	return evidence, nil
}

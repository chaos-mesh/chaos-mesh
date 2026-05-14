package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/factcheck/core"
)

// EventCollector checks for specific events (like 'Killing') related to a target.
type EventCollector struct {
	client        client.Client
	injectionTime time.Time
	reason        string
}

func NewEventCollector(c client.Client, injectionTime time.Time, reason string) *EventCollector {
	return &EventCollector{
		client:        c,
		injectionTime: injectionTime,
		reason:        reason,
	}
}

func (c *EventCollector) Name() string {
	return "EventCollector"
}

func (c *EventCollector) Collect(ctx context.Context, target core.Target) ([]core.Evidence, error) {
	var evidence []core.Evidence
	var events corev1.EventList
	
	if err := c.client.List(ctx, &events, client.InNamespace(target.Namespace)); err != nil {
		return evidence, err
	}

	for _, event := range events.Items {
		if event.InvolvedObject.Name == target.Name && event.InvolvedObject.Kind == target.Kind && event.Reason == c.reason {
			if event.LastTimestamp.Time.After(c.injectionTime) {
				evidence = append(evidence, core.Evidence{
					Type:        "KubernetesEvent",
					Source:      c.Name(),
					Target:      fmt.Sprintf("%s/%s", target.Namespace, target.Name),
					Timestamp:   time.Now().Format(time.RFC3339),
					Description: fmt.Sprintf("Found %s event: %s", c.reason, event.Message),
				})
			}
		}
	}

	return evidence, nil
}

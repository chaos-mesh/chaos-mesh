package core

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Target represents an infrastructure component under verification.
type Target struct {
	Namespace string
	Name      string
	Kind      string
}

// Evidence represents a verified runtime fact.
type Evidence struct {
	Type        string // e.g., "PodState", "KubernetesEvent", "ActiveProbe"
	Source      string // Specific source, e.g., "PingCollector"
	Target      string // String representation of the target it relates to
	Timestamp   string
	Description string
}

// Verdict represents the final outcome of the verification.
type Verdict string

const (
	Matched      Verdict = "MATCHED"
	Mismatched   Verdict = "MISMATCHED"
	Inconclusive Verdict = "INCONCLUSIVE"
)

// TargetResult aggregates evidence per target.
type TargetResult struct {
	Target   Target
	Success  bool
	ErrorMsg string
	Evidence []Evidence
}

// VerdictReport contains the final structured verification result.
type VerdictReport struct {
	ChaosName string     `json:"chaosName"`
	ChaosKind string     `json:"chaosKind"`
	Verdict   Verdict    `json:"verdict"`
	Reason    string     `json:"reason"`
	Evidence  []Evidence `json:"evidence"`
}

// Verifier implements capability-centric verification for a specific chaos kind.
type Verifier interface {
	Kind() string
	ResolveTargets(ctx context.Context, obj client.Object) ([]Target, error)
	CollectEvidence(ctx context.Context, obj client.Object, targets []Target) ([]TargetResult, error)
	Evaluate(results []TargetResult) (Verdict, string)
}

// Collector handles atomic evidence collection.
type Collector interface {
	Name() string
	Collect(ctx context.Context, target Target) ([]Evidence, error)
}

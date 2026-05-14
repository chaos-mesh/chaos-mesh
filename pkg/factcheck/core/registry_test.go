package core

import (
	"context"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type dummyVerifier struct{}

func (d *dummyVerifier) Kind() string {
	return "DummyChaos"
}

func (d *dummyVerifier) ResolveTargets(_ context.Context, _ client.Object) ([]Target, error) {
	return nil, nil
}

func (d *dummyVerifier) CollectEvidence(_ context.Context, _ client.Object, _ []Target) ([]TargetResult, error) {
	return nil, nil
}

func (d *dummyVerifier) Evaluate(_ []TargetResult) (Verdict, string) {
	return Matched, ""
}

func TestRegistryGet(t *testing.T) {
	r := NewRegistry()
	r.Register(&dummyVerifier{})

	v, err := r.Get("DummyChaos")
	if err != nil {
		t.Fatal(err)
	}

	if v.Kind() != "DummyChaos" {
		t.Fatalf("expected DummyChaos verifier, got %s", v.Kind())
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("MissingChaos")
	if err == nil {
		t.Fatal("expected error for missing verifier")
	}
}

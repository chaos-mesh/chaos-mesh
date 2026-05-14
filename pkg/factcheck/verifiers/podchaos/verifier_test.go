package podchaos

import (
	"context"
	"reflect"
	"testing"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/factcheck/core"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestResolveTargets(t *testing.T) {
	cli := fake.NewClientBuilder().Build()
	v := NewVerifier(cli)

	podChaos := &v1alpha1.PodChaos{}
	podChaos.Status.Experiment.Records = []*v1alpha1.Record{
		{Id: "default/nginx-1"},
		{Id: "test/web-0"},
	}

	targets, err := v.ResolveTargets(context.Background(), podChaos)
	if err != nil {
		t.Fatal(err)
	}

	expected := []core.Target{
		{Namespace: "default", Name: "nginx-1", Kind: "Pod"},
		{Namespace: "test", Name: "web-0", Kind: "Pod"},
	}

	if !reflect.DeepEqual(targets, expected) {
		t.Fatalf("expected %#v, got %#v", expected, targets)
	}
}

func TestEvaluate(t *testing.T) {
	cli := fake.NewClientBuilder().Build()
	v := NewVerifier(cli)

	t.Run("no results", func(t *testing.T) {
		verdict, reason := v.Evaluate(nil)
		if verdict != core.Mismatched {
			t.Fatalf("expected %s, got %s", core.Mismatched, verdict)
		}
		if reason == "" {
			t.Fatal("expected non-empty reason")
		}
	})

	t.Run("all matched", func(t *testing.T) {
		results := []core.TargetResult{{Target: core.Target{Name: "nginx-1"}, Success: true}}
		verdict, _ := v.Evaluate(results)
		if verdict != core.Matched {
			t.Fatalf("expected %s, got %s", core.Matched, verdict)
		}
	})

	t.Run("partial match", func(t *testing.T) {
		results := []core.TargetResult{
			{Target: core.Target{Name: "nginx-1"}, Success: true},
			{Target: core.Target{Name: "web-0"}, Success: false, ErrorMsg: "timed out"},
		}
		verdict, reason := v.Evaluate(results)
		if verdict != core.Mismatched {
			t.Fatalf("expected %s, got %s", core.Mismatched, verdict)
		}
		if reason == "" {
			t.Fatal("expected non-empty reason")
		}
	})
}

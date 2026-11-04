// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package netem

import (
	"testing"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestFromDelay_Valid(t *testing.T) {
	in := &v1alpha1.DelaySpec{
		Latency:     "100ms",
		Jitter:      "10ms",
		Correlation: "25",
	}
	out, err := FromDelay(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Time != "100ms" {
		t.Errorf("expected Time=100ms, got %s", out.Time)
	}
	if out.Jitter != "10ms" {
		t.Errorf("expected Jitter=10ms, got %s", out.Jitter)
	}
	if out.DelayCorr != float32(25) {
		t.Errorf("expected DelayCorr=25, got %f", out.DelayCorr)
	}
}

func TestFromDelay_WithReorder(t *testing.T) {
	in := &v1alpha1.DelaySpec{
		Latency:     "50ms",
		Jitter:      "0ms",
		Correlation: "0",
		Reorder: &v1alpha1.ReorderSpec{
			Reorder:     "10",
			Correlation: "5",
			Gap:         3,
		},
	}
	out, err := FromDelay(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Reorder != float32(10) {
		t.Errorf("expected Reorder=10, got %f", out.Reorder)
	}
	if out.ReorderCorr != float32(5) {
		t.Errorf("expected ReorderCorr=5, got %f", out.ReorderCorr)
	}
	if out.Gap != 3 {
		t.Errorf("expected Gap=3, got %d", out.Gap)
	}
}

func TestFromDelay_InvalidLatency(t *testing.T) {
	in := &v1alpha1.DelaySpec{
		Latency:     "notaduration",
		Jitter:      "0ms",
		Correlation: "0",
	}
	_, err := FromDelay(in)
	if err == nil {
		t.Error("expected error for invalid latency, got nil")
	}
}

func TestFromDelay_InvalidJitter(t *testing.T) {
	in := &v1alpha1.DelaySpec{
		Latency:     "100ms",
		Jitter:      "notaduration",
		Correlation: "0",
	}
	_, err := FromDelay(in)
	if err == nil {
		t.Error("expected error for invalid jitter, got nil")
	}
}

func TestFromDelay_InvalidCorrelation(t *testing.T) {
	in := &v1alpha1.DelaySpec{
		Latency:     "100ms",
		Jitter:      "0ms",
		Correlation: "abc",
	}
	_, err := FromDelay(in)
	if err == nil {
		t.Error("expected error for invalid correlation, got nil")
	}
}

func TestFromDelay_InvalidReorderPercentage(t *testing.T) {
	in := &v1alpha1.DelaySpec{
		Latency:     "100ms",
		Jitter:      "0ms",
		Correlation: "0",
		Reorder: &v1alpha1.ReorderSpec{
			Reorder:     "abc",
			Correlation: "0",
			Gap:         1,
		},
	}
	_, err := FromDelay(in)
	if err == nil {
		t.Error("expected error for invalid reorder percentage, got nil")
	}
}

func TestFromDelay_InvalidReorderCorrelation(t *testing.T) {
	in := &v1alpha1.DelaySpec{
		Latency:     "100ms",
		Jitter:      "0ms",
		Correlation: "0",
		Reorder: &v1alpha1.ReorderSpec{
			Reorder:     "10",
			Correlation: "abc",
			Gap:         1,
		},
	}
	_, err := FromDelay(in)
	if err == nil {
		t.Error("expected error for invalid reorder correlation, got nil")
	}
}

func TestFromLoss_Valid(t *testing.T) {
	in := &v1alpha1.LossSpec{
		Loss:        "50",
		Correlation: "10",
	}
	out, err := FromLoss(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Loss != float32(50) {
		t.Errorf("expected Loss=50, got %f", out.Loss)
	}
	if out.LossCorr != float32(10) {
		t.Errorf("expected LossCorr=10, got %f", out.LossCorr)
	}
}

func TestFromLoss_InvalidLoss(t *testing.T) {
	in := &v1alpha1.LossSpec{
		Loss:        "abc",
		Correlation: "0",
	}
	_, err := FromLoss(in)
	if err == nil {
		t.Error("expected error for invalid loss, got nil")
	}
}

func TestFromLoss_InvalidCorrelation(t *testing.T) {
	in := &v1alpha1.LossSpec{
		Loss:        "50",
		Correlation: "abc",
	}
	_, err := FromLoss(in)
	if err == nil {
		t.Error("expected error for invalid correlation, got nil")
	}
}

func TestFromDuplicate_Valid(t *testing.T) {
	in := &v1alpha1.DuplicateSpec{
		Duplicate:   "30",
		Correlation: "5",
	}
	out, err := FromDuplicate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Duplicate != float32(30) {
		t.Errorf("expected Duplicate=30, got %f", out.Duplicate)
	}
	if out.DuplicateCorr != float32(5) {
		t.Errorf("expected DuplicateCorr=5, got %f", out.DuplicateCorr)
	}
}

func TestFromDuplicate_InvalidDuplicate(t *testing.T) {
	in := &v1alpha1.DuplicateSpec{
		Duplicate:   "abc",
		Correlation: "0",
	}
	_, err := FromDuplicate(in)
	if err == nil {
		t.Error("expected error for invalid duplicate, got nil")
	}
}

func TestFromDuplicate_InvalidCorrelation(t *testing.T) {
	in := &v1alpha1.DuplicateSpec{
		Duplicate:   "30",
		Correlation: "abc",
	}
	_, err := FromDuplicate(in)
	if err == nil {
		t.Error("expected error for invalid correlation, got nil")
	}
}

func TestFromCorrupt_Valid(t *testing.T) {
	in := &v1alpha1.CorruptSpec{
		Corrupt:     "20",
		Correlation: "15",
	}
	out, err := FromCorrupt(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Corrupt != float32(20) {
		t.Errorf("expected Corrupt=20, got %f", out.Corrupt)
	}
	if out.CorruptCorr != float32(15) {
		t.Errorf("expected CorruptCorr=15, got %f", out.CorruptCorr)
	}
}

func TestFromCorrupt_InvalidCorrupt(t *testing.T) {
	in := &v1alpha1.CorruptSpec{
		Corrupt:     "abc",
		Correlation: "0",
	}
	_, err := FromCorrupt(in)
	if err == nil {
		t.Error("expected error for invalid corrupt, got nil")
	}
}

func TestFromCorrupt_InvalidCorrelation(t *testing.T) {
	in := &v1alpha1.CorruptSpec{
		Corrupt:     "20",
		Correlation: "abc",
	}
	_, err := FromCorrupt(in)
	if err == nil {
		t.Error("expected error for invalid correlation, got nil")
	}
}

func TestFromRate_Valid(t *testing.T) {
	in := &v1alpha1.RateSpec{
		Rate: "1mbps",
	}
	out, err := FromRate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Rate != "1mbps" {
		t.Errorf("expected Rate=1mbps, got %s", out.Rate)
	}
}

func TestFromBandwidth_Valid(t *testing.T) {
	in := &v1alpha1.BandwidthSpec{
		Rate:   "10mbps",
		Limit:  100,
		Buffer: 10000,
	}
	out, err := FromBandwidth(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Rate != "10mbps" {
		t.Errorf("expected Rate=10mbps, got %s", out.Rate)
	}
	if out.Limit != 100 {
		t.Errorf("expected Limit=100, got %d", out.Limit)
	}
	if out.Buffer != 10000 {
		t.Errorf("expected Buffer=10000, got %d", out.Buffer)
	}
	if out.PeakRate != 0 {
		t.Errorf("expected PeakRate=0 when not set, got %d", out.PeakRate)
	}
	if out.MinBurst != 0 {
		t.Errorf("expected MinBurst=0 when not set, got %d", out.MinBurst)
	}
}

func TestFromBandwidth_WithPeakrateAndMinburst(t *testing.T) {
	peakrate := uint64(5000000)
	minburst := uint32(1500)
	in := &v1alpha1.BandwidthSpec{
		Rate:     "10mbps",
		Limit:    100,
		Buffer:   10000,
		Peakrate: &peakrate,
		Minburst: &minburst,
	}
	out, err := FromBandwidth(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PeakRate != peakrate {
		t.Errorf("expected PeakRate=%d, got %d", peakrate, out.PeakRate)
	}
	if out.MinBurst != minburst {
		t.Errorf("expected MinBurst=%d, got %d", minburst, out.MinBurst)
	}
}

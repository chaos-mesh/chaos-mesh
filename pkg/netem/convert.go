// Copyright 2021 Chaos Mesh Authors.
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
	"strconv"
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	chaosdaemonpb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

// FromDelay convert delay to netem
func FromDelay(in *v1alpha1.DelaySpec) (*chaosdaemonpb.Netem, error) {
	delayTime, err := time.ParseDuration(in.Latency)
	if err != nil {
		return nil, err
	}
	jitter, err := time.ParseDuration(in.Jitter)
	if err != nil {
		return nil, err
	}

	corr, err := strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		return nil, err
	}

	netem := &chaosdaemonpb.Netem{
		Time:      uint32(delayTime.Nanoseconds() / 1e3),
		DelayCorr: float32(corr),
		Jitter:    uint32(jitter.Nanoseconds() / 1e3),
	}

	if in.Reorder != nil {
		reorderPercentage, err := strconv.ParseFloat(in.Reorder.Reorder, 32)
		if err != nil {
			return nil, err
		}

		corr, err := strconv.ParseFloat(in.Reorder.Correlation, 32)
		if err != nil {
			return nil, err
		}

		netem.Reorder = float32(reorderPercentage)
		netem.ReorderCorr = float32(corr)
		netem.Gap = uint32(in.Reorder.Gap)
	}

	return netem, nil
}

// FromLoss convert loss to netem
func FromLoss(in *v1alpha1.LossSpec) (*chaosdaemonpb.Netem, error) {
	lossPercentage, err := strconv.ParseFloat(in.Loss, 32)
	if err != nil {
		return nil, err
	}

	corr, err := strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		return nil, err
	}

	return &chaosdaemonpb.Netem{
		Loss:     float32(lossPercentage),
		LossCorr: float32(corr),
	}, nil
}

// FromDuplicate convert duplicate to netem
func FromDuplicate(in *v1alpha1.DuplicateSpec) (*chaosdaemonpb.Netem, error) {
	duplicatePercentage, err := strconv.ParseFloat(in.Duplicate, 32)
	if err != nil {
		return nil, err
	}

	corr, err := strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		return nil, err
	}

	return &chaosdaemonpb.Netem{
		Duplicate:     float32(duplicatePercentage),
		DuplicateCorr: float32(corr),
	}, nil
}

// FromCorrupt convert corrupt to netem
func FromCorrupt(in *v1alpha1.CorruptSpec) (*chaosdaemonpb.Netem, error) {
	corruptPercentage, err := strconv.ParseFloat(in.Corrupt, 32)
	if err != nil {
		return nil, err
	}

	corr, err := strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		return nil, err
	}

	return &chaosdaemonpb.Netem{
		Corrupt:     float32(corruptPercentage),
		CorruptCorr: float32(corr),
	}, nil
}

// FromRate convert RateSpec to netem
func FromRate(in *v1alpha1.RateSpec) (*chaosdaemonpb.Netem, error) {
	return &chaosdaemonpb.Netem{
		Rate: in.Rate,
	}, nil
}

// FromBandwidth converts BandwidthSpec to *chaosdaemonpb.Tbf
// Bandwidth action use TBF under the hood.
// TBF stands for Token Bucket Filter, is a classful queueing discipline available
// for traffic control with the tc command.
// http://man7.org/linux/man-pages/man8/tc-tbf.8.html
func FromBandwidth(in *v1alpha1.BandwidthSpec) (*chaosdaemonpb.Tbf, error) {
	tbf := &chaosdaemonpb.Tbf{
		Rate:   in.Rate,
		Limit:  in.Limit,
		Buffer: in.Buffer,
	}

	if in.Peakrate != nil && in.Minburst != nil {
		tbf.PeakRate = *in.Peakrate
		tbf.MinBurst = *in.Minburst
	}

	return tbf, nil
}

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
	"math"
	"strconv"
	"strings"

	chaosdaemon "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

// MergeNetem merges two Netem protos into a new one.
// REMEMBER to assign the return value, i.e. merged = utils.MergeNetm(merged, em)
// For each field it takes the bigger value of the two.
// Its main use case is merging netem of different types, e.g. delay and loss.
// It returns nil if both inputs are nil.
// Otherwise it returns a new Netem with merged values.
func MergeNetem(a, b *chaosdaemon.Netem) *chaosdaemon.Netem {
	if a == nil && b == nil {
		return nil
	}
	// NOTE: because proto getters check nil, we are good here even if one of them is nil.
	// But we just assign empty value to make IDE and linters happy.
	if a == nil {
		a = &chaosdaemon.Netem{}
	}
	if b == nil {
		b = &chaosdaemon.Netem{}
	}
	return &chaosdaemon.Netem{
		Time:          maxu32(a.GetTime(), b.GetTime()),
		Jitter:        maxu32(a.GetJitter(), b.GetJitter()),
		DelayCorr:     maxf32(a.GetDelayCorr(), b.GetDelayCorr()),
		Limit:         maxu32(a.GetLimit(), b.GetLimit()),
		Loss:          maxf32(a.GetLoss(), b.GetLoss()),
		LossCorr:      maxf32(a.GetLossCorr(), b.GetLossCorr()),
		Gap:           maxu32(a.GetGap(), b.GetGap()),
		Duplicate:     maxf32(a.GetDuplicate(), b.GetDuplicate()),
		DuplicateCorr: maxf32(a.GetDuplicateCorr(), b.GetDuplicateCorr()),
		Reorder:       maxf32(a.GetReorder(), b.GetReorder()),
		ReorderCorr:   maxf32(a.GetReorderCorr(), b.GetReorderCorr()),
		Corrupt:       maxf32(a.GetCorrupt(), b.GetCorrupt()),
		CorruptCorr:   maxf32(a.GetCorruptCorr(), b.GetCorruptCorr()),
		Rate:          maxRateString(a.GetRate(), b.GetRate()),
	}
}

func maxu32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func maxf32(a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}

func parseRate(nu string) uint64 {
	// normalize input
	s := strings.ToLower(strings.TrimSpace(nu))

	for i, u := range []string{"tbps", "gbps", "mbps", "kbps", "bps"} {
		if strings.HasSuffix(s, u) {
			ts := strings.TrimSuffix(s, u)
			s := strings.TrimSpace(ts)

			n, err := strconv.ParseUint(s, 10, 64)

			if err != nil {
				return 0
			}

			// convert unit to bytes
			for j := 4 - i; j > 0; j-- {
				n = n * 1024
			}

			return n
		}
	}

	for i, u := range []string{"tbit", "gbit", "mbit", "kbit", "bit"} {
		if strings.HasSuffix(s, u) {
			ts := strings.TrimSuffix(s, u)
			s := strings.TrimSpace(ts)

			n, err := strconv.ParseUint(s, 10, 64)

			if err != nil {
				return 0
			}

			// convert unit to bytes
			for j := 4 - i; j > 0; j-- {
				n = n * 1000
			}
			n = n / 8

			return n
		}
	}

	return 0
}

func maxRateString(a, b string) string {
	if parseRate(a) > parseRate(b) {
		return a
	}
	return b
}

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

package time

import (
	"github.com/pkg/errors"
)

// timeSkewFakeImage is the filename of fake image after compiling
const timeSkewFakeImage = "fake_clock_gettime.o"

// These three consts corresponding to the three extern variables in the fake_clock_gettime.c
const (
	externVarClockIdsMask = "CLOCK_IDS_MASK"
	externVarTvSecDelta   = "TV_SEC_DELTA"
	externVarTvNsecDelta  = "TV_NSEC_DELTA"
)

type TimeSkew struct {
	deltaSeconds     int64
	deltaNanoSeconds int64
	clockIDsMask     uint64
	fakeImage        *FakeImage
}

func NewTimeSkew(deltaSeconds int64, deltaNanoSeconds int64, clockIDsMask uint64) (*TimeSkew, error) {
	var image *FakeImage
	var err error

	if image, err = LoadFakeImageFromEmbedFs(timeSkewFakeImage); err != nil {
		return nil, errors.Wrap(err, "load fake image")
	}

	return NewTimeSkewWithCustomFakeImage(deltaSeconds, deltaNanoSeconds, clockIDsMask, image), nil
}

func NewTimeSkewWithCustomFakeImage(deltaSeconds int64, deltaNanoSeconds int64, clockIDsMask uint64, fakeImage *FakeImage) *TimeSkew {
	return &TimeSkew{deltaSeconds: deltaSeconds, deltaNanoSeconds: deltaNanoSeconds, clockIDsMask: clockIDsMask, fakeImage: fakeImage}
}

func (it *TimeSkew) Inject(pid int) error {
	return it.fakeImage.AttachToProcess(pid, map[string]uint64{
		externVarClockIdsMask: it.clockIDsMask,
		externVarTvSecDelta:   uint64(it.deltaSeconds),
		externVarTvNsecDelta:  uint64(it.deltaNanoSeconds),
	})
}

func (it *TimeSkew) Recover(pid int) error {
	zeroSkew := NewTimeSkewWithCustomFakeImage(0, 0, it.clockIDsMask, it.fakeImage)
	return zeroSkew.Inject(pid)
}

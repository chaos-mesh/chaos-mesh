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

package generic

import (
	"crypto/rand"
	"math"
	"math/big"
	"strconv"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// FilterObjectsByMode filters objects by mode
func FilterObjectsByMode(mode v1alpha1.SelectorMode, value string, count int) ([]uint, error) {
	if count == 0 {
		return nil, errors.New("cannot generate objects from empty list")
	}

	switch mode {
	case v1alpha1.OneMode:
		index := getRandomNumber(count)
		return []uint{uint(index)}, nil
	case v1alpha1.AllMode:
		return RandomFixedIndexes(0, uint(count), uint(count)+1), nil
	case v1alpha1.FixedMode:
		num, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		if count < num {
			num = count
		}

		if num <= 0 {
			return nil, errors.New("cannot select any object as value below or equal 0")
		}

		return RandomFixedIndexes(0, uint(count), uint(num)), nil
	case v1alpha1.FixedPercentMode:
		percentage, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		if percentage == 0 {
			return nil, errors.New("cannot select any object as value below or equal 0")
		}

		if percentage < 0 || percentage > 100 {
			return nil, errors.Errorf("fixed percentage value of %d is invalid, Must be (0,100]", percentage)
		}

		// at least one object should be selected
		num := int(math.Ceil(float64(count) * float64(percentage) / 100))

		return RandomFixedIndexes(0, uint(count), uint(num)), nil
	case v1alpha1.RandomMaxPercentMode:
		maxPercentage, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}

		if maxPercentage == 0 {
			return nil, errors.New("cannot select any object as value below or equal 0")
		}

		if maxPercentage < 0 || maxPercentage > 100 {
			return nil, errors.Errorf("fixed percentage value of %d is invalid, Must be [0-100]", maxPercentage)
		}

		percentage := getRandomNumber(maxPercentage + 1) // + 1 because Intn works with half open interval [0,n) and we want [0,n]
		num := int(math.Ceil(float64(count) * float64(percentage) / 100))

		return RandomFixedIndexes(0, uint(count), uint(num)), nil
	default:
		return nil, errors.Errorf("mode %s not supported", mode)
	}
}

// RandomFixedIndexes returns the `count` random indexes between `start` and `end`.
// [start, end)
func RandomFixedIndexes(start, end, count uint) []uint {
	var indexes []uint
	m := make(map[uint]uint, count)

	if end < start {
		return indexes
	}

	if count > end-start {
		for i := start; i < end; i++ {
			indexes = append(indexes, i)
		}

		return indexes
	}

	for i := 0; i < int(count); {
		index := uint(getRandomNumber(int(end-start))) + start

		_, exist := m[index]
		if exist {
			continue
		}

		m[index] = index
		indexes = append(indexes, index)
		i++
	}

	return indexes
}

func getRandomNumber(max int) uint64 {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return num.Uint64()
}

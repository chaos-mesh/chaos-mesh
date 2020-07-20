// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package netutils

import (
	"crypto/sha1"
	"fmt"
)

func CompressName(originalName string, targetLength int, namePostFix string) (name string) {
	if len(originalName) < 6 {
		name = originalName + "_" + namePostFix
	} else {
		namePrefix := originalName[0:5]
		nameRest := originalName[5:]

		hasher := sha1.New()
		hasher.Write([]byte(nameRest))
		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))

		// keep the length does not exceed targetLength
		name = namePrefix + "_" + hashValue[0:targetLength-7-len(namePostFix)] + "_" + namePostFix
	}

	return
}

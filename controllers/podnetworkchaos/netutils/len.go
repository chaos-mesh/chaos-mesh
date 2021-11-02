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

package netutils

import (
	"crypto/sha256"
	"fmt"
	"log"
)

// CompressName compresses name to targetLength with specified postfix
// targetLength < 7 or targetLength-7-len(namePostFix) < 0 are not allowed
func CompressName(originalName string, targetLength int, namePostFix string) (name string) {
	if targetLength < 7 {
		log.Fatal("targetLength shouldn't be less than 7")
	}
	if targetLength-7-len(namePostFix) < 0 {
		log.Fatalf("namePostFix longer than (targetLength-7) = %d: %s", targetLength-7, namePostFix)
	}

	if len(originalName) < 6 {
		// len(originalName) < 6 && 7 + len(namePostFix) < targetlength
		// => 1 + len(originalName) + len(namePostFix) < targetLength
		name = originalName + "_" + namePostFix
		return
	}

	namePrefix := originalName[0:5]
	nameRest := originalName[5:]

	hasher := sha256.New()
	hasher.Write([]byte(nameRest))
	hashValue := fmt.Sprintf("%x", hasher.Sum(nil))

	// keep the length does not exceed targetLength
	name = namePrefix + "_" + hashValue[0:targetLength-7-len(namePostFix)] + "_" + namePostFix

	return
}

// Copyright 2020 PingCAP, Inc.
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

package main

import (
	"fmt"
	"time"
)

// This program is for testing time chaos in the e2e test
func main() {
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-ticker.C:
			now := time.Now().Format(time.RFC3339Nano)
			fmt.Println(now)
		}
	}
}

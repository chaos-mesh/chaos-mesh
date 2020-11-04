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

package timechaos

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// get pod current time in nanosecond
func getPodTimeNS(c http.Client, port uint16) (*time.Time, error) {
	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/time", port))
	if err != nil {
		return nil, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	t, err := time.Parse(time.RFC3339Nano, string(out))
	if err != nil {
		return nil, err
	}
	return &t, nil
}

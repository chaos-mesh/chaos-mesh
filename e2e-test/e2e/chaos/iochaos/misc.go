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

package iochaos

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// get pod io delay
func getPodIODelay(c http.Client, port uint16) (time.Duration, error) {
	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/io", port))
	if err != nil {
		return 0, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}

	result := string(out)
	if strings.Contains(result, "failed to write file") {
		return 0, errors.New(result)
	}
	dur, err := time.ParseDuration(result)
	if err != nil {
		return 0, err
	}

	return dur, nil
}

func getPodIoMistake(c http.Client, port uint16) (bool, error) {
	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/mistake", port))
	if err != nil {
		return false, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return false, err
	}

	result := string(out)
	fmt.Println("e2e server io mistake test response: ", resp, result)
	if strings.Contains(result, "true") {
		return true, nil
	}
	if strings.Contains(result, "false") {
		return false, nil
	}
	if strings.Contains(result, "err") {
		return false, errors.New(result)
	}
	return false, errors.New("unexpected reply from e2e server")
}

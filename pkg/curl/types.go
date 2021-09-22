// Copyright 2021 Chaos Mesh Authors.
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

package curl

import "net/http"

const HeaderContentType = "Content-Type"
const ApplicationJson = "application/json"

// RequestFlags should contain all the fields shown on frontend, and it could be parsed from flags of curl command line.
type RequestFlags struct {
	Name           string      `json:"name"`
	Method         string      `json:"method"`
	URL            string      `json:"url"`
	Header         http.Header `json:"header"`
	Body           string      `json:"body"`
	FollowLocation bool        `json:"followLocation"`
	JsonContent    bool        `json:"jsonContent"`
}

type Commands []string

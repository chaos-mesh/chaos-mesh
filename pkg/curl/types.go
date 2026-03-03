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

package curl

const HeaderContentType = "Content-Type"
const ApplicationJson = "application/json"

// RequestForm should contain all the fields shown on frontend
type RequestForm struct {
	CommandFlags
	Name string `json:"name"`
}

// Header is copied from http.Header, for speed up swagger_spec code generator without --parseDependency
type Header map[string][]string

// CommandFlags could be parsed from flags of curl command line.
type CommandFlags struct {
	Method         string `json:"method"`
	URL            string `json:"url"`
	Header         Header `json:"header"`
	Body           string `json:"body"`
	FollowLocation bool   `json:"followLocation"`
	JsonContent    bool   `json:"jsonContent"`
}

type Commands []string

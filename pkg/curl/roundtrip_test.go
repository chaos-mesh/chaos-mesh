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

import (
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
)

// Maybe we need some fuzzing test

func TestRoundTrip(t *testing.T) {
	RegisterTestingT(t)
	tests := []struct {
		name  string
		flags CommandFlags
	}{
		{
			name: "simple get",
			flags: CommandFlags{
				Method:         http.MethodGet,
				URL:            "github.com/chaos-mesh/chaos-mesh",
				Header:         nil,
				Body:           "",
				FollowLocation: false,
				JsonContent:    false,
			},
		}, {
			name: "get with header",
			flags: CommandFlags{
				Method: http.MethodGet,
				URL:    "https://github.com/chaos-mesh/chaos-mesh",
				Header: Header{
					"User-Agent": []string{"Go-http-client/1.1"},
				},
				Body:           "",
				FollowLocation: false,
				JsonContent:    false,
			},
		}, {
			name: "post json",
			flags: CommandFlags{
				Method:         http.MethodPost,
				URL:            "https://jsonplaceholder.typicode.com/posts",
				Header:         nil,
				Body:           "{\"foo\": \"bar\"}",
				FollowLocation: false,
				JsonContent:    true,
			},
		}, {
			name: "post json with custom header",
			flags: CommandFlags{
				Method: http.MethodPost,
				URL:    "https://jsonplaceholder.typicode.com/posts",
				Header: Header{
					"User-Agent": []string{"Go-http-client/1.1"},
				},
				Body:           "{\"foo\": \"bar\"}",
				FollowLocation: false,
				JsonContent:    true,
			},
		}, {
			name: "get with following location",
			flags: CommandFlags{
				Method:         http.MethodGet,
				URL:            "www.google.com",
				Header:         nil,
				Body:           "",
				FollowLocation: true,
				JsonContent:    false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands, err := renderCommands(test.flags)
			Expect(err).ShouldNot(HaveOccurred())
			parsedFlags, err := parseCommands(commands)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(parsedFlags).To(Equal(&test.flags), "rendered commands %+v", commands)
		})
	}
}

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

import (
	"fmt"
	"net/http"
	"strings"
)

// some example usage of RenderCommands
// notice that the output could not be used in shell directly, you need quotes and escape
func ExampleRenderCommands() {
	commands, _ := RenderCommands(RequestFlags{
		Method:         http.MethodGet,
		URL:            "https://github.com/chaos-mesh/chaos-mesh",
		Header:         nil,
		Body:           "",
		FollowLocation: true,
		JsonContent:    false,
	})

	fmt.Println(strings.Join(commands, " "))
	// Output: curl -i -s -L https://github.com/chaos-mesh/chaos-mesh
}

func ExampleRenderCommands_withCustomHeader() {
	commands, _ := RenderCommands(RequestFlags{
		Method: http.MethodGet,
		URL:    "https://github.com/chaos-mesh/chaos-mesh",
		Header: http.Header{
			"User-Agent": []string{"Go-http-client/1.1"},
		},
		Body:           "",
		FollowLocation: true,
		JsonContent:    false,
	})

	fmt.Println(strings.Join(commands, " "))
	// Output: curl -i -s -L -H User-Agent: Go-http-client/1.1 https://github.com/chaos-mesh/chaos-mesh
}

func ExampleRenderCommands_postJson() {
	commands, _ := RenderCommands(RequestFlags{
		Method:         http.MethodPost,
		URL:            "https://jsonplaceholder.typicode.com/posts",
		Header:         nil,
		Body:           "{\"foo\": \"bar\"}",
		FollowLocation: false,
		JsonContent:    true,
	})

	fmt.Println(strings.Join(commands, " "))
	// Output: curl -i -s -X POST -d {"foo": "bar"} -H Content-Type: application/json https://jsonplaceholder.typicode.com/posts
}

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

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

	flag "github.com/spf13/pflag"
)

func ParseCommands(command Commands) (RequestFlags, error) {
	flagset := flag.NewFlagSet("curl", flag.ContinueOnError)
	flagset.ParseErrorsWhitelist.UnknownFlags = true

	// these parts of flags are referenced to the manual of curl
	flagset.BoolP("silent", "s", true, "silent mode")
	location := flagset.BoolP("location", "L", false, "follow the location")
	requestMethod := flagset.StringP("request", "X", "GET", "request method")
	rawHeader := flagset.StringArrayP("header", "H", []string{}, "HTTP extra header")
	data := flagset.StringP("data", "d", "", "data")
	err := flagset.Parse(command)

	// first non-flag arg is the command itself, use the second non-flag arg as the url.
	if flag.NArg() > 1 {
		return RequestFlags{}, fmt.Errorf("can not find the url")
	}
	url := flagset.Arg(1)

	if err != nil {
		return RequestFlags{}, nil
	}

	isJson := false
	var header http.Header
	if len(*rawHeader) > 0 {
		header = http.Header{}
	}
	for _, item := range *rawHeader {
		k, v := parseHeader(item)
		if k == HeaderContentType && v == ApplicationJson {
			isJson = true
			continue
		}
		header[k] = append(header[k], v)
	}
	if len(header) == 0 {
		header = nil
	}

	return RequestFlags{
		Method:         *requestMethod,
		URL:            url,
		Header:         header,
		Body:           *data,
		FollowLocation: *location,
		JsonContent:    isJson,
	}, nil
}

func ParseWorkflowTaskTemplate(template *v1alpha1.Template) (RequestFlags, error) {

	return RequestFlags{}, nil
}

func parseHeader(headerKV string) (string, string) {
	substring := strings.SplitN(headerKV, ":", 2)
	return strings.TrimSpace(substring[0]), strings.TrimSpace(substring[1])
}

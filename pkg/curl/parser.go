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
	"strings"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func parseCommands(command Commands) (*CommandFlags, error) {
	flagset := flag.NewFlagSet("curl", flag.ContinueOnError)
	flagset.ParseErrorsWhitelist.UnknownFlags = true

	// these parts of flags are referenced to the manual of curl
	flagset.BoolP("silent", "s", true, "silent mode")
	location := flagset.BoolP("location", "L", false, "follow the location")
	requestMethod := flagset.StringP("request", "X", "GET", "request method")
	rawHeader := flagset.StringArrayP("header", "H", []string{}, "HTTP extra header")
	data := flagset.StringP("data", "d", "", "data")
	err := flagset.Parse(command)
	if err != nil {
		return nil, err
	}

	// first non-flag arg is the command itself, use the second non-flag arg as the url.
	if flag.NArg() > 1 {
		return nil, errors.New("can not find the url")
	}
	url := flagset.Arg(1)

	isJson := false
	var header Header
	if len(*rawHeader) > 0 {
		header = Header{}
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

	return &CommandFlags{
		Method:         *requestMethod,
		URL:            url,
		Header:         header,
		Body:           *data,
		FollowLocation: *location,
		JsonContent:    isJson,
	}, nil
}

func ParseWorkflowTaskTemplate(template *v1alpha1.Template) (*RequestForm, error) {
	if !IsValidRenderedTask(template) {
		return nil, errors.New("invalid request, this task is not rendered by curl-render")
	}
	parsedFlags, err := parseCommands(template.Task.Container.Command)
	if err != nil {
		return nil, err
	}
	return &RequestForm{
		CommandFlags: *parsedFlags,
		Name:         template.Name,
	}, nil
}

func IsValidRenderedTask(template *v1alpha1.Template) bool {
	return template.Type == v1alpha1.TypeTask && strings.HasSuffix(template.Task.Container.Name, nameSuffix)
}

func parseHeader(headerKV string) (key, value string) {
	kv := strings.SplitN(headerKV, ":", 2)

	key = strings.TrimSpace(kv[0])
	value = strings.TrimSpace(kv[1])
	return
}

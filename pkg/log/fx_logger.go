// Copyright 2022 Chaos Mesh Authors.
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

package log

import (
	"fmt"

	"github.com/go-logr/logr"
)

type LogrPrinter struct {
	logger logr.Logger
}

func NewLogrPrinter(logger logr.Logger) *LogrPrinter {
	return &LogrPrinter{logger: logger}
}

func (it *LogrPrinter) Printf(s string, i ...interface{}) {
	it.logger.
		// Here are 2 level wrapper for this logger, one is LogrPrinter, another is fxlog.Logger,
		// so we use 2 here. It's a little tricky but would make fx logging better.
		WithCallDepth(2).
		Info(fmt.Sprintf(s, i...))
}

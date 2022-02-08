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
//

package log

import (
	"sync"

	"github.com/go-logr/logr"
)

// the way for management and access global logger is referenced from zap.
var (
	globalMu     sync.RWMutex
	globalLogger = logr.Discard()
)

// L is the way to access the global logger. You could use it if there are no logger in your code context. Please notice
// that the default value of "global logger" is a "discard logger" which means that all logs will be ignored. Make sure
// that initialize the global logger by ReplaceGlobals before using it, for example, calling ReplaceGlobals at the beginning
// of your main function.
//
// Do NOT save the global logger to a variable for long-term using, because it is possible that the global logger is
// replaced by another. Keep calling L at each time.
//
// Deprecated: Do not use global logger anymore. For more detail, see
// https://github.com/chaos-mesh/rfcs/blob/main/text/2021-12-09-logging.md#global-logger
func L() logr.Logger {
	globalMu.RLock()
	result := globalLogger
	globalMu.RUnlock()
	return result
}

// ReplaceGlobals would replace the global logger with the given logger. It should be used when your application starting.
//
// Deprecated: Do not use global logger anymore. For more detail, see
// https://github.com/chaos-mesh/rfcs/blob/main/text/2021-12-09-logging.md#global-logger
func ReplaceGlobals(logger logr.Logger) {
	globalMu.Lock()
	globalLogger = logger
	globalMu.Unlock()
}

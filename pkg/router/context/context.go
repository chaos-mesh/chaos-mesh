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

package context

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Context is the running context for a controller
type Context struct {
	client.Client
	client.Reader
	record.EventRecorder
	Log logr.Logger
}

// LogWithValues appends values for logger in the context
func (c *Context) LogWithValues(keysAndValues ...interface{}) Context {
	return Context{
		Client:        c.Client,
		Reader:        c.Reader,
		EventRecorder: c.EventRecorder,
		Log:           c.Log.WithValues(keysAndValues...),
	}
}

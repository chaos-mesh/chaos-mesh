// Copyright 2019 Chaos Mesh Authors.
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

package watcher

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config"
)

// Message is a message that describes a change and payload to a sidecar configuration
type Message struct {
	Event           Event
	InjectionConfig config.InjectionConfig
}

// Event is what happened to the config (add/delete/update)
type Event uint8

const (
	// EventAdd is a new ConfigMap
	EventAdd Event = iota
	// EventUpdate is an Updated ConfigMap
	EventUpdate
	// EventDelete is a deleted ConfigMap
	EventDelete
)

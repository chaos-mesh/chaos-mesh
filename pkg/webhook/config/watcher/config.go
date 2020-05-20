// Copyright 2019 PingCAP, Inc.
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
	"github.com/pingcap/errors"
)

// Config is a configuration struct for the Watcher type
type Config struct {
	Namespace      string
	TemplateLabels map[string]string
	ConfigLabels   map[string]string
}

// NewConfig returns a new initialized Config
func NewConfig() *Config {
	return &Config{
		Namespace:      "",
		TemplateLabels: map[string]string{},
		ConfigLabels:   map[string]string{},
	}
}

// InitLabels initializes labels in Config
func (c *Config) InitLabels(templateLabels, confLabels map[string]string) error {
	if len(templateLabels) == 0 {
		return errors.New("template labels must be set")
	}
	if len(confLabels) == 0 {
		return errors.New("conf labels must be set")
	}
	c.TemplateLabels, c.ConfigLabels = templateLabels, confLabels
	return nil
}

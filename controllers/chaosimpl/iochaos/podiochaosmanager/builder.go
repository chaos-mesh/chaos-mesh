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

package podiochaosmanager

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Builder struct {
	Log logr.Logger
	client.Client
	client.Reader
	scheme *runtime.Scheme
}

func NewBuilder(logger logr.Logger, client client.Client, reader client.Reader, scheme *runtime.Scheme) *Builder {
	return &Builder{
		Log:    logger,
		Client: client,
		Reader: reader,
		scheme: scheme,
	}
}

func (b *Builder) WithInit(source string, key types.NamespacedName) *PodIOManager {
	t := &PodIOTransaction{}
	t.Clear(source)

	return &PodIOManager{
		Source: source,
		Log:    b.Log,
		Client: b.Client,
		Reader: b.Reader,
		scheme: b.scheme,

		Key: key,
		T:   t,
	}
}

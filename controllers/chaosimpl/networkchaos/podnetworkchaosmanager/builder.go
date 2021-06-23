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

package podnetworkchaosmanager

import (
	"github.com/go-logr/logr"
	"go.uber.org/fx"
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

type Params struct {
	fx.In

	Logger logr.Logger
	Client client.Client
	Reader client.Reader `name:"no-cache"`
	Scheme *runtime.Scheme
}

func NewBuilder(params Params) *Builder {
	return &Builder{
		Log:    params.Logger,
		Client: params.Client,
		Reader: params.Reader,
		scheme: params.Scheme,
	}
}

func (b *Builder) Build(source string, key types.NamespacedName) *PodNetworkManager {
	t := &PodNetworkTransaction{}

	return &PodNetworkManager{
		Source: source,
		Log:    b.Log,
		Client: b.Client,
		Reader: b.Reader,
		scheme: b.scheme,

		Key: key,
		T:   t,
	}
}

func (b *Builder) WithInit(source string, key types.NamespacedName) *PodNetworkManager {
	t := &PodNetworkTransaction{}
	t.Clear(source)

	return &PodNetworkManager{
		Source: source,
		Log:    b.Log,
		Client: b.Client,
		Reader: b.Reader,
		scheme: b.scheme,

		Key: key,
		T:   t,
	}
}

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

package helm

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

var _ genericclioptions.RESTClientGetter = &restClientGetter{}

type restClientGetter struct {
	clientConfig clientcmd.ClientConfig
}

func NewRESTClientGetter(clientConfig clientcmd.ClientConfig) genericclioptions.RESTClientGetter {
	return &restClientGetter{
		clientConfig,
	}
}

func (getter *restClientGetter) ToRESTConfig() (*rest.Config, error) {
	return getter.clientConfig.ClientConfig()
}

func (getter *restClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	restConfig, err := getter.clientConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "get rest config from client config")
	}

	client, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "create discovery client")
	}
	return memory.NewMemCacheClient(client), nil
}

func (getter *restClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := getter.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

func (getter *restClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return getter.clientConfig
}

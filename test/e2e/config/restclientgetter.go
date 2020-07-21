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

package config

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	diskcached "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

var defaultCacheDir = filepath.Join(homedir.HomeDir(), ".kube", "http-cache")

// simpleRestClientGetter implements genericclioptions.RESTClientGetter
type simpleRestClientGetter struct {
	clientcmdapi.Config
}

// ToRESTConfig returns restconfig
func (getter *simpleRestClientGetter) ToRESTConfig() (*rest.Config, error) {
	return getter.ToRawKubeConfigLoader().ClientConfig()
}

// ToDiscoveryClient returns discovery client
func (getter *simpleRestClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := getter.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	config.Burst = 100

	httpCacheDir := defaultCacheDir
	discoveryCacheDir := computeDiscoverCacheDir(filepath.Join(homedir.HomeDir(), ".kube", "cache", "discovery"), config.Host)

	return diskcached.NewCachedDiscoveryClientForConfig(config, discoveryCacheDir, httpCacheDir, time.Duration(10*time.Minute))
}

// ToRESTMapper returns a restmapper
func (getter *simpleRestClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := getter.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

// ToRawKubeConfigLoader return kubeconfig loader as-is
func (getter *simpleRestClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return clientcmd.NewDefaultClientConfig(getter.Config, &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults})
}

// overlyCautiousIllegalFileCharacters matches characters that *might* not be supported.  Windows is really restrictive, so this is really restrictive
var overlyCautiousIllegalFileCharacters = regexp.MustCompile(`[^(\w/\.)]`)

// computeDiscoverCacheDir takes the parentDir and the host and comes up with a "usually non-colliding" name.
func computeDiscoverCacheDir(parentDir, host string) string {
	// strip the optional scheme from host if its there:
	schemelessHost := strings.Replace(strings.Replace(host, "https://", "", 1), "http://", "", 1)
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.  Even if we do collide the problem is short lived
	safeHost := overlyCautiousIllegalFileCharacters.ReplaceAllString(schemelessHost, "_")
	return filepath.Join(parentDir, safeHost)
}

// NewSimpleRESTClientGetter initializes a new genericclioptions.RESTClientGetter from clientcmdapi.Config.
func NewSimpleRESTClientGetter(config clientcmdapi.Config) genericclioptions.RESTClientGetter {
	return &simpleRestClientGetter{config}
}

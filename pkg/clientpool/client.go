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

package clientpool

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	pkgclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// K8sClients is an object of Clients
var K8sClients *Clients

// Clients is the client pool of k8s client
type Clients struct {
	sync.Mutex

	scheme      *runtime.Scheme
	localConfig *rest.Config
	clients     *lru.Cache
}

// New creates a new Clients
func New(localConfig *rest.Config, scheme *runtime.Scheme, maxClientNum int) (*Clients, error) {
	clients, err := lru.New(maxClientNum)
	if err != nil {
		return nil, err
	}

	return &Clients{
		localConfig: localConfig,
		scheme:      scheme,
		clients:     clients,
	}, nil
}

// Client returns a k8s client according to the token
func (c *Clients) Client(token string) (pkgclient.Client, error) {
	c.Lock()
	defer c.Unlock()

	if len(token) == 0 {
		return nil, errors.New("token is empty")
	}

	value, ok := c.clients.Get(token)
	if ok {
		return value.(pkgclient.Client), nil
	}

	config := rest.CopyConfig(c.localConfig)
	config.BearerToken = token
	config.BearerTokenFile = ""

	newFunc := pkgclient.New

	if mockNew := mock.On("MockCreateK8sClient"); mockNew != nil {
		newFunc = mockNew.(func(config *rest.Config, options pkgclient.Options) (pkgclient.Client, error))
	}

	client, err := newFunc(config, pkgclient.Options{
		Scheme: c.scheme,
	})
	if err != nil {
		return nil, err
	}

	_ = c.clients.Add(token, client)

	return client, nil
}

// ExtractTokenFromHeader extracts token from http header
func ExtractTokenFromHeader(header http.Header) string {
	auth := header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

// ExtractTokenAndGetClient extracts token from http header, and get the k8s client of this token
func ExtractTokenAndGetClient(header http.Header) (pkgclient.Client, error) {
	token := ExtractTokenFromHeader(header)
	return K8sClients.Client(token)
}

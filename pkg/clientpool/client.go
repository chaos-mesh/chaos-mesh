// Copyright 2021 Chaos Mesh Authors.
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

package clientpool

import (
	"net/http"
	"strings"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	pkgclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// K8sClients is an object of Clients
var K8sClients Clients

type Clients interface {
	Client(token string) (pkgclient.Client, error)
	AuthClient(token string) (authorizationv1.AuthorizationV1Interface, error)
	Num() int
	Contains(token string) bool
}

type LocalClient struct {
	client     pkgclient.Client
	authClient authorizationv1.AuthorizationV1Interface
}

func NewLocalClient(localConfig *rest.Config, scheme *runtime.Scheme) (Clients, error) {
	client, err := pkgclient.New(localConfig, pkgclient.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	authCli, err := authorizationv1.NewForConfig(localConfig)
	if err != nil {
		return nil, err
	}

	return &LocalClient{
		client:     client,
		authClient: authCli,
	}, nil
}

// Client returns the local k8s client
func (c *LocalClient) Client(token string) (pkgclient.Client, error) {
	return c.client, nil
}

func (c *LocalClient) AuthClient(token string) (authorizationv1.AuthorizationV1Interface, error) {
	return c.authClient, nil
}

// Num returns the num of clients
func (c *LocalClient) Num() int {
	return 1
}

// Contains return false for LocalClient
func (c *LocalClient) Contains(token string) bool {
	return false
}

// Clients is the client pool of k8s client
type ClientsPool struct {
	sync.RWMutex

	scheme      *runtime.Scheme
	localConfig *rest.Config
	clients     *lru.Cache
	authClients *lru.Cache
}

// New creates a new Clients
func NewClientPool(localConfig *rest.Config, scheme *runtime.Scheme, maxClientNum int) (Clients, error) {
	clients, err := lru.New(maxClientNum)
	if err != nil {
		return nil, err
	}

	authClients, err := lru.New(maxClientNum)
	if err != nil {
		return nil, err
	}

	return &ClientsPool{
		localConfig: localConfig,
		scheme:      scheme,
		clients:     clients,
		authClients: authClients,
	}, nil
}

// Client returns a k8s client according to the token
func (c *ClientsPool) Client(token string) (pkgclient.Client, error) {
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

func (c *ClientsPool) AuthClient(token string) (authorizationv1.AuthorizationV1Interface, error) {
	c.Lock()
	defer c.Unlock()

	if len(token) == 0 {
		return nil, errors.New("token is empty")
	}

	value, ok := c.authClients.Get(token)
	if ok {
		return value.(authorizationv1.AuthorizationV1Interface), nil
	}

	config := rest.CopyConfig(c.localConfig)
	config.BearerToken = token
	config.BearerTokenFile = ""

	authCli, err := authorizationv1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	_ = c.authClients.Add(token, authCli)

	return authCli, nil
}

// Num returns the num of clients
func (c *ClientsPool) Num() int {
	return c.clients.Len()
}

// Contains return true if have client for the token
func (c *ClientsPool) Contains(token string) bool {
	c.RLock()
	defer c.RUnlock()

	_, ok := c.clients.Get(token)
	return ok
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

// ExtractTokenAndGetAuthClient extracts token from http header, and get the authority client of this token
func ExtractTokenAndGetAuthClient(header http.Header) (authorizationv1.AuthorizationV1Interface, error) {
	token := ExtractTokenFromHeader(header)
	return K8sClients.AuthClient(token)
}

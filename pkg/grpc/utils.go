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

package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"
)

// DefaultRPCTimeout specifies default timeout of RPC between controller and chaos-operator
const DefaultRPCTimeout = 60 * time.Second

// RPCTimeout specifies timeout of RPC between controller and chaos-operator
var RPCTimeout = DefaultRPCTimeout

var log = ctrl.Log.WithName("util")

type GrpcBuilder struct {
	options []grpc.DialOption
	address string
	port    int
	err     error
}

func Builder() *GrpcBuilder {
	return &GrpcBuilder{options: []grpc.DialOption{}, address: "localhost"}
}

func (it *GrpcBuilder) WithDefaultTimeout() *GrpcBuilder {
	it.options = append(it.options, grpc.WithUnaryInterceptor(TimeoutClientInterceptor(DefaultRPCTimeout)))
	return it
}

func (it *GrpcBuilder) WithTimeout(timeout time.Duration) *GrpcBuilder {
	it.options = append(it.options, grpc.WithUnaryInterceptor(TimeoutClientInterceptor(timeout)))
	return it
}

func (it *GrpcBuilder) Address(address string) *GrpcBuilder {
	it.address = address
	return it
}

func (it *GrpcBuilder) Port(port int) *GrpcBuilder {
	it.port = port
	return it
}

func (it *GrpcBuilder) Insecure() *GrpcBuilder {
	it.options = append(it.options, grpc.WithInsecure())
	return it
}

func (it *GrpcBuilder) TryTLSFromFiles(caCertPath string, certPath string, keyPath string) *GrpcBuilder {
	if caCertPath == "" || certPath == "" || keyPath != "" {
		return it.Insecure()
	}
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return it.Insecure()
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return it.Insecure()
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
		ServerName:   "chaos-daemon.chaos-mesh.org",
	})
	it.options = append(it.options, grpc.WithTransportCredentials(creds))
	return it
}

func (it *GrpcBuilder) TLSFromFiles(caCertPath string, certPath string, keyPath string) *GrpcBuilder {
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		if it.err == nil {
			it.err = err
		}
		return it
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		if it.err == nil {
			it.err = err
		}
		return it
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
		ServerName:   "chaos-daemon.chaos-mesh.org",
	})
	it.options = append(it.options, grpc.WithTransportCredentials(creds))
	return it
}

func (it *GrpcBuilder) TLSFromRaw(caCert []byte, cert []byte, key []byte) *GrpcBuilder {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		if it.err == nil {
			it.err = err
		}
		return it
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
		ServerName:   "chaos-daemon.chaos-mesh.org",
	})
	it.options = append(it.options, grpc.WithTransportCredentials(creds))
	return it
}

func (it *GrpcBuilder) Build() (*grpc.ClientConn, error) {
	if it.err != nil {
		return nil, it.err
	}
	return grpc.Dial(fmt.Sprintf("%s:%d", it.address, it.port), it.options...)
}

func (it *GrpcBuilder) GetError() error {
	return it.err
}

// TimeoutClientInterceptor wraps the RPC with a timeout.
func TimeoutClientInterceptor(timeout time.Duration) func(context.Context, string, interface{}, interface{},
	*grpc.ClientConn, grpc.UnaryInvoker, ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// TimeoutServerInterceptor ensures the context is intact before handling over the
// request to application.
func TimeoutServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return handler(ctx, req)
}

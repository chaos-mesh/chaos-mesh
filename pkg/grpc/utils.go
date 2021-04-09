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

type TLSRaw struct {
	CaCert []byte
	Cert   []byte
	Key    []byte
}

type TLSFile struct {
	CaCert string
	Cert   string
	Key    string
}

type TLSFromType = string

const (
	RAW  TLSFromType = "RAW"
	FILE TLSFromType = "FILE"
)

type CredentialProvider struct {
	raw      TLSRaw
	file     TLSFile
	insecure bool
	fromType TLSFromType
}

func (it *CredentialProvider) getCredentialOption() (grpc.DialOption, error) {
	if it.insecure {
		return grpc.WithInsecure(), nil
	}
	if it.fromType == RAW {
		return it.TLSFromRaw()
	}
	if it.fromType == FILE {
		return it.TLSFromFile()
	}

	return nil, fmt.Errorf("an authorization method must be specified")
}

func (it *CredentialProvider) TLSFromFile() (grpc.DialOption, error) {
	caCert, err := ioutil.ReadFile(it.file.CaCert)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCert, err := tls.LoadX509KeyPair(it.file.Cert, it.file.Key)
	if err != nil {
		return nil, err
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
		ServerName:   "chaos-daemon.chaos-mesh.org",
	})
	return grpc.WithTransportCredentials(creds), nil
}

func (it *CredentialProvider) TLSFromRaw() (grpc.DialOption, error) {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(it.raw.CaCert)

	clientCert, err := tls.X509KeyPair(it.raw.Cert, it.raw.Key)
	if err != nil {
		return nil, err
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
		ServerName:   "chaos-daemon.chaos-mesh.org",
	})
	return grpc.WithTransportCredentials(creds), nil
}

type GrpcBuilder struct {
	options            []grpc.DialOption
	credentialProvider CredentialProvider
	address            string
	port               int
}

func Builder(address string, port int) *GrpcBuilder {
	return &GrpcBuilder{options: []grpc.DialOption{}, address: address, port: port}
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
	it.credentialProvider.insecure = true
	return it
}

func (it *GrpcBuilder) TLSFromRaw(caCert []byte, cert []byte, key []byte) *GrpcBuilder {
	it.credentialProvider.insecure = false
	it.credentialProvider.fromType = RAW
	it.credentialProvider.raw = TLSRaw{
		CaCert: caCert,
		Cert:   cert,
		Key:    key,
	}
	return it
}

func (it *GrpcBuilder) TLSFromFile(caCertPath string, certPath string, keyPath string) *GrpcBuilder {
	it.credentialProvider.insecure = false
	it.credentialProvider.fromType = FILE
	it.credentialProvider.file = TLSFile{
		CaCert: caCertPath,
		Cert:   certPath,
		Key:    keyPath,
	}
	return it
}

func (it *GrpcBuilder) Build() (*grpc.ClientConn, error) {
	option, err := it.credentialProvider.getCredentialOption()
	if err != nil {
		return nil, err
	}
	it.options = append(it.options, option)
	return grpc.Dial(fmt.Sprintf("%s:%d", it.address, it.port), it.options...)
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

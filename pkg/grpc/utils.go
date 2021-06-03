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
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// DefaultRPCTimeout specifies default timeout of RPC between controller and chaos-operator
const DefaultRPCTimeout = 60 * time.Second

// RPCTimeout specifies timeout of RPC between controller and chaos-operator
var RPCTimeout = DefaultRPCTimeout

const ChaosDaemonServerName = "chaos-daemon.chaos-mesh.org"

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

type FileProvider struct {
	file TLSFile
}

type RawProvider struct {
	raw TLSRaw
}

type InsecureProvider struct {
}

type CredentialProvider interface {
	getCredentialOption() (grpc.DialOption, error)
}

func (it *FileProvider) getCredentialOption() (grpc.DialOption, error) {
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
		ServerName:   ChaosDaemonServerName,
	})
	return grpc.WithTransportCredentials(creds), nil
}

func (it *RawProvider) getCredentialOption() (grpc.DialOption, error) {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(it.raw.CaCert)

	clientCert, err := tls.X509KeyPair(it.raw.Cert, it.raw.Key)
	if err != nil {
		return nil, err
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
		ServerName:   ChaosDaemonServerName,
	})
	return grpc.WithTransportCredentials(creds), nil
}

func (it *InsecureProvider) getCredentialOption() (grpc.DialOption, error) {
	return grpc.WithInsecure(), nil
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

func (it *GrpcBuilder) Insecure() *GrpcBuilder {
	it.credentialProvider = &InsecureProvider{}
	return it
}

func (it *GrpcBuilder) TLSFromRaw(caCert []byte, cert []byte, key []byte) *GrpcBuilder {
	it.credentialProvider = &RawProvider{
		raw: TLSRaw{
			CaCert: caCert,
			Cert:   cert,
			Key:    key,
		},
	}

	return it
}

func (it *GrpcBuilder) TLSFromFile(caCertPath string, certPath string, keyPath string) *GrpcBuilder {
	it.credentialProvider = &FileProvider{
		file: TLSFile{
			CaCert: caCertPath,
			Cert:   certPath,
			Key:    keyPath,
		},
	}
	return it
}

func (it *GrpcBuilder) Build() (*grpc.ClientConn, error) {
	if it.credentialProvider == nil {
		return nil, fmt.Errorf("an authorization method must be specified")
	}
	option, err := it.credentialProvider.getCredentialOption()
	if err != nil {
		return nil, err
	}
	it.options = append(it.options, option)
	return grpc.Dial(net.JoinHostPort(it.address, strconv.Itoa(it.port)), it.options...)
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

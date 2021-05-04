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
	"net"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"
)

// DefaultRPCTimeout specifies default timeout of RPC between controller and chaos-operator
const DefaultRPCTimeout = 60 * time.Second

// RPCTimeout specifies timeout of RPC between controller and chaos-operator
var RPCTimeout = DefaultRPCTimeout

var log = ctrl.Log.WithName("util")

// CreateGrpcConnection create a grpc connection with given port and address
func CreateGrpcConnection(address string, port int, caCertPath string, certPath string, keyPath string) (*grpc.ClientConn, error) {
	if caCertPath != "" && certPath != "" && keyPath != "" {
		var caCert, cert, key []byte
		var err error
		caCert, err = ioutil.ReadFile(caCertPath)
		if err != nil {
			return nil, err
		}
		cert, err = ioutil.ReadFile(certPath)
		if err != nil {
			return nil, err
		}
		key, err = ioutil.ReadFile(keyPath)
		if err != nil {
			return nil, err
		}
		return CreateGrpcConnectionFromRaw(address, port, caCert, cert, key)
	}
	options := []grpc.DialOption{grpc.WithUnaryInterceptor(TimeoutClientInterceptor)}
	options = append(options, grpc.WithInsecure())
	conn, err := grpc.Dial(net.JoinHostPort(address, fmt.Sprintf("%d", port)), options...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// CreateGrpcConnectionFromRaw create a grpc connection with given port and address, and use raw data instead of file path
func CreateGrpcConnectionFromRaw(address string, port int, caCert []byte, cert []byte, key []byte) (*grpc.ClientConn, error) {
	options := []grpc.DialOption{grpc.WithUnaryInterceptor(TimeoutClientInterceptor)}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	clientCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
		ServerName:   "chaos-daemon.chaos-mesh.org",
	})
	options = append(options, grpc.WithTransportCredentials(creds))

	conn, err := grpc.Dial(net.JoinHostPort(address, fmt.Sprintf("%d", port)),
		options...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// TimeoutClientInterceptor wraps the RPC with a timeout.
func TimeoutClientInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, RPCTimeout)
	defer cancel()
	return invoker(ctx, method, req, reply, cc, opts...)
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

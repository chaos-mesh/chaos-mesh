// Copyright 2019 PingCAP, Inc.
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

package chaosfs

import (
	"context"
	"math/rand"
	"net"
	"os"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosfs/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("fuse-server")

//go:generate protoc -I pb pb/injure.proto --go_out=plugins=grpc:pb

var (
	faultMap sync.Map

	methods map[string]bool
)

func init() {
	faultMap = sync.Map{}
	initMethods()
}

type faultContext struct {
	errno  error
	random bool
	pct    uint32
	path   string
	delay  time.Duration
}

func initMethods() {
	methods = make(map[string]bool)
	methods["open"] = true
	methods["read"] = true
	methods["write"] = true
	methods["mkdir"] = true
	methods["rmdir"] = true
	methods["opendir"] = true
	methods["fsync"] = true
	methods["flush"] = true
	methods["release"] = true
	methods["truncate"] = true
	methods["getattr"] = true
	methods["chown"] = true
	methods["chmod"] = true
	methods["utimens"] = true
	methods["allocate"] = true
	methods["getlk"] = true
	methods["setlk"] = true
	methods["setlkw"] = true
	methods["statfs"] = true
	methods["readlink"] = true
	methods["symlink"] = true
	methods["create"] = true
	methods["access"] = true
	methods["link"] = true
	methods["mknod"] = true
	methods["rename"] = true
	methods["unlink"] = true
	methods["getxattr"] = true
	methods["listxattr"] = true
	methods["removexattr"] = true
	methods["setxattr"] = true
}

func randomErrno() error {
	// from E2BIG to EXFULL, notice linux only
	return syscall.Errno(rand.Intn(0x36-0x7) + 0x7)
}

func probab(percentage uint32) bool {
	return rand.Intn(99) < int(percentage)
}

func faultInject(path, method string) error {
	val, ok := faultMap.Load(method)
	if !ok {
		return nil
	}

	fc := val.(*faultContext)
	if !probab(fc.pct) {
		return nil
	}

	if len(fc.path) > 0 {
		re, err := regexp.Compile(fc.path)
		if err != nil {
			log.Error(err, "failed to parse path", "path: ", fc.path)
			return nil
		}
		if !re.MatchString(path) {
			return nil
		}
	}

	log.V(6).Info("Inject fault", "method", method, "path", path)
	log.V(6).Info("Inject fault", "context", fc)

	var errno error = nil
	if fc.errno != nil {
		errno = fc.errno
	} else if fc.random {
		errno = randomErrno()
	}

	if fc.delay > 0 {
		time.Sleep(fc.delay)
	}

	return errno
}

type server struct {
}

func (s *server) methods() []string {
	ms := make([]string, 0)
	for k := range methods {
		ms = append(ms, k)
	}
	return ms
}

func (s *server) Methods(ctx context.Context, in *empty.Empty) (*pb.Response, error) {
	return &pb.Response{Methods: s.methods()}, nil
}

func (s *server) RecoverAll(ctx context.Context, in *empty.Empty) (*empty.Empty, error) {
	log.Info("Recover all fault")
	faultMap.Range(func(k, v interface{}) bool {
		faultMap.Delete(k)
		return true
	})
	return &empty.Empty{}, nil
}

func (s *server) RecoverMethod(ctx context.Context, in *pb.Request) (*empty.Empty, error) {
	ms := in.GetMethods()
	for _, v := range ms {
		faultMap.Delete(v)
	}
	return &empty.Empty{}, nil
}

func (s *server) setFault(ms []string, f *faultContext) {
	for _, v := range ms {
		faultMap.Store(v, f)
	}
}

func (s *server) SetFault(ctx context.Context, in *pb.Request) (*empty.Empty, error) {
	// TODO: use Errno(0), and handle Errno(0) in Hook interfaces
	log.Info("Set fault", "request", in)

	var errno error = nil
	if in.Errno != 0 {
		errno = syscall.Errno(in.Errno)
	}
	f := &faultContext{
		errno:  errno,
		random: in.Random,
		pct:    in.Pct,
		path:   in.Path,
		delay:  time.Duration(in.Delay) * time.Microsecond,
	}

	s.setFault(in.Methods, f)
	return &empty.Empty{}, nil
}

func (s *server) SetFaultAll(ctx context.Context, in *pb.Request) (*empty.Empty, error) {
	// TODO: use Errno(0), and handle Errno(0) in Hook interfaces
	log.Info("Set fault all methods", "request", in)

	var errno error = nil
	if in.Errno != 0 {
		errno = syscall.Errno(in.Errno)
	}
	f := &faultContext{
		errno:  errno,
		random: in.Random,
		pct:    in.Pct,
		path:   in.Path,
		delay:  time.Duration(in.Delay) * time.Microsecond,
	}

	s.setFault(s.methods(), f)
	return &empty.Empty{}, nil
}

func StartServer(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error(err, "failed to listen tcp server", "address", addr)
		os.Exit(1)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(utils.TimeoutServerInterceptor))
	pb.RegisterInjureServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Error(err, "failed to start serve")
			os.Exit(1)
		}
	}()
}

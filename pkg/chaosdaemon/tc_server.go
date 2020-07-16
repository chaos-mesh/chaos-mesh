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

package chaosdaemon

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *daemonServer) AddQdisc(ctx context.Context, in *pb.QdiscRequest) (*empty.Empty, error) {
	log.Info("Add Qdisc", "Request", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	args, err := generateQdiscArgs("add", in.Qdisc)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate qdisc args error: %v", err)
	}

	if err := applyTc(ctx, pid, args...); err != nil {
		return nil, status.Errorf(codes.Internal, "tbf apply error: %v", err)
	}

	return &empty.Empty{}, nil
}

func (s *daemonServer) DelQdisc(ctx context.Context, in *pb.QdiscRequest) (*empty.Empty, error) {
	log.Info("Del Qdisc", "Request", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	args, err := generateQdiscArgs("del", in.Qdisc)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate qdisc args error: %v", err)
	}

	if err := applyTc(ctx, pid, args...); err != nil {
		return nil, status.Errorf(codes.Internal, "tbf apply error: %v", err)
	}

	return &empty.Empty{}, nil
}

func (s *daemonServer) AddEmatchFilter(ctx context.Context, in *pb.EmatchFilterRequest) (*empty.Empty, error) {
	log.Info("Add ematch filter", "Request", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	args := []string{"filter", "add", "dev", "eth0"}

	args = append(args, "parent", fmt.Sprintf("%d:%d", in.Filter.Parent.Major, in.Filter.Parent.Minor))

	args = append(args, "basic", "match", in.Filter.Match)

	args = append(args, "classid", fmt.Sprintf("%d:%d", in.Filter.Classid.Major, in.Filter.Classid.Minor))

	if err := applyTc(ctx, pid, args...); err != nil {
		return nil, status.Errorf(codes.Internal, "tbf apply error: %v", err)
	}

	return &empty.Empty{}, nil

}

func (s *daemonServer) DelTcFilter(ctx context.Context, in *pb.TcFilterRequest) (*empty.Empty, error) {
	log.Info("Del tc filter", "Request", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}

	args := []string{"filter", "del", "dev", "eth0"}

	args = append(args, "parent", fmt.Sprintf("%d:%d", in.Filter.Parent.Major, in.Filter.Parent.Minor))

	if err := applyTc(ctx, pid, args...); err != nil {
		return nil, status.Errorf(codes.Internal, "tbf apply error: %v", err)
	}

	return &empty.Empty{}, nil
}

func generateQdiscArgs(action string, qdisc *pb.Qdisc) ([]string, error) {

	if qdisc == nil {
		return nil, fmt.Errorf("qdisc is required")
	}

	if qdisc.Type == "" {
		return nil, fmt.Errorf("qdisc.Type is required")
	}

	args := []string{"qdisc", action, "dev", "eth0"}

	if qdisc.Parent == nil {
		args = append(args, "root")
	} else if qdisc.Parent.Major == 1 && qdisc.Parent.Minor == 0 {
		args = append(args, "root")
	} else {
		args = append(args, "parent", fmt.Sprintf("%d:%d", qdisc.Parent.Major, qdisc.Parent.Minor))
	}

	if qdisc.Handle == nil {
		args = append(args, "handle", fmt.Sprintf("%d:%d", 1, 0))
	} else {
		args = append(args, "handle", fmt.Sprintf("%d:%d", qdisc.Handle.Major, qdisc.Handle.Minor))
	}

	args = append(args, qdisc.Type)

	if qdisc.Args != nil {
		args = append(args, qdisc.Args...)
	}

	return args, nil
}

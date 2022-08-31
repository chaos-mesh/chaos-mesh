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

// Since BlockChaos is not unimplemented in darwin os. This file is only used for debugging, for example if your editor has gopls activated automatically.

package chaosdaemon

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const chaosDaemonHelperCommand = "cdh"

func (s *DaemonServer) ApplyBlockChaos(ctx context.Context, req *pb.ApplyBlockChaosRequest) (*pb.ApplyBlockChaosResponse, error) {
	return nil, nil
}

func normalizeVolumeName(ctx context.Context, volumePath string) (string, error) {
	return "", nil
}

func enableIOEMElevator(volumeName string) error {
	return nil
}

func (s *DaemonServer) RecoverBlockChaos(ctx context.Context, req *pb.RecoverBlockChaosRequest) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

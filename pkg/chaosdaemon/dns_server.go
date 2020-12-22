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

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	// DNSServerConfFile is the default config file for DNS server
	DNSServerConfFile = "/etc/resolv.conf"
)

func (s *DaemonServer) SetDNSServer(ctx context.Context,
	req *pb.SetDNSServerRequest) (*empty.Empty, error) {
	log.Info("SetDNSServer", "request", req)
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "GetPidFromContainerID")
		return nil, err
	}

	if req.Enable {
		// set dns server to the chaos dns server's address

		if len(req.DnsServer) == 0 {
			return &empty.Empty{}, fmt.Errorf("invalid set dns server request %v", req)
		}

		// backup the /etc/resolv.conf
		processBuilder := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("ls %s.chaos.bak || cp %s %s.chaos.bak", DNSServerConfFile, DNSServerConfFile, DNSServerConfFile)).SetContext(ctx)
		if !req.WithoutNS {
			processBuilder = processBuilder.SetNS(pid, bpm.MountNS)
		}

		cmd := processBuilder.Build()
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(err, "execute command error", "command", cmd.String(), "output", output)
			return nil, err
		}
		if len(output) != 0 {
			log.Info("command output", "output", string(output))
		}

		// add chaos dns server to the first line of /etc/resolv.conf
		// Note: can not replace the /etc/resolv.conf like `mv temp resolv.conf`, will execute with error `Device or resource busy`
		processBuilder = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("cp %s temp && sed -i 's/.*nameserver.*/nameserver %s/' temp && cat temp > %s", DNSServerConfFile, req.DnsServer, DNSServerConfFile)).SetContext(ctx)
		if !req.WithoutNS {
			processBuilder = processBuilder.SetNS(pid, bpm.MountNS)
		}

		cmd = processBuilder.Build()
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Error(err, "execute command error", "command", cmd.String(), "output", output)
			return nil, err
		}
		if len(output) != 0 {
			log.Info("command output", "output", string(output))
		}
	} else {
		// recover the dns server's address
		processBuilder := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("ls %s.chaos.bak && cat %s.chaos.bak > %s || true", DNSServerConfFile, DNSServerConfFile, DNSServerConfFile)).SetContext(ctx)
		if !req.WithoutNS {
			processBuilder = processBuilder.SetNS(pid, bpm.MountNS)
		}

		cmd := processBuilder.Build()
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(err, "execute command error", "command", cmd.String(), "output", output)
			return nil, err
		}
		if len(output) != 0 {
			log.Info("command output", "output", string(output))
		}
	}

	return &empty.Empty{}, nil
}

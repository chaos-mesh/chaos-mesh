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

package chaosdaemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

var (
	// DNSServerConfFile is the default config file for DNS server
	DNSServerConfFile = "/etc/resolv.conf"
)

var ErrInvalidDNSServer = errors.New("invalid DNS server address")

func (s *DaemonServer) SetDNSServer(ctx context.Context,
	req *pb.SetDNSServerRequest) (*empty.Empty, error) {
	log := s.getLoggerFromContext(ctx)

	log.Info("SetDNSServer", "request", req)
	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "GetPidFromContainerID")
		return nil, err
	}

	targetResolvPath := DNSServerConfFile
	if req.EnterNS {
		targetResolvPath = fmt.Sprintf("/proc/%d/root%s", pid, DNSServerConfFile)
	}
	backupPath := targetResolvPath + ".chaos.bak"

	if req.Enable {
		// set dns server to the chaos dns server's address

		if net.ParseIP(req.DnsServer) == nil {
			return nil, ErrInvalidDNSServer
		}

		content, err := os.ReadFile(targetResolvPath)
		if err != nil {
			log.Error(err, "read resolv.conf error")
			return nil, errors.Wrap(err, "read resolv.conf error")
		}

		// backup the /etc/resolv.conf
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			err = os.WriteFile(backupPath, content, 0644)
			if err != nil {
				log.Error(err, "backup resolv.conf error")
				return nil, errors.Wrap(err, "backup resolv.conf error")
			}
			log.Info("backup resolv.conf successfully", "path", backupPath)
		}

		// add chaos dns server to the first line of /etc/resolv.conf

		lines := strings.Split(string(content), "\n")
		nameserverLine := fmt.Sprintf("nameserver %s", req.DnsServer)
		var newLines []string
		replaced := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "nameserver") {
				newLines = append(newLines, nameserverLine)
				replaced = true
			} else {
				newLines = append(newLines, line)
			}
		}
		if !replaced {
			newLines = append([]string{nameserverLine}, newLines...)
		}

		newContent := strings.Join(newLines, "\n")
		err = os.WriteFile(targetResolvPath, []byte(newContent), 0644)
		if err != nil {
			log.Error(err, "write resolv.conf error")
			return nil, errors.Wrap(err, "write resolv.conf error")
		}
		log.Info("write resolv.conf successfully", "path", targetResolvPath)
	} else {
		// recover the dns server's address
		if _, err := os.Stat(backupPath); err == nil {
			content, err := os.ReadFile(backupPath)
			if err != nil {
				log.Error(err, "read backup resolv.conf error")
				return nil, errors.Wrap(err, "read backup resolv.conf error")
			}
			err = os.WriteFile(targetResolvPath, content, 0644)
			if err != nil {
				log.Error(err, "restore resolv.conf error")
				return nil, errors.Wrap(err, "restore resolv.conf error")
			}
			_ = os.Remove(backupPath)
			log.Info("restore resolv.conf successfully", "path", targetResolvPath)
		}
	}

	return &empty.Empty{}, nil
}

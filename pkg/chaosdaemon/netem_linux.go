// Copyright 2019 Chaos Mesh Authors.
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
	"github.com/vishvananda/netlink"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

func applyNetem(netem *pb.Netem, pid uint32) error {
	// Mock point to return error in unit test
	if err := mock.On("NetemApplyError"); err != nil {
		if e, ok := err.(error); ok {
			return e
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}

	p, h := buildHandles(netem)

	return applyQdisc(pid, func(handle *netlink.Handle, link netlink.Link) netlink.Qdisc {
		return netlink.NewNetem(netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    h,
			Parent:    p,
		}, ToNetlinkNetemAttrs(netem))
	})
}

func deleteNetem(netem *pb.Netem, pid uint32) error {
	// Mock point to return error in unit test
	if err := mock.On("NetemCancelError"); err != nil {
		if e, ok := err.(error); ok {
			return e
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}

	p, h := buildHandles(netem)

	return deleteQdisc(pid, func(handle *netlink.Handle, link netlink.Link) netlink.Qdisc {
		return &netlink.Netem{
			QdiscAttrs: netlink.QdiscAttrs{
				LinkIndex: link.Attrs().Index,
				Handle:    h,
				Parent:    p,
			},
		}
	})
}

func buildHandles(netem *pb.Netem) (parent, handle uint32) {

	if netem == nil {
		parent = netlink.HANDLE_ROOT
		handle = netlink.MakeHandle(1, 0)
		return
	}

	if netem.Parent == nil {
		parent = netlink.HANDLE_ROOT
	} else if netem.Parent.Major == 1 && netem.Parent.Minor == 0 {
		parent = netlink.HANDLE_ROOT
	} else {
		parent = netlink.MakeHandle(uint16(netem.Parent.Major), uint16(netem.Parent.Minor))
	}

	if netem.Handle == nil {
		handle = netlink.MakeHandle(1, 0)
	} else {
		handle = netlink.MakeHandle(uint16(netem.Handle.Major), uint16(netem.Handle.Minor))
	}

	return
}

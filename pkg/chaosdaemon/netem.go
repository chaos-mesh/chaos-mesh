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

package chaosdaemon

import (
	"strings"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
)

// Apply applies a netem on eth0 in pid related namespace
func Apply(netem *pb.Netem, pid uint32) error {
	log.Info("Apply netem on PID", "pid", pid)

	ns, err := netns.GetFromPath(GenNetnsPath(pid))
	if err != nil {
		log.Error(err, "failed to find network namespace", "pid", pid)
		return err
	}
	defer ns.Close()

	handle, err := netlink.NewHandleAt(ns)
	if err != nil {
		log.Error(err, "failed to get handle at network namespace", "network namespace", ns)
		return err
	}

	link, err := handle.LinkByName("eth0") // TODO: check whether interface name is eth0
	if err != nil {
		log.Error(err, "failed to find eth0 interface")
		return err
	}

	netemQdisc := netlink.NewNetem(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	}, ToNetlinkNetemAttrs(netem))

	log.Info("add qdisc", "qdisc", netemQdisc)
	if err = handle.QdiscAdd(netemQdisc); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Error(err, "failed to add Qdisc", "qdisc", netemQdisc)
			return err
		}
	}

	return nil
}

// Cancel will remove netem on eth0 in pid related namespace
func Cancel(netem *pb.Netem, pid uint32) error {
	// WARN: This will delete all netem on this interface
	log.Info("Cancel netem on PID", "pid", pid)

	ns, err := netns.GetFromPath(GenNetnsPath(pid))
	if err != nil {
		log.Error(err, "failed to find network namespace", "pid", pid)
		return err
	}
	defer ns.Close()

	handle, err := netlink.NewHandleAt(ns)
	if err != nil {
		log.Error(err, "failed to create new handle at network namespace", "network namespace", ns)
		return err
	}

	link, err := handle.LinkByName("eth0") // TODO: check whether interface name is eth0
	if err != nil {
		log.Error(err, "failed to find eth0 interface")
		return err
	}

	netemQdisc := &netlink.Netem{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(1, 0),
			Parent:    netlink.HANDLE_ROOT,
		},
	}

	exist, err := qdiscExists(netemQdisc, handle, link)
	if err != nil {
		log.Error(err, "failed to check qdisc", "qdisc", netemQdisc, "link", link)
		return err
	}

	if !exist {
		log.Info("qdisc not exists", "qdisc", netemQdisc)
		return nil
	}

	log.Info("remove qdisc", "qdisc", netemQdisc)
	if err = handle.QdiscDel(netemQdisc); err != nil {
		log.Error(err, "failed to remove qdisc", "qdisc", netemQdisc)

		return err
	}

	return nil
}

func qdiscExists(qdisc netlink.Qdisc, handler *netlink.Handle, link netlink.Link) (bool, error) {
	qds, err := handler.QdiscList(link)
	if err != nil {
		log.Error(err, "failed to list qdiscs", "link", link)
		return false, err
	}

	for _, qd := range qds {
		if qd.Attrs().LinkIndex == qdisc.Attrs().LinkIndex &&
			qd.Attrs().Parent == qdisc.Attrs().Parent &&
			qd.Attrs().Handle == qdisc.Attrs().Handle {
			return true, nil
		}
	}

	return false, nil
}

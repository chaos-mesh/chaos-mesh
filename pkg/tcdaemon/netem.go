package tcdaemon

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/juju/errors"
	pb "github.com/pingcap/chaos-operator/pkg/tcdaemon/pb"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	defaultProcPrefix = "/mnt/proc"
)

//Apply applies a netem on eth0 in pid related namespace
func Apply(netem *pb.Netem, pid uint32) error {
	glog.Infof("Apply netem on PID: %d", pid)
	nsPath := fmt.Sprintf("%s/%d/ns/net", defaultProcPrefix, pid)
	ns, err := netns.GetFromPath(nsPath)
	if err != nil {
		glog.Errorf("error while finding network namespace %s", nsPath)
		return errors.Trace(err)
	}

	handle, err := netlink.NewHandleAt(ns)
	if err != nil {
		glog.Errorf("error while getting handle at netns %v: %v", ns, err)
	}

	link, err := handle.LinkByName("eth0") // TODO: check whether interface name is eth0
	if err != nil {
		glog.Errorf("error while finding eth0 interface %v", err)
		return errors.Trace(err)
	}

	netemQdisc := netlink.NewNetem(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	}, ToNetlinkNetemAttrs(netem))

	if err = handle.QdiscAdd(netemQdisc); err != nil {
		glog.Errorf("error while adding Qdisc %v", err)
		return errors.Trace(err)
	}

	return nil
}

// Cancel will remove netem on eth0 in pid related namespace
func Cancel(netem *pb.Netem, pid uint32) error {
	// WARN: This will delete all netem on this interface

	glog.Infof("Cancel netem on PID: %d", pid)

	nsPath := fmt.Sprintf("%s/%d/ns/net", defaultProcPrefix, pid)
	ns, err := netns.GetFromPath(nsPath)
	if err != nil {
		glog.Errorf("error while finding network namespace %s", nsPath)
		return errors.Trace(err)
	}

	handle, err := netlink.NewHandleAt(ns)

	link, err := handle.LinkByName("eth0") // TODO: check whether interface name is eth0
	if err != nil {
		glog.Errorf("error while finding eth0 interface %v", err)
		return errors.Trace(err)
	}

	netemQdisc := &netlink.Netem{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: link.Attrs().Index,
			Handle:    netlink.MakeHandle(1, 0),
			Parent:    netlink.HANDLE_ROOT,
		},
	}

	if err = handle.QdiscDel(netemQdisc); err != nil {
		glog.Errorf("error while removing Qdisc %v", err)
		return errors.Trace(err)
	}

	return nil
}

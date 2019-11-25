package chaosdaemon

import (
	"strings"

	"github.com/juju/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"

	pb "github.com/pingcap/chaos-operator/pkg/chaosdaemon/pb"
)

// Apply applies a netem on eth0 in pid related namespace
func Apply(netem *pb.Netem, pid uint32) error {
	log.Info("Apply netem on PID", "pid", pid)

	ns, err := netns.GetFromPath(GenNetnsPath(pid))
	if err != nil {
		log.Error(err, "failed to find network namespace", "pid", pid)
		return errors.Trace(err)
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
		return errors.Trace(err)
	}

	netemQdisc := netlink.NewNetem(netlink.QdiscAttrs{
		LinkIndex: link.Attrs().Index,
		Handle:    netlink.MakeHandle(1, 0),
		Parent:    netlink.HANDLE_ROOT,
	}, ToNetlinkNetemAttrs(netem))

	if err = handle.QdiscAdd(netemQdisc); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Error(err, "failed to add Qdisc")
			return errors.Trace(err)
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
		return errors.Trace(err)
	}
	defer ns.Close()

	handle, err := netlink.NewHandleAt(ns)
	if err != nil {
		log.Error(err, "failed to create new handle at network namespace", "network namespace", ns)
		return errors.Trace(err)
	}

	link, err := handle.LinkByName("eth0") // TODO: check whether interface name is eth0
	if err != nil {
		log.Error(err, "failed to find eth0 interface")
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
		log.Error(err, "failed to remove Qdisc")
		return errors.Trace(err)
	}

	return nil
}

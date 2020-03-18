package chaosdaemon

import (
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func newNetlinkHandle(pid uint32) (netns.NsHandle, *netlink.Handle, netlink.Link, error) {
	ns, err := netns.GetFromPath(GenNetnsPath(pid))
	if err != nil {
		log.Error(err, "failed to find network namespace", "pid", pid)
		return -1, nil, nil, err
	}

	handle, err := netlink.NewHandleAt(ns)
	if err != nil {
		log.Error(err, "failed to create new handle at network namespace", "network namespace", ns)
		return -1, nil, nil, err
	}

	link, err := handle.LinkByName("eth0") // TODO: check whether interface name is eth0
	if err != nil {
		log.Error(err, "failed to find eth0 interface")
		return -1, nil, nil, err
	}

	return ns, handle, link, nil
}

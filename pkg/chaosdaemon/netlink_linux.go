package chaosdaemon

import (
	"strings"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type toQdiscFunc func(*netlink.Handle, netlink.Link) netlink.Qdisc

func applyQdisc(pid uint32, toQdisc toQdiscFunc) error {
	log.Info("apply qdisc on PID", "pid", pid)

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

	qdisc := toQdisc(handle, link)

	log.Info("Add qdisc", "qdisc", qdisc)
	if err = handle.QdiscAdd(qdisc); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			log.Error(err, "failed to add Qdisc", "qdisc", qdisc)
			return err
		}
	}

	return nil
}

func deleteQdisc(pid uint32, toQdisc toQdiscFunc) error {
	// WARN: This will delete all qdisc on this interface
	log.Info("delete qdisc on PID", "pid", pid)

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

	qdisc := toQdisc(handle, link)

	exist, err := qdiscExists(qdisc, handle, link)
	if err != nil {
		log.Error(err, "failed to check qdisc", "qdisc", qdisc, "link", link)
		return err
	}

	if !exist {
		log.Error(nil, "qdisc not exists, qdisc may be deleted by mistake or not injected successfully, there may be bugs here", "qdisc", qdisc)
		return nil
	}

	log.Info("Remove qdisc", "qdisc", qdisc)
	if err = handle.QdiscDel(qdisc); err != nil {
		log.Error(err, "failed to remove qdisc", "qdisc", qdisc)

		return err
	}

	return nil
}

// TODO(vincent178): this is ok if we only apply either netem or tbf, revisit this if we want both working
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

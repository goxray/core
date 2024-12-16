//go:build linux

package tun

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

func (i *Interface) up(local *net.IPNet, gw net.IP) error {
	link, err := netlink.LinkByName(i.Name())
	if err != nil {
		return fmt.Errorf("failed to detect %s interface: %s", i.Name(), err)
	}

	ipv4Addr := &netlink.Addr{
		IPNet: local,
		Peer:  &net.IPNet{IP: gw, Mask: []byte{0, 0, 0, 0}},
	}
	err = netlink.AddrAdd(link, ipv4Addr)
	if err != nil {
		return fmt.Errorf("failed to set peer address on %s interface: %s", i.Name(), err)
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return fmt.Errorf("failed to set %s interface up: %s", i.Name(), err)
	}

	return nil
}

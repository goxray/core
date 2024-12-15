package route

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"

	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
)

func addDeleteRoutes(options Opts, delete bool) error {
	var err error
	if err := options.Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	socket, err := unix.Socket(unix.AF_ROUTE, unix.SOCK_RAW, unix.AF_UNSPEC)
	if err != nil {
		return fmt.Errorf("unix socket: %v", err)
	}
	defer func() {
		err = errors.Join(err, unix.Shutdown(socket, unix.SHUT_RDWR), unix.Close(socket))
	}()

	for _, cidr := range options.Routes {
		var addresses []route.Addr
		switch options.opType() {
		case typeIF:
			addresses, err = toIfaceAddresses(options.IfName, cidr)
			if err != nil {
				return fmt.Errorf("addDeleteRoutes %s to %s: %v", cidr, options.IfName, err)
			}
		case typeGW:
			addresses, err = toGWAddresses(options.Gateway, cidr)
			if err != nil {
				return fmt.Errorf("addDeleteRoutes %s to %s: %v", options.Gateway, options.IfName, err)
			}
		case typeUnknown:
			return errors.New("unknown route type")
		}

		action := uint8(syscall.RTM_ADD)
		if delete {
			action = syscall.RTM_DELETE
		}

		// TODO: refine flags
		flags := syscall.RTF_UP | /*syscall.RTF_CLONING |*/ syscall.RTF_GATEWAY | syscall.RTF_STATIC
		if err := processRouteCall(socket, addresses, action, flags); err != nil {
			return fmt.Errorf("failed to addDeleteRoutes %s route: %v", cidr, err)
		}
	}

	return nil
}

func processRouteCall(socket int, addrs []route.Addr, action uint8, flags int) error {
	msg := route.RouteMessage{
		Version: syscall.RTM_VERSION,
		Type:    int(action),
		Flags:   flags,
		Seq:     1,
		ID:      uintptr(os.Getpid()),
		Addrs:   addrs,
	}
	bin, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("marshal route message: %v", err)
	}

	_, err = unix.Write(socket, bin[:])
	if err != nil {
		return fmt.Errorf("write to socket %d: %v", socket, err)
	}

	return nil
}

func toIfaceAddresses(ifcName string, dest *Addr) ([]route.Addr, error) {
	switch {
	case dest == nil:
		return nil, fmt.Errorf("destination is nil")
	case ifcName == "":
		return nil, fmt.Errorf("ifcName is empty")
	default:
		_, err := net.InterfaceByName(ifcName)
		if err != nil {
			return nil, fmt.Errorf("interface %s not found: %v", ifcName, err)
		}
	}

	addresses := []route.Addr{
		inet4Addr(dest.IP),
		&route.LinkAddr{Name: ifcName},
	}
	if dest.Mask != nil {
		addresses = append(addresses, inet4Addr(net.IP(dest.Mask)))
	}

	return addresses, nil
}

func toGWAddresses(gw net.IP, dest *Addr) ([]route.Addr, error) {
	switch {
	case gw == nil:
		return nil, fmt.Errorf("gateway is nil")
	case dest == nil:
		return nil, fmt.Errorf("destination is nil")
	}

	addresses := []route.Addr{
		inet4Addr(dest.IP),
		inet4Addr(gw),
	}
	if dest.Mask != nil {
		addresses = append(addresses, inet4Addr(net.IP(dest.Mask)))
	}

	return addresses, nil
}

func inet4Addr(ip net.IP) *route.Inet4Addr {
	return &route.Inet4Addr{IP: ([4]byte)(ip.To4())}
}

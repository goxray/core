//go:build darwin

package tun

import (
	"fmt"
	"net"
	"unsafe"

	"golang.org/x/sys/unix"
)

// up brings the TUN interface up and assign addresses for it.
//
// TODO: IPV6
func (i *Interface) up(local *net.IPNet, gw net.IP) error {
	// https://github.com/freebsd/freebsd-src/blob/de1aa3dab23c06fec962a14da3e7b4755c5880cf/sys/net/if.h#L444
	type ifAliasReq struct {
		Name      [unix.IFNAMSIZ]byte /* if name, e.g. "utun123" */
		Addr      unix.RawSockaddrInet4
		BroadAddr unix.RawSockaddrInet4
		Mask      unix.RawSockaddrInet4
	}

	// https://github.com/freebsd/freebsd-src/blob/de1aa3dab23c06fec962a14da3e7b4755c5880cf/sys/net/if.h#L403
	type ifFlagsReq struct {
		Name  [unix.IFNAMSIZ]byte
		Flags uint16
	}

	// Open file descriptor for AF_INET (IPV4)
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return fmt.Errorf("create socket: %v", err)
	}
	defer unix.Close(fd)

	toAddr4Bytes := func(ip net.IP) [4]byte {
		ip4 := ip.To4()
		return [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]}
	}

	ifr := ifAliasReq{
		Name: i.nameBytes(),
		// Local IP
		Addr: unix.RawSockaddrInet4{
			Len:    unix.SizeofSockaddrInet4,
			Family: unix.AF_INET,
			Addr:   toAddr4Bytes(local.IP),
		},
		// Local Mask
		Mask: unix.RawSockaddrInet4{
			Len:    unix.SizeofSockaddrInet4,
			Family: unix.AF_INET,
			Addr: func() [4]byte {
				if local.Mask == nil {
					return [4]byte{}
				}

				return toAddr4Bytes(net.IP(local.Mask))
			}(),
		},
		// Peer destination address
		BroadAddr: unix.RawSockaddrInet4{
			Len:    unix.SizeofSockaddrInet4,
			Family: unix.AF_INET,
			Addr:   toAddr4Bytes(gw),
		},
	}

	// Set interface addr via IOCTL syscall.
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.SIOCAIFADDR),
		uintptr(unsafe.Pointer(&ifr)),
	)
	if errno != 0 {
		return fmt.Errorf("set address on %s interface: %v", i.Name(), errno)
	}

	// Bring interface UP
	ifrFlags := ifFlagsReq{
		Name:  ifr.Name,
		Flags: unix.IFF_UP | unix.IFF_RUNNING,
	}

	// Mark interface as UP via IOCTL syscall.
	_, _, errno = unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.SIOCSIFFLAGS),
		uintptr(unsafe.Pointer(&ifrFlags)),
	)
	if errno != 0 {
		return fmt.Errorf("activate %s interface: %v", i.Name(), errno)
	}

	return nil
}

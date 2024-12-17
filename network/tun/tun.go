/*
Package tun implements TUN interface operations on system level.
*/
package tun

import (
	"fmt"
	"net"

	"github.com/songgao/water"
)

// Interface is a TUN/TAP interface.
type Interface struct {
	ifc *water.Interface
}

// New creates a new TUN interface.
//
// On macOS the device name will be assigned dynamically if you pass an empty name string.
func New(name string, MTU int) (*Interface, error) {
	// We can use sing-tun as internal TUN driver in future (it has more granural setup), example:
	// 	options := tun.Options{Name: "utun123", MTU: 1500, Inet4Address: []netip.Prefix{netip.MustParsePrefix("192.18.0.1/0")}}
	//	tunDev, _ := tun.New(options)

	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = name

	ifc, err := water.New(config)
	if err != nil {
		return nil, fmt.Errorf("create tun interface: %v", err)
	}

	return &Interface{ifc: ifc}, nil
}

// Up brings the TUN interface up and assign addresses for it.
//
// Basically mimics "ifconfig {tun} {local} {gateway} up".
// But is implemented via system calls.
func (i *Interface) Up(local *net.IPNet, gw net.IP) error {
	return i.up(local, gw) // call OS specific up function.
}

// Name returns the interface name of ifce, e.g. tun0, tap1, tun0, etc..
func (i *Interface) Name() string {
	return i.ifc.Name()
}

// Read reads p bytes from the interface fd.
func (i *Interface) Read(p []byte) (n int, err error) {
	return i.ifc.Read(p)
}

// Write writes p bytes to the interface fd.
func (i *Interface) Write(p []byte) (n int, err error) {
	return i.ifc.Write(p)
}

// Close closes fd. Close call is recommended to close the socket and destroy the interface.
func (i *Interface) Close() error {
	return i.ifc.Close()
}

// nameBytes transforms Name() into fixed 16 byte array. For internal usage.
func (i *Interface) nameBytes() [16]byte {
	sb := make([]byte, 16)
	copy(sb[:len(i.Name())], i.Name())

	return [16]byte(sb)
}

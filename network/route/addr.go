package route

import (
	"fmt"
	"net"
)

// Addr represents IP and Port.
type Addr net.IPNet

func (a *Addr) String() string {
	if a.Mask == nil {
		return a.IP.String()
	}

	return (*net.IPNet)(a).String()
}

// ParseAddr parses the given addr and transforms it into Addr.
//
// First it tries to parse it as plain net.IP addr if no mask present.
func ParseAddr(addr string) (*Addr, error) {
	ip, _ := net.ResolveIPAddr("ip", addr)
	if ip != nil {
		return &Addr{
			IP: ip.IP,
		}, nil
	}

	_, ipNet, err := net.ParseCIDR(addr)
	if err != nil {
		return nil, fmt.Errorf("parse cidr: %w", err)
	}

	return (*Addr)(ipNet), nil
}

// MustParseAddr is the same as ParseAddr but panics on errors.
func MustParseAddr(addr string) *Addr {
	a, err := ParseAddr(addr)
	if err != nil {
		panic(err)
	}

	return a
}

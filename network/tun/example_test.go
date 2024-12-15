package tun_test

import (
	"log"
	"net"

	"github.com/goxray/core/network/tun"
)

func Example_create() {
	ifc, err := tun.New("utun123")
	if err != nil {
		log.Fatal(err)
	}
	defer ifc.Close()

	addr := &net.IPNet{IP: net.IPv4(192, 18, 0, 1)}

	if err := ifc.Up(addr, addr.IP); err != nil {
		log.Fatal(err)
	}

	// Interface will be destroyed upon exiting the program...
}

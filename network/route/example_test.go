package route_test

import (
	"log"
	"net"

	"github.com/goxray/core/network/route"
)

func Example_networkInterface() {
	routes := []*route.Addr{
		route.MustParseAddr("93.184.215.14"), // example.com IP address
	}

	opts := route.Opts{IfName: "en0", Routes: routes}

	// Route to default Mac OS interface `en0`
	if err := route.Add(opts); err != nil {
		log.Fatal(err)
	}

	// Remove route from the table
	if err := route.Delete(opts); err != nil {
		log.Fatal(err)
	}
}

func Example_gatewayIP() {
	routes := []*route.Addr{
		route.MustParseAddr("93.184.215.14"), // example.com IP address
	}

	opts := route.Opts{Gateway: net.IP{192, 0, 0, 1}, Routes: routes}

	// Route to 192.0.0.1
	if err := route.Add(opts); err != nil {
		log.Fatal(err)
	}

	// Remove route from the table
	if err := route.Delete(opts); err != nil {
		log.Fatal(err)
	}
}

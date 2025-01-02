/*
Package route implements basic api to interact with system IP tables.

Add and Delete operations are implemented for gateway/interface addresses.
*/
package route

import (
	"fmt"
	"net"
)

type opType int

const (
	typeUnknown opType = iota // Not expected.
	typeIF                    // Operation is performed on network interface.
	typeGW                    // Operation is performed on network gateway IP address.
)

// Opts represent route action options. You must specify either IfName or Gateway,
// specifying both will trigger Validate error.
type Opts struct {
	IfName  string
	Gateway net.IP
	Routes  []*Addr
}

// Route is used to add or delete system routes.
type Route struct{}

func New() (*Route, error) {
	return &Route{}, nil
}

// Add adds route to ip table.
func (r *Route) Add(options Opts) error {
	return addDeleteRoutes(options, false)
}

// Delete deletes route from ip table.
func (r *Route) Delete(options Opts) error {
	return addDeleteRoutes(options, true)
}

// Add adds route to ip table.
func Add(options Opts) error {
	return addDeleteRoutes(options, false)
}

// Delete deletes route from ip table.
func Delete(options Opts) error {
	return addDeleteRoutes(options, true)
}

func (o Opts) opType() opType {
	switch {
	case o.hasIfName():
		return typeIF
	case o.hasGateway():
		return typeGW
	}

	return typeUnknown
}

func (o Opts) hasIfName() bool {
	return o.IfName != ""
}

func (o Opts) hasGateway() bool {
	return o.Gateway != nil
}

func (o Opts) Validate() error {
	switch {
	case len(o.Routes) == 0:
		return fmt.Errorf("at least one address must be specified")
	case !o.hasIfName() && !o.hasGateway():
		return fmt.Errorf("either IfName or Gateway must be specified")
	case o.hasIfName() && o.hasGateway():
		return fmt.Errorf("cannot specify both IfName and Gateway")
	}

	return nil
}

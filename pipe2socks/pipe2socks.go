/*
Package pipe2socks implements a pipe to route IP packets stream from io.ReadWriteCloser to socks proxy and back.
You need a configured socks5 proxy address for the Copy function.

Basically it is a broad implementation of tun2socks. And yes, io.ReadWriteCloser could be a TUN device.
*/
package pipe2socks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device/iobased"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"github.com/xjasonlyu/tun2socks/v2/tunnel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// Opts contain options for the established connection between pipe and Socks server.
// DefaultOpts should be used for most cases.
type Opts struct {
	MTU        int  // MTU should be set according to your configuration to prevent data losses.
	UDP        bool // UDP enables UDP support. Recommended to be turned on.
	UDPTimeout time.Duration
}

// DefaultOpts represent the default connection settings suitable for most cases.
var DefaultOpts = &Opts{
	MTU:        1500,
	UDP:        true,
	UDPTimeout: 30 * time.Second,
}

// Pipe represents a pipe that connects io.ReadWriteCloser and sock5 proxy.
type Pipe struct {
	opts  *Opts
	stack *stack.Stack
	proxy proxy.Proxy
}

func NewPipe(opts *Opts) (*Pipe, error) {
	if opts == nil {
		opts = DefaultOpts
	}

	return &Pipe{opts: opts}, nil
}

// Copy connects io.ReadWriteCloser to socks5 server.
//
// It reads IP packets from pipe and routes them to socks5 and back.
// This function blocks for the duration of the whole transmission, and
// it is recommended to gracefully unlock it (ending the established connection) by cancelling the ctx.
func (p *Pipe) Copy(ctx context.Context, pipe io.ReadWriteCloser, socks5 string) error {
	proxyAddr, err := parseSocksAddr(socks5)
	if err != nil {
		return fmt.Errorf("parse socks addr: %v", err)
	}

	// Create SOCKS5 proxy
	p.proxy, err = proxy.NewSocks5(proxyAddr.String(), "", "")
	if err != nil {
		return fmt.Errorf("create socks proxy: %v", err)
	}

	// Set the proxy for tunnel
	tunnel.T().SetDialer(p.proxy)

	// Set UDP timeout if UDP is enabled
	if p.opts.UDP {
		tunnel.T().SetUDPTimeout(p.opts.UDPTimeout)
	}

	// Create device endpoint from io.ReadWriteCloser
	device, err := iobased.New(pipe, uint32(p.opts.MTU), 0)
	if err != nil {
		return fmt.Errorf("create device: %v", err)
	}

	// Create stack
	p.stack, err = core.CreateStack(&core.Config{
		LinkEndpoint:     device,
		TransportHandler: tunnel.T(),
	})
	if err != nil {
		return fmt.Errorf("create stack: %v", err)
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Cleanup
	if p.stack != nil {
		p.stack.Close()
		p.stack.Wait()
	}

	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

// Copy is a one-function version of Pipe.Copy().
func Copy(ctx context.Context, pipe io.ReadWriteCloser, socks5 string, options *Opts) error {
	p, err := NewPipe(options)
	if err != nil {
		return err
	}

	return p.Copy(ctx, pipe, socks5)
}

func parseSocksAddr(socks5 string) (*net.TCPAddr, error) {
	if !strings.Contains(socks5, "://") {
		socks5 = fmt.Sprintf("socks5://%s", socks5)
	}

	socksURL, err := url.Parse(socks5)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy address: %w", err)
	}

	address := socksURL.Host
	if address == "" {
		address = socksURL.Path
	}

	proxyAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("resolve proxy address: %w", err)
	}

	return proxyAddr, nil
}

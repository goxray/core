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

	"github.com/eycorsican/go-tun2socks/core"
	"github.com/eycorsican/go-tun2socks/proxy/socks"
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

// Copy connects io.ReadWriteCloser to socks5 server.
//
// It reads IP packets from pipe and routes them to socks5 and back.
// This function blocks for the duration of the whole transmission and
// it is recommended to gracefully unlock it (ending the established connection) by cancelling the ctx.
func Copy(ctx context.Context, pipe io.ReadWriteCloser, socks5 string, options *Opts) error {
	if options == nil {
		options = DefaultOpts
	}

	proxy, err := parseSocksAddr(socks5)
	if err != nil {
		return fmt.Errorf("parse socks addr: %v", err)
	}

	core.RegisterTCPConnHandler(socks.NewTCPHandler(proxy.IP.String(), uint16(proxy.Port)))
	if options.UDP {
		core.RegisterUDPConnHandler(socks.NewUDPHandler(proxy.IP.String(), uint16(proxy.Port), options.UDPTimeout))
	}

	// Register an output callback to write packets output from lwip stack to pipe
	// device, output function should be set before input any packets.
	core.RegisterOutputFn(pipe.Write)

	// Setup TCP/IP stack.
	lwipWriter := core.NewLWIPStack()
	_, err = io.CopyBuffer(lwipWriter, newCtxReader(ctx, pipe), make([]byte, options.MTU))
	if err != nil {
		if isErrorExpected(ctx, err) {
			return nil
		}

		err = fmt.Errorf("write lwip stack: %v", err)
		if ctx.Err() != nil {
			return errors.Join(err, ctx.Err())
		}

		return err
	}

	return nil
}

// isErrorExpected implements a hacky way to ensure we close the connection properly.
func isErrorExpected(ctx context.Context, err error) bool {
	closed := strings.Contains(err.Error(), "already closed") || errors.Is(err, io.EOF)
	if errors.Is(ctx.Err(), context.Canceled) && closed {
		return true
	}

	return false
}

func newCtxReader(ctx context.Context, r io.Reader) io.Reader {
	return &ctxReader{ctx: ctx, r: r}
}

type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (r *ctxReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, io.EOF
	default:
		return r.r.Read(p)
	}
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

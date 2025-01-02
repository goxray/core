# ⚙️ GoXRAY Core

![Static Badge](https://img.shields.io/badge/OS-macOS%20%7C%20Linux-blue?style=flat&logo=linux&logoColor=white&logoSize=auto&color=blue)
![Static Badge](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/goxray/core)](https://goreportcard.com/report/github.com/goxray/core)
[![Go Reference](https://pkg.go.dev/badge/github.com/goxray/core.svg)](https://pkg.go.dev/github.com/goxray/core)

Core libraries and packages for GoXRay VPN.

## Packages:
- [network/route](/network/route): basic api to interact with system IP tables.
- [network/tun](/network/tun): TUN interface operations on system level.
- [pipe2socks](/pipe2socks): a pipe to route IP packets stream from io.ReadWriteCloser to socks proxy and back.

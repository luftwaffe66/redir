# Redir — Multi-Protocol TCP/UDP/HTTP Port Redirector & Forwarder

[![CI](https://github.com/luftwaffe66/redir/actions/workflows/ci.yml/badge.svg)](https://github.com/luftwaffe66/redir/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/luftwaffe66/redir)](https://goreportcard.com/report/github.com/luftwaffe66/redir)

**Redir** is a simple, lightweight, zero-dependency multi-protocol redirector written in Go. It forwards incoming TCP, UDP, and HTTP traffic from a local port to any remote destination with full bidirectional support.

Use it for **port forwarding**, **traffic tunneling**, **local development proxies**, **network debugging**, or as a **Go library** in your own applications.

---

## Features

- **3 protocols in one binary** — TCP, UDP, and HTTP redirection
- **Full bidirectional** — request and response both forwarded
- **Zero external dependencies** — pure Go standard library only
- **Structured logging** — `log/slog` with debug and info levels
- **Graceful shutdown** — handles SIGINT/SIGTERM cleanly
- **Concurrent connections** — each connection handled in its own goroutine
- **HTTP reverse proxy** — proper `httputil.ReverseProxy` for HTTP traffic
- **UDP request-response** — bidirectional UDP with per-packet response forwarding
- **Portable** — cross-platform: Linux, macOS, Windows, ARM, x86
- **CLI flags** — scriptable interface, no interactive prompts
- **Go library API** — `StartRedirector()` for embedding in your projects

---

## Installation

### Via Go

```bash
go install github.com/luftwaffe66/redir/cmd/redir@latest
```

### From Source

```bash
git clone https://github.com/luftwaffe66/redir.git
cd redir
go build -o redir ./cmd/redir/
```

### Pre-built Binaries

Download from the [Releases](https://github.com/luftwaffe66/redir/releases) page for Linux, macOS, and Windows (amd64, arm64, 386).

---

## Usage

### Command Line

```bash
# Redirect TCP traffic from port 8080 to 10.0.0.1:9090
redir -proto tcp -listen 8080 -dest 10.0.0.1:9090

# UDP DNS forwarding to Cloudflare
redir -proto udp -listen 53 -dest 1.1.1.1:53

# HTTP reverse proxy to local dev server
redir -proto http -listen 3000 -dest 127.0.0.1:8080

# Verbose mode with debug logging
redir -proto tcp -listen 4444 -dest 10.0.0.1:5555 -v

# Using environment variable for verbose mode
REDIR_VERBOSE=1 redir -proto tcp -listen 8080 -dest example.com:80
```

```
Usage: redir -dest host:port [options]

Options:
  -proto string
        Protocol to redirect: tcp, udp, or http (default "tcp")
  -listen string
        Local port to listen on (default "8080")
  -dest string
        Destination address (host:port)
  -v    Verbose logging
  -help
        Show this help
```

### As a Go Library

```go
package main

import (
    "context"
    "log"

    "github.com/luftwaffe66/redir"
)

func main() {
    opts := redirector.RedirectOptions{
        ListenPort:      "8888",
        DestinationIP:   "192.168.1.100",
        DestinationPort: "8080",
        Protocol:        "tcp",
        Verbose:         true,
    }

    if err := redirector.StartRedirector(context.Background(), opts); err != nil {
        log.Fatalf("Redirector failed: %v", err)
    }
}
```

---

## API Reference

### `RedirectOptions`

| Field            | Type   | Description                          |
|------------------|--------|--------------------------------------|
| `ListenPort`     | string | Local port to bind (e.g. `"8080"`)   |
| `DestinationIP`  | string | Remote destination IP or hostname    |
| `DestinationPort`| string | Remote destination port              |
| `Protocol`       | string | Protocol: `"tcp"`, `"udp"`, `"http"` |
| `Verbose`        | bool   | Enable debug-level logging           |

### `StartRedirector(ctx context.Context, opts RedirectOptions) error`

Starts the redirector with the given configuration. Blocks until the context is cancelled or a fatal error occurs. Handles SIGINT/SIGTERM via `signal.NotifyContext`.

---

## How It Works

**TCP**: Listens on the local port, accepts connections, dials the destination, and runs bidirectional copy between the two sides. Each direction uses independent goroutines with manual read/write loops.

**UDP**: Listens on the local port, reads incoming packets, forwards each to the destination, waits for a response (with timeout), and sends it back to the original client. Supports concurrent request-response flows with a bounded goroutine pool (max 100).

**HTTP**: Acts as a reverse proxy using `httputil.ReverseProxy`. Forwards the full HTTP request (method, path, headers, body) to the destination and streams the response back. Configurable timeouts.

---

## Project Structure

```
redir/
├── cmd/
│   └── redir/
│       └── main.go              # CLI entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration and flag parsing
│   ├── tcp/
│   │   └── redirect.go          # TCP port forwarding
│   ├── udp/
│   │   └── redirect.go          # UDP packet forwarding
│   └── httpproxy/
│       └── redirect.go          # HTTP reverse proxy
├── redirector.go                 # Public API (StartRedirector, RedirectOptions)
├── example/
│   └── main.go                   # Usage example as Go library
├── .github/workflows/
│   └── ci.yml                    # CI with multi-platform builds
├── go.mod
└── README.md
```

---

## Benchmarks

> TODO: Benchmarks coming soon. For now, the redirector uses efficient goroutine-per-connection concurrency with manual read/write loops (avoiding kernel splice for reliable EOF detection).

---

## Use Cases

- **Local development** — forward traffic from a public port to a local dev server
- **Network debugging** — inspect or log traffic between services
- **DNS forwarding** — redirect DNS queries to a custom resolver
- **Port tunneling** — expose services behind NAT or firewalls
- **Container networking** — forward ports between containers
- **IoT** — bridge traffic between devices on different networks

---

## License

MIT — see [LICENSE](LICENSE).

---

## Contributing

Issues and pull requests welcome. Ensure tests pass before submitting:

```bash
go test -v -race -count=1 ./...
go vet ./...
```

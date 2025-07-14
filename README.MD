# Redir

**Redir** is a simple, lightweight, and robust multi-protocol redirector written in Go.

It allows you to redirect incoming connections on a local port to a remote IP and port, supporting multiple protocols including TCP, UDP, and HTTP, with full bidirectional support.

## ✨ Features

- 🔁 Full bidirectional support for TCP, UDP, and HTTP redirection
- ⚙️ Easy-to-use command-line interface for selecting protocols
- 🧱 Can be imported as a library into any Go project
- 🚨 Robust error handling and panic safety
- 📦 Clean and dependency-free (except for `github.com/fatih/color` for color logging)

## 📦 Installation

To use this package in your project:

```bash
go get github.com/luftwaffe66/redirector
```

## 🧪 Example Usage

You can use Redir by importing it into your Go application and initializing the redirector with the desired configuration.

```go
package main

import (
    "log"
    "github.com/luftwaffe66/redirector"
)

func main() {
    opts := redirector.RedirectOptions{
        ListenPort:      "8080",           // Local port to listen on
        DestinationIP:   "192.168.1.100",  // Remote destination IP
        DestinationPort: "9090",           // Remote port to forward to
        Verbose:         true,             // Show info logs
    }

    if err := redirector.StartRedirector(opts); err != nil {
        log.Fatalf("Redirector failed: %v", err)
    }
}
```

## 🧱 Struct Reference

The `RedirectOptions` struct is used to configure the redirector. Here's the full reference:

```go
type RedirectOptions struct {
    ListenPort      string // Local port to bind and accept incoming connections
    DestinationIP   string // Remote target IP address
    DestinationPort string // Remote target port
    Verbose         bool   // Whether to log connection events
}
```

## 🔧 How It Works

When `StartRedirector()` is called, it:

1. Listens on the specified local port
2. Accepts incoming connections for the selected protocol (TCP, UDP, or HTTP)
3. Forwards the connection to the specified remote IP and port
4. Maintains bidirectional data flow (client ↔ server)
5. Logs only important messages unless Verbose is set to true

## ⚙️ Protocols Supported

Currently, the following protocols are supported:

- **TCP**: Redirects bidirectional TCP traffic
- **UDP**: Redirects UDP packets, supporting basic data forwarding
- **HTTP**: Redirects HTTP requests to the target destination

## 🛡 License

MIT License

See LICENSE for details.
```


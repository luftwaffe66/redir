# redir

**Redir** is a simple, lightweight and robust TCP port redirector written in Go.

It allows you to redirect any incoming connection on a local port to a remote IP and port, with full bidirectional support.

---

## ✨ Features

- 🔁 Full bidirectional TCP redirection
- ⚙️ Easy-to-use configuration with Go struct
- 🧱 Can be imported as a library into any Go project
- 🚨 Robust error handling and panic safety
- 📦 Clean and dependency-free

---

## 📦 Installation

To use this package in your project:

```bash
go get github.com/luftwaffe66/redir
```

---

## 🧪 Example Usage

You can use `redir` by importing it into your Go application and initializing the redirector:

```go
package main

import (
    "log"
    "github.com/luftwaffe66/redir"
)

func main() {
    opts := redir.RedirectOptions{
        ListenPort:      "8888",           // Local port to listen on
        DestinationIP:   "192.168.1.100",  // Remote destination IP
        DestinationPort: "8080",           // Remote port to forward to
        Verbose:         true,             // Show info logs
    }

    if err := redir.StartRedirector(opts); err != nil {
        log.Fatalf("Redirector failed: %v", err)
    }
}
```

---

## 🧱 Struct Reference

```go
type RedirectOptions struct {
    ListenPort      string // Local port to bind and accept incoming connections
    DestinationIP   string // Remote target IP address
    DestinationPort string // Remote target port
    Verbose         bool   // Whether to log connection events
}
```

---

## 🔧 How It Works

When `StartRedirector()` is called, it:

1. Listens on the specified local port
2. Accepts incoming TCP connections
3. Forwards the connection to the specified remote IP and port
4. Maintains bidirectional data flow (client ↔ server)
5. Logs only important messages unless `Verbose` is true

---

## 🛡 License

MIT License

See [`LICENSE`](./LICENSE) for details.

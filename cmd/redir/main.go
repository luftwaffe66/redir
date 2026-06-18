// Command redir is a simple, lightweight multi-protocol redirector.
//
// Usage:
//
//	redir -dest host:port [-listen port] [-proto tcp|udp|http] [-v]
package main

import (
	"fmt"
	"os"

	"github.com/luftwaffe66/redir"
)

func main() {
	if err := redirector.RunCLI(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

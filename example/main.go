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

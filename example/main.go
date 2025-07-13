package main

import (
	"log"

	"github.com/tuusuario/redir"
)

func main() {
	opts := redir.RedirectOptions{
		ListenPort:      "8888",
		DestinationIP:   "192.168.1.100",
		DestinationPort: "8080",
		Verbose:         true,
	}

	if err := redir.StartRedirector(opts); err != nil {
		log.Fatalf("Redirector failed: %v", err)
	}
}

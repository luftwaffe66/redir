// Package config provides configuration parsing for the redirector.
package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Protocol enum values.
const (
	ProtocolTCP  = "tcp"
	ProtocolUDP  = "udp"
	ProtocolHTTP = "http"
)

// Config holds all redirector configuration.
type Config struct {
	Protocol    string
	ListenPort  string
	Destination string // host:port
	Verbose     bool
}

// ListenAddr returns the listen address (e.g. ":8080").
func (c *Config) ListenAddr() string {
	return ":" + c.ListenPort
}

// DestHost returns the destination host.
func (c *Config) DestHost() string {
	host, _, _ := net.SplitHostPort(c.Destination)
	return host
}

// DestPort returns the destination port.
func (c *Config) DestPort() string {
	_, port, _ := net.SplitHostPort(c.Destination)
	return port
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if _, err := strconv.Atoi(c.ListenPort); err != nil {
		return fmt.Errorf("invalid listen port %q: %w", c.ListenPort, err)
	}
	if _, _, err := net.SplitHostPort(c.Destination); err != nil {
		return fmt.Errorf("invalid destination %q, use host:port format: %w", c.Destination, err)
	}
	switch c.Protocol {
	case ProtocolTCP, ProtocolUDP, ProtocolHTTP:
	default:
		return fmt.Errorf("unsupported protocol %q, use tcp, udp, or http", c.Protocol)
	}
	return nil
}

// Parse reads configuration from command-line flags.
func Parse() (*Config, error) {
	proto := flag.String("proto", "tcp",
		"Protocol to redirect: tcp, udp, or http")
	listen := flag.String("listen", "8080",
		"Local port to listen on")
	dest := flag.String("dest", "",
		"Destination address (host:port)")
	verbose := flag.Bool("v", false,
		"Verbose logging")
	showHelp := flag.Bool("help", false,
		"Show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, strings.TrimSpace(`
Usage: redir -dest host:port [options]

Multi-protocol redirector — forwards connections from a local port
to a remote destination.

Options:
`))
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
	}

	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *dest == "" {
		return nil, errors.New("destination address is required (-dest host:port)")
	}

	isVerbose := *verbose
	if envVerbose := os.Getenv("REDIR_VERBOSE"); envVerbose == "1" || envVerbose == "true" {
		isVerbose = true
	}

	cfg := &Config{
		Protocol:    strings.ToLower(*proto),
		ListenPort:  *listen,
		Destination: *dest,
		Verbose:     isVerbose,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

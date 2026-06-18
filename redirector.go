// Package redirector provides a simple, lightweight multi-protocol redirector.
//
// It supports TCP, UDP, and HTTP redirection with full bidirectional support.
// Use StartRedirector to begin forwarding traffic from a local port to a
// remote destination.
package redirector

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/luftwaffe66/redir/internal/config"
	"github.com/luftwaffe66/redir/internal/httpproxy"
	"github.com/luftwaffe66/redir/internal/tcp"
	"github.com/luftwaffe66/redir/internal/udp"
)

// RedirectOptions configures the redirector.
type RedirectOptions struct {
	// Local port to listen on (e.g. "8080").
	ListenPort string

	// Remote destination IP address.
	DestinationIP string

	// Remote destination port.
	DestinationPort string

	// Protocol to use: "tcp", "udp", or "http".
	Protocol string

	// Verbose enables detailed logging.
	Verbose bool
}

// StartRedirector begins forwarding traffic according to the provided options.
// It sets up OS signal handling for graceful shutdown (SIGINT/SIGTERM).
func StartRedirector(ctx context.Context, opts RedirectOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := &config.Config{
		Protocol:    opts.Protocol,
		ListenPort:  opts.ListenPort,
		Destination: opts.DestinationIP + ":" + opts.DestinationPort,
		Verbose:     opts.Verbose,
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	logger := newLogger(cfg.Verbose)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("redirector starting",
		"protocol", cfg.Protocol,
		"listen", cfg.ListenAddr(),
		"dest", cfg.Destination,
	)

	return startByProtocol(ctx, cfg, logger)
}

// startByProtocol dispatches to the appropriate protocol handler.
func startByProtocol(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	switch cfg.Protocol {
	case config.ProtocolTCP:
		r := tcp.New(cfg.ListenAddr(), cfg.Destination, logger)
		return r.Start(ctx)
	case config.ProtocolUDP:
		r := udp.New(cfg.ListenAddr(), cfg.Destination, logger)
		return r.Start(ctx)
	case config.ProtocolHTTP:
		r, err := httpproxy.New(cfg.ListenAddr(), cfg.Destination, logger)
		if err != nil {
			return err
		}
		return r.Start(ctx)
	default:
		return fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}
}

// RunCLI is the entry point used by cmd/redir.
// It parses flags, configures logging, and starts the redirector.
// It returns nil on clean shutdown and an error on failure.
func RunCLI() error {
	cfg, err := config.Parse()
	if err != nil {
		return err
	}

	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "redir %s -> %s (%s)\n",
			cfg.ListenAddr(), cfg.Destination, cfg.Protocol)
	}

	logger := newLogger(cfg.Verbose)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("redirector starting",
		"protocol", cfg.Protocol,
		"listen", cfg.ListenAddr(),
		"dest", cfg.Destination,
	)

	return startByProtocol(ctx, cfg, logger)
}

func newLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level}
	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}



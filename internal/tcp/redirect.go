// Package tcp implements TCP port forwarding / redirection.
package tcp

import (
	"context"
	"io"
	"log/slog"
	"net"
	"sync"
)

const bufferSize = 32 * 1024

// Redirector handles TCP traffic forwarding.
type Redirector struct {
	listenAddr string
	destAddr   string
	logger     *slog.Logger
	mu         sync.Mutex
	addr       net.Addr // actual listener address, set after Start
}

// New creates a new TCP redirector.
func New(listenAddr, destAddr string, logger *slog.Logger) *Redirector {
	return &Redirector{
		listenAddr: listenAddr,
		destAddr:   destAddr,
		logger:     logger,
	}
}

// Addr returns the actual listening address after Start has been called.
func (r *Redirector) Addr() net.Addr {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.addr
}

// Start begins listening and forwarding TCP connections.
// Blocks until the context is cancelled or a fatal error occurs.
func (r *Redirector) Start(ctx context.Context) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", r.listenAddr)
	if err != nil {
		return err
	}
	r.addr = listener.Addr()
	defer listener.Close()

	r.mu.Lock()
	r.addr = listener.Addr()
	r.mu.Unlock()

	r.logger.Info("tcp redirector started",
		"listen", r.listenAddr,
		"dest", r.destAddr,
	)

	// Close listener on context cancellation to unblock Accept.
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil // shutdown requested
			}
			r.logger.Warn("accept error", "error", err)
			continue
		}
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			r.handleConn(c)
		}(conn)
	}
}

// handleConn manages a single proxied TCP connection.
func (r *Redirector) handleConn(client net.Conn) {
	defer client.Close()

	dest, err := net.Dial("tcp", r.destAddr)
	if err != nil {
		r.logger.Error("dial failed",
			"remote", client.RemoteAddr(),
			"dest", r.destAddr,
			"error", err,
		)
		return
	}
	defer dest.Close()

	r.logger.Debug("connection relayed",
		"remote", client.RemoteAddr(),
		"dest", r.destAddr,
	)

	// Bidirectional copy using channels to coordinate completion.
	// We use manual read/write instead of io.Copy to avoid kernel splice(2)
	// which may not properly detect EOF on some platforms (e.g. Android).
	done := make(chan error, 2)
	go func() {
		done <- copyBuf(dest, client)
		dest.Close()
	}()
	go func() {
		done <- copyBuf(client, dest)
		client.Close()
	}()
	<-done
	<-done
}

// copyBuf copies from src to dst using a reusable buffer.
// Unlike io.Copy, this avoids kernel splice/sendfile optimizations.
func copyBuf(dst io.Writer, src io.Reader) error {
	buf := make([]byte, bufferSize)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

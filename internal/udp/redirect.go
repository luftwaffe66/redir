// Package udp implements UDP packet forwarding / redirection.
package udp

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"time"
)

const (
	maxPacketSize = 65507
	readTimeout   = 30 * time.Second
)

// Redirector handles UDP traffic forwarding.
type Redirector struct {
	listenAddr string
	destAddr   string
	logger     *slog.Logger
	mu         sync.Mutex
	addr       net.Addr // actual listening address, set after Start
}

// Addr returns the actual listening address after Start has been called.
func (r *Redirector) Addr() net.Addr {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.addr
}

// New creates a new UDP redirector.
func New(listenAddr, destAddr string, logger *slog.Logger) *Redirector {
	return &Redirector{
		listenAddr: listenAddr,
		destAddr:   destAddr,
		logger:     logger,
	}
}

// Start begins listening and forwarding UDP packets.
// Blocks until the context is cancelled or a fatal error occurs.
func (r *Redirector) Start(ctx context.Context) error {
	localAddr, err := net.ResolveUDPAddr("udp", r.listenAddr)
	if err != nil {
		return err
	}
	remoteAddr, err := net.ResolveUDPAddr("udp", r.destAddr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	r.mu.Lock()
	r.addr = conn.LocalAddr()
	r.mu.Unlock()

	// Close conn on context cancellation to unblock ReadFromUDP.
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	// Establish a "connection" to the destination for sending.
	destConn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return err
	}
	defer destConn.Close()

	r.logger.Info("udp redirector started",
		"listen", r.listenAddr,
		"dest", r.destAddr,
	)

	// responses carries packets to be sent back to original clients.
	type packet struct {
		data []byte
		addr *net.UDPAddr
	}
	respCh := make(chan packet, 256)

	// Background writer: sends responses back to clients.
	var writerWg sync.WaitGroup
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case p, ok := <-respCh:
				if !ok {
					return
				}
				if _, err := conn.WriteToUDP(p.data, p.addr); err != nil {
					r.logger.Warn("writeback error", "error", err)
				}
			}
		}
	}()

	// Bounded goroutine pool for request-response forwarding.
	sem := make(chan struct{}, 100)

	buf := make([]byte, maxPacketSize)
	for {
		select {
		case <-ctx.Done():
			close(respCh)
			writerWg.Wait()
			return nil
		default:
		}

		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				close(respCh)
				writerWg.Wait()
				return nil
			}
			r.logger.Warn("read error", "error", err)
			continue
		}

		// Copy the packet data (buf will be reused).
		data := make([]byte, n)
		copy(data, buf[:n])

		select {
		case sem <- struct{}{}:
		default:
			r.logger.Warn("dropping packet: too many concurrent requests")
			continue
		}

		go func(clientAddr *net.UDPAddr, data []byte) {
			defer func() { <-sem }()

			if err := r.forwardAndReply(destConn, conn, clientAddr, data); err != nil {
				r.logger.Warn("forward error",
					"client", clientAddr,
					"error", err,
				)
			}
		}(clientAddr, data)
	}
}

// forwardAndReply sends a packet to the destination and forwards the response
// back to the client. This enables bidirectional request-response UDP forwarding.
func (r *Redirector) forwardAndReply(destConn, localConn *net.UDPConn, clientAddr *net.UDPAddr, data []byte) error {
	// Send to destination.
	if _, err := destConn.Write(data); err != nil {
		return err
	}
	r.logger.Debug("forwarded packet",
		"client", clientAddr,
		"dest", r.destAddr,
		"bytes", len(data),
	)

	// Read response from destination with timeout.
	destConn.SetReadDeadline(time.Now().Add(readTimeout))
	resp := make([]byte, maxPacketSize)
	n, err := destConn.Read(resp)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// No response — normal for fire-and-forget UDP.
			return nil
		}
		return err
	}

	// Forward response back to client.
	if _, err := localConn.WriteToUDP(resp[:n], clientAddr); err != nil {
		return err
	}
	r.logger.Debug("replied to client",
		"client", clientAddr,
		"bytes", n,
	)
	return nil
}

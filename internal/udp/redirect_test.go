package udp

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"
)

func TestUDPRedirect(t *testing.T) {
	// Start a test UDP echo server.
	echo := startUDPEchoServer(t)
	defer echo.stop()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	redir := New("127.0.0.1:0", echo.addr, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redirDone := make(chan error, 1)
	go func() {
		redirDone <- redir.Start(ctx)
	}()

	// Wait for redirector to start listening.
	var redirAddr string
	for i := 0; i < 50; i++ {
		if addr := redir.Addr(); addr != nil {
			redirAddr = addr.String()
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if redirAddr == "" {
		t.Fatal("redirector did not start in time")
	}

	// Send a packet via the redirector.
	conn, err := net.Dial("udp", redirAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	msg := []byte("hello from udp test")
	if _, err := conn.Write(msg); err != nil {
		t.Fatal(err)
	}

	// Read echoed response.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 65507)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf[:n]) != string(msg) {
		t.Errorf("got %q, want %q", string(buf[:n]), string(msg))
	}

	cancel()
	if err := <-redirDone; err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
}

type udpEchoServer struct {
	addr   string
	conn   *net.UDPConn
	doneCh chan struct{}
}

func startUDPEchoServer(t *testing.T) *udpEchoServer {
	t.Helper()
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	es := &udpEchoServer{addr: conn.LocalAddr().String(), conn: conn, doneCh: make(chan struct{})}
	go func() {
		defer close(es.doneCh)
		buf := make([]byte, 65507)
		for {
			n, clientAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			conn.WriteToUDP(buf[:n], clientAddr)
		}
	}()
	return es
}

func (es *udpEchoServer) stop() {
	es.conn.Close()
	<-es.doneCh
}

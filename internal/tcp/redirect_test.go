package tcp

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"
)

func TestTCPRedirect(t *testing.T) {
	// Start a test echo server.
	echo := startEchoServer(t)
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

	// Connect through the redirector.
	conn, err := net.Dial("tcp", redirAddr)
	if err != nil {
		t.Fatal(err)
	}

	msg := "hello from tcp test\n"
	if _, err := conn.Write([]byte(msg)); err != nil {
		t.Fatal(err)
	}

	// Read echoed response.
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if response != msg {
		t.Errorf("got %q, want %q", response, msg)
	}

	// Close the connection so the redirector's io.Copy goroutines can finish.
	conn.Close()
	cancel()
	if err := <-redirDone; err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
}

type echoServer struct {
	addr   string
	ln     net.Listener
	doneCh chan struct{}
}

func startEchoServer(t *testing.T) *echoServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	es := &echoServer{addr: ln.Addr().String(), ln: ln, doneCh: make(chan struct{})}
	go func() {
		defer close(es.doneCh)
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				scanner := bufio.NewScanner(c)
				for scanner.Scan() {
					c.Write([]byte(scanner.Text() + "\n"))
				}
			}(conn)
		}
	}()
	return es
}

func (es *echoServer) stop() {
	es.ln.Close()
	<-es.doneCh
}

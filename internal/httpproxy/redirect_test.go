package httpproxy

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPRedirect(t *testing.T) {
	// Start a test backend server.
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backend response"))
	}))
	defer backend.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Backend's host:port.
	destAddr := backend.Listener.Addr().String()

	redir, err := New("127.0.0.1:0", destAddr, logger)
	if err != nil {
		t.Fatal(err)
	}

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

	// Send a request through the redirector.
	resp, err := http.Get("http://" + redirAddr + "/test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "backend response" {
		t.Errorf("got %q, want %q", string(body), "backend response")
	}

	cancel()
	if err := <-redirDone; err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
}

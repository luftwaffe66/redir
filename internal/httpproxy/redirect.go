// Package httpproxy implements HTTP reverse proxying / redirection.
package httpproxy

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	readTimeout  = 30 * time.Second
	writeTimeout = 30 * time.Second
	idleTimeout  = 60 * time.Second
)

// Redirector handles HTTP traffic forwarding as a reverse proxy.
type Redirector struct {
	listenAddr string
	destURL    *url.URL
	logger     *slog.Logger
	addr       net.Addr // actual listener address, set after Start
}

// Addr returns the actual listening address after Start has been called.
func (r *Redirector) Addr() net.Addr { return r.addr }

// New creates a new HTTP redirector.
// destAddr must be in host:port format.
func New(listenAddr, destAddr string, logger *slog.Logger) (*Redirector, error) {
	destURL, err := url.Parse("http://" + destAddr)
	if err != nil {
		return nil, err
	}
	return &Redirector{
		listenAddr: listenAddr,
		destURL:    destURL,
		logger:     logger,
	}, nil
}

// Start begins listening and proxying HTTP requests.
// Blocks until the server is stopped or a fatal error occurs.
func (r *Redirector) Start(ctx context.Context) error {
	proxy := httputil.NewSingleHostReverseProxy(r.destURL)
	proxy.ErrorLog = slog.NewLogLogger(r.logger.Handler(), slog.LevelWarn)

	// Customize the Director to log requests.
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		r.logger.Debug("proxying request",
			"method", req.Method,
			"path", req.URL.Path,
			"dest", r.destURL.Host,
		)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         r.listenAddr,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		ErrorLog:     slog.NewLogLogger(r.logger.Handler(), slog.LevelWarn),
	}

	// Use an explicit listener so we can capture the actual address.
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", r.listenAddr)
	if err != nil {
		return err
	}
	r.addr = listener.Addr()
	defer listener.Close()

	r.logger.Info("http redirector started",
		"listen", listener.Addr().String(),
		"dest", r.destURL.Host,
	)

	// Shutdown on context cancellation.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			r.logger.Error("shutdown error", "error", err)
		}
	}()

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

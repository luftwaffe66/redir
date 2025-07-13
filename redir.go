package redir

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type RedirectOptions struct {
	ListenPort      string // Port to listen on locally
	DestinationIP   string // IP of the target machine
	DestinationPort string // Port of the target machine
	Verbose         bool   // Show info logs or not
}

// StartRedirector runs the TCP port redirector.
// It listens on ListenPort and forwards traffic to DestinationIP:DestinationPort.
func StartRedirector(opts RedirectOptions) error {
	destAddr := fmt.Sprintf("%s:%s", opts.DestinationIP, opts.DestinationPort)
	listenAddr := ":" + opts.ListenPort

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to bind to port %s: %w", opts.ListenPort, err)
	}
	defer listener.Close()

	log(opts, "[INFO]", fmt.Sprintf("Listening on :%s → %s", opts.ListenPort, destAddr))

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log(opts, "[WARN]", fmt.Sprintf("Failed to accept connection: %v", err))
			time.Sleep(time.Second)
			continue
		}

		go handleConnection(opts, clientConn, destAddr)
	}
}

func handleConnection(opts RedirectOptions, client net.Conn, dest string) {
	server, err := net.Dial("tcp", dest)
	if err != nil {
		log(opts, "[ERROR]", fmt.Sprintf("Failed to connect to %s: %v", dest, err))
		client.Close()
		return
	}

	log(opts, "[INFO]", fmt.Sprintf("New connection: %s → %s", client.RemoteAddr(), dest))

	go proxyData(opts, client, server)
	go proxyData(opts, server, client)
}

func proxyData(opts RedirectOptions, src net.Conn, dst net.Conn) {
	defer src.Close()
	defer dst.Close()

	_, err := io.Copy(dst, src)
	if err != nil && !isConnReset(err) {
		log(opts, "[WARN]", fmt.Sprintf("I/O error %s → %s: %v", src.RemoteAddr(), dst.RemoteAddr(), err))
	}
}

func isConnReset(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "use of closed network connection")
}

func log(opts RedirectOptions, level, message string) {
	if level == "[INFO]" && !opts.Verbose {
		return
	}
	fmt.Printf("%s %s\n", level, message)
}

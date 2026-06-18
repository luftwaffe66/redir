package config

import (
	"testing"
)

func TestValidateValid(t *testing.T) {
	tt := []struct {
		name string
		cfg  Config
	}{
		{"tcp", Config{Protocol: "tcp", ListenPort: "8080", Destination: "10.0.0.1:9090"}},
		{"udp", Config{Protocol: "udp", ListenPort: "53", Destination: "1.1.1.1:53"}},
		{"http", Config{Protocol: "http", ListenPort: "80", Destination: "192.168.1.1:3000"}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestValidateInvalid(t *testing.T) {
	tt := []struct {
		name string
		cfg  Config
	}{
		{"bad protocol", Config{Protocol: "quic", ListenPort: "8080", Destination: "10.0.0.1:9090"}},
		{"bad listen port", Config{Protocol: "tcp", ListenPort: "abc", Destination: "10.0.0.1:9090"}},
		{"bad destination", Config{Protocol: "tcp", ListenPort: "8080", Destination: "invalid"}},
		{"empty destination", Config{Protocol: "tcp", ListenPort: "8080", Destination: ""}},
		{"empty listen port", Config{Protocol: "tcp", ListenPort: "", Destination: "10.0.0.1:9090"}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); err == nil {
				t.Errorf("Validate() expected error, got nil")
			}
		})
	}
}

func TestListenAddr(t *testing.T) {
	cfg := Config{ListenPort: "9999"}
	if got := cfg.ListenAddr(); got != ":9999" {
		t.Errorf("ListenAddr() = %q, want %q", got, ":9999")
	}
}

func TestDestParts(t *testing.T) {
	cfg := Config{Destination: "10.0.0.1:53"}
	if got := cfg.DestHost(); got != "10.0.0.1" {
		t.Errorf("DestHost() = %q, want %q", got, "10.0.0.1")
	}
	if got := cfg.DestPort(); got != "53" {
		t.Errorf("DestPort() = %q, want %q", got, "53")
	}
}

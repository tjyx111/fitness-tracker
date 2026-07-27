package main

import (
	"crypto/tls"
	"strings"
	"testing"
)

func TestResolveTLSFiles(t *testing.T) {
	t.Setenv("TLS_CERT_FILE", "/tmp/server.crt")
	t.Setenv("TLS_KEY_FILE", "/tmp/server.key")

	config, err := resolveTLSFiles()
	if err != nil {
		t.Fatalf("resolveTLSFiles returned an error: %v", err)
	}
	if !config.enabled() {
		t.Fatal("expected TLS to be enabled")
	}
}

func TestResolveTLSFilesRejectsPartialConfiguration(t *testing.T) {
	t.Setenv("TLS_CERT_FILE", "/tmp/server.crt")
	t.Setenv("TLS_KEY_FILE", "")

	_, err := resolveTLSFiles()
	if err == nil || !strings.Contains(err.Error(), "configured together") {
		t.Fatalf("expected paired certificate error, got %v", err)
	}
}

func TestNewTLSConfigDoesNotRequestClientCertificate(t *testing.T) {
	config := newTLSConfig()
	if config.MinVersion != tls.VersionTLS12 {
		t.Fatalf("minimum TLS version = %d, want TLS 1.2", config.MinVersion)
	}
	if config.ClientAuth != tls.NoClientCert {
		t.Fatalf("client auth = %d, want no client certificate", config.ClientAuth)
	}
}

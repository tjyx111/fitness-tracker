package main

import (
	"crypto/tls"
	"fmt"
	"os"
)

type tlsFileConfig struct {
	certFile string
	keyFile  string
}

func resolveTLSFiles() (tlsFileConfig, error) {
	config := tlsFileConfig{
		certFile: os.Getenv("TLS_CERT_FILE"),
		keyFile:  os.Getenv("TLS_KEY_FILE"),
	}

	if (config.certFile == "") != (config.keyFile == "") {
		return tlsFileConfig{}, fmt.Errorf("TLS_CERT_FILE and TLS_KEY_FILE must be configured together")
	}
	return config, nil
}

func (config tlsFileConfig) enabled() bool {
	return config.certFile != "" && config.keyFile != ""
}

func newTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
}

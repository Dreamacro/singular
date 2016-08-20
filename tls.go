package singular

import (
	"crypto/tls"
)

// NewTLSConfig return tls.config
func NewTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return &tls.Config{}, err
	}
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return &config, nil
}

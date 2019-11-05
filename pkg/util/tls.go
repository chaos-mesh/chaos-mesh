package util

import "crypto/tls"

// config tls cert for server
func ConfigTLS(certFile string, keyFile string) (*tls.Config, error) {
	sCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}, nil
}

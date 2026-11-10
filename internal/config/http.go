package config

import (
	"net"
	"os"
)

type httpConfig struct {
	host string
	port string
}

func NewHTTPConfig() (HTTPConfig, error) {
	host := os.Getenv("HTTP_HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	return &httpConfig{host: host, port: port}, nil
}

func (cfg *httpConfig) Address() string {
	return net.JoinHostPort(cfg.host, cfg.port)
}

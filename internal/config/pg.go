package config

import (
	"errors"
	"os"
)

type pgConfig struct {
	dsn string
}

func NewPGConfig() (PGConfig, error) {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		return nil, errors.New("PG_DSN is not set")
	}
	return &pgConfig{dsn: dsn}, nil
}

func (cfg *pgConfig) DSN() string {
	return cfg.dsn
}

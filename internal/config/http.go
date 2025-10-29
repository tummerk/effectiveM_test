package config

import (
	"time"
)

type HTTP struct {
	Addr         string        `env:"HTTP_ADDR"        envDefault:":8080"`
	BaseURL      string        `env:"HTTP_BASE_URL"    envDefault:""`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT"  envDefault:"10s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"60s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT"  envDefault:"120s"`
}

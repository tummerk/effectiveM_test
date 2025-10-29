package config

import "fmt"

type Postgres struct {
	Username string `env:"POSTGRES_USERNAME" RenvDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD,required"`
	Database string `env:"POSTGRES_DB,required"`
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	SSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`
	Timeout  int    `env:"POSTGRES_TIMEOUT" envDefault:"30"`
}

func (p *Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		p.Username, p.Password, p.Host, p.Port, p.Database, p.SSLMode)
}

package config

type Logger struct {
	Debug bool `env:"DEBUG" envDefault:"false"`
}

package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"log"
)

type Config struct {
	Postgres Postgres
	Logger   Logger `envPrefix:"LOGGER_"`
	HTTP     HTTP   `envPrefix:"HTTP_"`
}

func Load() (Config, error) {
	var config Config
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file", err)
	}
	if err = env.Parse(&config); err != nil {
		return Config{}, err
	}
	return config, nil
}

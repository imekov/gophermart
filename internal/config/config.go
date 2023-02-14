package config

import (
	"crypto/rand"
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Secret               []byte
}

func GetConfig() Config {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "HTTP server start address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "the base address of the resulting shortened URL")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "the path to file with shortened URLs")
	flag.Parse()

	cfg.Secret = make([]byte, 16)
	_, err = rand.Read(cfg.Secret)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

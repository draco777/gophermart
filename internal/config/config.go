package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", "", "Server address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system address")
	flag.Parse()

	// Переменные окружения имеют приоритет над флагами
	if runAddr := os.Getenv("RUN_ADDRESS"); runAddr != "" {
		cfg.RunAddress = runAddr
	}
	if dbURI := os.Getenv("DATABASE_URI"); dbURI != "" {
		cfg.DatabaseURI = dbURI
	}
	if accrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualAddr != "" {
		cfg.AccrualSystemAddress = accrualAddr
	}

	if cfg.DatabaseURI == "" {
		return nil, fmt.Errorf("database URI is required")
	}
	if cfg.AccrualSystemAddress == "" {
		return nil, fmt.Errorf("accrual system address is required")
	}

	return cfg, nil
}

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
	RunMigrations        bool
}

func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "Server address")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Accrual system address")
	flag.BoolVar(&cfg.RunMigrations, "migrate", true, "Run database migrations on startup")
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
	// RUN_MIGRATIONS может быть "true", "1", "yes" для включения
	if runMigrations := os.Getenv("RUN_MIGRATIONS"); runMigrations != "" {
		cfg.RunMigrations = runMigrations == "true" || runMigrations == "1" || runMigrations == "yes"
	}

	if cfg.DatabaseURI == "" {
		return nil, fmt.Errorf("database URI is required")
	}
	if cfg.AccrualSystemAddress == "" {
		return nil, fmt.Errorf("accrual system address is required")
	}

	return cfg, nil
}

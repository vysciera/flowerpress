package config

import "os"

type Config struct {
	Address		string
	DatabasePath	string
	Environment	string
}

func Load() Config {
	return Config{
		Address:		getEnv("FLOWERPRESS_ADDRESS", ":8080"),
		DatabasePath:		getEnv("FLOWERPRESS_DATABASE", "data/flowerpress.db"),
		Environment:		getEnv("FLOWERPRESS_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

package config

import (
	"os"
)

// Config holds the application configuration.
type Config struct {
	Port           string
	DatabasePath   string
	ModDataPath    string
	CacheDuration  int // Seconds
}

// LoadConfig reads configuration from environment variables or sets defaults.
func LoadConfig() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		DatabasePath:   getEnv("DB_PATH", "./mc-servers.db"),
		ModDataPath:    getEnv("MOD_DATA_PATH", "data/mods"),
		CacheDuration:  60, 
	}
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

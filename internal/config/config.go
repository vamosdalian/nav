package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	ServerPort    string
	OSMDataPath   string
	GraphDataPath string
	LogLevel      string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		ServerPort:    getEnv("PORT", "8080"),
		OSMDataPath:   getEnv("OSM_DATA_PATH", ""),
		GraphDataPath: getEnv("GRAPH_DATA_PATH", "graph.bin.snappy"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.OSMDataPath == "" && c.GraphDataPath == "" {
		return fmt.Errorf("either OSM_DATA_PATH or GRAPH_DATA_PATH must be set")
	}
	return nil
}

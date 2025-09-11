package config

import (
	"os"
)

// Config holds all configuration for the BOAT agent
type Config struct {
	// Claude API configuration
	ClaudeAPIKey string

	// AWS configuration
	AWSRegion string

	// Container registry configuration
	ContainerRegistry string

	// Cost optimization settings
	MaxCostThreshold float64

	// Notification settings
	NotificationLevel string
}

// Load initializes configuration from environment variables
func Load() (*Config, error) {
	return &Config{
		ClaudeAPIKey:      getEnvOrDefault("CLAUDE_API_KEY", ""),
		AWSRegion:         getEnvOrDefault("AWS_REGION", "us-east-1"),
		ContainerRegistry: getEnvOrDefault("CONTAINER_REGISTRY", "ghcr.io/jrzesz33/"),
		MaxCostThreshold:  100.0, // Stay within free tier limits
		NotificationLevel: getEnvOrDefault("NOTIFICATION_LEVEL", "INFO"),
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
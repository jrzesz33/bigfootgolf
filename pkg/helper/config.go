// Package helper provides utility functions for the golf booking application.
package helper

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// ConfigValidationError represents a configuration validation error
type ConfigValidationError struct {
	MissingVars []string
	Warnings    []string
}

func (e *ConfigValidationError) Error() string {
	var msg strings.Builder
	if len(e.MissingVars) > 0 {
		msg.WriteString(fmt.Sprintf("Missing required environment variables: %s\n", strings.Join(e.MissingVars, ", ")))
	}
	if len(e.Warnings) > 0 {
		msg.WriteString(fmt.Sprintf("Warnings: %s", strings.Join(e.Warnings, "; ")))
	}
	return msg.String()
}

// ValidateConfig validates that all required environment variables are set
// Returns an error if any required variables are missing
func ValidateConfig() error {
	required := map[string]string{
		"DB_ADMIN":   "Neo4j database admin password",
		"JWT_SECRET": "JWT secret for token signing",
	}
	/*
		optional := map[string]string{
			"DB_URI":             "Neo4j database URI (defaults to bolt://localhost:7687)",
			"DB_USER":            "Neo4j username (defaults to neo4j)",
			"LLM_GATEWAY_URL":    "LLM Gateway URL for AI features",
			"LLM_MODEL":          "LLM model to use",
			"MCP_GATEWAY_URL":    "MCP Gateway URL",
			"MCP_SERVER_URL":     "MCP Server URL",
			"SESSION_KEY":        "Session encryption key",
			"GOOGLE_CLIENT_ID":   "Google OAuth client ID",
			"APPLE_CLIENT_ID":    "Apple OAuth client ID",
			"GMAIL_USER":         "Gmail SMTP user",
			"PORT":               "Server port (defaults to 8000)",
			"TZ":                 "Timezone (defaults to America/New_York)",
		}
	*/
	var missing []string
	var warnings []string

	// Check required variables
	for varName, description := range required {
		if os.Getenv(varName) == "" {
			missing = append(missing, fmt.Sprintf("%s (%s)", varName, description))
		}
	}

	// Check optional variables and warn if missing
	if os.Getenv("LLM_GATEWAY_URL") == "" {
		warnings = append(warnings, "LLM_GATEWAY_URL not set - AI features will be disabled")
	}
	if os.Getenv("MCP_GATEWAY_URL") == "" {
		warnings = append(warnings, "MCP_GATEWAY_URL not set - MCP features will be disabled")
	}
	if os.Getenv("SESSION_KEY") == "" {
		warnings = append(warnings, "SESSION_KEY not set - using default key (not secure for production)")
	}

	// Log optional variables for informational purposes
	if len(warnings) > 0 {
		log.Println("Configuration Warnings:")
		for _, warning := range warnings {
			log.Printf("  - %s\n", warning)
		}
	}

	if len(missing) > 0 {
		return &ConfigValidationError{
			MissingVars: missing,
			Warnings:    warnings,
		}
	}

	// Log successfully loaded config
	log.Println("✓ Configuration validation passed")
	return nil
}

// GetEnvOrDefault returns the value of an environment variable or a default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// MustGetEnv returns the value of an environment variable or panics if not set
func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("FATAL: Required environment variable %s is not set", key)
	}
	return value
}

// ListConfigVars logs all configured environment variables (without values for security)
func ListConfigVars() {
	configVars := []string{
		"DB_ADMIN", "DB_URI", "DB_USER", "JWT_SECRET", "SESSION_KEY",
		"LLM_GATEWAY_URL", "LLM_MODEL", "MCP_GATEWAY_URL", "MCP_SERVER_URL",
		"GOOGLE_CLIENT_ID", "APPLE_CLIENT_ID", "GMAIL_USER", "PORT", "TZ", "MODE",
	}

	log.Println("Environment Variables Status:")
	for _, varName := range configVars {
		status := "✗ NOT SET"
		if os.Getenv(varName) != "" {
			status = "✓ SET"
		}
		log.Printf("  %s: %s\n", varName, status)
	}
}

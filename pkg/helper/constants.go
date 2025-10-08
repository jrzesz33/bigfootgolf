// Package helper provides utility functions for the golf booking application.
package helper

import "time"

// HTTP Configuration
const (
	// DefaultHTTPTimeout is the default timeout for HTTP requests
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultPort is the default server port
	DefaultPort = "8000"

	// DefaultMCPPort is the default MCP server port
	DefaultMCPPort = "8081"
)

// Database Configuration
const (
	// DefaultDBPoolSize is the default maximum connection pool size for Neo4j
	DefaultDBPoolSize = 50

	// DefaultDBURI is the default Neo4j connection URI
	DefaultDBURI = "bolt://localhost:7687"

	// DefaultDBUser is the default Neo4j username
	DefaultDBUser = "neo4j"
)

// OAuth URLs
const (
	// GoogleTokenURL is the URL for Google OAuth token exchange
	GoogleTokenURL = "https://oauth2.googleapis.com/token"

	// GoogleUserInfoURL is the URL for fetching Google user information
	GoogleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

	// AppleTokenURL is the URL for Apple OAuth token exchange
	AppleTokenURL = "https://appleid.apple.com/auth/token"

	// AppleUserInfoURL is the URL for fetching Apple user information
	AppleUserInfoURL = "https://appleid.apple.com/auth/userinfo"
)

// Security Constants
const (
	// JWTSecretLength is the required length for JWT secrets in bytes
	JWTSecretLength = 32

	// SessionKeyLength is the required length for session keys in bytes
	SessionKeyLength = 32
)

// Application Defaults
const (
	// DefaultTimezone is the default application timezone
	DefaultTimezone = "America/New_York"

	// DevMode is the value for development mode
	DevMode = "dev"

	// ProductionMode is the value for production mode
	ProductionMode = "production"
)

// API Timeouts
const (
	// LLMRequestTimeout is the timeout for LLM API requests
	LLMRequestTimeout = 60 * time.Second

	// MCPRequestTimeout is the timeout for MCP requests
	MCPRequestTimeout = 30 * time.Second

	// WeatherAPITimeout is the timeout for weather API requests
	WeatherAPITimeout = 10 * time.Second
)

// HTTP Status Messages
const (
	// ErrInvalidRequest is returned when request body cannot be parsed
	ErrInvalidRequest = "Invalid request body"

	// ErrUnauthorized is returned when user is not authenticated
	ErrUnauthorized = "Unauthorized"

	// ErrForbidden is returned when user lacks permissions
	ErrForbidden = "Forbidden"

	// ErrNotFound is returned when resource is not found
	ErrNotFound = "Not found"

	// ErrInternalServer is returned for unexpected errors
	ErrInternalServer = "Internal server error"
)

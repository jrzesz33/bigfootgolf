// Package auth provides authentication and authorization services for the golf
// booking application. It supports multiple authentication methods including:
// - JWT token-based authentication for local accounts
// - OAuth integration with Google and Apple
// - Session management with secure cookie handling
//
// The package manages auth configuration persistence in Neo4j and provides
// utilities for token generation, validation, and user authentication flows.
package auth

import (
	"bigfoot/golf/common/helper"
	"bigfoot/golf/common/models/db"
	"crypto/rand"
	"encoding/json"
	"log"
	"os"
)

type AuthConfig struct {
	ID           string `json:"id"`
	GoogleConfig []byte `json:"googleConfig"`
	AppleConfig  []byte `json:"appleConfig"`
	LocalJSec    []byte `json:"localJSec"`
}

func NewAuthConfig(srv AuthServer) (AuthConfig, error) {
	var config AuthConfig

	config.LocalJSec = srv.jwtSecret
	apple, err := json.Marshal(srv.appleConfig)
	if err != nil {
		return config, err
	}
	google, err := json.Marshal(srv.googleConfig)
	if err != nil {
		return config, err
	}
	config.AppleConfig = apple
	config.GoogleConfig = google
	config.Save()
	return config, nil
}

func (a *AuthConfig) Save() error {
	_, err := db.Instance.SaveStruct(a, "AuthConfig")
	return err
}
func (a *AuthConfig) GetServer() (AuthServer, error) {
	var srv AuthServer
	var google OAuthConfig
	var apple OAuthConfig
	erg := json.Unmarshal(a.GoogleConfig, &google)
	if erg != nil {
		return srv, erg
	}
	era := json.Unmarshal(a.AppleConfig, &apple)
	if era != nil {
		return srv, era
	}
	srv.appleConfig = apple
	srv.googleConfig = google

	srv.jwtSecret = GetJWTSecret()
	return srv, nil
}
func LoadLocalConfig() (*AuthConfig, error) {
	_configs, err := db.Instance.QueryNodes("AuthConfig", nil)
	if len(_configs) > 0 {
		config := AuthConfig{
			ID:           _configs[0]["id"].(string),
			LocalJSec:    _configs[0]["localJSec"].([]byte),
			GoogleConfig: _configs[0]["googleConfig"].([]byte),
			AppleConfig:  _configs[0]["appleConfig"].([]byte),
		}

		return &config, err
	}
	return nil, err
}

func InitAuth() AuthServer {

	//Load variables for Auth
	config, err := LoadLocalConfig()
	if err != nil || config == nil {
		//no local config so make one

		// Load JWT secret from environment variable
		// If not set, generate a random one and warn the user
		var jwtSecret []byte
		jwtSecretEnv := os.Getenv("JWT_SECRET")
		if jwtSecretEnv != "" {
			jwtSecret = []byte(jwtSecretEnv)
		} else {
			// Generate random secret as fallback
			jwtSecret = make([]byte, helper.JWTSecretLength)
			rand.Read(jwtSecret)
			log.Println("WARNING: JWT_SECRET not set in environment. Using randomly generated secret. All existing sessions will be invalidated on restart.")
		}

		// Initialize OAuth configs (replace with your actual credentials)
		googleConfig := OAuthConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			TokenURL:     helper.GoogleTokenURL,
			UserInfoURL:  helper.GoogleUserInfoURL,
		}

		appleConfig := OAuthConfig{
			ClientID:     os.Getenv("APPLE_CLIENT_ID"),
			ClientSecret: os.Getenv("APPLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("APPLE_REDIRECT_URL"),
			TokenURL:     helper.AppleTokenURL,
			UserInfoURL:  helper.AppleUserInfoURL,
		}
		server := AuthServer{
			jwtSecret:    jwtSecret,
			googleConfig: googleConfig,
			appleConfig:  appleConfig,
		}
		_, _ = NewAuthConfig(server)
		return server
	}
	server, err := config.GetServer()
	if err != nil {
		log.Printf("ERROR: Failed to load auth config: %v. Creating new config...", err)
		// Try to create new config instead of fatal error
		jwtSecret := make([]byte, helper.JWTSecretLength)
		rand.Read(jwtSecret)
		server = AuthServer{
			jwtSecret:    jwtSecret,
			googleConfig: OAuthConfig{},
			appleConfig:  OAuthConfig{},
		}
		_, _ = NewAuthConfig(server)
	}
	return server

}

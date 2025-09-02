package auth

import (
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

	srv.jwtSecret = []byte(a.LocalJSec)
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

		// Generate JWT secret (in production, use environment variable)
		jwtSecret := make([]byte, 32)
		rand.Read(jwtSecret)

		// Initialize OAuth configs (replace with your actual credentials)
		googleConfig := OAuthConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		}

		appleConfig := OAuthConfig{
			ClientID:     os.Getenv("APPLE_CLIENT_ID"),
			ClientSecret: os.Getenv("APPLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("APPLE_REDIRECT_URL"),
			TokenURL:     "https://appleid.apple.com/auth/token",
			UserInfoURL:  "https://appleid.apple.com/auth/userinfo",
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
		log.Fatal("error loading auth config")
	}
	return server

}

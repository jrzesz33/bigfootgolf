package state

import (
	"birdsfoot/app/models/account"
	"birdsfoot/app/models/auth"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RefreshInterval defines how often tokens are refreshed in minutes
const RefreshInterval int64 = 5

// Token storage
type TokenBot struct {
	mu              sync.RWMutex
	authBody        *auth.AuthResponse
	refreshInterval time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
	eventChan       chan TokenEvent
	isRunning       bool
}
type TokenEvent struct {
	Type     string // "refresh_success", "refresh_failed", "token_expired"
	Message  string
	Error    error
	AuthResp *auth.AuthResponse
}

func NewTokenBot(_tokens *auth.AuthResponse) *TokenBot {
	actx, cancel := context.WithCancel(context.Background())
	tm := &TokenBot{
		refreshInterval: time.Duration(RefreshInterval) * time.Minute, // Check every 5 minutes
		ctx:             actx,
		cancel:          cancel,
		eventChan:       make(chan TokenEvent, 10),
		authBody:        _tokens,
	}
	//tm.loadTokensFromStorage()
	tm.checkAndRefreshToken(false)
	return tm
}

func (tm *TokenBot) Start() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.isRunning {
		return
	}

	tm.isRunning = true
	go tm.backgroundRefresh()
}

func (tm *TokenBot) Stop() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.isRunning {
		return
	}

	tm.cancel()
	tm.isRunning = false
}

func (tm *TokenBot) backgroundRefresh() {
	fmt.Println("Starting Background Refresh")
	ticker := time.NewTicker(tm.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tm.ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("Checking for Refresh Token on Timer")
			tm.checkAndRefreshToken(false)
		}
	}
}
func (tm *TokenBot) ForceRefresh() string {
	fmt.Println("Forcing Refresh")
	return tm.checkAndRefreshToken(true)
}
func (tm *TokenBot) checkAndRefreshToken(force bool) string {

	tm.mu.RLock()
	_authBody := tm.authBody
	tm.mu.RUnlock()

	if _authBody == nil {
		fmt.Println("No Refresh Token")
		return "" // No token to refresh
	}
	//force ignores the expiration and forces the refresh
	// Check if token expires within next 10 minutes
	if !force && time.Until(_authBody.ExpiresIn) > 10*time.Minute {
		return "" // Token is still valid
	}

	// Attempt to refresh token
	authResp, err := refreshTokens(_authBody.RefreshToken)
	if err != nil || authResp == nil {
		tm.eventChan <- TokenEvent{
			Type:    "refresh_failed",
			Message: "Failed to refresh token",
			Error:   err,
		}
		return ""
	}

	// Update tokens
	tm.mu.Lock()
	tm.authBody = authResp
	tm.mu.Unlock()

	// Store in localStorage (if available)
	//tm.saveTokensToStorage()

	tm.eventChan <- TokenEvent{
		Type:     "refresh_success",
		Message:  "Token refreshed successfully",
		AuthResp: authResp,
	}
	return authResp.Token
}

func refreshTokens(refreshToken string) (*auth.AuthResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	refreshReq := map[string]string{
		"refresh_token": refreshToken,
	}

	reqBody, _ := json.Marshal(refreshReq)

	resp, err := http.Post("/auth/refresh", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed with status: %d", resp.StatusCode)
	}

	var refreshResp auth.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		return nil, err
	}

	return &refreshResp, nil
}
func (tm *TokenBot) SetTokens(_tokens auth.AuthResponse) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.authBody = &_tokens
}

func (tm *TokenBot) UpdateUser(_user account.User) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.authBody != nil {
		tm.authBody.User = _user
		//publish event
		tm.eventChan <- TokenEvent{
			Type:     "update_user",
			Message:  "User Updated",
			AuthResp: tm.authBody,
		}
	}
}

func (tm *TokenBot) GetAuth() *auth.AuthResponse {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.authBody
}

func (tm *TokenBot) ClearTokens() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.authBody = nil

}

func (tm *TokenBot) EventChannel() <-chan TokenEvent {
	return tm.eventChan
}

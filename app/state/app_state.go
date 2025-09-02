package state

import (
	"birdsfoot/app/models"
	"birdsfoot/app/models/account"
	"birdsfoot/app/models/auth"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type AppState struct {
	mu           sync.RWMutex
	tokenManager *TokenBot
	subscribers  []chan StateEvent
	isAuth       bool
}

type StateEvent struct {
	Type string
	Data interface{}
}

var (
	appState *AppState
	once     sync.Once
)

func GetAppState(_localTokens *auth.AuthResponse) *AppState {
	once.Do(func() {
		appState = &AppState{
			tokenManager: NewTokenBot(_localTokens),
			subscribers:  make([]chan StateEvent, 0),
		}
		appState.init()
	})

	return appState
}

func (as *AppState) init() {
	// Start token manager
	as.tokenManager.Start()

	// Listen for token events
	go as.handleTokenEvents()
}

func (as *AppState) handleTokenEvents() {
	for event := range as.tokenManager.EventChannel() {
		fmt.Println("State Event Triggered ", event)
		switch event.Type {
		case "refresh_failed":
			as.mu.Lock()
			as.isAuth = false
			as.mu.Unlock()

			// Notify all subscribers
			as.notifySubscribers(StateEvent{
				Type: "auth_failed",
				Data: event.Message,
			})

		case "refresh_success":
			as.mu.Lock()
			as.isAuth = true
			as.mu.Unlock()
			var authResp auth.AuthResponse
			authResp = *event.AuthResp
			as.notifySubscribers(StateEvent{
				Type: "token_refreshed",
				Data: authResp,
			})
			//update_user
		case "update_user":
			as.mu.Lock()
			as.isAuth = true
			as.mu.Unlock()
			var authResp auth.AuthResponse
			authResp = *event.AuthResp
			as.notifySubscribers(StateEvent{
				Type: "update_user",
				Data: authResp,
			})

		default:
			fmt.Println("State Engine UNKNOWN EVENT ", event.Type)
		}
	}
}

func (as *AppState) Subscribe() chan StateEvent {
	as.mu.Lock()
	defer as.mu.Unlock()

	ch := make(chan StateEvent, 10)
	as.subscribers = append(as.subscribers, ch)
	return ch
}

func (as *AppState) Unsubscribe(ch chan StateEvent) {
	as.mu.Lock()
	defer as.mu.Unlock()

	for i, sub := range as.subscribers {
		if sub == ch {
			close(sub)
			as.subscribers = append(as.subscribers[:i], as.subscribers[i+1:]...)
			break
		}
	}
}

func (as *AppState) notifySubscribers(event StateEvent) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	for _, sub := range as.subscribers {
		select {
		case sub <- event:
		default:
			// Channel is full, skip this subscriber
		}
	}
}

func (as *AppState) TokenManager() *TokenBot {

	return as.tokenManager
}

func (as *AppState) IsAuthenticated() bool {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.tokenManager.authBody.AuthLevel > auth.NoAuthLevel
}
func (as *AppState) UpdateUser(_user account.User) {
	as.mu.RLock()
	defer as.mu.RUnlock()
	as.tokenManager.UpdateUser(_user)
}

func (as *AppState) CacheEvent(_type string, _data interface{}) {
	_event := StateEvent{Type: _type, Data: _data}
	as.notifySubscribers(_event)
}

func (as *AppState) ForceRefresh() string {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.tokenManager.ForceRefresh()
}

func (as *AppState) Login(authReq auth.LoginRequest) error {
	as.mu.Lock()
	as.isAuth = true
	as.mu.Unlock()

	url := "./auth/login"
	reqBody, _ := json.Marshal(authReq)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var authResp auth.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}
	as.tokenManager.SetTokens(authResp)

	as.notifySubscribers(StateEvent{
		Type: "login_success",
		Data: authResp,
	})

	return nil
}

func (as *AppState) RegisterUser(_user account.User) models.BError {
	url := "./auth/register"
	_err := models.BError{Request: url}
	_payload, err := json.Marshal(_user)
	if err != nil {
		fmt.Println(err, "Error marshalling")
		_err.BError = err
		return _err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(_payload))
	if err != nil {
		_err.BError = err
		return _err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_err.BError = fmt.Errorf(resp.Status)
		_err.Code = resp.StatusCode
		return _err
	}

	var authResp auth.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		_err.BError = err
		return _err
	}

	as.tokenManager.SetTokens(authResp)

	as.notifySubscribers(StateEvent{
		Type: "login_success",
		Data: authResp,
	})

	return _err

}

func (as *AppState) Logout() {
	as.mu.Lock()
	as.isAuth = false
	as.mu.Unlock()

	as.tokenManager.ClearTokens()

	as.notifySubscribers(StateEvent{
		Type: "logout",
		Data: "User logged out",
	})
}

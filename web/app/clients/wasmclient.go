package clients

import (
	"bigfoot/golf/common/models"
	"bigfoot/golf/web/app/state"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SendPostWithPayload sends a POST request with JSON payload to the specified endpoint
func SendPostWithPayload(baseURL, payload string) ([]byte, error) {
	// Build URL with ID in path
	fullURL := fmt.Sprintf("./%s", baseURL)

	// Create context with timeout (important for WASM)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// CORS headers might be needed depending on your setup
	req.Header.Set("Access-Control-Allow-Origin", "*")

	// Use default client (works in WASM)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	//fmt.Printf("Response: %s\n", string(body))

	return body, nil
}

// SendGetReq sends a GET request to the specified endpoint
func SendGetReq(baseURL string) ([]byte, error) {
	// Build URL with ID in path
	fullURL := fmt.Sprintf("./%s", baseURL)

	// Create context with timeout (important for WASM)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// CORS headers might be needed depending on your setup
	req.Header.Set("Access-Control-Allow-Origin", "*")

	// Use default client (works in WASM)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	//fmt.Printf("Response: %s\n", string(body))

	return body, nil
}

// SendPostWithAuth sends a POST request with authentication token and handles token refresh
func SendPostWithAuth(baseURL, payload string) ([]byte, models.BError) {

	stMgr := state.GetAppState(nil)
	accessToken := stMgr.TokenManager().GetAuth().Token
	// Build URL with ID in path
	fullURL := fmt.Sprintf("./%s", baseURL)
	_err := models.BError{Request: fullURL}
	var byteOut []byte
	resp, statusCd, err := sendBirdRequest(payload, accessToken, fullURL)
	if err != nil {
		_err.BError = err
		return nil, _err
	}
	_err.Code = statusCd
	fmt.Printf("Status: %d\n", statusCd)
	byteOut = resp
	if statusCd == 401 {
		accessToken = stMgr.ForceRefresh()
		resp, statusCd, err := sendBirdRequest(payload, accessToken, fullURL)
		if err != nil {
			_err.BError = err
			return nil, _err
		}
		fmt.Printf("ReTry Status: %d\n", statusCd)
		byteOut = resp
	}
	return byteOut, _err
}

func sendBirdRequest(payload, accessToken, fullURL string) ([]byte, int, error) {
	// Create context with timeout (important for WASM)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, strings.NewReader(payload))
	if err != nil {
		return nil, -1, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// CORS headers might be needed depending on your setup
	req.Header.Set("Access-Control-Allow-Origin", "*")
	// Use default client (works in WASM)
	resp, erp := http.DefaultClient.Do(req)
	if erp != nil {
		return nil, -1, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to read response: %w", err)
	}

	return body, resp.StatusCode, nil
}

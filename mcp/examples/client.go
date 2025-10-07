// Example stateful MCP client
// Run with: go run examples/client.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Simple stateful MCP client for testing
type StatefulClient struct {
	baseURL   string
	sessionID string
	client    *http.Client
}

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewStatefulClient(baseURL string) *StatefulClient {
	return &StatefulClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *StatefulClient) sendRequest(req JSONRPCRequest) (*JSONRPCResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Add session ID if we have one
	if c.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Capture session ID from response headers
	if sessionID := resp.Header.Get("Mcp-Session-Id"); sessionID != "" {
		c.sessionID = sessionID
		fmt.Printf("📌 Session ID: %s\n", sessionID)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal(respBody, &jsonResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w (body: %s)", err, string(respBody))
	}

	if jsonResp.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error %d: %s", jsonResp.Error.Code, jsonResp.Error.Message)
	}

	return &jsonResp, nil
}

func (c *StatefulClient) Initialize() error {
	fmt.Println("\n🔧 Initializing MCP session...")

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]string{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	fmt.Println("✅ Session initialized")
	fmt.Printf("Response: %s\n", string(resp.Result))
	return nil
}

func (c *StatefulClient) ListTools() error {
	fmt.Println("\n📋 Listing available tools...")

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, resp.Result, "", "  ")
	fmt.Println(prettyJSON.String())
	return nil
}

func (c *StatefulClient) CallWeatherTool(days int) error {
	fmt.Printf("\n🌤️  Calling weather forecast tool (days: %d)...\n", days)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "get_weather_forecast",
			"arguments": map[string]interface{}{
				"days": days,
			},
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, resp.Result, "", "  ")
	fmt.Println(prettyJSON.String())
	return nil
}

func (c *StatefulClient) CallTeeTimesTool(date string) error {
	fmt.Printf("\n⛳ Calling tee times tool (date: %s)...\n", date)

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "get_available_tee_times",
			"arguments": map[string]interface{}{
				"date": date,
			},
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return err
	}

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, resp.Result, "", "  ")
	fmt.Println(prettyJSON.String())
	return nil
}

func main() {
	serverURL := os.Getenv("MCP_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8081/mcp"
	}

	fmt.Println("=== MCP Stateful Client Test ===")
	fmt.Printf("Server: %s\n", serverURL)

	client := NewStatefulClient(serverURL)

	// Step 1: Initialize session
	if err := client.Initialize(); err != nil {
		log.Fatalf("Initialize failed: %v", err)
	}

	// Give server a moment
	time.Sleep(100 * time.Millisecond)

	// Step 2: List tools
	if err := client.ListTools(); err != nil {
		log.Fatalf("List tools failed: %v", err)
	}

	// Step 3: Call weather tool
	if err := client.CallWeatherTool(3); err != nil {
		log.Fatalf("Weather tool failed: %v", err)
	}

	// Step 4: Call tee times tool
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	if err := client.CallTeeTimesTool(tomorrow); err != nil {
		log.Fatalf("Tee times tool failed: %v", err)
	}

	// Step 5: Call weather again (reusing session)
	fmt.Println("\n🔄 Reusing session for another weather call...")
	if err := client.CallWeatherTool(7); err != nil {
		log.Fatalf("Second weather tool failed: %v", err)
	}

	fmt.Println("\n✅ All tests completed successfully!")
}

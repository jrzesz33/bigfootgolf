package anthropic

import (
	"bigfoot/golf/common/models/auth"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// MCPClient handles communication with the MCP server
type MCPClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	UseProxy   bool
}

// MCPToolCall represents a tool call to the MCP server
type MCPToolCall struct {
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
}

// MCPResponse represents the response from the MCP server
type MCPResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
	Error   string      `json:"error,omitempty"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient(userID, userEmail string) (*MCPClient, error) {
	// Generate JWT token for the user
	token, err := auth.GenerateToken(userID, userEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}
	
	// Check if we should use proxy for development
	useProxy := os.Getenv("MCP_USE_PROXY") == "true"
	baseURL := os.Getenv("MCP_SERVER_URL")
	if baseURL == "" {
		if useProxy {
			baseURL = "http://localhost:8082" // Proxy server
		} else {
			baseURL = "http://localhost:8081" // Direct MCP server
		}
	}
	
	return &MCPClient{
		BaseURL:  baseURL,
		Token:    token,
		UseProxy: useProxy,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// CallTool calls a specific tool on the MCP server
func (c *MCPClient) CallTool(toolName string, params map[string]interface{}) (*MCPResponse, error) {
	// Prepare the request
	toolCall := MCPToolCall{
		Tool:   toolName,
		Params: params,
	}
	
	jsonData, err := json.Marshal(toolCall)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	
	// Create HTTP request
	endpoint := c.BaseURL + "/mcp"
	if c.UseProxy {
		endpoint = c.BaseURL + "/proxy/mcp"
	}
	
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	
	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP server returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	return &mcpResp, nil
}

// GetReservations gets user reservations via MCP
func (c *MCPClient) GetReservations(userID string) (interface{}, error) {
	resp, err := c.CallTool("manage_reservations", map[string]interface{}{
		"action":  "get",
		"user_id": userID,
	})
	if err != nil {
		return nil, err
	}
	
	if !resp.Success {
		return nil, fmt.Errorf("MCP error: %s", resp.Error)
	}
	
	return resp.Result, nil
}

// BookReservation books a new reservation via MCP
func (c *MCPClient) BookReservation(userID, teeTime string, players int) (interface{}, error) {
	resp, err := c.CallTool("manage_reservations", map[string]interface{}{
		"action":   "book",
		"user_id":  userID,
		"tee_time": teeTime,
		"players":  players,
	})
	if err != nil {
		return nil, err
	}
	
	if !resp.Success {
		return nil, fmt.Errorf("MCP error: %s", resp.Error)
	}
	
	return resp.Result, nil
}

// CancelReservation cancels a reservation via MCP
func (c *MCPClient) CancelReservation(userID, reservationID string) error {
	resp, err := c.CallTool("manage_reservations", map[string]interface{}{
		"action":         "cancel",
		"user_id":        userID,
		"reservation_id": reservationID,
	})
	if err != nil {
		return err
	}
	
	if !resp.Success {
		return fmt.Errorf("MCP error: %s", resp.Error)
	}
	
	return nil
}

// FindTeeTimes finds available tee times via MCP
func (c *MCPClient) FindTeeTimes(date, timeRange string, players int) (interface{}, error) {
	params := map[string]interface{}{
		"date": date,
	}
	
	if timeRange != "" {
		params["time_range"] = timeRange
	}
	if players > 0 {
		params["players"] = players
	}
	
	resp, err := c.CallTool("find_tee_times", params)
	if err != nil {
		return nil, err
	}
	
	if !resp.Success {
		return nil, fmt.Errorf("MCP error: %s", resp.Error)
	}
	
	return resp.Result, nil
}

// GetWeatherConditions gets weather and course conditions via MCP
func (c *MCPClient) GetWeatherConditions(date string) (interface{}, error) {
	params := map[string]interface{}{}
	if date != "" {
		params["date"] = date
	}
	
	resp, err := c.CallTool("get_conditions", params)
	if err != nil {
		return nil, err
	}
	
	if !resp.Success {
		return nil, fmt.Errorf("MCP error: %s", resp.Error)
	}
	
	return resp.Result, nil
}

// IntegrateWithClaude adds MCP capabilities to Claude conversations
func IntegrateMCPWithClaude(client *ClaudeClient, mcpClient *MCPClient, enableMCP bool) {
	// This function would be called to enhance Claude with MCP tools
	// The actual integration would happen in the chat handler
	if enableMCP && mcpClient != nil {
		// Register MCP tools with Claude
		// This would be done through the tools parameter in Claude requests
	}
}
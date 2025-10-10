// HTTP client utilities for testing MCP server
// This is a library file - import it from other examples
package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TestClient provides utilities for testing the MCP server endpoints
type TestClient struct {
	BaseURL string
	Client  *http.Client
}

// NewTestClient creates a new test client
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// HealthCheck checks if the server is healthy
func (c *TestClient) HealthCheck() error {
	resp, err := c.Client.Get(c.BaseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetServerInfo retrieves server information
func (c *TestClient) GetServerInfo() (map[string]interface{}, error) {
	resp, err := c.Client.Get(c.BaseURL + "/info")
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server info returned %d: %s", resp.StatusCode, string(body))
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode server info: %w", err)
	}

	return info, nil
}

// MCPRequest represents an MCP tool call request
type MCPRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CallTool calls an MCP tool endpoint
func (c *TestClient) CallTool(toolName string, args map[string]interface{}) (interface{}, error) {
	request := MCPRequest{
		Method: "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.Client.Post(
		c.BaseURL+"/mcp",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if mcpResp.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", mcpResp.Error.Code, mcpResp.Error.Message)
	}

	return mcpResp.Result, nil
}

// TestGetWeatherForecast tests the get_weather_forecast tool
func (c *TestClient) TestGetWeatherForecast(days int) error {
	args := map[string]interface{}{}
	if days > 0 {
		args["days"] = days
	}

	result, err := c.CallTool("get_weather_forecast", args)
	if err != nil {
		return err
	}

	fmt.Printf("Weather Forecast Result:\n%+v\n", result)
	return nil
}

// TestGetAvailableTeeTimes tests the get_available_tee_times tool
func (c *TestClient) TestGetAvailableTeeTimes(date string) error {
	args := map[string]interface{}{
		"date": date,
	}

	result, err := c.CallTool("get_available_tee_times", args)
	if err != nil {
		return err
	}

	fmt.Printf("Tee Times Result:\n%+v\n", result)
	return nil
}

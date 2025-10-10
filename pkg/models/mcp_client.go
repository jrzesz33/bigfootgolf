package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// Simple stateful MCP client for testing
type StatefulClient struct {
	baseURL     string
	sessionID   string
	client      *http.Client
	transaction int
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
	sc := &StatefulClient{
		baseURL:     baseURL,
		transaction: 1,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	sc.Initialize()
	return sc
}
func (c *StatefulClient) CallMCPServer(method string, input map[string]interface{}) (string, error) {

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.transaction,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      method,
			"arguments": input,
		},
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return "", err
	}
	c.transaction++

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, resp.Result, "", "  ")
	fmt.Println(prettyJSON.String())
	return prettyJSON.String(), nil
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
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
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

	// Strip SSE "data: " prefix if present
	bodyStr := string(respBody)
	if len(bodyStr) >= 6 && bodyStr[:6] == "data: " {
		bodyStr = bodyStr[6:]
	}

	var jsonResp JSONRPCResponse
	if err := json.Unmarshal([]byte(bodyStr), &jsonResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w (body: %s)", err, bodyStr)
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

func (c *StatefulClient) ListTools() ([]BigfootTool, error) {
	fmt.Println("\n📋 Listing available tools...")

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}

	//var prettyJSON bytes.Buffer
	//json.Indent(&prettyJSON, resp.Result, "", "  ")
	//fmt.Println(prettyJSON.String())

	type loadTools struct {
		Tools []mcp.Tool `json:"tools"`
	}

	var mcpTools loadTools
	err = json.Unmarshal(resp.Result, &mcpTools)
	if err != nil {
		return nil, err
	}

	return convertMCPtoOpenAICompletion(mcpTools.Tools), nil
}
func convertMCPtoOpenAICompletion(mcpTools []mcp.Tool) []BigfootTool {
	var toolsOut []BigfootTool
	for _, tool := range mcpTools {
		toolsOut = append(toolsOut, BigfootTool{
			Type: "function",
			Function: BigfootFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: BigfootParameter{
					Type:       tool.InputSchema.Type,
					Properties: tool.InputSchema.Properties,
				},
			},
		})
	}
	return toolsOut
}

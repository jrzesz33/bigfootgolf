package anthropic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"time"
)

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Tool represents a function/tool definition
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

// ClaudeRequest represents the request payload to Claude API
type ClaudeRequest struct {
	Model         string    `json:"model"`
	MaxTokens     int       `json:"max_tokens"`
	Temperature   float64   `json:"temperature,omitempty"`
	Messages      []Message `json:"messages"`
	Tools         []Tool    `json:"tools,omitempty"`
	ToolChoice    interface{} `json:"tool_choice,omitempty"`
	System        string    `json:"system,omitempty"`
}

// ClaudeResponse represents the response from Claude API
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
		Name string `json:"name,omitempty"`
		ID   string `json:"id,omitempty"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ChatRequest represents the incoming request to our API
type ChatRequest struct {
	Message          string    `json:"message"`
	ConversationID   string    `json:"conversation_id,omitempty"`
	ConversationHist []Message `json:"conversation_history,omitempty"`
	SystemMessage    string    `json:"system_message,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	EnableFunctions  bool      `json:"enable_functions,omitempty"`
}

// ChatResponse represents our API response
type ChatResponse struct {
	Response         string    `json:"response"`
	ConversationID   string    `json:"conversation_id"`
	ConversationHist []Message `json:"conversation_history"`
	HasFunctionCall  bool      `json:"has_function_call"`
	FunctionCalls    []string  `json:"function_calls,omitempty"`
	Usage            struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// ClaudeClient handles communication with Claude API
type ClaudeClient struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient(apiKey string) *ClaudeClient {
	return &ClaudeClient{
		APIKey:  apiKey,
		BaseURL: "https://api.anthropic.com/v1",
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GetAvailableTools returns the tools/functions available to Claude
func GetAvailableTools() []Tool {
	return []Tool{
		{
			Name:        "calculate",
			Description: "Perform mathematical calculations",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]interface{}{
						"type":        "string",
						"description": "The mathematical expression to evaluate",
					},
				},
				"required": []string{"expression"},
			},
		},
	}
}

// SendMessage sends a message to Claude API
func (c *ClaudeClient) SendMessage(req ClaudeRequest) (*ClaudeResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Debug: Print request JSON
	fmt.Printf("Claude API Request: %s\n", string(jsonData))

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &claudeResp, nil
}

// ChatHandler handles the main chat endpoint
func ChatHandler(claudeClient *ClaudeClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var chatReq ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Set defaults
		if chatReq.MaxTokens == 0 {
			chatReq.MaxTokens = 4096
		}
		if chatReq.Temperature == 0 {
			chatReq.Temperature = 0.7
		}

		// Build conversation history
		messages := chatReq.ConversationHist
		messages = append(messages, Message{
			Role:    "user",
			Content: chatReq.Message,
		})

		// Prepare Claude request
		claudeReq := ClaudeRequest{
			Model:       "claude-3-5-sonnet-20241022", // Using latest available model
			MaxTokens:   chatReq.MaxTokens,
			Temperature: chatReq.Temperature,
			Messages:    messages,
			System:      chatReq.SystemMessage,
		}

		// Add tools if functions are enabled
		var hasFunctionCall bool
		var functionCalls []string

		if chatReq.EnableFunctions {
			claudeReq.Tools = GetAvailableTools()
			claudeReq.ToolChoice = map[string]string{"type": "auto"}
		}

		// Send request to Claude
		claudeResp, err := claudeClient.SendMessage(claudeReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Claude API error: %v", err), http.StatusInternalServerError)
			return
		}

		// Process response and check for function calls
		var responseText string
		for _, content := range claudeResp.Content {
			if content.Type == "text" {
				responseText += content.Text
			} else if content.Type == "tool_use" {
				hasFunctionCall = true
				if content.Name != "" {
					functionCalls = append(functionCalls, content.Name)
				}
			}
		}

		// Update conversation history
		updatedHistory := append(messages, Message{
			Role:    "assistant",
			Content: responseText,
		})

		// Generate conversation ID if not provided
		conversationID := chatReq.ConversationID
		if conversationID == "" {
			conversationID = fmt.Sprintf("conv_%d", time.Now().Unix())
		}

		// Prepare response
		response := ChatResponse{
			Response:         responseText,
			ConversationID:   conversationID,
			ConversationHist: updatedHistory,
			HasFunctionCall:  hasFunctionCall,
			FunctionCalls:    functionCalls,
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{
				InputTokens:  claudeResp.Usage.InputTokens,
				OutputTokens: claudeResp.Usage.OutputTokens,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// HealthHandler provides a health check endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (r *ChatRequest) AddNewMessage(message string) {
	// Build conversation history
	r.ConversationHist = append(r.ConversationHist, Message{
		Role:    "user",
		Content: message,
	})
	r.Message = message
	// Don't override SystemMessage - let it be set by the controller
}

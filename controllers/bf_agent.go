package controllers

import (
	"birdsfoot/app/models/anthropic"
	"fmt"
	"os"
	"time"
)

type AgentController struct {
	ChatHistory string
	Client      anthropic.ClaudeClient
	Request     anthropic.ChatRequest
}

func NewAgentController() AgentController {
	var _agent AgentController

	// Get API key from environment variable
	apiKey := os.Getenv("ANTHROPIC_API_KEY")

	if apiKey == "" {
		fmt.Println("ANTHROPIC_API_KEY environment variable is required")
		os.Exit(1)
	}

	// Create Claude client
	_agent.Client = *anthropic.NewClaudeClient(apiKey)
	return _agent
}

func (a *AgentController) HandleChat(message anthropic.ChatRequest) (*anthropic.ChatResponse, error) {

	a.Request = message

	// Set defaults
	if a.Request.MaxTokens == 0 {
		a.Request.MaxTokens = 4096
	}
	if a.Request.Temperature == 0 {
		a.Request.Temperature = 0.7
	}
	a.Request.EnableFunctions = true

	a.Request.SystemMessage = fmt.Sprintf(anthropic.SystemMessage, 1, "")

	// Ensure we have at least one message
	if len(a.Request.ConversationHist) == 0 {
		return nil, fmt.Errorf("conversation history is empty")
	}

	// Validate messages have required fields
	for i, msg := range a.Request.ConversationHist {
		if msg.Role == "" || msg.Content == "" {
			return nil, fmt.Errorf("message %d missing role or content", i)
		}
		if msg.Role != "user" && msg.Role != "assistant" {
			return nil, fmt.Errorf("message %d has invalid role: %s", i, msg.Role)
		}
	}

	// Prepare Claude request
	claudeReq := anthropic.ClaudeRequest{
		Model:       "claude-3-5-sonnet-20241022", // Using valid available model
		MaxTokens:   a.Request.MaxTokens,
		Temperature: a.Request.Temperature,
		Messages:    a.Request.ConversationHist,
		System:      a.Request.SystemMessage,
	}

	// Add tools if functions are enabled
	var hasFunctionCall bool
	var functionCalls []string

	if a.Request.EnableFunctions {
		claudeReq.Tools = anthropic.GetAvailableTools()
		claudeReq.ToolChoice = map[string]string{"type": "auto"}
	}

	// Send request to Claude
	claudeResp, err := a.Client.SendMessage(claudeReq)
	if err != nil {
		fmt.Printf("Claude API Error: %v\n", err)
		fmt.Printf("Request details - Model: %s, MaxTokens: %d, Messages: %d\n", 
			claudeReq.Model, claudeReq.MaxTokens, len(claudeReq.Messages))
		return nil, err
	}

	// Process response and check for function calls
	var responseText string
	for _, content := range claudeResp.Content {
		switch content.Type {
		case "text":
			responseText += content.Text
		case "tool_use":
			hasFunctionCall = true
			if content.Name != "" {
				functionCalls = append(functionCalls, content.Name)
			}
		}
	}

	// Update conversation history
	updatedHistory := append(a.Request.ConversationHist, anthropic.Message{
		Role:    "assistant",
		Content: responseText,
	})

	// Generate conversation ID if not provided
	conversationID := a.Request.ConversationID
	if conversationID == "" {
		conversationID = fmt.Sprintf("conv_%d", time.Now().Unix())
	}

	// Prepare response
	response := anthropic.ChatResponse{
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
	return &response, nil
}

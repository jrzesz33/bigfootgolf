package controllers

import (
	"bigfoot/golf/common/models/anthropic"
	"bigfoot/golf/common/models/teetimes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type AgentController struct {
	ChatHistory  string
	Client       anthropic.ClaudeClient
	Request      anthropic.ChatRequest
	ToolExecutor *anthropic.ToolExecutor
	UserID       string
	UserEmail    string
	MCPClient    *anthropic.MCPClient
	UseMCP       bool
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

// SetUserID sets the user ID for the agent controller
func (a *AgentController) SetUserID(userID string) {
	a.UserID = userID
	a.ToolExecutor = anthropic.NewToolExecutor(userID)
}

// SetUserInfo sets the user information and optionally enables MCP
func (a *AgentController) SetUserInfo(userID, userEmail string, enableMCP bool) error {
	a.UserID = userID
	a.UserEmail = userEmail
	a.UseMCP = enableMCP
	a.ToolExecutor = anthropic.NewToolExecutor(userID)

	// Initialize MCP client if enabled
	if enableMCP {
		mcpClient, err := anthropic.NewMCPClient(userID, userEmail)
		if err != nil {
			return fmt.Errorf("failed to create MCP client: %v", err)
		}
		a.MCPClient = mcpClient

		// Enhance tool executor with MCP capabilities
		a.ToolExecutor.SetMCPClient(mcpClient)
	}

	return nil
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

	// Initialize tool executor if not set
	if a.ToolExecutor == nil {
		a.ToolExecutor = anthropic.NewToolExecutor(a.UserID)
	}

	// Get user's current reservations for context
	userReservations, err := teetimes.GetUserReservations(a.UserID, false)
	if err != nil {
		fmt.Printf("Warning: Could not get user reservations for context: %v\n", err)
		userReservations = []teetimes.Reservation{}
	}

	// Format reservations for system message
	reservationsText := ""
	if len(userReservations) == 0 {
		reservationsText = "No current reservations."
	} else {
		for _, res := range userReservations {
			reservationsText += fmt.Sprintf("- %s at %s | ID: %s | %d players | $%.2f | %s\n",
				res.TeeTime.Format("January 2, 2006"),
				res.TeeTime.Format("3:04 PM"),
				res.ID,
				len(res.Players)+1,
				res.Price,
				res.Group,
			)
		}
	}

	// Add tee time context and user reservations to system message
	teeTimeContext := anthropic.GetTeeTimeContext()
	systemMessage := fmt.Sprintf(anthropic.SystemMessage, a.UserID, reservationsText) +
		"\n\nCurrent Available Tee Times:\n" + teeTimeContext

	a.Request.SystemMessage = systemMessage

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

	// Process response and handle tool calls
	var responseText string
	var toolResults []map[string]interface{}

	for _, content := range claudeResp.Content {
		switch content.Type {
		case "text":
			responseText += content.Text
		case "tool_use":
			hasFunctionCall = true
			if content.Name != "" {
				functionCalls = append(functionCalls, content.Name)

				// Execute the tool
				toolInput := make(map[string]interface{})
				if inputData, ok := content.Input.(map[string]interface{}); ok {
					toolInput = inputData
				}

				toolResult, err := a.ToolExecutor.ExecuteTool(content.Name, toolInput)
				if err != nil {
					toolResult = fmt.Sprintf("Error executing tool %s: %v", content.Name, err)
				}

				// Add tool result for follow-up if needed
				toolResults = append(toolResults, map[string]interface{}{
					"tool_use_id": content.ID,
					"type":        "tool_result",
					"content":     toolResult,
				})

				// Include tool result in response
				responseText += fmt.Sprintf("\n\n[Tool Result: %s]\n%s", content.Name, toolResult)
			}
		}
	}

	// If there were tool calls, we might need a follow-up call to Claude
	if hasFunctionCall && len(toolResults) > 0 {
		// Add the assistant's message with tool calls
		followUpMessages := append(a.Request.ConversationHist, anthropic.Message{
			Role:    "assistant",
			Content: responseText,
		})

		// Add tool results as user messages
		for _, toolResult := range toolResults {
			resultJSON, _ := json.Marshal(toolResult)
			followUpMessages = append(followUpMessages, anthropic.Message{
				Role:    "user",
				Content: string(resultJSON),
			})
		}

		// Make follow-up call to Claude with tool results
		followUpReq := anthropic.ClaudeRequest{
			Model:       claudeReq.Model,
			MaxTokens:   claudeReq.MaxTokens,
			Temperature: claudeReq.Temperature,
			Messages:    followUpMessages,
			System:      claudeReq.System,
			Tools:       claudeReq.Tools,
			ToolChoice:  claudeReq.ToolChoice,
		}

		followUpResp, err := a.Client.SendMessage(followUpReq)
		if err == nil && len(followUpResp.Content) > 0 {
			// Use the follow-up response
			claudeResp = followUpResp
			responseText = ""
			for _, content := range followUpResp.Content {
				if content.Type == "text" {
					responseText += content.Text
				}
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

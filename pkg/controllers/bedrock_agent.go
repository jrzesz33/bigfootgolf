package controllers

import (
	"bigfoot/golf/common/models/anthropic"
	"bigfoot/golf/common/models/teetimes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type BigfootAgentController struct {
	ChatHistory   string
	Client        anthropic.LLMClient
	Request       anthropic.ChatRequest
	ToolExecutor  *anthropic.ToolExecutor
	UserID        string
	UserEmail     string
	MCPClient     *anthropic.MCPClient
	UseMCP        bool
	Region        string
	AgentPlatform string
	ModelID       string
	ModelVersion  string
}

func NewBigfootAgentController() *BigfootAgentController {
	var _agent BigfootAgentController

	// Get AWS region from environment variable (optional, defaults to us-east-1)
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	_agent.Region = region
	_agent.AgentPlatform = os.Getenv("AGENT_PLATFORM")
	_agent.ModelID = os.Getenv("LLM_MODEL")
	_agent.ModelVersion = os.Getenv("LLM_VERSION")

	if _agent.AgentPlatform == "" || _agent.ModelID == "" {
		return nil
	}

	switch _agent.AgentPlatform {
	case "bedrock":
		// Create Bedrock client using AWS SDK
		client, err := anthropic.NewBedrockClient(region)
		if err != nil {
			fmt.Printf("Failed to create Bedrock client: %v\n", err)
			fmt.Println("Make sure AWS credentials are configured (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)")
			os.Exit(1)
		}
		_agent.Client = client
	default:
		// Get API key from environment variable
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			fmt.Println("ANTHROPIC_API_KEY environment variable is required")
			return nil
		}
		client := anthropic.NewClaudeClient(apiKey)
		_agent.Client = client
	}

	return &_agent
}

// SetUserID sets the user ID for the agent controller
func (a *BigfootAgentController) SetUserID(userID string) {
	a.UserID = userID
	a.ToolExecutor = anthropic.NewToolExecutor(userID)
}

// SetUserInfo sets the user information and optionally enables MCP
func (a *BigfootAgentController) SetUserInfo(userID, userEmail string, enableMCP bool) error {
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

func (a *BigfootAgentController) HandleChat(message anthropic.ChatRequest) (*anthropic.ChatResponse, error) {
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

	// Prepare Bedrock request
	bedrockReq := anthropic.BedrockRequest{
		Model:       "anthropic.claude-sonnet-4-5-20250929-v1:0",
		MaxTokens:   a.Request.MaxTokens,
		Temperature: a.Request.Temperature,
		Messages:    a.Request.ConversationHist,
		System:      a.Request.SystemMessage,
	}

	// Add tools if functions are enabled
	var hasFunctionCall bool
	var functionCalls []string

	if a.Request.EnableFunctions {
		bedrockReq.Tools = anthropic.GetAvailableTools()
		bedrockReq.ToolChoice = map[string]string{"type": "auto"}
	}

	// Send request to Bedrock
	bedrockResp, err := a.Client.SendMessage(bedrockReq)
	if err != nil {
		fmt.Printf("Bedrock API Error: %v\n", err)
		fmt.Printf("Request details - Model: %s, MaxTokens: %d, Messages: %d\n",
			bedrockReq.Model, bedrockReq.MaxTokens, len(bedrockReq.Messages))
		return nil, err
	}

	// Process response and handle tool calls
	var responseText string
	var toolResults []map[string]interface{}

	for _, content := range bedrockResp.Content {
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

	// If there were tool calls, we might need a follow-up call to Bedrock
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

		// Make follow-up call to Bedrock with tool results
		followUpReq := anthropic.ClaudeRequest{
			Model:       bedrockReq.Model,
			MaxTokens:   bedrockReq.MaxTokens,
			Temperature: bedrockReq.Temperature,
			Messages:    followUpMessages,
			System:      bedrockReq.System,
			Tools:       bedrockReq.Tools,
			ToolChoice:  bedrockReq.ToolChoice,
		}

		followUpResp, err := a.Client.SendMessage(followUpReq)
		if err == nil && len(followUpResp.Content) > 0 {
			// Use the follow-up response
			bedrockResp = followUpResp
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
			InputTokens:  bedrockResp.Usage.InputTokens,
			OutputTokens: bedrockResp.Usage.OutputTokens,
		},
	}
	return &response, nil
}

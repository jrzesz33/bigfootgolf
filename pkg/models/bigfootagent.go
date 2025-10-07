package models

import (
	"bigfoot/golf/common/models/anthropic"
	"bigfoot/golf/common/models/teetimes"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type GolfAgent struct {
	GatewayURL string
	Model      string
	ClaudeReq  AgentRequest
	//Tools        []anthropic.Tool
	ToolChoice   map[string]string
	ToolExecutor *anthropic.ToolExecutor
}

// ChatRequest represents the incoming request to our API
type AgentChatRequest struct {
	Message          string    `json:"message"`
	ConversationID   string    `json:"conversation_id,omitempty"`
	ConversationHist []Message `json:"conversation_history,omitempty"`
	SystemMessage    string    `json:"system_message,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	EnableFunctions  bool      `json:"enable_functions,omitempty"`
}

func (r *AgentChatRequest) AddNewMessage(message string) {
	// Build conversation history
	r.ConversationHist = append(r.ConversationHist, Message{
		Role:    "user",
		Content: message,
	})
	r.Message = message
	// Don't override SystemMessage - let it be set by the controller
}

// ChatResponse represents our API response
type AgentChatResponse struct {
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

const DEVGATEWAY_URL = "http://localhost:3000"

func NewGolfAgent() GolfAgent {
	var ga GolfAgent
	ga.GatewayURL = os.Getenv("GATEWAY_URL")
	if ga.GatewayURL == "" {
		ga.GatewayURL = DEVGATEWAY_URL
	}
	ga.Model = os.Getenv("LLM_MODEL")
	// Build the request body
	ga.ClaudeReq = AgentRequest{
		MaxTokens:        4096,
		Temperature:      0.7,
		AnthropicVersion: "bedrock-2023-05-31",
		ToolChoice:       "auto",
		Tools:            GetBigfootTools(false), //anthropic.GetAvailableTools(),
	}
	return ga
}
func (g *GolfAgent) SendWithMessage(userId string, prompt AgentChatRequest) (*AgentChatResponse, error) {

	// Prepare response
	if prompt.ConversationID == "" {
		prompt.ConversationID = fmt.Sprintf("conv_%d", time.Now().Unix())
	}
	response := AgentChatResponse{
		ConversationID:  prompt.ConversationID,
		HasFunctionCall: false,
	}

	//loop thru llm requests for tool processing
	for {
		//log the Prompts
		fmt.Println("**INVOKING LLM WITH**")
		for _, msg := range prompt.ConversationHist {
			fmt.Println("PROMPT: ", msg)
		}
		resp, err := g.InvokeWithMessage(userId, prompt.ConversationHist)
		if err != nil {
			return nil, err
		}
		// Process response and handle tool calls
		var toolResults []map[string]interface{}
		var responseText string
		var llmResponse string
		//review response
		for _, content := range resp.Choices {
			switch content.FinishReason {
			case "stop":
				responseText += content.Message.Content
				response.HasFunctionCall = false
			case "tool_calls":
				//initialize the Tool Executor
				g.ToolExecutor = anthropic.NewToolExecutor(userId)
				response.HasFunctionCall = true
				for _, toolCalls := range content.Message.ToolCalls {
					//llmToolResponse +=
					//add the function call
					llmResponse += fmt.Sprintf("\n[Tool Call: %s]%s(%s)", toolCalls.Type, toolCalls.Function.Name, toolCalls.Function.Arguments)
					// Parse the arguments string into a map
					var args map[string]interface{}
					if err := json.Unmarshal([]byte(toolCalls.Function.Arguments), &args); err != nil {
						toolResult := fmt.Sprintf("Error parsing arguments for tool %s: %v", toolCalls.Function.Name, err)
						toolResults = append(toolResults, map[string]interface{}{
							"tool_use_id": toolCalls.ID,
							"type":        "tool_result",
							"content":     toolResult,
						})
						//responseText += fmt.Sprintf("\n\n[Tool Result: %s]\n%s", toolCalls.Function.Name, toolResult)
						continue
					}

					toolResult, err := g.ToolExecutor.ExecuteTool(toolCalls.Function.Name, args)
					if err != nil {
						toolResult = fmt.Sprintf("Error executing tool %s: %v", toolCalls.Function.Name, err)
					}
					// Add tool result for follow-up if needed
					toolResults = append(toolResults, map[string]interface{}{
						"tool_use_id": toolCalls.ID,
						"type":        "tool_result",
						"content":     toolResult,
					})

					// Include tool result in response
					//responseText += fmt.Sprintf("\n\n[Tool Result: %s]\n%s", toolCalls.Function.Name, toolResult)
				}

			default:
				fmt.Println("UNKNOWN LLM RESPONSE ACTION.. ", resp)
			}

		}

		// If there were tool calls, we might need a follow-up call to Bedrock
		if !response.HasFunctionCall {

			prompt.ConversationHist = append(prompt.ConversationHist, Message{
				Role:    "assistant",
				Content: responseText,
			})
			response.ConversationHist = prompt.ConversationHist
			response.Response = responseText
			response.Usage.InputTokens = resp.Usage.PromptTokens
			response.Usage.OutputTokens = resp.Usage.TotalTokens
			return &response, nil
		} else {
			//add the llm response to the history
			prompt.ConversationHist = append(prompt.ConversationHist, Message{
				Role:    "assistant",
				Content: llmResponse,
			})
			// Add tool results as user messages
			for _, tr := range toolResults {
				resultJSON, _ := json.Marshal(tr)
				prompt.ConversationHist = append(prompt.ConversationHist, Message{
					Role:    "user",
					Content: string(resultJSON),
				})
			}
		}
	}

}
func (g *GolfAgent) InvokeWithMessage(userId string, messages []Message) (*GlobalAIResponse, error) {

	//add the users Message to the Request
	g.ClaudeReq.Messages = messages

	//add systemMessage
	g.ClaudeReq.System = getUserSystemMessage(userId)

	resp, err := g.postJSONRequest()
	if err != nil {
		return nil, err
	}

	// Parse the response
	var claudeResp GlobalAIResponse
	if err := json.Unmarshal(resp, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &claudeResp, nil

}

func getUserSystemMessage(userId string) string {

	// Get user's current reservations for context
	userReservations, err := teetimes.GetUserReservations(userId, false)
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
	systemMessage := fmt.Sprintf(anthropic.SystemMessage, userId, reservationsText) +
		"\n\nCurrent Available Tee Times:\n" + teeTimeContext
	return systemMessage
}

// PostJSONRequest sends an HTTP POST request with a JSON body and specified headers.
func (g *GolfAgent) postJSONRequest() ([]byte, error) {
	// Marshal request body
	jsonBody, err := json.Marshal(g.ClaudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	//fmt.Println(string(jsonBody))
	// Create a new HTTP POST request
	req, err := http.NewRequest(http.MethodPost, g.GatewayURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set Content-Type and Accept headers
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Accept", "application/json")

	// Create an HTTP client with a timeout
	client := &http.Client{Timeout: 30 * time.Second}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close() // Ensure the response body is closed

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

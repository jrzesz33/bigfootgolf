package models

type BigfootTool struct {
	Type     string          `json:"type,omitempty"`
	Function BigfootFunction `json:"function,omitempty"`
}
type BigfootFunction struct {
	Name        string           `json:"name,omitempty"`
	Description string           `json:"description,omitempty"`
	Parameters  BigfootParameter `json:"parameters,omitempty"`
}
type BigfootParameter struct {
	Type       string                 `json:"type,omitempty"`
	Required   []string               `json:"required,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// ClaudeRequest represents the request payload to Claude API
type AgentRequest struct {
	Model            string        `json:"model,omitempty"`
	MaxTokens        int           `json:"max_tokens"`
	Temperature      float64       `json:"temperature,omitempty"`
	Messages         []Message     `json:"messages"`
	Tools            []BigfootTool `json:"tools,omitempty"`
	ToolChoice       string        `json:"tool_choice,omitempty"`
	System           string        `json:"system,omitempty"`
	AnthropicVersion string        `json:"anthropic_version,omitempty"`
}

type GlobalAIResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Object  string   `json:"object"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Message struct {
	ToolCalls []ToolCall `json:"tool_calls"`
	Role      string     `json:"role"`
	Content   string     `json:"content"`
}

type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GetBigfootTools returns the tools/functions available to Claude
func GetBigfootTools(useMCP bool) []BigfootTool {
	if !useMCP {
		return []BigfootTool{
			{
				Type: "function",
				Function: BigfootFunction{
					Name:        "get_weather_forecast",
					Description: "Get weather forecast for golf course area",
					Parameters: BigfootParameter{
						Type: "object",
						Properties: map[string]interface{}{
							"days": map[string]interface{}{
								"type":        "integer",
								"description": "Number of days to forecast (1-7)",
								"minimum":     1,
								"maximum":     7,
								"default":     3,
							},
						},
					},
				},
			},
			{
				Type: "function",
				Function: BigfootFunction{
					Name:        "get_available_tee_times",
					Description: "Get available tee times for a specific date",
					Parameters: BigfootParameter{
						Type:     "object",
						Required: []string{"date"},
						Properties: map[string]interface{}{
							"date": map[string]interface{}{
								"type":        "string",
								"description": "Date in YYYY-MM-DD format (e.g., 2024-01-15)",
							},
						},
					},
				},
			},
		}
	} else {
		//CLAUDE CODE CAN YOU FINISH THIS LOGIC TO CONNECT TO MY LOCAL MCP Server
		return []BigfootTool{
			{
				Type: "function",
				Function: BigfootFunction{
					Name:        "get_weather_forecast",
					Description: "Get weather forecast for golf course area",
					Parameters: BigfootParameter{
						Type: "object",
						Properties: map[string]interface{}{
							"days": map[string]interface{}{
								"type":        "integer",
								"description": "Number of days to forecast (1-7)",
								"minimum":     1,
								"maximum":     7,
								"default":     3,
							},
						},
					},
				},
			},
			{
				Type: "function",
				Function: BigfootFunction{
					Name:        "get_available_tee_times",
					Description: "Get available tee times for a specific date",
					Parameters: BigfootParameter{
						Type:     "object",
						Required: []string{"date"},
						Properties: map[string]interface{}{
							"date": map[string]interface{}{
								"type":        "string",
								"description": "Date in YYYY-MM-DD format (e.g., 2024-01-15)",
							},
						},
					},
				},
			},
		}
	}
}

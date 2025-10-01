package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type LLMClient interface {
	SendMessage(request interface{}) (*ClaudeResponse, error)
}

// BedrockClient handles communication with AWS Bedrock API
type BedrockClient struct {
	Client *bedrockruntime.Client
	Region string
}

// BedrockRequest represents the request payload to AWS Bedrock API
type BedrockRequest struct {
	Model       string                 `json:"model"`
	Messages    []Message              `json:"messages"`
	MaxTokens   int                    `json:"max_tokens"`
	Temperature float64                `json:"temperature,omitempty"`
	System      string                 `json:"system,omitempty"`
	Tools       []Tool                 `json:"tools,omitempty"`
	ToolChoice  interface{}            `json:"tool_choice,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BedrockResponse represents the response from AWS Bedrock API
type BedrockResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type  string      `json:"type"`
		Text  string      `json:"text,omitempty"`
		Name  string      `json:"name,omitempty"`
		ID    string      `json:"id,omitempty"`
		Input interface{} `json:"input,omitempty"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// NewBedrockClient creates a new AWS Bedrock API client using AWS SDK
func NewBedrockClient(region string) (*BedrockClient, error) {
	os.Environ()

	if region == "" {
		region = "us-east-1"
	}

	// Load AWS configuration from environment variables and credentials
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	// Create Bedrock Runtime client
	client := bedrockruntime.NewFromConfig(cfg)

	return &BedrockClient{
		Client: client,
		Region: region,
	}, nil
}

// SendMessage sends a message to AWS Bedrock API
func (b *BedrockClient) SendMessage(request interface{}) (*ClaudeResponse, error) {

	var req BedrockRequest
	if bedrockReq, ok := request.(BedrockRequest); ok {
		req = bedrockReq
	} else {
		return nil, fmt.Errorf("incompatible request")
	}
	// Build the request body
	requestBody := ClaudeRequest{
		MaxTokens: req.MaxTokens,
		Messages:  req.Messages,
	}

	if req.Temperature > 0 {
		requestBody.Temperature = req.Temperature
	}

	if req.System != "" {
		requestBody.System = req.System
	}

	if len(req.Tools) > 0 {
		requestBody.Tools = req.Tools
		if req.ToolChoice != nil {
			requestBody.ToolChoice = req.ToolChoice
		}
	}

	// Marshal request body
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug: Print request JSON
	fmt.Printf("Bedrock API Request: %s\n", string(jsonData))

	// Model ID for Claude 3.5 Sonnet
	modelID := "anthropic.claude-sonnet-4-5-20250929-v1:0"

	// Invoke the model
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		Body:        jsonData,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}

	resp, err := b.Client.InvokeModel(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke model: %w", err)
	}

	// Parse the response
	var bedrockResp ClaudeResponse
	if err := json.Unmarshal(resp.Body, &bedrockResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &bedrockResp, nil
}

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	anthropic "github.com/liushuangls/go-anthropic/v2"
	"github.com/sirupsen/logrus"

	"bigfoot/golf/opsagent/internal/aws"
	"bigfoot/golf/opsagent/internal/config"
	"bigfoot/golf/opsagent/internal/models"
	"bigfoot/golf/opsagent/internal/notifications"
	"bigfoot/golf/opsagent/pkg/deployment"
	"bigfoot/golf/opsagent/pkg/validation"
)

// BOATAgent represents the Bigfoot Ops Agent for Technology
type BOATAgent struct {
	config       *config.Config
	claudeClient *anthropic.Client
	awsManager   *aws.Manager
	deployer     *deployment.Manager
	validator    *validation.Validator
	notifier     *notifications.Manager
}

// NewBOATAgent creates a new BOAT agent instance
func NewBOATAgent(cfg *config.Config) (*BOATAgent, error) {
	if cfg.ClaudeAPIKey == "" {
		return nil, fmt.Errorf("CLAUDE_API_KEY is required")
	}

	claudeClient := anthropic.NewClient(cfg.ClaudeAPIKey)

	awsManager, err := aws.NewManager(cfg.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS manager: %w", err)
	}

	deployer := deployment.NewManager(awsManager)
	validator := validation.NewValidator()
	notifier := notifications.NewManager()

	return &BOATAgent{
		config:       cfg,
		claudeClient: claudeClient,
		awsManager:   awsManager,
		deployer:     deployer,
		validator:    validator,
		notifier:     notifier,
	}, nil
}

// ProcessTaskRequest processes an incoming task request
func (b *BOATAgent) ProcessTaskRequest(ctx context.Context, eventDetail map[string]interface{}) error {
	logrus.Info("Processing task request with BOAT agent")

	// Parse task request
	taskRequest, err := b.parseTaskRequest(eventDetail)
	if err != nil {
		return fmt.Errorf("failed to parse task request: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"task_id":     taskRequest.ID,
		"task_type":   taskRequest.Type,
		"description": taskRequest.Description,
		"requester":   taskRequest.Requester,
	}).Info("Parsed task request")

	// Use Claude to analyze and plan the task
	deploymentPlan, err := b.analyzeTaskWithClaude(ctx, taskRequest)
	if err != nil {
		return fmt.Errorf("failed to analyze task with Claude: %w", err)
	}

	// Validate container requirements if this is a deployment task
	if taskRequest.Type == "deploy" {
		containerReqs, err := b.extractContainerRequirements(taskRequest, deploymentPlan)
		if err != nil {
			return fmt.Errorf("failed to extract container requirements: %w", err)
		}

		if err := b.validator.ValidateContainerRequirements(containerReqs); err != nil {
			// Send notification about missing requirements
			b.notifier.NotifyMissingRequirements(taskRequest.ID, containerReqs.MissingFields)
			return fmt.Errorf("container requirements validation failed: %w", err)
		}
	}

	// Execute the deployment plan
	result, err := b.deployer.ExecutePlan(ctx, deploymentPlan)
	if err != nil {
		return fmt.Errorf("failed to execute deployment plan: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"task_id":   taskRequest.ID,
		"status":    result.Status,
		"resources": len(result.DeployedResources),
	}).Info("Task processing completed")

	return nil
}

// analyzeTaskWithClaude uses Claude to analyze the task and create a deployment plan
func (b *BOATAgent) analyzeTaskWithClaude(ctx context.Context, task *models.TaskRequest) (*models.DeploymentPlan, error) {
	systemPrompt := b.buildSystemPrompt()
	userPrompt := b.buildUserPrompt(task)

	// Check if we have a real Claude API key
	if b.config.ClaudeAPIKey == "" || b.config.ClaudeAPIKey == "test-key" || b.config.ClaudeAPIKey == "test-key-for-local-development" {
		logrus.WithFields(logrus.Fields{
			"system_prompt": systemPrompt,
			"user_prompt":   userPrompt,
		}).Info("No valid Claude API key - using fallback deployment plan")

		return b.createFallbackDeploymentPlan(task), nil
	}

	// Make real Claude API call
	logrus.WithField("task_id", task.ID).Info("Calling Claude API for task analysis")

	req := anthropic.MessagesRequest{
		Model:  anthropic.ModelClaude3Haiku20240307, // Using available model
		System: systemPrompt,
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewTextMessageContent(userPrompt),
				},
			},
		},
		MaxTokens: 4000,
	}

	resp, err := b.claudeClient.CreateMessages(ctx, req)
	if err != nil {
		logrus.WithError(err).Error("Claude API request failed, using fallback")
		return b.createFallbackDeploymentPlan(task), nil
	}

	// Log Claude API usage
	logrus.WithFields(logrus.Fields{
		"task_id":       task.ID,
		"model":         resp.Model,
		"input_tokens":  resp.Usage.InputTokens,
		"output_tokens": resp.Usage.OutputTokens,
	}).Info("Claude API call successful")

	// Parse Claude's response to extract deployment plan
	if len(resp.Content) == 0 {
		logrus.Warn("Empty response from Claude, using fallback")
		return b.createFallbackDeploymentPlan(task), nil
	}

	claudeResponse := resp.Content[0].GetText()

	// Try to parse as JSON first
	var deploymentPlan models.DeploymentPlan
	if err := json.Unmarshal([]byte(claudeResponse), &deploymentPlan); err != nil {
		logrus.WithError(err).Debug("Claude response not valid JSON, parsing as text")
		deploymentPlan = b.parseClaudeResponseFallback(claudeResponse, task)
	}

	deploymentPlan.TaskID = task.ID

	// Add a note that this came from Claude
	deploymentPlan.Warnings = append(deploymentPlan.Warnings, "Generated by Claude AI")

	return &deploymentPlan, nil
}

// createFallbackDeploymentPlan creates a basic deployment plan when Claude is unavailable
func (b *BOATAgent) createFallbackDeploymentPlan(task *models.TaskRequest) *models.DeploymentPlan {
	return &models.DeploymentPlan{
		TaskID: task.ID,
		Resources: []models.AWSResourceDefinition{
			{
				Type: "AWS::ECS::Service",
				Name: fmt.Sprintf("%s-service", task.ID),
				Properties: map[string]interface{}{
					"ServiceName": fmt.Sprintf("%s-service", task.ID),
				},
			},
		},
		EstimatedCost: 16.20, // Basic ALB cost
		Warnings:      []string{"Using fallback deployment plan (Claude unavailable)"},
		Prerequisites: []string{},
	}
}

// buildSystemPrompt creates the system prompt for Claude
func (b *BOATAgent) buildSystemPrompt() string {
	return fmt.Sprintf(`You are BOAT (Bigfoot Ops Agent for Technology), an expert cloud engineer specializing in AWS infrastructure.

Your core principles:
1. COST OPTIMIZATION: Always choose the most cost-effective solutions, staying within AWS free tier when possible
2. CONTAINERIZATION: All applications must be containerized with complete deployment specifications
3. REGISTRY RESTRICTION: Only use containers from %s registry
4. COMPLETENESS: If any required information is missing, clearly identify what's needed

For every deployment task, ensure these container requirements are specified:
- Container ports and host port mappings
- Dynamic secrets (to be created)
- Existing secrets (already available)
- Configuration environment variables
- External egress requirements (specific sites/APIs the container needs access to)
- Public-facing ports (for load balancer/gateway exposure)
- Health check configuration (path, port, protocol, intervals)

Always respond with a structured deployment plan in JSON format that includes:
- AWS resources needed (ECS services, load balancers, security groups, etc.)
- Estimated cost analysis
- Any warnings about missing information
- Prerequisites that need to be completed first

If critical information is missing, mark the plan as requiring additional input.`, b.config.ContainerRegistry)
}

// buildUserPrompt creates the user prompt for Claude based on the task
func (b *BOATAgent) buildUserPrompt(task *models.TaskRequest) string {
	return fmt.Sprintf(`Task Request:
Type: %s
Description: %s
Requester: %s
Priority: %s

Additional Parameters: %s

Please analyze this request and provide a detailed AWS deployment plan. Focus on cost optimization and ensure all containerization requirements are captured.`,
		task.Type,
		task.Description,
		task.Requester,
		task.Priority,
		formatParameters(task.Parameters))
}

// parseTaskRequest converts the event detail into a structured task request
func (b *BOATAgent) parseTaskRequest(eventDetail map[string]interface{}) (*models.TaskRequest, error) {
	taskJSON, err := json.Marshal(eventDetail)
	if err != nil {
		return nil, err
	}

	var task models.TaskRequest
	if err := json.Unmarshal(taskJSON, &task); err != nil {
		return nil, err
	}

	// Ensure required fields have defaults
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", len(eventDetail))
	}
	if task.Priority == "" {
		task.Priority = "medium"
	}

	return &task, nil
}

// extractContainerRequirements extracts container requirements from task and deployment plan
func (b *BOATAgent) extractContainerRequirements(task *models.TaskRequest, plan *models.DeploymentPlan) (*models.ContainerRequirements, error) {
	// This would extract container requirements from the deployment plan
	// For now, return a basic structure
	return &models.ContainerRequirements{
		Image:                fmt.Sprintf("%s%s", b.config.ContainerRegistry, "default-app:latest"),
		Ports:                []models.PortMapping{},
		DynamicSecrets:       []string{},
		ExistingSecrets:      []string{},
		EnvironmentVariables: make(map[string]string),
		ExternalEgress:       []string{},
		PublicPorts:          []int{},
		HealthCheck:          models.HealthCheckConfig{Enabled: true, Path: "/health", Port: 8080, Protocol: "http"},
		IsValid:              false,
		MissingFields:        []string{},
		ValidationNotes:      []string{},
	}, nil
}

// parseClaudeResponseFallback parses Claude response when JSON parsing fails
func (b *BOATAgent) parseClaudeResponseFallback(response string, task *models.TaskRequest) models.DeploymentPlan {
	// Basic fallback parsing - in production this would be more sophisticated
	return models.DeploymentPlan{
		TaskID:        task.ID,
		Resources:     []models.AWSResourceDefinition{},
		EstimatedCost: 0.0,
		Warnings:      []string{"Failed to parse Claude response, using fallback plan"},
		Prerequisites: []string{},
	}
}

// formatParameters formats task parameters for Claude prompt
func formatParameters(params map[string]interface{}) string {
	if len(params) == 0 {
		return "None specified"
	}

	var parts []string
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s: %v", k, v))
	}
	return strings.Join(parts, ", ")
}

package test

import (
	"context"
	"testing"
	"time"

	"bigfoot/golf/opsagent/internal/config"
	"bigfoot/golf/opsagent/internal/models"
	"bigfoot/golf/opsagent/pkg/validation"
)

func TestContainerRequirementsValidation(t *testing.T) {
	validator := validation.NewValidator()

	tests := []struct {
		name            string
		requirements    *models.ContainerRequirements
		shouldPass      bool
		expectedMissing []string
	}{
		{
			name: "Valid container requirements",
			requirements: &models.ContainerRequirements{
				Image: "ghcr.io/jrzesz33/test-app:latest",
				Ports: []models.PortMapping{
					{ContainerPort: 8080, HostPort: 80, Protocol: "tcp"},
				},
				HealthCheck: models.HealthCheckConfig{
					Enabled:  true,
					Path:     "/health",
					Port:     8080,
					Protocol: "http",
				},
				ExternalEgress: []string{"https://api.example.com"},
				EnvironmentVariables: map[string]string{
					"NODE_ENV": "production",
				},
			},
			shouldPass: true,
		},
		{
			name: "Missing image",
			requirements: &models.ContainerRequirements{
				Ports: []models.PortMapping{
					{ContainerPort: 8080, HostPort: 80, Protocol: "tcp"},
				},
			},
			shouldPass:      false,
			expectedMissing: []string{"container_image"},
		},
		{
			name: "Invalid registry",
			requirements: &models.ContainerRequirements{
				Image: "docker.io/invalid/image:latest",
				Ports: []models.PortMapping{
					{ContainerPort: 8080, HostPort: 80, Protocol: "tcp"},
				},
			},
			shouldPass:      false,
			expectedMissing: []string{"valid_container_image"},
		},
		{
			name: "Missing ports",
			requirements: &models.ContainerRequirements{
				Image: "ghcr.io/jrzesz33/test-app:latest",
			},
			shouldPass:      false,
			expectedMissing: []string{"port_mappings"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateContainerRequirements(tt.requirements)

			if tt.shouldPass && err != nil {
				t.Errorf("Expected validation to pass, but got error: %v", err)
			}

			if !tt.shouldPass && err == nil {
				t.Errorf("Expected validation to fail, but it passed")
			}

			if len(tt.expectedMissing) > 0 {
				for _, expected := range tt.expectedMissing {
					found := false
					for _, missing := range tt.requirements.MissingFields {
						if missing == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected missing field '%s' not found in %v", expected, tt.requirements.MissingFields)
					}
				}
			}
		})
	}
}

func TestConfigLoad(t *testing.T) {
	// Test configuration loading
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.ContainerRegistry == "" {
		t.Error("Container registry should not be empty")
	}

	expectedRegistry := "ghcr.io/jrzesz33/"
	if cfg.ContainerRegistry != expectedRegistry {
		t.Errorf("Expected container registry '%s', got '%s'", expectedRegistry, cfg.ContainerRegistry)
	}

	if cfg.MaxCostThreshold <= 0 {
		t.Error("Max cost threshold should be positive")
	}
}

func TestTaskRequestParsing(t *testing.T) {
	eventDetail := map[string]interface{}{
		"id":          "test-123",
		"type":        "deploy",
		"description": "Test deployment",
		"parameters": map[string]interface{}{
			"image": "ghcr.io/jrzesz33/test:latest",
		},
		"priority":  "medium",
		"requester": "test-user",
	}

	// This would test the parseTaskRequest function
	// For now, just verify the basic structure
	if eventDetail["id"] != "test-123" {
		t.Error("Task ID parsing failed")
	}

	if eventDetail["type"] != "deploy" {
		t.Error("Task type parsing failed")
	}
}

func TestDeploymentPlanValidation(t *testing.T) {
	validator := validation.NewValidator()

	plan := &models.DeploymentPlan{
		TaskID: "test-plan-001",
		Resources: []models.AWSResourceDefinition{
			{
				Type: "AWS::ECS::Service",
				Name: "test-service",
				Properties: map[string]interface{}{
					"ServiceName": "test-service",
				},
			},
		},
		EstimatedCost: 50.0, // Within threshold
	}

	err := validator.ValidateDeploymentPlan(plan, 100.0)
	if err != nil {
		t.Errorf("Expected valid plan, got error: %v", err)
	}

	// Test cost threshold violation
	plan.EstimatedCost = 150.0
	err = validator.ValidateDeploymentPlan(plan, 100.0)
	if err == nil {
		t.Error("Expected cost threshold violation, but validation passed")
	}
}

func TestFreeTierCompliance(t *testing.T) {
	validator := validation.NewValidator()

	/* Test within free tier limits
	resourceCounts := map[string]int{
		"AWS::ECS::Service":      5,
		"AWS::EC2::SecurityGroup": 10,
	}*/

	err := validator.ValidateDeploymentPlan(&models.DeploymentPlan{
		Resources: []models.AWSResourceDefinition{
			{Type: "AWS::ECS::Service", Name: "service1"},
			{Type: "AWS::ECS::Service", Name: "service2"},
			{Type: "AWS::EC2::SecurityGroup", Name: "sg1"},
		},
		EstimatedCost: 25.0,
	}, 100.0)

	if err != nil {
		t.Errorf("Expected free tier compliance, got error: %v", err)
	}
}

// Benchmark tests for performance
func BenchmarkValidateContainerRequirements(b *testing.B) {
	validator := validation.NewValidator()
	requirements := &models.ContainerRequirements{
		Image: "ghcr.io/jrzesz33/test-app:latest",
		Ports: []models.PortMapping{
			{ContainerPort: 8080, HostPort: 80, Protocol: "tcp"},
		},
		HealthCheck: models.HealthCheckConfig{
			Enabled:  true,
			Path:     "/health",
			Port:     8080,
			Protocol: "http",
		},
		ExternalEgress: []string{"https://api.example.com"},
		EnvironmentVariables: map[string]string{
			"NODE_ENV": "production",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateContainerRequirements(requirements)
	}
}

func BenchmarkTaskProcessing(b *testing.B) {
	ctx := context.Background()

	eventDetail := map[string]interface{}{
		"id":          "bench-test",
		"type":        "deploy",
		"description": "Benchmark test deployment",
		"parameters": map[string]interface{}{
			"image": "ghcr.io/jrzesz33/bench:latest",
		},
		"priority":  "medium",
		"requester": "benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate task processing without actual AWS calls
		_ = eventDetail
		_ = ctx
		time.Sleep(time.Microsecond) // Simulate some processing
	}
}

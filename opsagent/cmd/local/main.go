package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"bigfoot/golf/opsagent/internal/agent"
	"bigfoot/golf/opsagent/internal/config"
)

func main() {
	// Configure logging for local development
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Info("Starting BOAT Agent in local development mode")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override for local development
	cfg.ClaudeAPIKey = getEnvOrDefault("CLAUDE_API_KEY", "test-key")
	cfg.AWSRegion = getEnvOrDefault("AWS_REGION", "us-east-1")

	// Initialize BOAT agent
	boatAgent, err := agent.NewBOATAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize BOAT agent: %v", err)
	}

	// Check command line arguments
	if len(os.Args) < 2 {
		showUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "test":
		runTestScenarios(boatAgent)
	case "interactive":
		runInteractiveMode(boatAgent)
	case "event":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ./boat-agent event <event-file.json>")
			return
		}
		runEventFromFile(boatAgent, os.Args[2])
	case "sample":
		generateSampleEvents()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showUsage()
	}
}

func showUsage() {
	fmt.Println("BOAT Agent Local Testing Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  ./boat-agent test         - Run predefined test scenarios")
	fmt.Println("  ./boat-agent interactive  - Interactive testing mode")
	fmt.Println("  ./boat-agent event <file> - Process event from JSON file")
	fmt.Println("  ./boat-agent sample       - Generate sample event files")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  CLAUDE_API_KEY  - Claude API key (optional for testing)")
	fmt.Println("  AWS_REGION      - AWS region (default: us-east-1)")
}

func runTestScenarios(agent *agent.BOATAgent) {
	logrus.Info("Running predefined test scenarios")

	scenarios := []struct {
		name  string
		event map[string]any
	}{
		{
			name: "Simple Container Deployment",
			event: map[string]any{
				"id":          "test-001",
				"type":        "deploy",
				"description": "Deploy a simple web application",
				"parameters": map[string]any{
					"image": "ghcr.io/jrzesz33/web-app:latest",
					"ports": []map[string]any{
						{"container": 8080, "host": 80},
					},
				},
				"priority":  "medium",
				"requester": "local-test",
			},
		},
		{
			name: "Microservice with Database",
			event: map[string]any{
				"id":          "test-002",
				"type":        "deploy",
				"description": "Deploy microservice with Redis cache",
				"parameters": map[string]any{
					"image":          "ghcr.io/jrzesz33/api-service:v2.1.0",
					"replicas":       2,
					"cache_required": true,
				},
				"priority":  "high",
				"requester": "local-test",
			},
		},
		{
			name: "Troubleshooting Request",
			event: map[string]any{
				"id":          "test-003",
				"type":        "troubleshoot",
				"description": "Service is returning 500 errors intermittently",
				"parameters": map[string]any{
					"service_name": "user-auth",
					"error_rate":   "15%",
				},
				"priority":  "urgent",
				"requester": "local-test",
			},
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("\n=== Test Scenario %d: %s ===\n", i+1, scenario.name)
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		
		start := time.Now()
		err := agent.ProcessTaskRequest(ctx, scenario.event)
		duration := time.Since(start)
		
		if err != nil {
			logrus.WithError(err).Errorf("Scenario %d failed", i+1)
		} else {
			logrus.Infof("Scenario %d completed successfully in %v", i+1, duration)
		}
		
		cancel()
		
		// Brief pause between scenarios
		time.Sleep(2 * time.Second)
	}
}

func runInteractiveMode(agent *agent.BOATAgent) {
	fmt.Println("\n=== BOAT Agent Interactive Mode ===")
	fmt.Println("Enter task descriptions (type 'quit' to exit):")

	for {
		fmt.Print("\nBOAT> ")
		
		var input string
		fmt.Scanln(&input)
		
		if input == "quit" || input == "exit" {
			break
		}
		
		if input == "" {
			continue
		}

		// Create event from user input
		event := map[string]any{
			"id":          fmt.Sprintf("interactive-%d", time.Now().Unix()),
			"type":        "deploy",
			"description": input,
			"parameters":  map[string]any{},
			"priority":    "medium",
			"requester":   "interactive-user",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		
		start := time.Now()
		err := agent.ProcessTaskRequest(ctx, event)
		duration := time.Since(start)
		
		if err != nil {
			fmt.Printf("❌ Error: %v (took %v)\n", err, duration)
		} else {
			fmt.Printf("✅ Task processed successfully (took %v)\n", duration)
		}
		
		cancel()
	}
	
	fmt.Println("Goodbye!")
}

func runEventFromFile(agent *agent.BOATAgent, filename string) {
	logrus.WithField("file", filename).Info("Processing event from file")

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read event file: %v", err)
	}

	var event map[string]any
	if err := json.Unmarshal(data, &event); err != nil {
		log.Fatalf("Failed to parse event JSON: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	err = agent.ProcessTaskRequest(ctx, event)
	duration := time.Since(start)

	if err != nil {
		logrus.WithError(err).Errorf("Event processing failed (took %v)", duration)
		os.Exit(1)
	}

	logrus.Infof("Event processed successfully (took %v)", duration)
}

func generateSampleEvents() {
	fmt.Println("Generating sample event files...")

	samples := map[string]map[string]any{
		"deploy-webapp.json": {
			"id":          "sample-webapp-001",
			"type":        "deploy",
			"description": "Deploy a React web application with nginx",
			"parameters": map[string]any{
				"image": "ghcr.io/jrzesz33/react-app:v1.0.0",
				"ports": []map[string]any{
					{"container_port": 80, "host_port": 8080, "protocol": "tcp"},
				},
				"environment_variables": map[string]string{
					"NODE_ENV":   "production",
					"API_URL":    "https://api.example.com",
					"LOG_LEVEL":  "info",
				},
				"external_egress": []string{
					"https://api.example.com",
					"https://cdn.jsdelivr.net",
				},
				"public_ports": []int{8080},
				"health_check": map[string]any{
					"enabled":  true,
					"path":     "/health",
					"port":     80,
					"protocol": "http",
					"interval": 30,
					"timeout":  10,
					"retries":  3,
				},
			},
			"priority":  "medium",
			"requester": "development-team",
		},
		"deploy-api.json": {
			"id":          "sample-api-001",
			"type":        "deploy",
			"description": "Deploy Go API service with Redis cache",
			"parameters": map[string]any{
				"image": "ghcr.io/jrzesz33/api-server:v2.3.1",
				"replicas": 3,
				"ports": []map[string]any{
					{"container_port": 8080, "host_port": 8080, "protocol": "tcp"},
				},
				"dynamic_secrets": []string{"database-password", "jwt-secret"},
				"existing_secrets": []string{"redis-connection-string"},
				"environment_variables": map[string]string{
					"PORT":        "8080",
					"GIN_MODE":    "release",
					"LOG_FORMAT":  "json",
				},
				"external_egress": []string{
					"https://redis-cluster.amazonaws.com:6379",
					"https://api.stripe.com",
				},
				"public_ports": []int{8080},
				"health_check": map[string]any{
					"enabled":  true,
					"path":     "/v1/health",
					"port":     8080,
					"protocol": "http",
					"interval": 15,
					"timeout":  5,
					"retries":  2,
				},
			},
			"priority":  "high",
			"requester": "backend-team",
		},
		"troubleshoot-service.json": {
			"id":          "sample-troubleshoot-001",
			"type":        "troubleshoot",
			"description": "Service experiencing high memory usage and slow response times",
			"parameters": map[string]any{
				"service_name":    "user-management-api",
				"symptoms":        []string{"high memory usage", "slow response times", "occasional timeouts"},
				"error_rate":      "12%",
				"avg_response_time": "2.5s",
				"memory_usage":    "85%",
				"cpu_usage":       "45%",
			},
			"priority":  "urgent",
			"requester": "ops-team",
		},
	}

	for filename, event := range samples {
		data, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			logrus.WithError(err).Errorf("Failed to marshal %s", filename)
			continue
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			logrus.WithError(err).Errorf("Failed to write %s", filename)
			continue
		}

		fmt.Printf("✅ Generated %s\n", filename)
	}

	fmt.Println("\nSample files generated! Use them with:")
	fmt.Println("  ./boat-agent event deploy-webapp.json")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
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
	// Configure logging for simulation
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	if len(os.Args) < 2 {
		showUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "aws-discovery":
		runAWSDiscovery()
	case "cost-analysis":
		runCostAnalysis()
	case "deployment-simulation": 
		runDeploymentSimulation()
	case "interactive":
		runInteractiveSimulation()
	default:
		fmt.Printf("Unknown simulation command: %s\n", command)
		showUsage()
	}
}

func showUsage() {
	fmt.Println("BOAT Agent AWS Simulation Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  ./simulate aws-discovery       - Discover your AWS infrastructure")
	fmt.Println("  ./simulate cost-analysis        - Analyze costs for deployments")
	fmt.Println("  ./simulate deployment-simulation - Simulate container deployments")
	fmt.Println("  ./simulate interactive          - Interactive AWS exploration")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  AWS_ACCESS_KEY_ID      - Your AWS access key")
	fmt.Println("  AWS_SECRET_ACCESS_KEY  - Your AWS secret key")
	fmt.Println("  AWS_REGION            - AWS region (default: us-east-1)")
	fmt.Println("  CLAUDE_API_KEY        - Optional Claude API key")
}

func runAWSDiscovery() {
	logrus.Info("🔍 Starting AWS Infrastructure Discovery")
	
	agent := initializeAgent()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	discoveryEvent := map[string]any{
		"id":          "aws-discovery-simulation",
		"type":        "deploy",
		"description": "Discover and analyze my current AWS infrastructure setup including VPCs, subnets, ECS clusters, and security groups",
		"parameters": map[string]any{
			"image": "ghcr.io/jrzesz33/infrastructure-analyzer:latest",
			"ports": []map[string]any{
				{"container_port": 8080, "host_port": 80, "protocol": "tcp"},
			},
			"health_check": map[string]any{
				"enabled":  true,
				"path":     "/health",
				"port":     8080,
				"protocol": "http",
			},
			"discovery_mode": true,
		},
		"priority":  "high",
		"requester": "aws-simulation-tool",
	}

	start := time.Now()
	err := agent.ProcessTaskRequest(ctx, discoveryEvent)
	duration := time.Since(start)

	if err != nil {
		logrus.WithError(err).Error("AWS discovery failed")
		fmt.Printf("❌ Discovery failed after %v: %v\n", duration, err)
	} else {
		fmt.Printf("✅ AWS discovery completed successfully in %v\n", duration)
	}
}

func runCostAnalysis() {
	logrus.Info("💰 Starting Cost Analysis Simulation")
	
	agent := initializeAgent()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	costEvent := map[string]any{
		"id":          "cost-analysis-simulation",
		"type":        "deploy", 
		"description": "Analyze costs for deploying a production web application with high availability, load balancer, and auto-scaling",
		"parameters": map[string]any{
			"image": "ghcr.io/jrzesz33/production-webapp:v2.1.0",
			"ports": []map[string]any{
				{"container_port": 8080, "host_port": 80, "protocol": "tcp"},
				{"container_port": 8443, "host_port": 443, "protocol": "tcp"},
			},
			"public_ports": []int{80, 443},
			"health_check": map[string]any{
				"enabled":  true,
				"path":     "/api/health",
				"port":     8080,
				"protocol": "http",
				"interval": 30,
				"timeout":  10,
				"retries":  3,
			},
			"environment_variables": map[string]string{
				"NODE_ENV":    "production",
				"LOG_LEVEL":   "info",
				"REDIS_URL":   "redis://redis-cluster:6379",
			},
			"external_egress": []string{
				"https://api.stripe.com",
				"https://api.sendgrid.com",
				"https://cloudflare.com",
			},
			"high_availability": true,
			"auto_scaling":     true,
			"replicas":         3,
		},
		"priority":  "high",
		"requester": "cost-analysis-simulation",
	}

	start := time.Now()
	err := agent.ProcessTaskRequest(ctx, costEvent)
	duration := time.Since(start)

	if err != nil {
		logrus.WithError(err).Error("Cost analysis failed")
		fmt.Printf("❌ Cost analysis failed after %v: %v\n", duration, err)
	} else {
		fmt.Printf("✅ Cost analysis completed successfully in %v\n", duration)
	}
}

func runDeploymentSimulation() {
	logrus.Info("🚀 Starting Deployment Simulation")
	
	agent := initializeAgent()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	deploymentEvent := map[string]any{
		"id":          "deployment-simulation",
		"type":        "deploy",
		"description": "Deploy a complex microservices architecture with API gateway, user service, payment service, and Redis cache. Handle 5000+ concurrent users with proper monitoring.",
		"parameters": map[string]any{
			"services": []string{"api-gateway", "user-service", "payment-service", "redis-cache"},
			"images": []string{
				"ghcr.io/jrzesz33/api-gateway:v1.5.0",
				"ghcr.io/jrzesz33/user-service:v2.0.0", 
				"ghcr.io/jrzesz33/payment-service:v1.3.0",
				"ghcr.io/jrzesz33/redis:alpine",
			},
			"ports": []map[string]any{
				{"container_port": 3000, "host_port": 3000, "protocol": "tcp", "service": "api-gateway"},
				{"container_port": 3001, "host_port": 3001, "protocol": "tcp", "service": "user-service"},
				{"container_port": 3002, "host_port": 3002, "protocol": "tcp", "service": "payment-service"},
				{"container_port": 6379, "host_port": 6379, "protocol": "tcp", "service": "redis-cache"},
			},
			"public_ports": []int{80, 443},
			"health_check": map[string]any{
				"enabled":  true,
				"path":     "/health",
				"port":     3000,
				"protocol": "http",
				"interval": 15,
				"timeout":  5,
				"retries":  2,
			},
			"dynamic_secrets": []string{
				"database-password",
				"jwt-secret",
				"stripe-api-key",
			},
			"environment_variables": map[string]string{
				"NODE_ENV":           "production",
				"LOG_LEVEL":          "info",
				"DATABASE_URL":       "postgres://db:5432/app",
				"REDIS_URL":          "redis://redis-cache:6379",
				"PAYMENT_GATEWAY":    "stripe",
			},
			"external_egress": []string{
				"https://api.stripe.com",
				"https://api.twilio.com",
				"https://hooks.slack.com",
				"https://api.datadog.com",
			},
			"expected_users":      "5000+",
			"high_availability":   true,
			"auto_scaling":       true,
			"monitoring":         "comprehensive",
		},
		"priority":  "urgent",
		"requester": "deployment-simulation",
	}

	start := time.Now()
	err := agent.ProcessTaskRequest(ctx, deploymentEvent)
	duration := time.Since(start)

	if err != nil {
		logrus.WithError(err).Error("Deployment simulation failed")
		fmt.Printf("❌ Deployment simulation failed after %v: %v\n", duration, err)
	} else {
		fmt.Printf("✅ Deployment simulation completed successfully in %v\n", duration)
	}
}

func runInteractiveSimulation() {
	fmt.Println("\n🎮 Interactive AWS Simulation Mode")
	fmt.Println("Enter deployment scenarios to see how BOAT would handle them with your AWS account:")
	fmt.Println("Type 'quit' to exit")

	agent := initializeAgent()

	for {
		fmt.Print("\nBOAT AWS Simulation> ")
		
		var input string
		fmt.Scanln(&input)
		
		if input == "quit" || input == "exit" {
			break
		}
		
		if input == "" {
			continue
		}

		// Create simulation event
		event := map[string]any{
			"id":          fmt.Sprintf("interactive-sim-%d", time.Now().Unix()),
			"type":        "deploy",
			"description": input,
			"parameters": map[string]any{
				"image": "ghcr.io/jrzesz33/app:latest",
				"ports": []map[string]any{
					{"container_port": 8080, "host_port": 80, "protocol": "tcp"},
				},
				"health_check": map[string]any{
					"enabled":  true,
					"path":     "/health",
					"port":     8080,
					"protocol": "http",
				},
				"interactive_mode": true,
			},
			"priority":  "medium",
			"requester": "interactive-simulation",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		
		start := time.Now()
		err := agent.ProcessTaskRequest(ctx, event)
		duration := time.Since(start)
		
		if err != nil {
			fmt.Printf("❌ Error: %v (took %v)\n", err, duration)
		} else {
			fmt.Printf("✅ Simulation completed successfully (took %v)\n", duration)
		}
		
		cancel()
	}
	
	fmt.Println("\n👋 AWS simulation session ended!")
}

func initializeAgent() *agent.BOATAgent {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override for simulation
	if cfg.ClaudeAPIKey == "" {
		cfg.ClaudeAPIKey = "simulation-test-key"
	}

	// Initialize BOAT agent
	boatAgent, err := agent.NewBOATAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize BOAT agent: %v", err)
	}

	return boatAgent
}

func formatJSON(data any) string {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(bytes)
}
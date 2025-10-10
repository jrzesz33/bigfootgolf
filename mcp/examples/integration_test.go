// Integration test example using HTTP client utilities
// Run with: go run examples/integration_test.go examples/http_client.go
//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// This is a manual test runner that can be executed with:
// go run test_main.go test_client.go

func main() {
	// Get server URL from environment or use default
	serverURL := os.Getenv("MCP_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8081"
	}

	client := NewTestClient(serverURL)

	fmt.Println("=== MCP Server Test Suite ===\n")

	// Test 1: Health Check
	fmt.Println("1. Testing Health Check...")
	if err := client.HealthCheck(); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Println("✓ Health check passed\n")

	// Test 2: Server Info
	fmt.Println("2. Testing Server Info...")
	info, err := client.GetServerInfo()
	if err != nil {
		log.Fatalf("Server info failed: %v", err)
	}
	fmt.Printf("✓ Server Info: %+v\n\n", info)

	// Give server a moment to be fully ready
	time.Sleep(500 * time.Millisecond)

	// Test 3: Get Weather Forecast (default 3 days)
	fmt.Println("3. Testing Weather Forecast (default 3 days)...")
	if err := client.TestGetWeatherForecast(0); err != nil {
		log.Printf("✗ Weather forecast test failed: %v\n", err)
	} else {
		fmt.Println("✓ Weather forecast test passed\n")
	}

	// Test 4: Get Weather Forecast (7 days)
	fmt.Println("4. Testing Weather Forecast (7 days)...")
	if err := client.TestGetWeatherForecast(7); err != nil {
		log.Printf("✗ Weather forecast test failed: %v\n", err)
	} else {
		fmt.Println("✓ Weather forecast test passed\n")
	}

	// Test 5: Get Available Tee Times (today)
	fmt.Println("5. Testing Tee Times (today)...")
	today := time.Now().Format("2006-01-02")
	if err := client.TestGetAvailableTeeTimes(today); err != nil {
		log.Printf("✗ Tee times test failed: %v\n", err)
	} else {
		fmt.Println("✓ Tee times test passed\n")
	}

	// Test 6: Get Available Tee Times (specific date)
	fmt.Println("6. Testing Tee Times (2024-12-25)...")
	if err := client.TestGetAvailableTeeTimes("2024-12-25"); err != nil {
		log.Printf("✗ Tee times test failed: %v\n", err)
	} else {
		fmt.Println("✓ Tee times test passed\n")
	}

	// Test 7: Invalid date format
	fmt.Println("7. Testing Invalid Date Format...")
	if err := client.TestGetAvailableTeeTimes("12/25/2024"); err != nil {
		fmt.Printf("✓ Correctly rejected invalid date: %v\n\n", err)
	} else {
		fmt.Println("✗ Should have rejected invalid date format\n")
	}

	fmt.Println("=== Test Suite Complete ===")
}

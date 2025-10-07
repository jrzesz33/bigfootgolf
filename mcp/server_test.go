package main

import (
	"bigfoot/golf/common/models/db"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestMain(m *testing.M) {
	// Set timezone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.Local
	}
	time.Local = loc
	db.TimeLocation = loc

	// Initialize database for tests if DB credentials are available
	ctx := context.Background()
	db.InitDB(ctx)

	// Run tests
	code := m.Run()

	os.Exit(code)
}

func TestNewMCPServer(t *testing.T) {
	server := NewMCPServer()
	if server == nil {
		t.Fatal("NewMCPServer returned nil")
	}
	if server.server == nil {
		t.Fatal("MCP server instance is nil")
	}
}

func TestHandleGetWeatherForecast(t *testing.T) {
	srv := NewMCPServer()
	ctx := context.Background()

	tests := []struct {
		name        string
		args        map[string]interface{}
		wantError   bool
		errorMsg    string
		validateFn  func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name: "default days (3)",
			args: map[string]interface{}{},
			wantError: false,
			validateFn: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("expected no error, got: %v", result.Content)
					return
				}

				contentStr := getTextContent(result)
				if contentStr == "" {
					t.Error("expected weather forecast text, got empty string")
				}

				// Check that the response contains expected text
				if !strings.Contains(contentStr, "weather forecast") && !strings.Contains(contentStr, "Weather forecast") {
					t.Errorf("expected forecast to contain 'weather forecast', got: %s", contentStr)
				}
			},
		},
		{
			name: "specific days (5)",
			args: map[string]interface{}{
				"days": 5,
			},
			wantError: false,
			validateFn: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("expected no error, got: %v", result.Content)
					return
				}

				contentStr := getTextContent(result)
				if contentStr == "" {
					t.Error("expected weather forecast text, got empty string")
				}
			},
		},
		{
			name: "invalid days - too low",
			args: map[string]interface{}{
				"days": 0,
			},
			wantError: true,
			errorMsg:  "days must be between 1 and 7",
		},
		{
			name: "invalid days - too high",
			args: map[string]interface{}{
				"days": 8,
			},
			wantError: true,
			errorMsg:  "days must be between 1 and 7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "get_weather_forecast",
					Arguments: tt.args,
				},
			}

			result, err := srv.handleGetWeatherForecast(ctx, request)
			if err != nil {
				t.Fatalf("handleGetWeatherForecast returned error: %v", err)
			}

			if tt.wantError {
				if !result.IsError {
					t.Errorf("expected error result, got success")
				}
				contentStr := getTextContent(result)
				if contentStr != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, contentStr)
				}
			} else {
				if tt.validateFn != nil {
					tt.validateFn(t, result)
				}
			}
		})
	}
}

func TestHandleGetAvailableTeeTimes(t *testing.T) {
	// Skip if database is not available
	if db.Instance == nil || db.Instance.Err != nil {
		t.Skip("Skipping tee times test: database not available")
	}

	srv := NewMCPServer()
	ctx := context.Background()

	tests := []struct {
		name       string
		args       map[string]interface{}
		wantError  bool
		errorMsg   string
		validateFn func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name: "valid date",
			args: map[string]interface{}{
				"date": time.Now().AddDate(0, 0, 1).Format("2006-01-02"), // Tomorrow
			},
			wantError: false,
			validateFn: func(t *testing.T, result *mcp.CallToolResult) {
				if result.IsError {
					t.Errorf("expected no error, got: %v", result.Content)
					return
				}

				contentStr := getTextContent(result)
				if contentStr == "" {
					t.Error("expected tee times text, got empty string")
				}

				// Should contain either tee times or a message about availability
				hasTeeTimes := strings.Contains(contentStr, "Available tee times") ||
					strings.Contains(contentStr, "No tee times available") ||
					strings.Contains(contentStr, "fully booked") ||
					strings.Contains(contentStr, "failed to get tee times")

				if !hasTeeTimes {
					t.Errorf("expected tee times information, got: %s", contentStr)
				}
			},
		},
		{
			name:      "missing date parameter",
			args:      map[string]interface{}{},
			wantError: true,
			errorMsg:  "date parameter is required and must be a string",
		},
		{
			name: "invalid date format",
			args: map[string]interface{}{
				"date": "01/15/2024",
			},
			wantError: true,
			errorMsg:  "invalid date format",
		},
		{
			name: "invalid date type",
			args: map[string]interface{}{
				"date": 12345,
			},
			wantError: true,
			errorMsg:  "date parameter is required and must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "get_available_tee_times",
					Arguments: tt.args,
				},
			}

			result, err := srv.handleGetAvailableTeeTimes(ctx, request)
			if err != nil {
				t.Fatalf("handleGetAvailableTeeTimes returned error: %v", err)
			}

			if tt.wantError {
				if !result.IsError {
					t.Errorf("expected error result, got success")
				}
				contentStr := getTextContent(result)
				if tt.errorMsg != "" && contentStr != tt.errorMsg {
					// Check if error message contains expected text
					if len(contentStr) < len(tt.errorMsg) || contentStr[:len(tt.errorMsg)] != tt.errorMsg {
						t.Errorf("expected error message to start with %q, got %q", tt.errorMsg, contentStr)
					}
				}
			} else {
				if tt.validateFn != nil {
					tt.validateFn(t, result)
				}
			}
		})
	}
}

// Helper function to extract text content from CallToolResult
func getTextContent(result *mcp.CallToolResult) string {
	if len(result.Content) == 0 {
		return ""
	}

	// Content can be either TextContent or ImageContent
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			return textContent.Text
		}
	}

	return ""
}

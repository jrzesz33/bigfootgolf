package main

import (
	"bigfoot/golf/common/models/teetimes"
	"bigfoot/golf/common/models/weather"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer represents the golf booking MCP server
type MCPServer struct {
	server *server.MCPServer
}

// NewMCPServer creates a new MCP server instance with StreamableHTTP transport
func NewMCPServer() *MCPServer {
	s := &MCPServer{}

	// Create MCP server with server info
	mcpServer := server.NewMCPServer(
		"golf-booking-server",
		"1.0.0",
	)

	// Register tools
	s.registerTools(mcpServer)

	s.server = mcpServer
	return s
}

// registerTools registers all available tools with the MCP server
func (s *MCPServer) registerTools(mcpServer *server.MCPServer) {
	// Register get_weather_forecast tool
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_weather_forecast",
		Description: "Get weather forecast for golf course area",
		InputSchema: mcp.ToolInputSchema{
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
	}, s.handleGetWeatherForecast)

	// Register get_available_tee_times tool
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_available_tee_times",
		Description: "Get available tee times for a specific date",
		InputSchema: mcp.ToolInputSchema{
			Type:     "object",
			Required: []string{"date"},
			Properties: map[string]interface{}{
				"date": map[string]interface{}{
					"type":        "string",
					"description": "Date in YYYY-MM-DD format (e.g., 2024-01-15)",
				},
			},
		},
	}, s.handleGetAvailableTeeTimes)
}

// handleGetWeatherForecast handles the get_weather_forecast tool call
func (s *MCPServer) handleGetWeatherForecast(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract days parameter with default value of 3
	days := request.GetInt("days", 3)

	// Validate days range
	if days < 1 || days > 7 {
		return mcp.NewToolResultError("days must be between 1 and 7"), nil
	}

	// Use the existing weather handler
	weatherHandler := weather.NewWeatherHandler(
		"https://api.weather.gov/gridpoints/PBZ/77,65/forecast",
		30*time.Minute,
	)

	weatherData, err := weatherHandler.GetWeatherData()
	if err != nil {
		// Fallback to basic forecast if weather API fails
		var result strings.Builder
		result.WriteString(fmt.Sprintf("Weather forecast for the next %d days:\n", days))

		for i := 0; i < days; i++ {
			date := time.Now().AddDate(0, 0, i)
			result.WriteString(fmt.Sprintf("- %s: Check local weather services for conditions\n",
				date.Format("January 2")))
		}
		return mcp.NewToolResultText(result.String()), nil
	}

	var result strings.Builder
	result.WriteString("Golf course weather forecast:\n")

	// Limit to requested number of days
	periodsToShow := days * 2 // Day and night periods
	if len(weatherData.Properties.Periods) < periodsToShow {
		periodsToShow = len(weatherData.Properties.Periods)
	}

	for i := 0; i < periodsToShow && i < len(weatherData.Properties.Periods); i++ {
		period := weatherData.Properties.Periods[i]
		result.WriteString(fmt.Sprintf("- %s: %s, %d°%s, %s %s\n",
			period.Name,
			period.ShortForecast,
			period.Temperature,
			period.TemperatureUnit,
			period.WindSpeed,
			period.WindDirection,
		))
	}

	result.WriteString("\nPerfect for planning your golf outing!")
	return mcp.NewToolResultText(result.String()), nil
}

// handleGetAvailableTeeTimes handles the get_available_tee_times tool call
func (s *MCPServer) handleGetAvailableTeeTimes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract date parameter
	dateStr, err := request.RequireString("date")
	if err != nil {
		return mcp.NewToolResultError("date parameter is required and must be a string"), nil
	}

	// Validate date format
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid date format: %v. Use YYYY-MM-DD", err)), nil
	}

	// Use the existing booking engine to get tee times
	var booking teetimes.BookingEngine
	days, err := booking.GetDayTeeTimes(date)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get tee times: %v", err)), nil
	}

	if len(days) == 0 {
		return mcp.NewToolResultText("No tee times available for " + dateStr), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Available tee times for %s:\n", dateStr))

	for _, day := range days {
		for _, slot := range day.Times {
			if len(slot.Players) < 4 { // Only show available slots
				result.WriteString(fmt.Sprintf("- %s | Slot %d | %d spots available | $%.2f | %s\n",
					slot.TeeTime.Format("3:04 PM"),
					slot.Slot,
					4-len(slot.Players),
					slot.Price,
					slot.Group,
				))
			}
		}
	}

	if result.Len() == len(fmt.Sprintf("Available tee times for %s:\n", dateStr)) {
		return mcp.NewToolResultText("All tee times are fully booked for " + dateStr), nil
	}

	return mcp.NewToolResultText(result.String()), nil
}

// GetServer returns the underlying MCP server instance
func (s *MCPServer) GetServer() *server.MCPServer {
	return s.server
}

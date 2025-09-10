package main

import (
	"bigfoot/golf/common/models/auth"
	"bigfoot/golf/common/models/db"
	"bigfoot/golf/common/models/teetimes"
	"bigfoot/golf/common/models/weather"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type MCPServer struct {
	server *server.MCPServer
	router *mux.Router
}

func NewMCPServer() *MCPServer {
	return &MCPServer{
		router: mux.NewRouter(),
	}
}

func (m *MCPServer) Initialize(ctx context.Context) error {
	// Initialize database connection
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return fmt.Errorf("error loading location: %v", err)
	}
	time.Local = loc
	db.TimeLocation = loc

	// Initialize DB
	db.InitDB(ctx)

	// Create MCP server with standard configuration
	mcpServer := server.NewMCPServer("Golf Booking MCP Server", "1.0.0")

	// Register tools
	if err := m.registerTools(mcpServer); err != nil {
		return fmt.Errorf("failed to register tools: %v", err)
	}

	m.server = mcpServer

	return nil
}

func (m *MCPServer) registerTools(s *server.MCPServer) error {
	// Tool: Manage Reservations
	s.AddTool(
		mcp.Tool{
			Name:        "manage_reservations",
			Description: "Get, book, or cancel user reservations",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"action": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"get", "book", "cancel"},
						"description": "Action to perform on reservations",
					},
					"user_id": map[string]interface{}{
						"type":        "string",
						"description": "User ID for the reservation",
					},
					"reservation_id": map[string]interface{}{
						"type":        "string",
						"description": "Reservation ID (required for cancel)",
					},
					"tee_time": map[string]interface{}{
						"type":        "string",
						"description": "Tee time to book (ISO format, required for book)",
					},
					"players": map[string]interface{}{
						"type":        "integer",
						"description": "Number of players (required for book)",
					},
				},
				Required: []string{"action", "user_id"},
			},
		},
		m.handleReservations,
	)

	// Tool: Find Tee Times
	s.AddTool(
		mcp.Tool{
			Name:        "find_tee_times",
			Description: "Find available tee times",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"date": map[string]interface{}{
						"type":        "string",
						"description": "Date to search for tee times (YYYY-MM-DD)",
					},
					"time_range": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"morning", "midday", "afternoon", "all"},
						"description": "Preferred time range",
					},
					"players": map[string]interface{}{
						"type":        "integer",
						"description": "Number of players",
					},
				},
				Required: []string{"date"},
			},
		},
		m.handleFindTeeTimes,
	)

	// Tool: Get Weather/Conditions
	s.AddTool(
		mcp.Tool{
			Name:        "get_conditions",
			Description: "Get current weather and course conditions",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"date": map[string]interface{}{
						"type":        "string",
						"description": "Date for weather forecast (YYYY-MM-DD)",
					},
				},
			},
		},
		m.handleWeatherConditions,
	)

	return nil
}

func (m *MCPServer) handleReservations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	arguments := request.Params.Arguments
	var argMap map[string]any
	if aMap, ok := arguments.(map[string]any); ok {
		argMap = aMap
	}
	action, _ := argMap["action"].(string)
	userID, _ := argMap["user_id"].(string)

	switch action {
	case "get":
		reservations, err := teetimes.GetUserReservationsForMCP(userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get reservations: %v", err)
		}

		content, _ := json.Marshal(reservations)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(content),
				},
			},
		}, nil

	case "book":
		teeTime, _ := argMap["tee_time"].(string)
		players, _ := argMap["players"].(float64)

		teeTimeDate, err := time.Parse(time.RFC3339, teeTime)
		if err != nil {
			return nil, fmt.Errorf("invalid tee time format: %v", err)
		}

		reservation, err := teetimes.CreateReservation(userID, teeTimeDate, int(players))
		if err != nil {
			return nil, fmt.Errorf("failed to book reservation: %v", err)
		}

		content, _ := json.Marshal(reservation)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(content),
				},
			},
		}, nil

	case "cancel":
		reservationID, _ := argMap["reservation_id"].(string)

		err := teetimes.CancelReservation(userID, reservationID)
		if err != nil {
			return nil, fmt.Errorf("failed to cancel reservation: %v", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "Reservation cancelled successfully",
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("invalid action: %s", action)
	}
}

func (m *MCPServer) handleFindTeeTimes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	var arguments map[string]any
	if aMap, ok := args.(map[string]any); ok {
		arguments = aMap
	}
	dateStr, _ := arguments["date"].(string)
	timeRange, _ := arguments["time_range"].(string)
	players, _ := arguments["players"].(float64)

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %v", err)
	}

	availableTimes, err := teetimes.GetAvailableTeeTimes(date, timeRange, int(players))
	if err != nil {
		return nil, fmt.Errorf("failed to get available tee times: %v", err)
	}

	content, _ := json.Marshal(availableTimes)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(content),
			},
		},
	}, nil
}

func (m *MCPServer) handleWeatherConditions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	var arguments map[string]any
	if aMap, ok := args.(map[string]any); ok {
		arguments = aMap
	}

	dateStr, _ := arguments["date"].(string)

	var date time.Time
	var err error

	if dateStr != "" {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %v", err)
		}
	} else {
		date = time.Now()
	}

	weatherData, err := weather.GetWeatherForecast(date)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather data: %v", err)
	}

	conditions, err := teetimes.GetCourseConditions()
	if err != nil {
		return nil, fmt.Errorf("failed to get course conditions: %v", err)
	}

	result := map[string]interface{}{
		"weather":    weatherData,
		"conditions": conditions,
	}

	content, _ := json.Marshal(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(content),
			},
		},
	}, nil
}

// HTTP Handler for MCP over HTTP with authentication
func (m *MCPServer) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract and validate token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
		return
	}

	// Validate JWT token
	token, err := jwt.ParseWithClaims(tokenString, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return auth.GetJWTSecret(), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse MCP request
	var request struct {
		Method string                 `json:"method"`
		Params map[string]interface{} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Handle based on method
	var response interface{}

	switch request.Method {
	case "initialize":
		response = map[string]interface{}{
			"protocolVersion": "1.0",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
		}

	case "tools/list":
		tools := []map[string]interface{}{}
		// Add tool descriptions
		tools = append(tools, map[string]interface{}{
			"name":        "manage_reservations",
			"description": "Get, book, or cancel user reservations",
		})
		tools = append(tools, map[string]interface{}{
			"name":        "find_tee_times",
			"description": "Find available tee times",
		})
		tools = append(tools, map[string]interface{}{
			"name":        "get_conditions",
			"description": "Get weather and course conditions",
		})
		response = map[string]interface{}{
			"tools": tools,
		}

	case "tools/call":
		toolName, _ := request.Params["name"].(string)
		arguments, _ := request.Params["arguments"].(map[string]interface{})

		// Add user context from JWT claims
		if claims, ok := token.Claims.(*auth.Claims); ok {
			if arguments == nil {
				arguments = make(map[string]interface{})
			}
			arguments["user_id"] = claims.UserID
		}

		var result *mcp.CallToolResult

		// Create a CallToolRequest
		toolRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      toolName,
				Arguments: arguments,
			},
		}

		switch toolName {
		case "manage_reservations":
			result, err = m.handleReservations(r.Context(), toolRequest)
		case "find_tee_times":
			result, err = m.handleFindTeeTimes(r.Context(), toolRequest)
		case "get_conditions":
			result, err = m.handleWeatherConditions(r.Context(), toolRequest)
		default:
			err = fmt.Errorf("unknown tool: %s", toolName)
		}

		if err != nil {
			response = map[string]interface{}{
				"error": err.Error(),
			}
		} else {
			response = result
		}

	default:
		http.Error(w, "Unknown method", http.StatusBadRequest)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func StartMCPServer() {
	ctx := context.Background()

	// Create MCP server
	mcpServer := NewMCPServer()

	// Initialize
	if err := mcpServer.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize MCP server: %v", err)
	}

	// Setup HTTP routes
	mcpServer.router.HandleFunc("/mcp", mcpServer.handleHTTP).Methods("POST")
	mcpServer.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Start server
	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("MCP Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mcpServer.router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

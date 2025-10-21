package main

import (
	"bigfoot/golf/common/models/db"
	"bigfoot/golf/mcp/mcp"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Set timezone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Printf("Warning: Error loading timezone: %v. Using local time.", err)
		loc = time.Local
	}
	time.Local = loc
	db.TimeLocation = loc
	fmt.Println("Application timezone set to:", time.Local.String())

	// Initialize database connection
	ctx := context.Background()
	db.InitDB(ctx)
	if db.Instance.Err != nil {
		log.Printf("Warning: Database initialization failed: %v. MCP server will have limited functionality.", db.Instance.Err)
	}

	// Create new MCP server
	mcpServer := mcp.NewMCPServer()

	// Create StreamableHTTP server with stateful sessions
	streamableServer := server.NewStreamableHTTPServer(
		mcpServer.GetServer(),
		server.WithEndpointPath("/mcp"),
	)

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.Handle("/mcp", streamableServer)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting MCP server on %s", addr)
	log.Printf("MCP endpoint: http://localhost%s/mcp", addr)
	log.Printf("Health check: http://localhost%s/health", addr)
	log.Printf("Info: http://localhost%s/info", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

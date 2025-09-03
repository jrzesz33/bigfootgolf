package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

// ProxyServer handles proxying requests for development
type ProxyServer struct {
	mcpServerURL string
	router       *mux.Router
}

// NewProxyServer creates a new proxy server for development
func NewProxyServer(mcpServerURL string) *ProxyServer {
	if mcpServerURL == "" {
		mcpServerURL = "http://localhost:8081"
	}
	
	return &ProxyServer{
		mcpServerURL: mcpServerURL,
		router:       mux.NewRouter(),
	}
}

// Initialize sets up the proxy routes
func (p *ProxyServer) Initialize() {
	// Proxy MCP requests
	p.router.HandleFunc("/proxy/mcp", p.handleMCPProxy).Methods("POST", "OPTIONS")
	
	// Health check
	p.router.HandleFunc("/proxy/health", p.handleHealth).Methods("GET")
	
	// CORS middleware for development
	p.router.Use(p.corsMiddleware)
}

// corsMiddleware adds CORS headers for development
func (p *ProxyServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// handleMCPProxy proxies requests to the MCP server
func (p *ProxyServer) handleMCPProxy(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request for debugging
	log.Printf("Proxy received request: %s %s", r.Method, r.URL.Path)
	
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Log request body for debugging
	log.Printf("Request body: %s", string(body))
	
	// Check for development token or generate one
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		// Generate a development token for testing
		// In real development, this would use a test user account
		authHeader = "Bearer dev-token-for-testing"
		log.Printf("Using development token")
	}
	
	// Create new request to MCP server
	mcpURL := p.mcpServerURL + "/mcp"
	proxyReq, err := http.NewRequest("POST", mcpURL, bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create proxy request: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Copy headers
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Authorization", authHeader)
	
	// Send request to MCP server
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Failed to proxy request: %v", err)
		http.Error(w, fmt.Sprintf("Failed to proxy request: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Log response for debugging
	log.Printf("MCP server response status: %d", resp.StatusCode)
	log.Printf("MCP server response body: %s", string(respBody))
	
	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	
	// Write response
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// handleHealth handles health check requests
func (p *ProxyServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "healthy",
		"type":   "proxy",
		"mcp":    p.mcpServerURL,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartProxyServer starts the development proxy server
func StartProxyServer() {
	mcpURL := os.Getenv("MCP_SERVER_URL")
	if mcpURL == "" {
		mcpURL = "http://localhost:8081"
	}
	
	proxy := NewProxyServer(mcpURL)
	proxy.Initialize()
	
	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "8082"
	}
	
	log.Printf("Starting MCP Proxy Server on port %s", port)
	log.Printf("Proxying to MCP server at %s", mcpURL)
	
	if err := http.ListenAndServe(":"+port, proxy.router); err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}
}
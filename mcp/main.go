package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Parse command line flags
	var mode string
	flag.StringVar(&mode, "mode", "server", "Mode to run in: server or proxy")
	flag.Parse()
	
	// Check environment variable as well
	if envMode := os.Getenv("MCP_MODE"); envMode != "" {
		mode = envMode
	}
	
	switch mode {
	case "server":
		// Run as MCP server
		fmt.Println("Starting MCP Server...")
		StartMCPServer()
		
	case "proxy":
		// Run as development proxy
		fmt.Println("Starting MCP Proxy Server for development...")
		StartProxyServer()
		
	default:
		fmt.Printf("Unknown mode: %s. Use 'server' or 'proxy'\n", mode)
		os.Exit(1)
	}
}
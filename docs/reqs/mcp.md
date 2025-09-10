# MCP Server for Reservation Management 

## REQUIREMENTS:
1. Security
   - The Service shall require a valid OAuth token from auth.AuthServer to Authenticate the MCP Request

2. Available Tools for the MCP Server
    - A tool to get, book, or cancel a list of users reservations should be available
    - A tool to find available tee times
    - A tool to get the course weather and conditions

3. Agent Integration
   - The AI Agent within the go-app application should utilize this MCP Server
   - For development and testing purposes the Agent should have a flag to utilize a proxy for local development

## TECHNICAL PREFERENCES:
- Follow Existing Techincal Stacks
- go-app Progressive Web Application Front End
- Go Backend services
- neo4j Graph Database
- Claude API Integration
- MCP Framework github.com/mark3labs/mcp-go/mcp
- MCP Server module should be built within the mcp folder
# Golf Booking App

A modern, mobile-friendly golf tee time booking application built with Go and WebAssembly.

## Overview

This application provides a comprehensive golf course management system with features for booking tee times, user authentication, course administration, and an AI-powered chat assistant for customer support.

## Features

- **User Authentication**: Secure registration, login, and password reset
- **Tee Time Booking**: Search and book available tee times
- **User Profiles**: Manage personal information and booking history
- **Course Administration**: Admin interface for managing seasons, pricing, and reservations
- **AI Chat Assistant**: Claude-powered chat bot for customer support
- **Mobile-First Design**: Responsive interface optimized for mobile devices
- **Progressive Web App (PWA)**: Installable web application with offline capabilities

## Technology Stack

- **Backend**: Go 1.23.4
- **Frontend**: Go + WebAssembly (WASM)
- **Database**: Neo4j (with SQLite support)
- **Authentication**: JWT tokens with OAuth integration
- **UI Framework**: go-app (Progressive Web App framework)
- **AI Integration**: Anthropic Claude API
- **Routing**: Gorilla Mux
- **Session Management**: Gorilla Sessions

## Project Structure

This project uses Go workspaces with two main modules:

```
golf_app/
├── pkg/                    # Backend module (bigfoot/golf/common)
│   ├── controllers/       # Business logic controllers
│   ├── handlers/         # HTTP request handlers
│   │   ├── admin/       # Admin-specific handlers
│   │   ├── sessionmgr/  # Session management
│   │   └── transactions/# Transaction handlers
│   ├── helper/          # Utility functions and configuration
│   ├── models/          # Data models and business logic
│   │   ├── account/    # User account models
│   │   ├── anthropic/  # AI chat integration
│   │   ├── auth/       # Authentication models
│   │   ├── db/         # Database utilities
│   │   ├── teetimes/   # Tee time booking models
│   │   └── weather/    # Weather integration
│   └── go.mod          # Backend module dependencies
├── web/                 # Frontend module (bigfoot/golf/web)
│   ├── app/            # WebAssembly application
│   │   ├── clients/   # API client utilities
│   │   ├── components/# Reusable UI components
│   │   ├── pages/     # Application pages
│   │   ├── routes/    # Frontend routing
│   │   └── state/     # Application state management
│   ├── public/        # Static assets (CSS, images)
│   ├── main.go        # WebAssembly entry point
│   └── go.mod         # Frontend module dependencies
├── mcp/                # Model Context Protocol server
│   ├── server.go      # MCP server implementation
│   ├── proxy.go       # Development proxy
│   └── README.md      # MCP-specific documentation
├── gateway/            # agentgateway configuration
│   └── config.yaml    # Gateway routing configuration
├── go.work            # Go workspace configuration
├── .env.example       # Environment variable template
└── README.md          # This file
```

## Installation

### Prerequisites

- Go 1.23.4 or higher
- Neo4j database
- (Optional) agentgateway for LLM proxying

### Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd golf_app
   ```

2. **Install Go workspace dependencies**
   ```bash
   go work sync
   ```

3. **Set up environment variables**

   Copy the example environment file and configure it:
   ```bash
   cp .env.example .env
   ```

   Edit `.env` and set required values:
   - `DB_ADMIN` - Neo4j password (required)
   - `JWT_SECRET` - Secret for JWT signing (required)
   - See `.env.example` for all available options

4. **Initialize the database**
   - Ensure Neo4j is running on `bolt://localhost:7687` (or configure `DB_URI`)
   - The application will automatically create the necessary schema

5. **Build WebAssembly frontend**
   ```bash
   GOOS=js GOARCH=wasm go build -o web/public/app.wasm web/main.go
   ```

6. **Run the server**
   ```bash
   MODE=dev go run web/main.go
   ```

## Development

### Go Workspace Structure

This project uses Go workspaces. All commands should be run from the project root:

```bash
# Sync workspace dependencies
go work sync

# Build WebAssembly (note the path to web/main.go!)
GOOS=js GOARCH=wasm go build -o web/public/app.wasm web/main.go

# Run server
go run web/main.go
```

### Building Components

**Backend Server:**
```bash
go build -o bin/web-server web/main.go
```

**WebAssembly Frontend:**
```bash
GOOS=js GOARCH=wasm go build -o web/public/app.wasm web/main.go
```

**MCP Server:**
```bash
cd mcp && go build -o mcp .
./mcp -mode server  # Run on port 8081
```

### Running in Development Mode

```bash
MODE=dev go run web/main.go
```

This will:
- Set up test/development data
- Enable verbose logging
- Skip certain validations

### API Endpoints

#### Public Endpoints
- `GET /papi/teetimes` - Get available tee times
- `POST /auth/register` - User registration
- `POST /auth/login` - User login

#### Authenticated Endpoints
- `GET /api/profile` - Get user profile
- `POST /api/booking` - Book a tee time
- `POST /api/chat` - Chat with AI assistant

#### Admin Endpoints
- `GET /admin/seasons` - Manage golf seasons
- `POST /admin/settings` - Update course settings

### Database Models

#### Key Entities
- **User**: User accounts with authentication
- **Season**: Golf season configuration
- **Reservation**: Tee time bookings
- **ReservationBlock**: Time slot availability
- **DetailedBlockSettings**: Pricing and availability rules

## Configuration

### Authentication
The application supports multiple authentication methods:
- Email/password authentication
- Google OAuth integration
- Apple Sign-In integration

### Pricing
Pricing is configured through `DetailedBlockSettings` with support for:
- Weekday/weekend pricing
- Morning/midday/afternoon rates
- Seasonal adjustments
- Special event pricing

### AI Chat & MCP Integration

The application includes an AI-powered chat assistant using Anthropic's Claude API with Model Context Protocol (MCP) for secure tool access:

**Features:**
- Help users find tee times
- Answer questions about bookings
- Provide course information
- Handle cancellations and modifications
- Access real-time reservation data via MCP

**MCP Server:**

The MCP server provides secure, OAuth-authenticated access to reservation tools:

```bash
# Start MCP server
cd mcp && go run . -mode server

# MCP server runs on port 8081 by default
# See mcp/README.md for detailed configuration
```

**Available MCP Tools:**
- `manage_reservations` - Get, book, or cancel user reservations
- `find_tee_times` - Find available tee times
- `get_conditions` - Get weather and course conditions

For detailed MCP documentation, see [mcp/README.md](mcp/README.md)

## Code Quality & Testing

### Linting

The project includes a golangci-lint configuration:

```bash
# Run linters
golangci-lint run

# Run linters on specific module
cd pkg && golangci-lint run
cd web && golangci-lint run
```

### Code Formatting

```bash
# Format code
gofmt -s -w .

# Check imports
goimports -w .
```

### Static Analysis

```bash
# Run from module directories (pkg/ or web/)
cd pkg && go vet ./...
cd web && go vet ./...
```

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Deployment

### Production Build

1. **Build the WebAssembly frontend**
   ```bash
   GOOS=js GOARCH=wasm go build -o web/public/app.wasm web/main.go
   ```

2. **Build the server binary**
   ```bash
   go build -o bin/golf-app web/main.go
   ```

3. **Set production environment variables**
   ```bash
   export MODE="production"
   export DB_ADMIN="production-db-password"
   export JWT_SECRET="your-secure-jwt-secret"
   export SESSION_KEY="your-secure-session-key"
   export COOKIE_SECURE="true"
   ```

4. **Deploy with your preferred method** (Docker, systemd, etc.)

### Docker Deployment

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o golf-app main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/golf-app .
COPY --from=builder /app/web ./web
CMD ["./golf-app"]
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting: `go vet ./...` and `staticcheck ./...`
5. Submit a pull request

## License

[Add your license information here]

## Support

For support or questions, please [create an issue](https://github.com/your-repo/golf-app/issues) or contact the development team.
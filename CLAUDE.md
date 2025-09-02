# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based golf tee time booking application using a multi-module workspace structure. The application consists of a WebAssembly frontend and a Go backend with Neo4j database integration.

## Architecture

### Module Structure
The project uses Go workspaces with two main modules:
- `pkg/` (module: `bigfoot/golf/common`) - Backend logic, handlers, models, database interactions
- `web/` (module: `bigfoot/golf/web`) - WebAssembly frontend application using go-app framework

### Key Components
- **Frontend**: WebAssembly-based Progressive Web App (PWA) using go-app framework
- **Backend**: Go HTTP server with Gorilla Mux routing
- **Database**: Neo4j graph database with bolt protocol
- **Authentication**: JWT tokens with session management via Gorilla Sessions
- **AI Integration**: Anthropic Claude API for chat assistant functionality

### Directory Layout
```
pkg/
├── controllers/     # Business logic (bf_agent.go, testdata.go)
├── handlers/        # HTTP request handlers
│   ├── admin/      # Admin-specific handlers
│   ├── sessionmgr/ # Session management
│   └── transactions/# Transaction handlers
├── models/         # Data models
│   ├── account/    # User account models
│   ├── anthropic/  # Claude AI integration
│   ├── auth/       # Authentication models
│   ├── db/         # Database utilities
│   ├── teetimes/   # Tee time booking models
│   └── weather/    # Weather integration
└── helper/         # Utility functions

web/
├── app/
│   ├── clients/    # API client utilities
│   ├── components/ # Reusable UI components
│   │   └── userui/ # User UI components
│   ├── pages/      # Application pages (login, register, bookings, etc.)
│   ├── routes/     # Frontend routing
│   └── state/      # Application state management
├── public/         # Static assets (CSS, images)
└── main.go        # WebAssembly entry point
```

## Development Commands

### Building the WebAssembly Frontend
```bash
GOOS=js GOARCH=wasm go build -o web/app.wasm web/main.go
```

### Running the Server
```bash
# Development mode (sets up test data)
MODE=dev go run web/main.go

# Production mode
go run web/main.go
```

### Working with Go Workspace
```bash
# Sync workspace dependencies
go work sync

# Add/update dependencies in specific module
cd pkg && go get <package>
cd web && go get <package>
```

### Code Quality
```bash
# Run Go vet (must be run from module directories)
cd pkg && go vet ./...
cd web && go vet ./...

# Run staticcheck (if available)
cd pkg && staticcheck ./...
cd web && staticcheck ./...
```

### Building the Complete Application
```bash
# Build WebAssembly
GOOS=js GOARCH=wasm go build -o web/app.wasm web/main.go

# Build server binary
go build -o golf-app web/main.go
```

## Environment Configuration

### Required Environment Variables
- `DB_ADMIN`: Neo4j database password
- `MODE`: Set to "dev" for development mode (enables test data setup)

### Database Configuration
- Neo4j connection: `bolt://localhost:7687`
- Timezone: America/New_York (hardcoded in main.go)

## API Routes

### Public Routes (`/papi/`)
- Tee times availability and public data

### Authenticated Routes (`/api/`)
- User profile, bookings, chat functionality

### Auth Routes (`/auth/`)
- Login, register, password reset, verification

### Admin Routes (`/admin/`)
- Season management, settings, reservations

## Static Assets
- Served from `/public/` path
- CSS files: app.css, app_add.css, nav.css, agent.css

## Key Dependencies
- `github.com/gorilla/mux` - HTTP routing
- `github.com/maxence-charriere/go-app/v10` - PWA framework
- `github.com/neo4j/neo4j-go-driver/v5` - Database driver
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `golang.org/x/crypto` - Password hashing

## Development Notes
- The application uses server port 8000
- WebAssembly handles client-side routing for PWA navigation
- Session management uses Gorilla Sessions
- Development mode (`MODE=dev`) automatically sets up test data via `controllers.SetupDevEnvironment()`
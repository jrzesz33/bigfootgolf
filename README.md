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

```
golf_app/
├── app/                    # Frontend WebAssembly application
│   ├── clients/           # API client utilities
│   ├── components/        # Reusable UI components
│   ├── pages/            # Application pages
│   ├── routes/           # Frontend routing
│   └── state/            # Application state management
├── controllers/           # Business logic controllers
├── handlers/             # HTTP request handlers
│   ├── admin/           # Admin-specific handlers
│   ├── sessionmgr/      # Session management
│   └── transactions/    # Transaction handlers
├── helper/               # Utility functions
├── models/              # Data models and database interactions
│   ├── account/         # User account models
│   ├── anthropic/       # AI chat integration
│   ├── auth/           # Authentication models
│   ├── db/             # Database utilities
│   ├── teetimes/       # Tee time booking models
│   └── weather/        # Weather integration
├── web/                 # Static web assets (CSS, images)
├── main.go             # Application entry point
└── go.mod              # Go module dependencies
```

## Installation

### Prerequisites

- Go 1.23.4 or higher
- Neo4j database (or SQLite for development)
- Node.js (for web asset compilation)

### Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd golf_app
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   export DB_ADMIN="your-neo4j-password"
   export MODE="dev"  # for development
   ```

4. **Initialize the database**
   - Ensure Neo4j is running on `bolt://localhost:7687`
   - The application will automatically create the necessary schema

5. **Build and run**
   ```bash
   go build -o golf-app main.go
   ./golf-app
   ```

## Development

### Building for WebAssembly

The application uses Go's WebAssembly support for the frontend:

```bash
GOOS=js GOARCH=wasm go build -o web/app.wasm main.go
```

### Running in Development Mode

```bash
MODE=dev go run main.go
```

This will:
- Set up development data
- Enable debug logging
- Use local configuration

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

### AI Chat
The chat assistant uses Anthropic's Claude API and can:
- Help users find tee times
- Answer questions about bookings
- Provide course information
- Handle cancellations and modifications

## Deployment

### Production Build

1. **Build the WebAssembly frontend**
   ```bash
   GOOS=js GOARCH=wasm go build -o web/app.wasm main.go
   ```

2. **Build the server binary**
   ```bash
   go build -o golf-app main.go
   ```

3. **Set production environment variables**
   ```bash
   export MODE="production"
   export DB_ADMIN="production-db-password"
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
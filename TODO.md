# TODO Items

This file tracks outstanding TODO items and future improvements for the golf booking application.

## Active TODOs (From Code)

### High Priority

1. **Mobile Error Popup** (`web/app/pages/times.go:201`)
   - **File**: `/workspaces/golf_app/web/app/pages/times.go`
   - **Line**: 201
   - **Issue**: Build mobile popup to highlight booking errors
   - **Current State**: Errors may be silently ignored on mobile devices
   - **Impact**: Poor UX - users may not see booking failures
   - **Recommendation**: Implement modal/toast notification system for errors

2. **Add Existing User to Reservations** (`pkg/models/teetimes/reservations.go:48`)
   - **File**: `/workspaces/golf_app/pkg/models/teetimes/reservations.go`
   - **Line**: 48
   - **Issue**: Cannot add existing users as players to reservations
   - **Current State**: Only supports adding new guest players
   - **Impact**: Users cannot invite registered users to join their tee times
   - **Recommendation**: Implement user search and invitation system

## Code Quality Improvements

### Logging

- **Replace fmt.Println with structured logging**
  - Currently 30+ files use `fmt.Println` instead of proper logging
  - Files affected: Most files in `pkg/models/`, `pkg/handlers/`, `web/app/`
  - Recommended: Migrate to `log/slog` or `zerolog`
  - Benefits: Log levels, structured data, better production debugging

### Testing

- **Achieve 60% test coverage**
  - Current coverage: ~5%
  - Priority areas:
    - `pkg/models/teetimes/` (booking logic)
    - `pkg/handlers/` (API endpoints)
    - `pkg/models/auth/` (authentication)
  - Add test infrastructure (mocks, fixtures)
  - Set up CI/CD pipeline

### Architecture

- **Implement Repository Pattern**
  - Current: Models directly call database
  - Proposed: Add repository layer for data access
  - Benefits: Better testability, separation of concerns

- **Add Service Layer**
  - Current: Handlers contain business logic
  - Proposed: Extract to service layer
  - Benefits: Reusable business logic, cleaner handlers

- **Dependency Injection**
  - Current: Global `db.Instance` singleton
  - Proposed: Pass dependencies explicitly
  - Benefits: Better testing, looser coupling

## Feature Enhancements

### Authentication

- [ ] Add refresh token support
- [ ] Implement rate limiting on auth endpoints
- [ ] Add password strength validation
- [ ] Support password recovery via email
- [ ] Add two-factor authentication (2FA)

### Booking System

- [ ] Add waitlist functionality
- [ ] Support recurring bookings
- [ ] Implement booking notifications (email/SMS)
- [ ] Add booking confirmation emails
- [ ] Support group bookings with multiple players

### Admin Features

- [ ] Add analytics dashboard
- [ ] Implement revenue reporting
- [ ] Add customer management tools
- [ ] Support bulk operations on reservations
- [ ] Add audit logging for admin actions

### AI Assistant

- [ ] Add conversation history persistence
- [ ] Implement context-aware responses
- [ ] Add multi-language support
- [ ] Improve tool calling accuracy
- [ ] Add fallback responses for edge cases

## Documentation

### Missing Documentation

- [ ] API documentation (OpenAPI/Swagger spec)
- [ ] Database schema documentation
- [ ] Architecture decision records (ADRs)
- [ ] Deployment guide
- [ ] Troubleshooting guide
- [ ] Security best practices guide

### Code Documentation

- [ ] Complete function documentation (currently ~30% covered)
- [ ] Add more code examples in package docs
- [ ] Document error codes and meanings
- [ ] Add inline comments for complex logic

## Security

### High Priority

- [ ] Add rate limiting on all public endpoints
- [ ] Implement CSRF protection
- [ ] Add input validation middleware
- [ ] Audit SQL injection risks (Cypher queries)
- [ ] Add security headers (CSP, HSTS, etc.)

### Medium Priority

- [ ] Implement API key rotation
- [ ] Add request signing for sensitive operations
- [ ] Set up security scanning in CI/CD
- [ ] Add secrets management (Vault, AWS Secrets Manager)
- [ ] Implement audit logging for sensitive operations

## Performance

- [ ] Add database query optimization
- [ ] Implement caching layer (Redis)
- [ ] Add response compression
- [ ] Optimize WebAssembly bundle size
- [ ] Add database connection pooling monitoring
- [ ] Implement lazy loading for large datasets

## Infrastructure

- [ ] Set up CI/CD pipeline (GitHub Actions)
- [ ] Add health check endpoints
- [ ] Implement graceful shutdown
- [ ] Add metrics collection (Prometheus)
- [ ] Set up distributed tracing
- [ ] Add log aggregation (ELK stack)

## Mobile Experience

- [ ] Add offline support for PWA
- [ ] Implement push notifications
- [ ] Optimize for slow network connections
- [ ] Add app install prompts
- [ ] Improve touch interactions

## Completed Items ✓

- [x] Fix `birdsfoot` → `bigfoot` naming bug
- [x] Create `.env.example` file
- [x] Add environment variable validation
- [x] Fix JWT secret persistence
- [x] Extract magic numbers to constants
- [x] Update `.gitignore` for binary files
- [x] Add golangci-lint configuration
- [x] Add package-level documentation
- [x] Update README with correct structure
- [x] Remove commented code
- [x] Fix error handling gaps

## Notes

- Review and update this file regularly
- Create GitHub issues for items you're actively working on
- Mark items as completed with date when finished
- Add new TODOs as they're discovered

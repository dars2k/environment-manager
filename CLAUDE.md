# Environment Manager - Claude Development Guide

## Project Overview
This is a full-stack application environment manager with Go backend and React frontend. It provides environment management, real-time monitoring, SSH operations, and comprehensive audit logging.

## Architecture
- **Backend**: Go 1.23+ with clean architecture (Gorilla Mux, MongoDB, WebSocket)
- **Frontend**: React 18 + TypeScript + Redux Toolkit + Material-UI + Vite
- **Database**: MongoDB 7.0
- **Deployment**: Docker Compose

## Development Commands

### Quick Start
```bash
# Start all services
make up

# Start with logs
make up-logs

# Stop all services
make down
```

### Testing
```bash
# Run all tests (backend + frontend)
make test

# Backend tests only
cd backend && go test ./...

# Frontend tests only
cd frontend && npm test

# Frontend tests with coverage
cd frontend && npm run test:coverage
```

### Building
```bash
# Build all Docker images
make build

# Build specific service
make build-backend
make build-frontend

# Docker compose build (with output truncation)
docker compose build && docker compose build 2>&1 | tail -50
```

### Linting & Formatting
```bash
# Frontend linting
cd frontend && npm run lint

# Frontend formatting
cd frontend && npm run format
```

### Development Workflow
After completing tasks, follow the workflow in `.clinerules/workflow_rules.md`:
Each feature or fix request only if said that using the main branch it will use the main beach. new branch name and prefix like fix/ feature/
When done will open Pull Request in GitHub using github skills.

1. **Run tests**: `make test`
2. **Build Docker image**: `docker compose build` (show last 50 lines)
3. **Start application**: `docker compose up -d`
4. **Verify running**: `make health`
5. **Commit changes**: Use commit messages starting with: `added`, `removed`, `fix`, `changed`

### Service Management
```bash
# Check service health
make health

# View logs
make logs
make logs-backend
make logs-frontend
make logs-db

# Restart services
make restart-backend
make restart-frontend

# Access shells
make backend-shell
make frontend-shell
make db-shell
```

### Development URLs
- Frontend: http://localhost (port 80)
- Backend API: http://localhost:8080
- MongoDB Express: http://localhost:8081 (debug profile)
- Default credentials: admin / admin123

## Key Files & Directories

### Backend Structure
- `backend/cmd/server/main.go` - Application entry point
- `backend/internal/api/` - HTTP handlers and routes
- `backend/internal/domain/` - Business logic and entities
- `backend/internal/repository/` - Data access layer
- `backend/internal/service/` - Application services
- `backend/internal/websocket/` - WebSocket handlers
- `backend/config/config.yaml` - Configuration

### Frontend Structure
- `frontend/src/api/` - API client services
- `frontend/src/components/` - Reusable UI components
- `frontend/src/pages/` - Page components
- `frontend/src/store/` - Redux store and slices
- `frontend/src/types/` - TypeScript types

### Configuration
- `docker-compose.yml` - Main Docker configuration
- `Makefile` - Development commands
- `.env` - Environment variables (copy from `.env.example`)

## Testing Strategy
The project includes comprehensive testing:
- Backend: Go unit tests with `go test ./...`
- Frontend: Vitest with React Testing Library
- Integration: Docker Compose test environment

## Security Notes
- JWT authentication with secure key management
- SSH key encryption for secure credential storage
- Input validation and sanitization
- CORS configuration for allowed origins
- MongoDB authentication and authorization
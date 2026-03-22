# Environment Manager

A modern, full-stack application for managing deployment environments with real-time monitoring, SSH-based operations, role-based access control, and comprehensive audit logging.

[![CI/CD Pipeline](https://github.com/dars2k/environment-manager/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/dars2k/environment-manager/actions/workflows/ci-cd.yml)
![Backend Coverage](https://img.shields.io/badge/Backend%20Coverage-89%25-yellow)
![Frontend Coverage](https://img.shields.io/badge/Frontend%20Coverage-92%25-brightgreen)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)
![MongoDB](https://img.shields.io/badge/MongoDB-7.0-47A248?logo=mongodb)
![License](https://img.shields.io/badge/License-MIT-blue)

## Features

- **Environment Management** — Create, update, and delete application environments
- **Real-time Monitoring** — Live health status via WebSocket connections
- **SSH & HTTP Operations** — Execute restart and upgrade commands over SSH or HTTP
- **Role-based Access Control** — Admin, user, and viewer roles with enforced permissions
- **Audit Logging** — Full audit trail for all actions with user attribution
- **Automated Health Checks** — Configurable periodic checks with status code or JSON regex validation
- **Version Management** — Fetch available versions and upgrade environments
- **Dark Theme UI** — Responsive interface built with Material-UI

## Architecture

| Layer | Technology |
|-------|-----------|
| Frontend | React 18 + TypeScript + Redux Toolkit + Material-UI + Vite |
| Backend | Go 1.26 + Gorilla Mux (Clean Architecture) |
| Database | MongoDB 7.0 |
| Real-time | WebSocket (Gorilla WebSocket) |
| Deployment | Docker Compose |

## Project Structure

```
environment-manager/
├── frontend/                 # React frontend
│   └── src/
│       ├── api/              # API client services
│       ├── components/       # Reusable UI components
│       ├── contexts/         # React contexts (WebSocket)
│       ├── hooks/            # Custom React hooks
│       ├── layouts/          # Page layouts
│       ├── pages/            # Page components
│       ├── routes/           # Route definitions
│       ├── store/            # Redux store and slices
│       └── theme/            # MUI theme configuration
├── backend/                  # Go backend
│   ├── cmd/server/           # Application entry point
│   └── internal/
│       ├── api/              # HTTP handlers, middleware, routes
│       ├── domain/           # Business entities and errors
│       ├── repository/       # Data access layer
│       ├── service/          # Application services
│       └── websocket/        # WebSocket hub
├── docker/                   # Docker configuration notes
├── docs/                     # Documentation
│   ├── api/                  # API reference
│   ├── architecture/         # Architecture docs
│   └── deployment/           # Deployment guides
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose

### 1. Clone the repository

```bash
git clone https://github.com/dars2k/environment-manager.git
cd environment-manager
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and set secure values for `JWT_SECRET` and `SSH_KEY_ENCRYPTION_KEY` (see [Configuration](#configuration)).

### 3. Start the application

```bash
make up
# or
docker-compose up -d
```

### 4. Access the application

| Service | URL |
|---------|-----|
| Frontend | http://localhost |
| Backend API | http://localhost:8080 |
| MongoDB Express (debug) | http://localhost:8081 |

> **Admin credentials**: The admin password is auto-generated on first startup and printed to the backend container logs. Run `make logs-backend` to retrieve it.

## Manual Setup (without Docker)

**Prerequisites:** Go 1.26+, Node.js 22+, MongoDB 7.0+

```bash
# Start MongoDB
docker run -d -p 27017:27017 --name env-manager-mongo \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin123 \
  mongo:7.0

# Run backend
cd backend
go mod download
go run cmd/server/main.go

# Run frontend (separate terminal)
cd frontend
npm install
npm run dev
```

## Configuration

Copy `.env.example` to `.env` and update the values:

```env
# MongoDB
MONGO_ROOT_USER=admin
MONGO_ROOT_PASSWORD=admin123
MONGO_DATABASE=app-env-manager
MONGO_PORT=27017
MONGODB_URI=mongodb://admin:admin123@mongodb:27017/app-env-manager?authSource=admin
MONGODB_DATABASE=app-env-manager

# Security — generate with: openssl rand -hex 32
JWT_SECRET=CHANGE_ME_USE_openssl_rand_hex_32
JWT_EXPIRY=24h
# Must be exactly 32 bytes — generate with: openssl rand -hex 16
SSH_KEY_ENCRYPTION_KEY=CHANGE_ME_32BYTES_USE_openssl_rand

# Frontend
FRONTEND_PORT=80
FRONTEND_HTTPS_PORT=443
ALLOWED_ORIGINS=http://localhost,https://localhost
VITE_API_URL=https://localhost/api
VITE_WS_URL=wss://localhost/ws

# Application
LOG_LEVEL=info
SESSION_TIMEOUT=24h
HEALTH_CHECK_INTERVAL=60s
HEALTH_CHECK_TIMEOUT=10s
SSH_CONNECTION_TIMEOUT=30s
SSH_COMMAND_TIMEOUT=300s

# MongoDB Express (optional debug UI)
ME_USERNAME=admin
ME_PASSWORD=admin123
ME_PORT=8081
```

## API Reference

All endpoints are prefixed with `/api/v1`. JWT token required in `Authorization: Bearer <token>` header for all endpoints except login and health check.

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/auth/login` | Login, receive JWT token |
| `GET` | `/auth/me` | Get current user info |
| `POST` | `/auth/logout` | Logout |

### Environments (admin: create/update/delete; all roles: read/restart/upgrade)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/environments` | List all environments |
| `GET` | `/environments/:id` | Get environment details |
| `POST` | `/environments` | Create environment *(admin only)* |
| `PUT` | `/environments/:id` | Update environment *(admin only)* |
| `DELETE` | `/environments/:id` | Delete environment *(admin only)* |
| `POST` | `/environments/:id/restart` | Restart environment |
| `POST` | `/environments/:id/check-health` | Trigger health check |
| `GET` | `/environments/:id/versions` | List available versions |
| `POST` | `/environments/:id/upgrade` | Upgrade environment version |
| `GET` | `/environments/:id/logs` | Get environment logs |

### Users *(admin only)*

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/users` | List all users |
| `GET` | `/users/:id` | Get user details |
| `POST` | `/users` | Create user |
| `PUT` | `/users/:id` | Update user (role, active status) |
| `DELETE` | `/users/:id` | Delete user |
| `POST` | `/users/password/change` | Change own password |
| `POST` | `/users/:id/password/reset` | Reset user password |

### Logs & Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/logs` | List all logs |
| `GET` | `/logs/count` | Log counts by type/level |
| `GET` | `/logs/:id` | Get log entry |
| `GET` | `/health` | API health check |
| `WS` | `/ws` | WebSocket real-time updates |

See [API documentation](./docs/api/README.md) for full request/response schemas.

## Development

### Available Make commands

```bash
make build           # Build Docker images
make up              # Start all services (detached)
make up-logs         # Start with logs streaming
make down            # Stop all services
make logs            # Stream all service logs
make logs-backend    # Stream backend logs
make logs-frontend   # Stream frontend logs
make logs-db         # Stream database logs
make health          # Check service health
make test            # Run all tests (backend + frontend)
make clean           # Remove containers, volumes, and images
make backend-shell   # Open shell in backend container
make frontend-shell  # Open shell in frontend container
make db-shell        # Open MongoDB shell
make mongo-express   # Start MongoDB Express UI (debug)
make restart-backend # Restart backend service
make restart-frontend# Restart frontend service
make init            # Initialize .env from .env.example
```

### Frontend development

```bash
cd frontend
npm run dev           # Start development server
npm run build         # Build for production
npm run test          # Run tests
npm run test:coverage # Run tests with coverage report
npm run lint          # Run ESLint
npm run format        # Format with Prettier
```

### Backend development

```bash
cd backend
go run cmd/server/main.go  # Run development server
go test ./...              # Run all tests
go test -cover ./...       # Run with coverage
```

## Testing

See [TESTING.md](./TESTING.md) for the full testing guide.

```bash
# Run all tests
make test

# Backend only
cd backend && go test ./...

# Frontend only
cd frontend && npm test
```

## Role-based Access Control

| Permission | Admin | User | Viewer |
|-----------|-------|------|--------|
| View environments | ✓ | ✓ | ✓ |
| Restart / Upgrade | ✓ | ✓ | ✓ |
| Create / Edit / Delete environments | ✓ | | |
| View logs | ✓ | ✓ | ✓ |
| Manage users | ✓ | | |

## Documentation

- [Architecture Overview](./docs/architecture/overview.md)
- [Backend Design](./docs/architecture/backend-design.md)
- [Frontend Design](./docs/architecture/frontend-design.md)
- [Data Models](./docs/architecture/data-models.md)
- [UI/UX Design](./docs/architecture/ui-ux-design.md)
- [API Reference](./docs/api/README.md)
- [Deployment Guide](./docs/deployment/README.md)
- [Testing Guide](./TESTING.md)
- [Docker Setup](./docker/README.md)

## Contributing

1. Fork the repository
2. Create a branch: `feature/your-feature` or `fix/your-fix`
3. Run tests: `make test`
4. Build and verify: `docker compose build && docker compose up -d`
5. Open a pull request

Please ensure all tests pass and the Docker build succeeds before submitting a PR.

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.

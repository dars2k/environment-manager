# Architecture Overview

## Design Principles

1. **Modularity** — Clear separation between components with well-defined interfaces
2. **Scalability** — Stateless API servers ready for horizontal scaling
3. **Maintainability** — Clean Architecture with dependency injection throughout
4. **Security** — JWT authentication, RBAC enforcement, encrypted SSH credential storage
5. **Observability** — Comprehensive audit logging and structured log output
6. **Flexibility** — SSH and HTTP command types for diverse environment configurations

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Frontend (React + Redux)                    │
│  ┌──────────┐  ┌────────────┐  ┌──────────┐  ┌─────────┐  │
│  │Dashboard │  │Environment │  │  Logs    │  │  Users  │  │
│  │          │  │  Details   │  │  View    │  │ (admin) │  │
│  └──────────┘  └────────────┘  └──────────┘  └─────────┘  │
└───────────────────────┬─────────────────────────────────────┘
                        │ HTTP / WebSocket
┌───────────────────────┴─────────────────────────────────────┐
│                  API Gateway (Go / Gorilla Mux)              │
│  ┌────────────┐  ┌────────────┐  ┌──────────────────────┐  │
│  │  REST API  │  │ WebSocket  │  │ Auth + RBAC          │  │
│  │  Handlers  │  │   Server   │  │ Middleware            │  │
│  └────────────┘  └────────────┘  └──────────────────────┘  │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────────────────┐
│                  Business Logic Layer                        │
│  ┌────────────┐  ┌────────────┐  ┌───────────┐             │
│  │Environment │  │  Health    │  │   Auth    │             │
│  │  Service   │  │  Checker   │  │  Service  │             │
│  └────────────┘  └────────────┘  └───────────┘             │
│  ┌────────────┐  ┌────────────┐  ┌───────────┐             │
│  │    SSH     │  │    Log     │  │   User    │             │
│  │  Manager   │  │  Service   │  │  Service  │             │
│  └────────────┘  └────────────┘  └───────────┘             │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────────────────┐
│                   Data Access Layer                          │
│  ┌────────────┐  ┌────────────┐  ┌───────────┐             │
│  │Environment │  │    Log     │  │   User    │             │
│  │ Repository │  │ Repository │  │Repository │             │
│  └────────────┘  └────────────┘  └───────────┘             │
└───────────────────────┬─────────────────────────────────────┘
                        │
                  ┌─────┴─────┐
                  │  MongoDB  │
                  └───────────┘
```

## Component Details

### Frontend

| Component | Responsibility |
|-----------|---------------|
| Dashboard | Real-time environment status grid, WebSocket-driven updates |
| Environment Details | Full environment info, logs, health status, action buttons |
| Environment Form | Create/edit with SSH/HTTP command config and health check setup |
| Users (admin only) | CRUD for users, role assignment, active/disable toggle |
| Redux Store | Environment, logs, UI, and notification slices |
| WebSocket Context | Persistent connection, auto-reconnect, status event dispatch |

### Backend Services

| Service | Responsibility |
|---------|---------------|
| Environment Service | CRUD, SSH/HTTP command execution, version management, event emission |
| Health Checker | Periodic checks, HTTP endpoint validation, JSON regex validation, status notifications |
| SSH Manager | Secure SSH connection management, command execution, key-based auth |
| Auth Service | JWT generation and validation, bcrypt password hashing |
| User Service | CRUD, role management, password operations |
| Log Service | Audit log creation, filtering, pagination, user attribution |
| WebSocket Hub | Client connection management, real-time event broadcasting |

## Data Flow

```
# Authentication
Login → Auth Handler → Auth Service → JWT Generation → Client

# Real-time status
Health Check → Status Change → Log Creation → WebSocket Broadcast → UI Update

# Command execution
UI Action → API Request → RBAC Check → Command Service → SSH/HTTP → Log → Response

# CRUD (admin only)
UI Form → API Request → JWT Validation → RequireAdmin → Service → Repository → MongoDB
```

## Security Architecture

### Authentication and Authorization

- JWT-based authentication with configurable expiry
- `RequireAdmin` middleware enforces admin-only routes
- Role extracted from JWT claims and stored in request context via `ctxutil`
- Three roles: `admin` (full access), `user` (read + operations), `viewer` (read only)

### Protected Routes

| Route | Minimum role |
|-------|-------------|
| `GET /environments`, `GET /logs` | viewer |
| `POST /environments/:id/restart`, `/upgrade` | user |
| `POST /environments`, `PUT`, `DELETE` | admin |
| All `/users` routes | admin |

### Credential Management

- SSH keys are stored encrypted (AES-256) using `SSH_KEY_ENCRYPTION_KEY`
- Passwords hashed with bcrypt
- No secrets returned in API responses

### API Security

- CORS configuration via `ALLOWED_ORIGINS`
- Request validation on all mutation endpoints
- Structured error responses (no stack traces in production)
- Request logging middleware for audit trail

## Scalability

- **Stateless API** — no server-side session state, scales horizontally behind a load balancer
- **MongoDB** — connection pooling, indexed queries, supports replica sets
- **WebSocket** — hub-based broadcast architecture, can be extended with pub/sub for multi-instance
- **Health checks** — goroutine-pool-based concurrent checking

## Technology Stack

### Backend

| Technology | Version | Purpose |
|-----------|---------|---------|
| Go | 1.26 | Language |
| Gorilla Mux | 1.8.1 | HTTP router |
| Gorilla WebSocket | 1.5.3 | WebSocket server |
| MongoDB driver | 1.17 | Database |
| Logrus | latest | Structured logging |
| golang.org/x/crypto | latest | SSH + bcrypt |
| golang-jwt/jwt | v5 | JWT tokens |

### Frontend

| Technology | Version | Purpose |
|-----------|---------|---------|
| React | 18 | UI framework |
| TypeScript | 5 | Type safety |
| Redux Toolkit | 1.9 | State management |
| Material-UI | v5 | UI component library |
| React Router | v6 | Client-side routing |
| Axios | 1.13 | HTTP client |
| Vite | 7 | Build tool |

### Infrastructure

| Technology | Purpose |
|-----------|---------|
| Docker Compose | Local orchestration |
| Nginx | Reverse proxy and static file serving |
| MongoDB 8.2 | Primary database |

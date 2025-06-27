# Architecture Overview

## System Design Principles

The Application Environment Manager is designed with the following principles:

1. **Modularity**: Clear separation between components with well-defined interfaces
2. **Scalability**: Horizontal scaling capabilities for both frontend and backend
3. **Maintainability**: Clean code architecture with dependency injection
4. **Security**: JWT authentication, secure credential storage and SSH key management
5. **Observability**: Comprehensive logging and monitoring capabilities
6. **Flexibility**: Support for both SSH and HTTP-based environment management

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Frontend (React + Redux)                 │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐   │
│  │  Dashboard  │  │ Environment  │  │    Users &      │   │
│  │    View     │  │   Details    │  │   Settings      │   │
│  └─────────────┘  └──────────────┘  └─────────────────┘   │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐   │
│  │    Logs     │  │ Create/Edit  │  │     Login       │   │
│  │    View     │  │ Environment  │  │                 │   │
│  └─────────────┘  └──────────────┘  └─────────────────┘   │
└─────────────────────────┬───────────────────────────────────┘
                          │ HTTP/WebSocket
┌─────────────────────────┴───────────────────────────────────┐
│                    API Gateway (Go/Mux)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐   │
│  │ REST API     │  │  WebSocket   │  │ Authentication │   │
│  │ Handlers     │  │   Server     │  │   Middleware   │   │
│  └──────────────┘  └──────────────┘  └────────────────┘   │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────┴───────────────────────────────────┐
│                   Business Logic Layer                       │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐   │
│  │ Environment  │  │   Health     │  │     Auth       │   │
│  │  Service     │  │   Checker    │  │   Service      │   │
│  └──────────────┘  └──────────────┘  └────────────────┘   │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐   │
│  │     SSH      │  │     Log      │  │     User       │   │
│  │   Manager    │  │   Service    │  │   Service      │   │
│  └──────────────┘  └──────────────┘  └────────────────┘   │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────┴───────────────────────────────────┐
│                    Data Access Layer                         │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐   │
│  │ Environment  │  │     Log      │  │     User       │   │
│  │ Repository   │  │  Repository  │  │  Repository    │   │
│  └──────────────┘  └──────────────┘  └────────────────┘   │
└─────────────────────────┬───────────────────────────────────┘
                          │
                    ┌─────┴─────┐
                    │  MongoDB  │
                    └───────────┘
```

## Component Details

### Frontend Components

1. **Dashboard Component**
   - Real-time environment status display
   - WebSocket connection for live updates
   - Environment cards with quick actions
   - Health status indicators

2. **Environment Details Component**
   - Comprehensive environment information
   - Environment logs viewer
   - Health check status and history
   - Action buttons (restart, upgrade, check health)

3. **Environment Form Component**
   - Dynamic form validation
   - Command configuration (SSH/HTTP)
   - Health check configuration
   - Upgrade configuration

4. **User Management Components**
   - User list with CRUD operations
   - Password management
   - Role-based access control

5. **Redux State Management**
   - Environment slice for environment data
   - Logs slice for log management
   - UI slice for UI state
   - Notification slice for alerts

### Backend Services

1. **Environment Service**
   - CRUD operations for environments
   - Command execution (SSH/HTTP)
   - Version management
   - Event emission for audit logging

2. **Health Checker Service**
   - Periodic health checks
   - HTTP endpoint validation
   - JSON response validation
   - Status change notifications

3. **SSH Manager**
   - Secure SSH connection management
   - Command execution
   - Error handling and retries
   - Key-based authentication

4. **Auth Service**
   - JWT token generation
   - User authentication
   - Password hashing (bcrypt)
   - Session management

5. **User Service**
   - User CRUD operations
   - Password management
   - Role validation

6. **Log Service**
   - Audit log creation
   - Log filtering and pagination
   - Log aggregation

7. **WebSocket Hub**
   - Client connection management
   - Real-time event broadcasting
   - Connection state tracking

## Data Flow

### Authentication Flow
```
Login Request → Auth Handler → Auth Service → JWT Generation → Client Storage
```

### Real-time Updates
```
Health Check → Status Change → Log Creation → WebSocket Broadcast → UI Update
```

### Command Execution
```
UI Action → API Request → Validation → Command Service → SSH/HTTP Execution → Log → Response
```

### CRUD Operations
```
UI Form → API Request → JWT Validation → Service Layer → Repository → MongoDB → Response
```

## Security Architecture

1. **Authentication & Authorization**
   - JWT-based authentication
   - Middleware for protected routes
   - Role-based access control
   - Token expiration

2. **Credential Management**
   - Encrypted SSH keys storage
   - Separate credential references
   - Password hashing with bcrypt

3. **API Security**
   - CORS configuration
   - Request validation
   - Error handling middleware
   - Logging middleware

4. **Command Security**
   - Command type validation
   - SSH connection security
   - HTTP request validation

## Data Models

### Core Entities

1. **Environment**
   - Basic info (name, description, URL)
   - Target configuration
   - Credentials reference
   - Health check config
   - Command configuration
   - Upgrade configuration
   - Status and timestamps

2. **User**
   - Authentication credentials
   - Role and permissions
   - Timestamps

3. **Log**
   - Environment and user references
   - Log type and level
   - Action type
   - Detailed message and metadata

## Scalability Strategy

1. **Horizontal Scaling**
   - Stateless API servers
   - MongoDB replica sets
   - WebSocket clustering support
   - Load balancer ready

2. **Performance Optimization**
   - Connection pooling (MongoDB)
   - Efficient query patterns
   - Pagination for large datasets
   - Caching strategy (future)

3. **Monitoring & Observability**
   - Structured logging with Logrus
   - Health check endpoints
   - Metrics collection ready
   - Error tracking

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Web Framework**: Gorilla Mux
- **Database**: MongoDB
- **Authentication**: JWT
- **WebSocket**: Gorilla WebSocket
- **Logging**: Logrus
- **SSH**: golang.org/x/crypto/ssh

### Frontend
- **Framework**: React 18 with TypeScript
- **State Management**: Redux Toolkit
- **UI Library**: Material-UI (MUI)
- **Routing**: React Router v6
- **HTTP Client**: Axios
- **WebSocket**: Native WebSocket API
- **Build Tool**: Vite

### Infrastructure
- **Containerization**: Docker
- **Orchestration**: Docker Compose
- **Reverse Proxy**: Nginx
- **Database**: MongoDB 7.0

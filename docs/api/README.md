# API Documentation

## Overview

The Application Environment Manager API is a RESTful API built with Go that provides endpoints for managing environments, executing operations, and streaming real-time updates.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All API endpoints (except `/auth/login` and `/health`) require JWT authentication.

```http
Authorization: Bearer <jwt-token>
```

## Response Format

All responses follow a consistent format:

### Success Response
```json
{
  "success": true,
  "data": { ... },
  "metadata": {
    "timestamp": "2025-06-11T18:00:00Z",
    "version": "1.0.0"
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ENV_NOT_FOUND",
    "message": "Environment not found",
    "details": { ... }
  },
  "metadata": {
    "timestamp": "2025-06-11T18:00:00Z",
    "version": "1.0.0"
  }
}
```

## Endpoints

### Health Check

#### GET /health
Check the API health status.

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

### Authentication

#### POST /auth/login
Login with credentials to receive JWT token.

**Request:**
```json
{
  "username": "admin",
  "password": "password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "user-123",
    "username": "admin",
    "email": "admin@example.com",
    "role": "admin",
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-06-11T12:00:00Z"
  },
  "expiresAt": "2025-06-12T18:00:00Z"
}
```

#### GET /auth/me
Get current authenticated user information.

**Response:**
```json
{
  "id": "user-123",
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin",
  "createdAt": "2025-01-01T00:00:00Z",
  "updatedAt": "2025-06-11T12:00:00Z"
}
```

#### POST /auth/logout
Logout the current user.

**Response:**
```json
{
  "message": "Logged out successfully"
}
```

### Users

#### GET /users
List all users.

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)

**Response:**
```json
{
  "users": [
    {
      "id": "user-123",
      "username": "admin",
      "email": "admin@example.com",
      "role": "admin",
      "createdAt": "2025-01-01T00:00:00Z",
      "updatedAt": "2025-06-11T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "totalPages": 3
  }
}
```

#### POST /users
Create a new user.

**Request:**
```json
{
  "username": "newuser",
  "password": "securepassword",
  "role": "user"
}
```

#### GET /users/:id
Get a specific user.

#### PUT /users/:id
Update a user.

**Request:**
```json
{
  "username": "updateduser",
  "email": "updated@example.com",
  "role": "admin"
}
```

#### DELETE /users/:id
Delete a user.

#### POST /users/password/change
Change current user's password.

**Request:**
```json
{
  "currentPassword": "oldpassword",
  "newPassword": "newpassword"
}
```

#### POST /users/:id/password/reset
Reset a user's password (admin only).

**Request:**
```json
{
  "newPassword": "resetpassword"
}
```

### Environments

#### GET /environments
List all environments with their current status.

**Query Parameters:**
- `status` (optional): Filter by health status (healthy, unhealthy, unknown)
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20)

**Response:**
```json
{
  "environments": [
    {
      "id": "env-123",
      "name": "production-api",
      "description": "Production API server",
      "environmentURL": "https://api.example.com",
      "target": {
        "host": "192.168.1.100",
        "port": 22,
        "domain": "api.example.com"
      },
      "credentials": {
        "type": "key",
        "username": "deploy",
        "keyId": "key-456"
      },
      "healthCheck": {
        "enabled": true,
        "endpoint": "/health",
        "method": "GET",
        "interval": 30,
        "timeout": 5,
        "validation": {
          "type": "statusCode",
          "value": 200
        },
        "headers": {
          "X-Health-Token": "secret-token"
        }
      },
      "status": {
        "health": "healthy",
        "lastCheck": "2025-06-11T17:55:00Z",
        "message": "All systems operational",
        "responseTime": 145
      },
      "systemInfo": {
        "osVersion": "Ubuntu 22.04 LTS",
        "appVersion": "2.1.0",
        "lastUpdated": "2025-06-11T12:00:00Z"
      },
      "timestamps": {
        "createdAt": "2025-01-01T00:00:00Z",
        "updatedAt": "2025-06-11T12:00:00Z",
        "lastRestartAt": "2025-06-10T15:30:00Z",
        "lastUpgradeAt": "2025-06-05T10:00:00Z",
        "lastHealthyAt": "2025-06-11T17:55:00Z"
      },
      "commands": {
        "type": "ssh",
        "restart": {
          "command": "sudo systemctl restart myapp"
        }
      },
      "upgradeConfig": {
        "enabled": true,
        "type": "http",
        "versionListURL": "https://api.example.com/versions",
        "versionListMethod": "GET",
        "jsonPathResponse": "$.versions[*].version",
        "upgradeCommand": {
          "url": "https://api.example.com/upgrade",
          "method": "POST",
          "headers": {
            "Authorization": "Bearer upgrade-token"
          }
        }
      },
      "metadata": {
        "team": "backend",
        "tier": "production"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "totalPages": 3
  }
}
```

#### GET /environments/:id
Get detailed information about a specific environment.

#### POST /environments
Create a new environment.

**Request:**
```json
{
  "name": "staging-api",
  "description": "Staging API server",
  "environmentURL": "https://staging-api.example.com",
  "target": {
    "host": "192.168.1.101",
    "port": 22,
    "domain": "staging-api.example.com"
  },
  "credentials": {
    "type": "key",
    "username": "deploy",
    "keyId": "key-789"
  },
  "healthCheck": {
    "enabled": true,
    "endpoint": "/health",
    "method": "GET",
    "interval": 60,
    "timeout": 10,
    "validation": {
      "type": "jsonRegex",
      "value": "\"status\":\\s*\"ok\""
    },
    "headers": {
      "X-Health-Token": "secret-token"
    }
  },
  "commands": {
    "type": "ssh",
    "restart": {
      "command": "sudo systemctl restart myapp"
    }
  },
  "upgradeConfig": {
    "enabled": true,
    "type": "ssh",
    "versionListURL": "https://api.example.com/versions",
    "versionListMethod": "GET",
    "jsonPathResponse": "$.versions[*].version",
    "upgradeCommand": {
      "command": "sudo /opt/myapp/upgrade.sh {{version}}"
    }
  },
  "metadata": {
    "team": "backend",
    "tier": "staging"
  }
}
```

#### PUT /environments/:id
Update an existing environment.

**Request:** Same structure as POST /environments

#### DELETE /environments/:id
Delete an environment.

**Response:**
```json
{
  "message": "Environment deleted successfully"
}
```

### Environment Operations

#### POST /environments/:id/restart
Restart an environment.

**Request:**
```json
{
  "force": false,
  "gracefulTimeout": 30
}
```

**Response:**
```json
{
  "operationId": "op-456",
  "status": "in_progress",
  "startedAt": "2025-06-11T18:00:00Z"
}
```

#### POST /environments/:id/check-health
Manually trigger a health check for an environment.

**Response:**
```json
{
  "health": "healthy",
  "message": "Health check passed",
  "responseTime": 145,
  "details": {
    "statusCode": 200,
    "response": "{\"status\":\"ok\"}"
  }
}
```

#### GET /environments/:id/versions
Get available versions for upgrade.

**Response:**
```json
{
  "currentVersion": "2.1.0",
  "availableVersions": ["2.2.0", "2.1.1", "2.1.0", "2.0.5"]
}
```

#### POST /environments/:id/upgrade
Upgrade an environment to a new version.

**Request:**
```json
{
  "version": "2.2.0",
  "backupFirst": true,
  "rollbackOnFailure": true
}
```

**Response:**
```json
{
  "operationId": "op-789",
  "status": "in_progress",
  "startedAt": "2025-06-11T18:00:00Z"
}
```

#### GET /environments/:id/logs
Get logs for a specific environment.

**Query Parameters:**
- `type` (optional): Filter by log type (health_check, action, system, error, auth)
- `level` (optional): Filter by log level (info, warning, error, success)
- `action` (optional): Filter by action type (create, update, delete, restart, upgrade, etc.)
- `startDate` (optional): Start date (ISO 8601)
- `endDate` (optional): End date (ISO 8601)
- `page` (optional): Page number
- `limit` (optional): Items per page

**Response:**
```json
{
  "logs": [
    {
      "id": "log-789",
      "timestamp": "2025-06-11T17:30:00Z",
      "environmentId": "env-123",
      "environmentName": "production-api",
      "userId": "user-123",
      "username": "admin",
      "type": "action",
      "level": "info",
      "action": "restart",
      "message": "Environment restarted successfully",
      "details": {
        "duration": 45000,
        "reason": "Apply configuration changes"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "totalPages": 8
  }
}
```

### Logs

#### GET /logs
Get all system logs.

**Query Parameters:**
- `type` (optional): Filter by log type (health_check, action, system, error, auth)
- `level` (optional): Filter by log level (info, warning, error, success)
- `environmentId` (optional): Filter by environment ID
- `userId` (optional): Filter by user ID
- `startDate` (optional): Start date (ISO 8601)
- `endDate` (optional): End date (ISO 8601)
- `page` (optional): Page number
- `limit` (optional): Items per page

#### GET /logs/count
Get log counts by type and level.

**Response:**
```json
{
  "total": 1500,
  "byType": {
    "health_check": 800,
    "action": 500,
    "system": 150,
    "error": 30,
    "auth": 20
  },
  "byLevel": {
    "info": 1200,
    "warning": 200,
    "error": 50,
    "success": 50
  }
}
```

#### GET /logs/:id
Get a specific log entry.

### WebSocket Endpoints

#### WS /ws
WebSocket connection for real-time updates.

**Connection:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
```

**Message Types:**

1. **Subscribe to environment updates:**
```json
{
  "type": "subscribe",
  "payload": {
    "environments": ["env-123", "env-456"]
  }
}
```

2. **Receive status updates:**
```json
{
  "type": "status_update",
  "payload": {
    "environmentId": "env-123",
    "status": {
      "health": "unhealthy",
      "message": "Connection timeout",
      "timestamp": "2025-06-11T18:00:00Z"
    }
  }
}
```

3. **Receive operation updates:**
```json
{
  "type": "operation_update",
  "payload": {
    "operationId": "op-456",
    "status": "completed",
    "output": "Service restarted successfully"
  }
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `AUTH_INVALID` | Invalid authentication credentials |
| `AUTH_EXPIRED` | Authentication token expired |
| `AUTH_UNAUTHORIZED` | Unauthorized access |
| `ENV_NOT_FOUND` | Environment not found |
| `ENV_DUPLICATE` | Environment name already exists |
| `USER_NOT_FOUND` | User not found |
| `USER_DUPLICATE` | Username or email already exists |
| `VALIDATION_ERROR` | Request validation failed |
| `SSH_CONNECTION_FAILED` | SSH connection failed |
| `HEALTH_CHECK_FAILED` | Health check failed |
| `OPERATION_FAILED` | Operation execution failed |
| `INTERNAL_ERROR` | Internal server error |

## Environment Command Configuration

The environment manager supports two types of command execution:

### SSH Commands
For environments that support SSH access:
```json
{
  "commands": {
    "type": "ssh",
    "restart": {
      "command": "sudo systemctl restart myapp"
    }
  }
}
```

### HTTP Commands
For environments that expose HTTP endpoints for management:
```json
{
  "commands": {
    "type": "http",
    "restart": {
      "url": "https://api.example.com/admin/restart",
      "method": "POST",
      "headers": {
        "Authorization": "Bearer admin-token"
      },
      "body": {
        "graceful": true
      }
    }
  }
}
```

## Health Check Configuration

Health checks support two validation types:

### Status Code Validation
```json
{
  "validation": {
    "type": "statusCode",
    "value": 200
  }
}
```

### JSON Regex Validation
```json
{
  "validation": {
    "type": "jsonRegex",
    "value": "\"status\":\\s*\"(ok|healthy)\""
  }
}
```

## CORS Configuration

The API supports Cross-Origin Resource Sharing (CORS) with the following settings:
- Allowed Methods: GET, POST, PUT, DELETE, OPTIONS
- Allowed Headers: All headers (*)
- Credentials: Allowed
- Max Age: 86400 seconds (24 hours)

The allowed origins are configured in the server configuration.

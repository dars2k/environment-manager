# API Reference

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All endpoints except `POST /auth/login` and `GET /health` require a JWT token:

```http
Authorization: Bearer <token>
```

## Response Format

### Success

```json
{
  "data": { ... },
  "metadata": {
    "timestamp": "2026-03-20T12:00:00Z"
  }
}
```

### Error

```json
{
  "error": {
    "code": "ENV_NOT_FOUND",
    "message": "Environment not found"
  }
}
```

---

## Health Check

### `GET /health`

```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

---

## Authentication

### `POST /auth/login`

**Request:**
```json
{
  "username": "admin",
  "password": "your-password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "username": "admin",
    "role": "admin",
    "active": true,
    "createdAt": "2026-01-01T00:00:00Z",
    "updatedAt": "2026-03-20T12:00:00Z"
  },
  "expiresAt": "2026-03-21T12:00:00Z"
}
```

### `GET /auth/me`

**Response:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "username": "admin",
  "role": "admin",
  "active": true,
  "createdAt": "2026-01-01T00:00:00Z",
  "updatedAt": "2026-03-20T12:00:00Z"
}
```

### `POST /auth/logout`

**Response:**
```json
{ "message": "Logged out successfully" }
```

---

## Users *(admin only)*

### `GET /users`

**Query parameters:** `page` (default: 1), `limit` (default: 20)

**Response:**
```json
{
  "users": [
    {
      "id": "507f1f77bcf86cd799439011",
      "username": "admin",
      "role": "admin",
      "active": true,
      "lastLoginAt": "2026-03-20T11:00:00Z",
      "createdAt": "2026-01-01T00:00:00Z",
      "updatedAt": "2026-03-20T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 5,
    "totalPages": 1
  }
}
```

### `GET /users/:id`

Returns a single user object (same schema as above).

### `POST /users`

**Request:**
```json
{
  "username": "newuser",
  "password": "Securepassword1!",
  "role": "user"
}
```

Password requirements: minimum 12 characters, must include uppercase, lowercase, and a digit.

### `PUT /users/:id`

**Request:**
```json
{
  "role": "viewer",
  "active": false
}
```

### `DELETE /users/:id`

**Response:**
```json
{ "message": "User deleted successfully" }
```

### `POST /users/password/change`

Change the currently authenticated user's password.

**Request:**
```json
{
  "currentPassword": "OldPassword1!",
  "newPassword": "NewPassword1!"
}
```

### `POST /users/:id/password/reset` *(admin only)*

**Request:**
```json
{
  "newPassword": "ResetPassword1!"
}
```

---

## Environments

### `GET /environments`

**Query parameters:**
- `status`: `healthy` | `unhealthy` | `unknown`
- `page`, `limit`

**Response:**
```json
{
  "environments": [
    {
      "id": "507f1f77bcf86cd799439012",
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
        "keyId": "507f1f77bcf86cd799439013"
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
        }
      },
      "status": {
        "health": "healthy",
        "lastCheck": "2026-03-20T11:55:00Z",
        "message": "All systems operational",
        "responseTime": 145
      },
      "systemInfo": {
        "osVersion": "Ubuntu 22.04 LTS",
        "appVersion": "2.1.0",
        "lastUpdated": "2026-03-20T10:00:00Z"
      },
      "timestamps": {
        "createdAt": "2026-01-01T00:00:00Z",
        "updatedAt": "2026-03-20T10:00:00Z",
        "lastRestartAt": "2026-03-19T15:30:00Z",
        "lastUpgradeAt": "2026-03-15T10:00:00Z",
        "lastHealthyAt": "2026-03-20T11:55:00Z"
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
          "method": "POST"
        }
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 12,
    "totalPages": 1
  }
}
```

### `GET /environments/:id`

Returns a single environment (same schema).

### `POST /environments` *(admin only)*

**Request:**
```json
{
  "name": "staging-api",
  "description": "Staging API server",
  "environmentURL": "https://staging.example.com",
  "target": {
    "host": "192.168.1.101",
    "port": 22
  },
  "credentials": {
    "type": "key",
    "username": "deploy",
    "keyId": "507f1f77bcf86cd799439013"
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
    }
  },
  "commands": {
    "type": "ssh",
    "restart": {
      "command": "sudo systemctl restart myapp"
    }
  }
}
```

### `PUT /environments/:id` *(admin only)*

Same structure as `POST /environments`.

### `DELETE /environments/:id` *(admin only)*

```json
{ "message": "Environment deleted successfully" }
```

---

## Environment Operations

### `POST /environments/:id/restart`

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
  "operationId": "op-abc123",
  "status": "in_progress",
  "startedAt": "2026-03-20T12:00:00Z"
}
```

### `POST /environments/:id/check-health`

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

### `GET /environments/:id/versions`

**Response:**
```json
{
  "currentVersion": "2.1.0",
  "availableVersions": ["2.2.0", "2.1.1", "2.1.0", "2.0.5"]
}
```

### `POST /environments/:id/upgrade`

**Request:**
```json
{
  "version": "2.2.0"
}
```

**Response:**
```json
{
  "operationId": "op-def456",
  "status": "in_progress",
  "startedAt": "2026-03-20T12:00:00Z"
}
```

### `GET /environments/:id/logs`

**Query parameters:**
- `type`: `health_check` | `action` | `system` | `error` | `auth`
- `level`: `info` | `warning` | `error` | `success`
- `action`: `create` | `update` | `delete` | `restart` | `upgrade` | `login` | `logout`
- `startDate`, `endDate`: ISO 8601
- `page`, `limit`

---

## Logs

### `GET /logs`

**Query parameters:**
- `type`, `level`, `action`
- `environmentId`, `userId`
- `startDate`, `endDate`
- `page`, `limit`

**Response:**
```json
{
  "logs": [
    {
      "id": "507f1f77bcf86cd799439014",
      "timestamp": "2026-03-20T11:30:00Z",
      "environmentId": "507f1f77bcf86cd799439012",
      "environmentName": "production-api",
      "userId": "507f1f77bcf86cd799439011",
      "username": "admin",
      "type": "action",
      "level": "info",
      "action": "restart",
      "message": "Environment restarted successfully",
      "details": {
        "duration": 45000
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

### `GET /logs/count`

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

### `GET /logs/:id`

Returns a single log entry.

---

## WebSocket

### `WS /ws`

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
```

**Subscribe to environments:**
```json
{
  "type": "subscribe",
  "payload": { "environments": ["507f1f77bcf86cd799439012"] }
}
```

**Receive status update:**
```json
{
  "type": "status_update",
  "payload": {
    "environmentId": "507f1f77bcf86cd799439012",
    "status": {
      "health": "unhealthy",
      "message": "Connection timeout",
      "timestamp": "2026-03-20T12:00:00Z"
    }
  }
}
```

**Receive operation update:**
```json
{
  "type": "operation_update",
  "payload": {
    "operationId": "op-abc123",
    "status": "completed",
    "output": "Service restarted successfully"
  }
}
```

---

## Command Configuration

### SSH

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

### HTTP

```json
{
  "commands": {
    "type": "http",
    "restart": {
      "url": "https://api.example.com/admin/restart",
      "method": "POST",
      "headers": { "Authorization": "Bearer token" },
      "body": { "graceful": true }
    }
  }
}
```

---

## Health Check Validation

### Status code

```json
{ "type": "statusCode", "value": 200 }
```

### JSON regex

```json
{ "type": "jsonRegex", "value": "\"status\":\\s*\"(ok|healthy)\"" }
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `AUTH_INVALID` | 401 | Invalid credentials |
| `AUTH_EXPIRED` | 401 | Token expired |
| `AUTH_UNAUTHORIZED` | 401 | Missing or invalid token |
| `AUTH_FORBIDDEN` | 403 | Insufficient role permissions |
| `ENV_NOT_FOUND` | 404 | Environment not found |
| `ENV_DUPLICATE` | 409 | Environment name already exists |
| `USER_NOT_FOUND` | 404 | User not found |
| `USER_DUPLICATE` | 409 | Username already exists |
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `SSH_CONNECTION_FAILED` | 500 | SSH connection failed |
| `HEALTH_CHECK_FAILED` | 500 | Health check failed |
| `OPERATION_FAILED` | 500 | Operation execution failed |
| `INTERNAL_ERROR` | 500 | Internal server error |

---

## CORS

- **Allowed origins**: configured via `ALLOWED_ORIGINS` environment variable
- **Allowed methods**: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`
- **Allowed headers**: all
- **Credentials**: allowed
- **Max age**: 86400 seconds

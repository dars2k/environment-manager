# Data Models

## MongoDB Collections

### 1. `environments`

Stores all environment configurations and current state.

```javascript
{
  "_id": ObjectId,
  "name": String,                    // Unique environment name
  "description": String,             // Optional description
  "environmentURL": String,          // URL to access the environment
  "target": {
    "host": String,                  // IP address or hostname
    "port": Number,                  // SSH port (default: 22)
    "domain": String                 // Optional domain name
  },
  "credentials": {
    "type": String,                  // "key" or "password"
    "username": String,              // SSH username
    "keyId": ObjectId                // Reference to SSH key (if type is "key")
  },
  "healthCheck": {
    "enabled": Boolean,
    "endpoint": String,              // HTTP endpoint path
    "method": String,                // HTTP method (GET, POST, etc.)
    "interval": Number,              // Check interval in seconds
    "timeout": Number,               // Request timeout in seconds
    "validation": {
      "type": String,                // "statusCode" or "jsonRegex"
      "value": Mixed                 // Expected status code or regex pattern
    },
    "headers": Object                // Optional HTTP headers
  },
  "status": {
    "health": String,                // "healthy" | "unhealthy" | "unknown"
    "lastCheck": Date,
    "message": String,               // Last health check message
    "responseTime": Number           // Last response time in ms
  },
  "systemInfo": {
    "osVersion": String,
    "appVersion": String,
    "lastUpdated": Date
  },
  "timestamps": {
    "createdAt": Date,
    "updatedAt": Date,
    "lastRestartAt": Date,
    "lastUpgradeAt": Date,
    "lastHealthyAt": Date
  },
  "commands": {
    "type": String,                  // "ssh" or "http"
    "restart": {
      "command": String,             // SSH: shell command to execute
      "url": String,                 // HTTP: endpoint URL
      "method": String,              // HTTP: method
      "headers": Object,             // HTTP: request headers
      "body": Object                 // HTTP: request body
    }
  },
  "upgradeConfig": {
    "enabled": Boolean,
    "type": String,                  // "ssh" or "http"
    "versionListURL": String,        // URL to fetch available versions
    "versionListMethod": String,
    "versionListHeaders": Object,
    "versionListBody": String,
    "jsonPathResponse": String,      // JSONPath to extract versions from response
    "upgradeCommand": {
      "command": String,             // SSH: upgrade command
      "url": String,                 // HTTP: upgrade endpoint
      "method": String,
      "headers": Object,
      "body": Object
    }
  },
  "metadata": Object                 // Custom key/value fields
}
```

**Indexes:**
- `name`: unique
- `status.health`: for health-based filtering
- `timestamps.createdAt`: for sort by creation date

---

### 2. `users`

Stores user accounts and authentication information.

```javascript
{
  "_id": ObjectId,
  "username": String,                // Unique username (alphanum, 3–50 chars)
  "passwordHash": String,            // bcrypt hash, never returned in API responses
  "role": String,                    // "admin" | "user" | "viewer"
  "active": Boolean,                 // Account enabled/disabled
  "lastLogin": Date,                 // Last successful login (optional)
  "createdAt": Date,
  "updatedAt": Date
}
```

**Role permissions:**

| Role | Environments (read) | Environments (write) | Operations (restart/upgrade) | Users |
|------|----|----|----|----|
| `admin` | ✓ | ✓ | ✓ | ✓ |
| `user` | ✓ | | ✓ | |
| `viewer` | ✓ | | | |

**Indexes:**
- `username`: unique
- `active`: for filtering active users

**Validation rules:**
- `username`: alphanumeric, 3–50 characters
- `password`: minimum 12 characters, must contain uppercase, lowercase, and a digit
- `role`: enum `["admin", "user", "viewer"]`

---

### 3. `logs`

Audit log for all actions, health checks, and system events.

```javascript
{
  "_id": ObjectId,
  "timestamp": Date,
  "environmentId": ObjectId,         // Optional reference to environment
  "environmentName": String,         // Denormalized for query performance
  "userId": ObjectId,                // Optional reference to user
  "username": String,                // Denormalized for query performance
  "type": String,                    // Log type (see below)
  "level": String,                   // Log level (see below)
  "action": String,                  // Action type (see below)
  "message": String,
  "details": Object                  // Additional context
}
```

**Log types:** `health_check` | `action` | `system` | `error` | `auth`

**Log levels:** `info` | `warning` | `error` | `success`

**Action types:** `create` | `update` | `delete` | `restart` | `shutdown` | `upgrade` | `login` | `logout`

**Indexes:**
- `timestamp`: for time-based queries (desc)
- `environmentId`: for environment-specific log views
- `userId`: for user activity tracking
- `type`, `level`: for filtering

---

## Design Decisions

### Denormalization

`environmentName` and `username` are stored directly in log documents to avoid joins when rendering the log view. The trade-off is that these values must be captured at write time (they reflect the name at the time of the action).

### Embedded vs Referenced

- **Embedded**: health check config, command config, and status — always accessed together with the environment document, so embedding avoids extra queries and allows atomic updates.
- **Referenced**: users and environments in logs — kept independent so either can be deleted without cascading issues.

### Dual Command Configuration

Both `commands` and `upgradeConfig` support SSH and HTTP execution types, using the same nested structure. This allows the same model to manage both traditional server environments (SSH) and modern microservices (HTTP management endpoints).

---

## Common Query Patterns

### Dashboard — all environments

```javascript
db.environments.find({}, {
  name: 1,
  description: 1,
  environmentURL: 1,
  "status.health": 1,
  "status.lastCheck": 1,
  "status.message": 1,
  "systemInfo.appVersion": 1,
  "timestamps.lastRestartAt": 1,
  "timestamps.lastUpgradeAt": 1
}).sort({ name: 1 })
```

### Environment logs (last 24 hours)

```javascript
db.logs.find({
  environmentId: ObjectId("..."),
  timestamp: { $gte: new Date(Date.now() - 24 * 60 * 60 * 1000) }
}).sort({ timestamp: -1 }).limit(100)
```

### Health status summary

```javascript
db.environments.aggregate([
  { $group: { _id: "$status.health", count: { $sum: 1 } } }
])
```

### Log counts by type and level

```javascript
db.logs.aggregate([
  {
    $facet: {
      byType: [{ $group: { _id: "$type", count: { $sum: 1 } } }],
      byLevel: [{ $group: { _id: "$level", count: { $sum: 1 } } }]
    }
  }
])
```

---

## Environment Validation Rules

```
name:              required, /^[a-zA-Z0-9-_]+$/, max 50 chars
target.host:       required, hostname or IP
target.port:       1–65535
healthCheck.interval: 10–3600 seconds
healthCheck.timeout:  1–60 seconds
```

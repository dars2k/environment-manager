# Data Models

## MongoDB Collections

### 1. environments Collection

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
    "keyId": ObjectId               // Reference to SSH key (if type is "key")
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
    "health": String,                // "healthy", "unhealthy", "unknown"
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
      "command": String,             // For SSH type
      "url": String,                 // For HTTP type
      "method": String,              // For HTTP type
      "headers": Object,             // For HTTP type
      "body": Object                 // For HTTP type
    }
  },
  "upgradeConfig": {
    "enabled": Boolean,
    "type": String,                  // "ssh" or "http"
    "versionListURL": String,        // URL to fetch available versions
    "versionListMethod": String,     // HTTP method for version list
    "versionListHeaders": Object,    // Headers for version list request
    "versionListBody": String,       // Body for version list request
    "jsonPathResponse": String,      // JSONPath to extract versions
    "upgradeCommand": {
      "command": String,             // For SSH type
      "url": String,                 // For HTTP type
      "method": String,              // For HTTP type
      "headers": Object,             // For HTTP type
      "body": Object                 // For HTTP type
    }
  },
  "metadata": Object                 // Custom metadata fields
}
```

**Indexes:**
- `name`: Unique index for fast lookups
- `status.health`: For filtering by health status
- `timestamps.createdAt`: For sorting by creation date

### 2. users Collection

Stores user accounts and authentication information.

```javascript
{
  "_id": ObjectId,
  "username": String,                // Unique username
  "email": String,                   // Unique email address
  "passwordHash": String,            // Bcrypt hashed password
  "role": String,                    // User role (e.g., "admin", "user")
  "active": Boolean,                 // Account active status
  "lastLogin": Date,                 // Last successful login timestamp
  "createdAt": Date,
  "updatedAt": Date
}
```

**Indexes:**
- `username`: Unique index
- `email`: Unique index
- `active`: For filtering active users

### 3. logs Collection

Event logging for all actions, health checks, and system events.

```javascript
{
  "_id": ObjectId,
  "timestamp": Date,
  "environmentId": ObjectId,         // Optional reference to environment
  "environmentName": String,         // Denormalized environment name
  "userId": ObjectId,                // Optional reference to user
  "username": String,                // Denormalized username
  "type": String,                    // Log type enum
  "level": String,                   // Log level enum
  "action": String,                  // Action type enum (optional)
  "message": String,                 // Log message
  "details": Object                  // Additional context data
}
```

**Log Types:**
- `health_check`: Health check results
- `action`: User-initiated actions
- `system`: System events
- `error`: Error logs
- `auth`: Authentication events

**Log Levels:**
- `info`: Informational messages
- `warning`: Warning messages
- `error`: Error messages
- `success`: Success messages

**Action Types:**
- `create`: Resource creation
- `update`: Resource update
- `delete`: Resource deletion
- `restart`: Environment restart
- `shutdown`: Environment shutdown
- `upgrade`: Environment upgrade
- `login`: User login
- `logout`: User logout

**Indexes:**
- `timestamp`: For time-based queries
- `environmentId`: For environment-specific logs
- `userId`: For user activity tracking
- `type`: For filtering by log type
- `level`: For filtering by severity

## Data Modeling Decisions

### 1. Denormalization Strategy
- Environment name and username are denormalized in logs for performance
- This avoids expensive joins when displaying logs
- Trade-off: Need to handle updates if names change

### 2. Embedded vs Referenced
- **Embedded**: 
  - Health check configuration in environments (always accessed together)
  - Command configuration in environments (atomic updates)
  - Status in environments (frequently accessed)
- **Referenced**: 
  - Users and environments in logs (maintains independence)
  - SSH keys (security isolation - not implemented yet)

### 3. Schema Flexibility
- Using `metadata` and `details` fields for extensibility
- Allows custom fields without schema changes
- Useful for environment-specific configurations

### 4. Command Configuration Design
The dual SSH/HTTP command support allows flexibility:
- SSH commands for traditional server management
- HTTP endpoints for modern microservices
- Same pattern for both restart and upgrade operations

## Query Patterns

### Dashboard Query
```javascript
// Get all environments with latest status
db.environments.find(
  {},
  {
    name: 1,
    description: 1,
    environmentURL: 1,
    "status.health": 1,
    "status.lastCheck": 1,
    "status.message": 1,
    "systemInfo.appVersion": 1,
    "timestamps.lastRestartAt": 1,
    "timestamps.lastUpgradeAt": 1
  }
).sort({ name: 1 })
```

### Environment Details Query
```javascript
// Get full environment details
db.environments.findOne({ _id: ObjectId("...") })

// Get recent logs for environment
db.logs.find({
  environmentId: ObjectId("..."),
  timestamp: { $gte: new Date(Date.now() - 24*60*60*1000) }
}).sort({ timestamp: -1 }).limit(100)
```

### User Activity Query
```javascript
// Get user's recent actions
db.logs.find({
  userId: ObjectId("..."),
  type: "action",
  timestamp: { $gte: new Date(Date.now() - 7*24*60*60*1000) }
}).sort({ timestamp: -1 })
```

### Health Status Summary
```javascript
// Count environments by health status
db.environments.aggregate([
  {
    $group: {
      _id: "$status.health",
      count: { $sum: 1 }
    }
  }
])
```

### Log Aggregation
```javascript
// Get log counts by type and level
db.logs.aggregate([
  {
    $facet: {
      byType: [
        { $group: { _id: "$type", count: { $sum: 1 } } }
      ],
      byLevel: [
        { $group: { _id: "$level", count: { $sum: 1 } } }
      ]
    }
  }
])
```

## Data Validation

### Environment Validation Rules
```javascript
{
  name: {
    required: true,
    pattern: /^[a-zA-Z0-9-_]+$/,
    maxLength: 50
  },
  "target.host": {
    required: true,
    pattern: /^[a-zA-Z0-9.-]+$/
  },
  "target.port": {
    type: "number",
    min: 1,
    max: 65535
  },
  "healthCheck.interval": {
    type: "number",
    min: 10,
    max: 3600
  },
  "healthCheck.timeout": {
    type: "number",
    min: 1,
    max: 60
  }
}
```

### User Validation Rules
```javascript
{
  username: {
    required: true,
    pattern: /^[a-zA-Z0-9_-]+$/,
    minLength: 3,
    maxLength: 30
  },
  email: {
    required: true,
    pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  },
  role: {
    enum: ["admin", "user", "viewer"]
  }
}
```

## Performance Considerations

### 1. Index Strategy
- Cover queries for dashboard views
- Compound indexes for complex queries
- TTL indexes for log retention (future)

### 2. Query Optimization
- Use projections to limit returned fields
- Paginate large result sets
- Cache frequently accessed data (future)

### 3. Write Performance
- Batch log inserts where possible
- Use upserts for status updates
- Consider write concern for critical operations

## Future Enhancements

### 1. Time-Series Collections
For MongoDB 5.0+, consider using time-series collections for:
- Health metrics with automatic compression
- Performance data with optimized queries
- Log data with automatic retention

### 2. Additional Collections
- **ssh_keys**: Separate collection for SSH key management
- **api_keys**: For API authentication
- **notifications**: For alert configurations
- **metrics**: For detailed performance metrics

### 3. Data Archival
- Move old logs to archive collections
- Implement data retention policies
- Consider cold storage for historical data

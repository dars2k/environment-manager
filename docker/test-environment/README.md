# Test Environment for App Environment Manager

Docker containers that simulate real environments with SSH access, health checks, and upgrade support.

## Credentials

| User       | Password      | Notes           |
|------------|---------------|-----------------|
| root       | testpassword  | Full root access |
| testuser   | testpass      | Standard user   |

## Quick Start

```bash
cd docker/test-environment
docker compose up -d --build
```

Available after startup:

| Container        | SSH Port | Health URL                     | Initial Version |
|-----------------|----------|-------------------------------|-----------------|
| test-environment-1 | 2221  | http://localhost:8001/health  | 1.0.0           |
| test-environment-2 | 2222  | http://localhost:8002/health  | 2.0.0           |

## HTTP Endpoints

| Method | Path                | Description                            |
|--------|---------------------|----------------------------------------|
| GET    | /health             | Health status + current version        |
| GET    | /version            | Current version + available versions   |
| GET    | /info               | System information                     |
| GET    | /metrics            | CPU, memory, disk stats                |
| POST   | /restart            | Simulate service restart               |
| POST   | /upgrade            | Upgrade to a version (updates version) |
| POST   | /toggle-health      | Toggle healthy/unhealthy state         |
| POST   | /set-response-time  | Set artificial response delay          |

## Environment Configuration Examples

### Minimal (SSH + Health Check)

```json
{
  "name": "Test Environment 1",
  "description": "Local test container with SSH",
  "environmentURL": "http://localhost:8001",
  "target": {
    "host": "localhost",
    "port": 2221
  },
  "credentials": {
    "type": "password",
    "username": "root"
  },
  "healthCheck": {
    "enabled": true,
    "endpoint": "/health",
    "method": "GET",
    "interval": 30,
    "timeout": 10,
    "validation": {
      "type": "statusCode",
      "value": 200
    }
  },
  "commands": {
    "type": "ssh",
    "restart": {
      "enabled": true,
      "command": "echo 'service restarted'"
    }
  },
  "metadata": {
    "password": "testpassword"
  }
}
```

### Full Configuration (All Available Options)

```json
{
  "name": "Test Environment 1 Full",
  "description": "Full-featured test environment with upgrade support",
  "environmentURL": "http://localhost:8001",
  "target": {
    "host": "localhost",
    "port": 2221,
    "domain": "test-env-1.local"
  },
  "credentials": {
    "type": "password",
    "username": "root"
  },
  "healthCheck": {
    "enabled": true,
    "endpoint": "http://localhost:8001/health",
    "method": "GET",
    "interval": 30,
    "timeout": 10,
    "validation": {
      "type": "jsonRegex",
      "value": "\"status\":\"healthy\""
    },
    "headers": {
      "Accept": "application/json"
    }
  },
  "commands": {
    "type": "ssh",
    "restart": {
      "enabled": true,
      "command": "echo 'restarting service'"
    }
  },
  "upgradeConfig": {
    "enabled": true,
    "type": "ssh",
    "versionListURL": "http://localhost:8001/version",
    "versionListMethod": "GET",
    "jsonPathResponse": "available",
    "upgradeCommand": {
      "command": "echo {VERSION} > /tmp/app_version"
    }
  },
  "metadata": {
    "password": "testpassword"
  }
}
```

### SSH Key Authentication Example

```json
{
  "name": "Test Environment Key Auth",
  "target": {
    "host": "localhost",
    "port": 2221
  },
  "credentials": {
    "type": "key",
    "username": "root",
    "keyId": "<SSH_KEY_OBJECT_ID>"
  },
  "commands": {
    "type": "ssh",
    "restart": {
      "enabled": true,
      "command": "echo 'restarted'"
    }
  }
}
```

### HTTP Command Type (no SSH needed)

```json
{
  "name": "Test Environment HTTP",
  "environmentURL": "http://localhost:8001",
  "target": {
    "host": "localhost",
    "port": 8001
  },
  "credentials": {
    "type": "password",
    "username": "testuser"
  },
  "healthCheck": {
    "enabled": true,
    "endpoint": "/health",
    "method": "GET",
    "interval": 60,
    "timeout": 5,
    "validation": {
      "type": "statusCode",
      "value": 200
    }
  },
  "commands": {
    "type": "http",
    "restart": {
      "enabled": true,
      "url": "http://localhost:8001/restart",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {}
    }
  },
  "upgradeConfig": {
    "enabled": true,
    "type": "http",
    "versionListURL": "http://localhost:8001/version",
    "versionListMethod": "GET",
    "jsonPathResponse": "available",
    "upgradeCommand": {
      "url": "http://localhost:8001/upgrade",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "version": "{VERSION}"
      }
    }
  }
}
```

## Testing Scenarios

### Check health status
```bash
curl http://localhost:8001/health
```

### Get available versions
```bash
curl http://localhost:8001/version
```

### Simulate upgrade via SSH
```bash
ssh root@localhost -p 2221 "echo 1.1.0 > /tmp/app_version"
curl http://localhost:8001/health   # version should now show 1.1.0
```

### Simulate upgrade via HTTP
```bash
curl -X POST http://localhost:8001/upgrade \
  -H "Content-Type: application/json" \
  -d '{"version": "2.0.0"}'
curl http://localhost:8001/health   # version should now show 2.0.0
```

### Toggle unhealthy state (for testing alerts)
```bash
curl -X POST http://localhost:8001/toggle-health
curl http://localhost:8001/health   # returns 503
curl -X POST http://localhost:8001/toggle-health   # toggle back
```

### Simulate slow response
```bash
curl -X POST http://localhost:8001/set-response-time \
  -H "Content-Type: application/json" \
  -d '{"ms": 2000}'
```

## Stopping

```bash
cd docker/test-environment
docker compose down
```

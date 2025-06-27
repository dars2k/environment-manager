# Test Environment for App Environment Manager

This directory contains Docker containers that simulate test environments with SSH access and health check endpoints.

## Features

- SSH access with password authentication
- Health check endpoints (`/health` and `/api/health`)
- Ability to toggle health status for testing
- Configurable response times

## Quick Start

1. Start the test environments:
   ```bash
   cd docker/test-environment
   docker-compose up -d
   ```

2. The test environments will be available at:
   - **Test Environment 1**: 
     - SSH: `localhost:2221` (user: `testuser`, password: `testpassword`)
     - Health Check: `http://localhost:8001/health`
   - **Test Environment 2**:
     - SSH: `localhost:2222` (user: `testuser`, password: `testpassword`)
     - Health Check: `http://localhost:8002/health`

## Adding to App Environment Manager

You can add these test environments to your App Environment Manager with the following configuration:

### Test Environment 1
- **Name**: Test Environment 1
- **Host**: localhost (or your Docker host IP)
- **Port**: 2221
- **Username**: testuser
- **Password**: testpassword
- **Health Check Endpoint**: http://localhost:8001/health
- **Health Check Method**: GET
- **Expected Status Code**: 200

### Test Environment 2
- **Name**: Test Environment 2
- **Host**: localhost (or your Docker host IP)
- **Port**: 2222
- **Username**: testuser
- **Password**: testpassword
- **Health Check Endpoint**: http://localhost:8002/health
- **Health Check Method**: GET
- **Expected Status Code**: 200

## Testing Features

### Toggle Health Status
```bash
curl -X POST http://localhost:8001/toggle-health
```

### Set Response Time
```bash
curl -X POST http://localhost:8001/set-response-time \
  -H "Content-Type: application/json" \
  -d '{"ms": 500}'
```

## Stopping Test Environments

```bash
cd docker/test-environment
docker-compose down

# Docker Setup

This directory contains supporting files for running the Environment Manager with Docker.

## Prerequisites

- [Docker Engine](https://docs.docker.com/engine/install/) 24.0+
- [Docker Compose](https://docs.docker.com/compose/install/) v2.0+ (included with Docker Desktop)
- `make` (optional, for Makefile shortcuts)

## Quick Start

```bash
# 1. Copy and configure environment variables
cp .env.example .env
# Edit .env — set secure JWT_SECRET and SSH_KEY_ENCRYPTION_KEY

# 2. Build and start all services
make build
make up

# 3. Check that everything is running
make health
```

> **Admin credentials**: The admin password is auto-generated on first startup and printed to the backend logs. Retrieve it with `make logs-backend`.

## Services

| Service | Description | Default port |
|---------|-------------|-------------|
| `mongodb` | MongoDB 8.2 database | 27017 |
| `backend` | Go API server with WebSocket support | 8080 |
| `frontend` | React app served by Nginx | 80 |
| `mongo-express` | MongoDB admin UI (debug profile only) | 8081 |

## Access Points

- **Frontend**: http://localhost
- **Backend API**: http://localhost:8080
- **MongoDB Express** (optional): http://localhost:8081

## Make Commands

```bash
# Lifecycle
make build           # Build all Docker images
make up              # Start services in background
make up-logs         # Start services with log streaming
make down            # Stop services
make clean           # Remove containers, volumes, and images

# Logs
make logs            # Stream all service logs
make logs-backend    # Stream backend logs
make logs-frontend   # Stream frontend logs
make logs-db         # Stream database logs

# Service management
make restart-backend # Restart backend
make restart-frontend# Restart frontend
make build-backend   # Rebuild backend image only
make build-frontend  # Rebuild frontend image only

# Debugging
make health          # Check service health status
make backend-shell   # Open shell in backend container
make frontend-shell  # Open shell in frontend container
make db-shell        # Open MongoDB shell
make mongo-express   # Start MongoDB Express UI
```

## Docker Compose Commands (direct)

```bash
# Start services
docker compose up -d

# Stream logs
docker compose logs -f

# Stop services
docker compose down

# Rebuild a specific service
docker compose build backend
docker compose up -d backend

# Execute commands in containers
docker compose exec backend /bin/sh
docker compose exec mongodb mongosh
```

## Key Environment Variables

See `.env.example` for the full list. Critical values to set for production:

```bash
# Generate with: openssl rand -hex 32
JWT_SECRET=<strong-secret>

# Must be exactly 32 bytes — generate with: openssl rand -hex 16
SSH_KEY_ENCRYPTION_KEY=<32-byte-key>

# Strong MongoDB credentials
MONGO_ROOT_PASSWORD=<strong-password>

# Comma-separated allowed origins
ALLOWED_ORIGINS=https://your-domain.com
```

## Troubleshooting

**Services not starting**
```bash
make logs           # Check for error messages
docker compose ps   # Check container status
```

**Port already in use**
```bash
# Check what's using the port
lsof -i :80
lsof -i :8080
lsof -i :27017
```

**Database connection issues**
```bash
make logs-db        # Check MongoDB logs
# Verify .env credentials match docker-compose.yml
```

**Frontend not loading**
```bash
docker compose logs frontend
curl http://localhost:8080/api/v1/health  # Check if backend is up
```

## Backup and Restore

### Backup MongoDB

```bash
docker compose exec mongodb mongodump \
  --username=admin \
  --password=<password> \
  --authenticationDatabase=admin \
  --db=app-env-manager \
  --out=/backup

docker cp $(docker compose ps -q mongodb):/backup ./mongodb-backup
```

### Restore MongoDB

```bash
docker cp ./mongodb-backup $(docker compose ps -q mongodb):/backup

docker compose exec mongodb mongorestore \
  --username=admin \
  --password=<password> \
  --authenticationDatabase=admin \
  --db=app-env-manager \
  /backup/app-env-manager
```

## Security Notes

1. Always use strong, unique passwords and secrets in production
2. `SSH_KEY_ENCRYPTION_KEY` must be exactly 32 bytes
3. Services communicate over an isolated Docker network (`app-network`)
4. All containers run as non-root users
5. Built-in health checks for all services

## Monitoring

| Check | Command |
|-------|---------|
| Service status | `docker compose ps` |
| Resource usage | `docker stats` |
| Backend health API | `curl http://localhost:8080/api/v1/health` |
| All service logs | `make logs` |

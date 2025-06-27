# Docker Setup for App Environment Manager

This directory contains all the necessary files to run the Application Environment Manager using Docker containers.

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- Make (optional, for using Makefile commands)

## Quick Start

1. **Clone the repository and navigate to the project root**

2. **Initialize the environment**
   ```bash
   make init
   # or manually:
   cp .env.example .env
   ```

3. **Update the `.env` file** with your configuration:
   - Set strong passwords for MongoDB
   - Generate a secure JWT secret (at least 32 characters)
   - Set a 32-byte SSH encryption key

4. **Build and start all services**
   ```bash
   make build
   make up
   # or with docker-compose:
   docker-compose build
   docker-compose up -d
   ```

5. **Access the application**
   - Frontend: http://localhost
   - Backend API: http://localhost:8080
   - MongoDB Express (optional): http://localhost:8081

## Architecture

The Docker setup includes:

- **MongoDB**: Database for storing environments and audit logs
- **Backend**: Go API server with WebSocket support
- **Frontend**: React application served by Nginx
- **MongoDB Express**: Optional web-based MongoDB admin interface

## Environment Variables

Key environment variables (see `.env.example`):

```bash
# MongoDB
MONGO_ROOT_USER=admin
MONGO_ROOT_PASSWORD=<secure-password>
MONGO_DATABASE=app-env-manager

# Backend
JWT_SECRET=<32+ character secret>
SSH_KEY_ENCRYPTION_KEY=<exactly 32 bytes>

# Ports
FRONTEND_PORT=80
BACKEND_PORT=8080
MONGO_PORT=27017
```

## Common Commands

### Using Make

```bash
# Start services
make up

# View logs
make logs
make logs-backend
make logs-frontend

# Stop services
make down

# Clean everything (including volumes)
make clean

# Access containers
make backend-shell
make db-shell

# Health check
make health

# Debug with MongoDB Express
make mongo-express
```

### Using Docker Compose

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild a specific service
docker-compose build backend
docker-compose up -d backend

# Execute commands in containers
docker-compose exec backend /bin/sh
docker-compose exec mongodb mongosh
```

## Development vs Production

### Development Mode
- Source code mounted as volumes
- Hot reloading enabled
- Debug tools available

```bash
make dev
# or
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up
```

### Production Mode
- Optimized builds
- Security hardening
- No source code mounting

```bash
make prod
# or
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## Troubleshooting

### Services not starting
1. Check logs: `make logs`
2. Verify port availability: `netstat -tulpn | grep -E '80|8080|27017'`
3. Ensure Docker daemon is running: `docker ps`

### Database connection issues
1. Verify MongoDB is healthy: `docker-compose ps`
2. Check credentials in `.env` match those in `docker-compose.yml`
3. Inspect MongoDB logs: `make logs-db`

### Frontend not loading
1. Check Nginx logs: `docker-compose logs frontend`
2. Verify backend is accessible: `curl http://localhost:8080/api/v1/health`
3. Check browser console for errors

### Permission issues
- Ensure proper ownership: Files are owned by non-root user in containers
- Check Docker socket permissions if using bind mounts

## Security Considerations

1. **Passwords**: Always use strong, unique passwords in production
2. **Encryption**: The SSH_KEY_ENCRYPTION_KEY must be exactly 32 bytes
3. **Network**: The app-network is isolated by default
4. **Non-root**: All containers run as non-root users
5. **Health checks**: Built-in health monitoring for all services

## Backup and Restore

### Backup MongoDB
```bash
docker-compose exec mongodb mongodump \
  --username=admin \
  --password=<password> \
  --authenticationDatabase=admin \
  --db=app-env-manager \
  --out=/backup

docker cp app-env-manager-db:/backup ./mongodb-backup
```

### Restore MongoDB
```bash
docker cp ./mongodb-backup app-env-manager-db:/backup

docker-compose exec mongodb mongorestore \
  --username=admin \
  --password=<password> \
  --authenticationDatabase=admin \
  --db=app-env-manager \
  /backup/app-env-manager
```

## Monitoring

- Backend health: http://localhost:8080/api/v1/health
- Frontend health: http://localhost/health
- Container status: `docker-compose ps`
- Resource usage: `docker stats`

## Advanced Configuration

### Custom SSL Certificates
1. Mount certificates in `frontend` service
2. Update `nginx.conf` for HTTPS
3. Update `ALLOWED_ORIGINS` in backend

### Scaling
```bash
# Scale backend instances
docker-compose up -d --scale backend=3
```

### Custom Networks
Modify `docker-compose.yml` to use external networks for integration with other services.

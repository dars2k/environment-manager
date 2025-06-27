# Makefile for App Environment Manager

.PHONY: help build up down logs clean test dev prod

# Default target
help:
	@echo "Available commands:"
	@echo "  make build      - Build all Docker images"
	@echo "  make up         - Start all services"
	@echo "  make down       - Stop all services"
	@echo "  make logs       - View logs from all services"
	@echo "  make clean      - Remove containers, volumes, and images"
	@echo "  make test       - Run tests"
	@echo "  make dev        - Start in development mode"
	@echo "  make prod       - Start in production mode"
	@echo "  make db-shell   - Open MongoDB shell"
	@echo "  make backend-shell - Open backend container shell"

# Build all Docker images
build:
	docker-compose build --no-cache

# Start all services
up:
	docker-compose up -d

# Start with logs
up-logs:
	docker-compose up

# Stop all services
down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# View specific service logs
logs-backend:
	docker-compose logs -f backend

logs-frontend:
	docker-compose logs -f frontend

logs-db:
	docker-compose logs -f mongodb

# Clean everything
clean:
	docker-compose down -v --rmi all

# Run tests
test:
	cd backend && go test ./...
	cd frontend && npm test

# Development mode
dev:
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up

# Production mode
prod:
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Database shell
db-shell:
	docker-compose exec mongodb mongosh -u admin -p admin123 --authenticationDatabase admin

# Backend shell
backend-shell:
	docker-compose exec backend /bin/sh

# Frontend shell
frontend-shell:
	docker-compose exec frontend /bin/sh

# Restart specific service
restart-backend:
	docker-compose restart backend

restart-frontend:
	docker-compose restart frontend

# Build specific service
build-backend:
	docker-compose build backend

build-frontend:
	docker-compose build frontend

# Initialize environment
init:
	@echo "Initializing environment..."
	@cp -n .env.example .env || true
	@echo "Environment file created. Please update .env with your configuration."

# Health check
health:
	@echo "Checking service health..."
	@docker-compose ps
	@echo "\nBackend health:"
	@curl -s http://localhost:8080/api/v1/health || echo "Backend not responding"
	@echo "\n\nFrontend health:"
	@curl -s http://localhost/health || echo "Frontend not responding"

# MongoDB Express (debug tool)
mongo-express:
	docker-compose --profile debug up -d mongo-express
	@echo "MongoDB Express available at http://localhost:8081"

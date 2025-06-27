# Application Environment Manager

A modern, scalable system for managing application environments with real-time monitoring, SSH-based operations, and comprehensive audit logging.

## 🚀 Features

- **Environment Management**: Create, update, and delete application environments with ease
- **Real-time Monitoring**: Live health status updates via WebSocket connections
- **Command Execution**: Support for both SSH and HTTP-based commands
- **User Authentication**: Secure JWT-based authentication system with role-based access control
- **Audit Logging**: Comprehensive logging of all actions and changes
- **Health Monitoring**: Automated health checks with configurable intervals
- **Version Management**: Track and upgrade environment versions
- **Dark Theme UI**: Modern, responsive interface built with Material-UI
- **Docker Support**: Easy deployment with Docker Compose

## 🏗️ Architecture Overview

The Application Environment Manager follows a clean, modular architecture with clear separation of concerns:

- **Frontend**: React 18 + TypeScript + Redux Toolkit + Material-UI + Vite
- **Backend**: Go 1.23+ with clean architecture pattern (Gorilla Mux)
- **Database**: MongoDB 7.0 for flexible data storage
- **Real-time Updates**: WebSocket connections for live status updates
- **Operations**: SSH and HTTP command execution with secure credential management

## 📁 Project Structure

```
environment-manager/
├── frontend/                 # React frontend application
│   ├── src/
│   │   ├── api/            # API client services
│   │   ├── components/     # Reusable UI components
│   │   ├── contexts/       # React contexts (WebSocket)
│   │   ├── hooks/          # Custom React hooks
│   │   ├── layouts/        # Page layouts
│   │   ├── pages/          # Page components
│   │   ├── routes/         # Route definitions
│   │   ├── store/          # Redux store and slices
│   │   ├── theme/          # MUI theme configuration
│   │   ├── types/          # TypeScript types
│   │   └── test/           # Test utilities
│   └── public/
├── backend/                 # Go backend application
│   ├── cmd/                # Application entrypoints
│   ├── internal/           # Private application code
│   │   ├── api/           # HTTP handlers and routes
│   │   ├── domain/        # Business logic and entities
│   │   ├── repository/    # Data access layer
│   │   ├── service/       # Application services
│   │   └── websocket/     # WebSocket handlers
│   └── config/            # Configuration files
├── docker/                # Docker configurations
├── docs/                  # Documentation
│   ├── architecture/      # Architecture diagrams and docs
│   ├── api/              # API documentation
│   └── deployment/       # Deployment guides
└── init-mongo.js         # MongoDB initialization script
```

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone https://github.com/dars2k/environment-manager.git
   cd environment-manager
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start the application**
   ```bash
   docker-compose up -d
   ```

4. **Access the application**
   - Frontend: http://localhost (port 80)
   - Backend API: http://localhost:8080
   - MongoDB Express (debug profile): http://localhost:8081
   - Default credentials: admin / admin123

### Manual Setup

1. **Prerequisites**
   - Go 1.23+
   - Node.js 18+
   - MongoDB 7.0+

2. **Backend Setup**
   ```bash
   cd backend
   go mod download
   go run cmd/server/main.go
   ```

3. **Frontend Setup**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **MongoDB Setup**
   ```bash
   docker run -d -p 27017:27017 --name env-manager-mongo mongo:7.0
   ```

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the root directory:

```env
# MongoDB Configuration
MONGO_ROOT_USER=admin
MONGO_ROOT_PASSWORD=admin123
MONGO_DATABASE=app-env-manager
MONGO_PORT=27017

# Backend Configuration
BACKEND_PORT=8080
JWT_SECRET=your-super-secret-jwt-key-change-this
SSH_KEY_ENCRYPTION_KEY=your-32-byte-encryption-key-here!
ALLOWED_ORIGINS=http://localhost

# Frontend Configuration
FRONTEND_PORT=80

# MongoDB Express (Optional)
ME_USERNAME=admin
ME_PASSWORD=admin123
ME_PORT=8081
```

## 📚 API Documentation

All API endpoints are prefixed with `/api/v1`

### Authentication Endpoints

- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/auth/me` - Get current user info

### Environment Endpoints

- `GET /api/v1/environments` - List all environments
- `GET /api/v1/environments/:id` - Get environment details
- `POST /api/v1/environments` - Create new environment
- `PUT /api/v1/environments/:id` - Update environment
- `DELETE /api/v1/environments/:id` - Delete environment
- `POST /api/v1/environments/:id/restart` - Restart environment
- `POST /api/v1/environments/:id/check-health` - Check environment health
- `GET /api/v1/environments/:id/versions` - Get available versions
- `POST /api/v1/environments/:id/upgrade` - Upgrade environment
- `GET /api/v1/environments/:id/logs` - Get environment logs

### User Management Endpoints

- `GET /api/v1/users` - List all users
- `GET /api/v1/users/:id` - Get user details
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `POST /api/v1/users/password/change` - Change password
- `POST /api/v1/users/:id/password/reset` - Reset user password

### Log Endpoints

- `GET /api/v1/logs` - List all logs
- `GET /api/v1/logs/count` - Get log count
- `GET /api/v1/logs/:id` - Get log by ID

### WebSocket Endpoint

- `WS /ws` - WebSocket connection for real-time updates

### Health Check

- `GET /api/v1/health` - Application health check

## 🛠️ Development

### Frontend Development

```bash
cd frontend
npm run dev          # Start development server
npm run build        # Build for production
npm run test         # Run tests
npm run test:ui      # Run tests with UI
npm run test:coverage # Run tests with coverage
npm run lint         # Run ESLint
npm run format       # Format code with Prettier
```

### Backend Development

```bash
cd backend
go run cmd/server/main.go  # Run development server
go test ./...              # Run tests
go build -o app-env-manager cmd/server/main.go  # Build binary
```

### Docker Development

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Start with debug profile (includes MongoDB Express)
docker-compose --profile debug up -d
```

## 🧪 Testing

See [TESTING.md](./TESTING.md) for comprehensive testing guide.

### Quick Test Commands

```bash
# Run all unit tests
make test

# Run with Docker
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## 📖 Documentation

- [Architecture Documentation](./docs/architecture/overview.md)
- [Backend Design](./docs/architecture/backend-design.md)
- [Frontend Design](./docs/architecture/frontend-design.md)
- [Data Models](./docs/architecture/data-models.md)
- [UI/UX Design](./docs/architecture/ui-ux-design.md)
- [API Documentation](./docs/api/README.md)
- [Deployment Guide](./docs/deployment/README.md)
- [Testing Guide](./TESTING.md)

## 🏭 Production Deployment

### Docker Production Setup

1. Update environment variables in `.env` for production
2. Ensure secure JWT_SECRET and SSH_KEY_ENCRYPTION_KEY
3. Configure proper ALLOWED_ORIGINS
4. Set strong MongoDB credentials
5. Deploy using:
   ```bash
   docker-compose -f docker-compose.yml up -d
   ```

### Development Guidelines

- Follow Go best practices and conventions
- Use TypeScript strict mode
- Write unit tests for new features
- Update documentation as needed
- Ensure all tests pass before submitting PR

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with ❤️ using Go, React, and MongoDB
- UI powered by Material-UI
- Real-time updates via Gorilla WebSocket
- HTTP routing with Gorilla Mux
- State management with Redux Toolkit
- Form handling with Formik and Yup
- Charts with Recharts and Chart.js

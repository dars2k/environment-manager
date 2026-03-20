# Backend Architecture Design

## Go Project Structure

The backend follows Clean Architecture principles with clear separation of concerns:

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP request handlers
│   │   ├── middleware/          # Auth, RBAC, logging, security middleware
│   │   ├── routes/              # Route definitions
│   │   ├── dto/                 # Data Transfer Objects
│   │   └── adapter/             # Framework adapters
│   ├── domain/
│   │   ├── entities/            # Business entities (Environment, User, Log)
│   │   └── errors/              # Domain-specific errors
│   ├── service/
│   │   ├── environment/         # Environment business logic
│   │   ├── health/              # Health checking service
│   │   ├── ssh/                 # SSH operations service
│   │   ├── auth/                # Authentication service
│   │   ├── user/                # User management service
│   │   └── log/                 # Audit logging service
│   ├── repository/
│   │   ├── mongodb/             # MongoDB implementations
│   │   └── interfaces/          # Repository interfaces
│   ├── infrastructure/
│   │   ├── config/              # Configuration loading
│   │   └── database/            # Database connections
│   └── websocket/
│       ├── hub/                 # WebSocket connection hub
│       └── client/              # WebSocket client handling
├── config/
│   └── config.yaml              # Default configuration
├── go.mod
└── go.sum
```

## Core Components

### 1. Domain Entities

```go
// internal/domain/entities/environment.go
type Environment struct {
    ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
    Name           string                 `bson:"name" json:"name"`
    Description    string                 `bson:"description" json:"description"`
    EnvironmentURL string                 `bson:"environmentURL" json:"environmentURL"`
    Target         Target                 `bson:"target" json:"target"`
    Credentials    CredentialRef          `bson:"credentials" json:"credentials"`
    HealthCheck    HealthCheckConfig      `bson:"healthCheck" json:"healthCheck"`
    Status         Status                 `bson:"status" json:"status"`
    SystemInfo     SystemInfo             `bson:"systemInfo" json:"systemInfo"`
    Timestamps     Timestamps             `bson:"timestamps" json:"timestamps"`
    Commands       CommandConfig          `bson:"commands" json:"commands"`
    UpgradeConfig  UpgradeConfig          `bson:"upgradeConfig" json:"upgradeConfig"`
    Metadata       map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// Command configuration supports both SSH and HTTP
type CommandConfig struct {
    Type    CommandType    `bson:"type" json:"type"`       // "ssh" or "http"
    Restart CommandDetails `bson:"restart" json:"restart"`
}

type CommandDetails struct {
    Command string                 `bson:"command,omitempty" json:"command,omitempty"` // SSH
    URL     string                 `bson:"url,omitempty" json:"url,omitempty"`         // HTTP
    Method  string                 `bson:"method,omitempty" json:"method,omitempty"`   // HTTP
    Headers map[string]string      `bson:"headers,omitempty" json:"headers,omitempty"` // HTTP
    Body    map[string]interface{} `bson:"body,omitempty" json:"body,omitempty"`       // HTTP
}

// internal/domain/entities/user.go
type User struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username     string             `bson:"username" json:"username"`
    PasswordHash string             `bson:"passwordHash" json:"-"` // never serialized
    Role         UserRole           `bson:"role" json:"role"`      // admin | user | viewer
    Active       bool               `bson:"active" json:"active"`
    LastLogin    *time.Time         `bson:"lastLogin,omitempty" json:"lastLogin,omitempty"`
    CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// internal/domain/entities/log.go
type Log struct {
    ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
    Timestamp       time.Time              `bson:"timestamp" json:"timestamp"`
    EnvironmentID   *primitive.ObjectID    `bson:"environmentId,omitempty" json:"environmentId,omitempty"`
    EnvironmentName string                 `bson:"environmentName,omitempty" json:"environmentName,omitempty"`
    UserID          *primitive.ObjectID    `bson:"userId,omitempty" json:"userId,omitempty"`
    Username        string                 `bson:"username,omitempty" json:"username,omitempty"`
    Type            LogType                `bson:"type" json:"type"`
    Level           LogLevel               `bson:"level" json:"level"`
    Action          ActionType             `bson:"action,omitempty" json:"action,omitempty"`
    Message         string                 `bson:"message" json:"message"`
    Details         map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
}
```

### 2. Repository Interfaces

```go
// internal/repository/interfaces/environment.go
type EnvironmentRepository interface {
    Create(ctx context.Context, env *entities.Environment) error
    GetByID(ctx context.Context, id string) (*entities.Environment, error)
    GetByName(ctx context.Context, name string) (*entities.Environment, error)
    List(ctx context.Context, opts ListOptions) ([]*entities.Environment, int64, error)
    Update(ctx context.Context, id string, env *entities.Environment) error
    UpdateStatus(ctx context.Context, id string, status entities.Status) error
    Delete(ctx context.Context, id string) error
    Count(ctx context.Context) (int64, error)
}

// internal/repository/interfaces/user.go
type UserRepository interface {
    Create(ctx context.Context, user *entities.User) error
    GetByID(ctx context.Context, id string) (*entities.User, error)
    GetByUsername(ctx context.Context, username string) (*entities.User, error)
    List(ctx context.Context, opts ListOptions) ([]*entities.User, int64, error)
    Update(ctx context.Context, id string, user *entities.User) error
    UpdatePassword(ctx context.Context, id string, passwordHash string) error
    Delete(ctx context.Context, id string) error
}

// internal/repository/interfaces/log.go
type LogRepository interface {
    Create(ctx context.Context, log *entities.Log) error
    GetByID(ctx context.Context, id string) (*entities.Log, error)
    List(ctx context.Context, filter LogFilter) ([]*entities.Log, int64, error)
    Count(ctx context.Context, filter LogFilter) (int64, error)
    GetByType(ctx context.Context) (map[string]int64, error)
    GetByLevel(ctx context.Context) (map[string]int64, error)
}
```

### 3. Service Layer

```go
// internal/service/environment/service.go
type Service struct {
    repo          interfaces.EnvironmentRepository
    logService    *log.Service
    sshManager    *ssh.Manager
    healthChecker *health.Checker
    logger        *logrus.Logger
}

func (s *Service) RestartEnvironment(ctx context.Context, envID string, force bool) error {
    env, err := s.repo.GetByID(ctx, envID)
    if err != nil {
        return err
    }

    s.logService.LogAction(ctx, &log.ActionLog{
        EnvironmentID:   &env.ID,
        EnvironmentName: env.Name,
        Action:          entities.ActionTypeRestart,
        Level:           entities.LogLevelInfo,
        Message:         "Restarting environment",
    })

    switch env.Commands.Type {
    case entities.CommandTypeSSH:
        return s.executeSSHCommand(ctx, env, env.Commands.Restart.Command)
    case entities.CommandTypeHTTP:
        return s.executeHTTPCommand(ctx, env, env.Commands.Restart)
    default:
        return fmt.Errorf("unsupported command type: %s", env.Commands.Type)
    }
}
```

### 4. HTTP Handlers

```go
// internal/api/handlers/environment.go
type EnvironmentHandler struct {
    service    *environment.Service
    logService *log.Service
    logger     *logrus.Logger
}

func (h *EnvironmentHandler) List(w http.ResponseWriter, r *http.Request) {
    filter := parseListFilter(r)

    envs, total, err := h.service.ListEnvironments(r.Context(), filter)
    if err != nil {
        h.respondError(w, err)
        return
    }

    h.respondJSON(w, http.StatusOK, map[string]interface{}{
        "environments": envs,
        "pagination": map[string]interface{}{
            "page":       filter.Page,
            "limit":      filter.Limit,
            "total":      total,
            "totalPages": (total + filter.Limit - 1) / filter.Limit,
        },
    })
}
```

### 5. Middleware

```go
// internal/api/middleware/mux_auth.go — JWT validation
func MuxAuthMiddleware(jwtSecret string, logger *logrus.Logger) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("unexpected signing method")
                }
                return []byte(jwtSecret), nil
            })

            if err != nil || !token.Valid {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            claims, ok := token.Claims.(jwt.MapClaims)
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            ctx := ctxutil.WithUser(r.Context(), claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// internal/api/middleware/auth.go — RBAC admin enforcement
func RequireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        role := ctxutil.GetRole(r.Context())
        if role != string(entities.UserRoleAdmin) {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// internal/api/middleware/logging.go — request logging
func LoggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

            next.ServeHTTP(wrapped, r)

            logger.WithFields(logrus.Fields{
                "method":      r.Method,
                "path":        r.URL.Path,
                "status":      wrapped.statusCode,
                "duration":    time.Since(start),
                "remote_addr": r.RemoteAddr,
            }).Info("HTTP request completed")
        })
    }
}
```

### 6. WebSocket Hub

```go
// internal/websocket/hub/hub.go
type Hub struct {
    clients    map[string]*client.Client
    broadcast  chan interface{}
    register   chan *client.Client
    unregister chan *client.Client
    logger     *logrus.Logger
    mu         sync.RWMutex
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client.ID] = client
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client.ID]; ok {
                delete(h.clients, client.ID)
                close(client.Send)
            }
            h.mu.Unlock()

        case message := <-h.broadcast:
            h.broadcastMessage(message)
        }
    }
}

func (h *Hub) BroadcastEnvironmentUpdate(envID string, status interface{}) {
    h.broadcast <- map[string]interface{}{
        "type": "status_update",
        "payload": map[string]interface{}{
            "environmentId": envID,
            "status":        status,
        },
    }
}
```

### 7. Configuration

```go
// internal/infrastructure/config/config.go
type Config struct {
    Server       ServerConfig   `yaml:"server"`
    Database     DatabaseConfig `yaml:"database"`
    JWT          JWTConfig      `yaml:"jwt"`
    AllowOrigins []string       `yaml:"allowOrigins"`
}

type ServerConfig struct {
    Port            string        `yaml:"port" envconfig:"PORT"`
    ReadTimeout     time.Duration `yaml:"readTimeout"`
    WriteTimeout    time.Duration `yaml:"writeTimeout"`
    ShutdownTimeout time.Duration `yaml:"shutdownTimeout"`
}

type DatabaseConfig struct {
    URI      string `yaml:"uri" envconfig:"MONGODB_URI"`
    Database string `yaml:"database" envconfig:"MONGODB_DATABASE"`
}

type JWTConfig struct {
    Secret     string        `yaml:"secret" envconfig:"JWT_SECRET"`
    ExpiryTime time.Duration `yaml:"expiryTime"`
}
```

## Error Handling

```go
// internal/domain/errors/errors.go
var (
    ErrNotFound           = errors.New("resource not found")
    ErrAlreadyExists      = errors.New("resource already exists")
    ErrValidation         = errors.New("validation error")
    ErrUnauthorized       = errors.New("unauthorized")
    ErrForbidden          = errors.New("forbidden")
    ErrInternalServer     = errors.New("internal server error")
    ErrInvalidCredentials = errors.New("invalid credentials")
)

func RespondError(w http.ResponseWriter, err error) {
    var status int
    var code string

    switch {
    case errors.Is(err, ErrNotFound):
        status, code = http.StatusNotFound, "NOT_FOUND"
    case errors.Is(err, ErrAlreadyExists):
        status, code = http.StatusConflict, "ALREADY_EXISTS"
    case errors.Is(err, ErrValidation):
        status, code = http.StatusBadRequest, "VALIDATION_ERROR"
    case errors.Is(err, ErrUnauthorized):
        status, code = http.StatusUnauthorized, "UNAUTHORIZED"
    case errors.Is(err, ErrForbidden):
        status, code = http.StatusForbidden, "FORBIDDEN"
    default:
        status, code = http.StatusInternalServerError, "INTERNAL_ERROR"
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": map[string]interface{}{"code": code, "message": err.Error()},
    })
}
```

## Testing Strategy

```go
// internal/service/environment/service_test.go
func TestService_RestartEnvironment(t *testing.T) {
    mockRepo := &mockEnvironmentRepository{}
    mockLogService := &mockLogService{}

    service := &Service{
        repo:       mockRepo,
        logService: mockLogService,
        logger:     logrus.New(),
    }

    env := &entities.Environment{
        Name: "test-env",
        Commands: entities.CommandConfig{
            Type:    entities.CommandTypeSSH,
            Restart: entities.CommandDetails{Command: "sudo systemctl restart app"},
        },
    }

    mockRepo.On("GetByID", mock.Anything, "test-id").Return(env, nil)
    mockLogService.On("LogAction", mock.Anything, mock.Anything).Return(nil)

    err := service.RestartEnvironment(context.Background(), "test-id", false)

    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

## Performance Considerations

1. **Database** — Indexed fields on `name`, `status.health`, `timestamps.createdAt`; connection pooling with configurable limits; efficient pagination
2. **Concurrency** — Goroutine pools for health checks; non-blocking WebSocket broadcasts; async log writes
3. **Resource management** — SSH connection reuse; HTTP client connection pooling; graceful shutdown with timeout

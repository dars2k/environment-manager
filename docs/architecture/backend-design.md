# Backend Architecture Design

## Go Project Structure

The backend follows Clean Architecture principles with clear separation of concerns:

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/                    # Private application code
│   ├── api/
│   │   ├── handlers/           # HTTP request handlers
│   │   ├── middleware/         # HTTP middleware (auth, logging, recovery)
│   │   ├── routes/             # Route definitions
│   │   ├── dto/                # Data Transfer Objects
│   │   └── adapter/            # Framework adapters (Gin adapter)
│   ├── domain/
│   │   ├── entities/           # Business entities
│   │   └── errors/             # Domain-specific errors
│   ├── service/
│   │   ├── environment/        # Environment business logic
│   │   ├── health/             # Health checking service
│   │   ├── ssh/                # SSH operations service
│   │   ├── auth/               # Authentication service
│   │   ├── user/               # User management service
│   │   └── log/                # Logging service
│   ├── repository/
│   │   ├── mongodb/            # MongoDB implementations
│   │   └── interfaces/         # Repository interfaces
│   ├── infrastructure/
│   │   ├── config/             # Configuration loading
│   │   └── database/           # Database connections
│   └── websocket/
│       ├── hub/                # WebSocket connection hub
│       └── client/             # WebSocket client handling
├── config/
│   └── config.yaml             # Default configuration
├── go.mod                      # Go module definition
└── go.sum                      # Go module checksums
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
    UpgradeConfig  UpgradeConfig         `bson:"upgradeConfig" json:"upgradeConfig"`
    Metadata       map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// Command configuration supports both SSH and HTTP
type CommandConfig struct {
    Type    CommandType    `bson:"type" json:"type"`        // "ssh" or "http"
    Restart CommandDetails `bson:"restart" json:"restart"`
}

type CommandDetails struct {
    Command string                 `bson:"command,omitempty" json:"command,omitempty"` // For SSH
    URL     string                 `bson:"url,omitempty" json:"url,omitempty"`         // For HTTP
    Method  string                 `bson:"method,omitempty" json:"method,omitempty"`   // For HTTP
    Headers map[string]string      `bson:"headers,omitempty" json:"headers,omitempty"` // For HTTP
    Body    map[string]interface{} `bson:"body,omitempty" json:"body,omitempty"`       // For HTTP
}

// internal/domain/entities/user.go
type User struct {
    ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username       string             `bson:"username" json:"username"`
    Email          string             `bson:"email" json:"email"`
    PasswordHash   string             `bson:"passwordHash" json:"-"`
    Role           UserRole           `bson:"role" json:"role"`
    Active         bool               `bson:"active" json:"active"`
    LastLogin      *time.Time         `bson:"lastLogin,omitempty" json:"lastLogin,omitempty"`
    CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
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
    GetByEmail(ctx context.Context, email string) (*entities.User, error)
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
    repo         interfaces.EnvironmentRepository
    logService   *log.Service
    sshManager   *ssh.Manager
    healthChecker *health.Checker
    logger       *logrus.Logger
}

func (s *Service) RestartEnvironment(ctx context.Context, envID string, force bool) error {
    env, err := s.repo.GetByID(ctx, envID)
    if err != nil {
        return err
    }

    // Log the action
    s.logService.LogAction(ctx, &log.ActionLog{
        EnvironmentID: &env.ID,
        EnvironmentName: env.Name,
        Action: entities.ActionTypeRestart,
        Level: entities.LogLevelInfo,
        Message: "Restarting environment",
    })

    // Execute based on command type
    switch env.Commands.Type {
    case entities.CommandTypeSSH:
        return s.executeSSHCommand(ctx, env, env.Commands.Restart.Command)
    case entities.CommandTypeHTTP:
        return s.executeHTTPCommand(ctx, env, env.Commands.Restart)
    default:
        return fmt.Errorf("unsupported command type: %s", env.Commands.Type)
    }
}

// internal/service/auth/service.go
type Service struct {
    userRepo   interfaces.UserRepository
    logService *log.Service
    jwtSecret  string
    logger     *logrus.Logger
}

func (s *Service) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
    user, err := s.userRepo.GetByUsername(ctx, username)
    if err != nil {
        s.logService.LogAuth(ctx, username, entities.ActionTypeLogin, false, "User not found")
        return nil, ErrInvalidCredentials
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        s.logService.LogAuth(ctx, username, entities.ActionTypeLogin, false, "Invalid password")
        return nil, ErrInvalidCredentials
    }

    token, expiresAt, err := s.generateJWT(user)
    if err != nil {
        return nil, err
    }

    s.logService.LogAuth(ctx, username, entities.ActionTypeLogin, true, "Login successful")

    return &LoginResponse{
        Token:     token,
        User:      user,
        ExpiresAt: expiresAt,
    }, nil
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
    // Parse query parameters
    filter := parseListFilter(r)
    
    // Get environments
    envs, total, err := h.service.ListEnvironments(r.Context(), filter)
    if err != nil {
        h.respondError(w, err)
        return
    }
    
    // Build response
    response := map[string]interface{}{
        "environments": envs,
        "pagination": map[string]interface{}{
            "page":       filter.Page,
            "limit":      filter.Limit,
            "total":      total,
            "totalPages": (total + filter.Limit - 1) / filter.Limit,
        },
    }
    
    h.respondJSON(w, http.StatusOK, response)
}

func (h *EnvironmentHandler) Restart(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    envID := vars["id"]
    
    var req dto.RestartRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondError(w, NewValidationError("Invalid request body"))
        return
    }
    
    if err := h.service.RestartEnvironment(r.Context(), envID, req.Force); err != nil {
        h.respondError(w, err)
        return
    }
    
    h.respondJSON(w, http.StatusOK, map[string]interface{}{
        "operationId": generateOperationID(),
        "status":      "in_progress",
        "startedAt":   time.Now(),
    })
}
```

### 5. Middleware

```go
// internal/api/middleware/mux_auth.go
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

            ctx := context.WithValue(r.Context(), "user", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// internal/api/middleware/logging.go
func LoggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            wrapped := &responseWriter{
                ResponseWriter: w,
                statusCode:    http.StatusOK,
            }
            
            next.ServeHTTP(wrapped, r)
            
            logger.WithFields(logrus.Fields{
                "method":       r.Method,
                "path":         r.URL.Path,
                "status":       wrapped.statusCode,
                "duration":     time.Since(start),
                "remote_addr":  r.RemoteAddr,
            }).Info("HTTP request completed")
        })
    }
}
```

### 6. WebSocket Implementation

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
            h.logger.WithField("clientID", client.ID).Info("Client registered")
            
        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client.ID]; ok {
                delete(h.clients, client.ID)
                close(client.Send)
            }
            h.mu.Unlock()
            h.logger.WithField("clientID", client.ID).Info("Client unregistered")
            
        case message := <-h.broadcast:
            h.broadcastMessage(message)
        }
    }
}

// Broadcast environment status update
func (h *Hub) BroadcastEnvironmentUpdate(envID string, status interface{}) {
    message := map[string]interface{}{
        "type": "status_update",
        "payload": map[string]interface{}{
            "environmentId": envID,
            "status":        status,
        },
    }
    h.broadcast <- message
}
```

### 7. Configuration Management

```go
// internal/infrastructure/config/config.go
type Config struct {
    Server       ServerConfig       `yaml:"server"`
    Database     DatabaseConfig     `yaml:"database"`
    JWT          JWTConfig          `yaml:"jwt"`
    AllowOrigins []string           `yaml:"allowOrigins"`
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

func Load(configPath string) (*Config, error) {
    config := &Config{}
    
    // Load from file if provided
    if configPath != "" {
        file, err := os.Open(configPath)
        if err != nil {
            return nil, err
        }
        defer file.Close()
        
        if err := yaml.NewDecoder(file).Decode(config); err != nil {
            return nil, err
        }
    }
    
    // Override with environment variables
    if err := envconfig.Process("", config); err != nil {
        return nil, err
    }
    
    // Set defaults
    config.setDefaults()
    
    return config, nil
}
```

## Error Handling

```go
// internal/domain/errors/errors.go
var (
    ErrNotFound             = errors.New("resource not found")
    ErrAlreadyExists        = errors.New("resource already exists")
    ErrValidation           = errors.New("validation error")
    ErrUnauthorized         = errors.New("unauthorized")
    ErrInternalServer       = errors.New("internal server error")
    ErrInvalidCredentials   = errors.New("invalid credentials")
)

// HTTP error response handling
func RespondError(w http.ResponseWriter, err error) {
    var status int
    var code string
    
    switch {
    case errors.Is(err, ErrNotFound):
        status = http.StatusNotFound
        code = "NOT_FOUND"
    case errors.Is(err, ErrAlreadyExists):
        status = http.StatusConflict
        code = "ALREADY_EXISTS"
    case errors.Is(err, ErrValidation):
        status = http.StatusBadRequest
        code = "VALIDATION_ERROR"
    case errors.Is(err, ErrUnauthorized):
        status = http.StatusUnauthorized
        code = "UNAUTHORIZED"
    default:
        status = http.StatusInternalServerError
        code = "INTERNAL_ERROR"
    }
    
    response := map[string]interface{}{
        "error": map[string]interface{}{
            "code":    code,
            "message": err.Error(),
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(response)
}
```

## Testing Strategy

### Unit Tests

```go
// internal/service/environment/service_test.go
func TestService_CreateEnvironment(t *testing.T) {
    // Mock repository implementation
    mockRepo := &mockEnvironmentRepository{}
    mockLogService := &mockLogService{}
    
    service := &Service{
        repo:       mockRepo,
        logService: mockLogService,
        logger:     logrus.New(),
    }
    
    env := &entities.Environment{
        Name: "test-env",
        Target: entities.Target{
            Host: "192.168.1.100",
            Port: 22,
        },
    }
    
    // Set expectations on mock
    mockRepo.On("GetByName", mock.Anything, "test-env").Return(nil, nil)
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Environment")).Return(nil)
    mockLogService.On("LogAction", mock.Anything, mock.Anything).Return(nil)
    
    err := service.CreateEnvironment(context.Background(), env)
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
    mockLogService.AssertExpectations(t)
}
```

### Integration Tests

```go
// internal/api/handlers/environment_test.go
func TestEnvironmentHandler_List(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Create test data
    repo := mongodb.NewEnvironmentRepository(db)
    createTestEnvironments(t, repo)
    
    // Setup handler
    handler := &EnvironmentHandler{
        service: environment.NewService(repo, nil, nil),
        logger:  logrus.New(),
    }
    
    // Make request
    req := httptest.NewRequest("GET", "/api/v1/environments", nil)
    w := httptest.NewRecorder()
    
    handler.List(w, req)
    
    // Assert response
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    
    envs := response["environments"].([]interface{})
    assert.Len(t, envs, 3)
}
```

## Performance Considerations

1. **Database Optimization**
   - Indexed fields: name, status.health, timestamps.createdAt
   - Connection pooling with configurable limits
   - Efficient pagination queries

2. **Caching Strategy**
   - Environment status caching (future)
   - User session caching (future)
   - Configuration caching

3. **Concurrent Operations**
   - Goroutine pools for health checks
   - Non-blocking WebSocket broadcasts
   - Async log writing

4. **Resource Management**
   - SSH connection reuse
   - HTTP client connection pooling
   - Graceful shutdown handling

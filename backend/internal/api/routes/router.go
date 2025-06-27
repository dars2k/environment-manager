package routes

import (
	"fmt"
	"net/http"
	"time"

	"app-env-manager/internal/api/adapter"
	"app-env-manager/internal/api/handlers"
	"app-env-manager/internal/api/middleware"
	"app-env-manager/internal/websocket/hub"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

// Config contains router configuration
type Config struct {
	EnvironmentHandler *handlers.EnvironmentHandler
	LogHandler         *handlers.LogHandler
	AuthHandler        *handlers.AuthHandler
	UserHandler        *handlers.UserHandler
	AuthService        interface{}
	UserService        interface{}
	WebSocketHub       *hub.Hub
	Logger             *logrus.Logger
	JWTSecret          string
	AllowedOrigins     []string
}

// NewRouter creates and configures a new router
func NewRouter(cfg Config) http.Handler {
	r := mux.NewRouter()

	// WebSocket endpoint (registered before middleware to avoid wrapping issues)
	r.HandleFunc("/ws", HandleWebSocket(cfg.WebSocketHub, cfg.Logger)).Methods("GET")

	// Setup middleware for non-WebSocket routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.LoggingMiddleware(cfg.Logger))
	apiRouter.Use(middleware.RecoveryMiddleware(cfg.Logger))

	// API routes
	api := apiRouter.PathPrefix("/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", HealthCheck).Methods("GET")

	// Public routes (no auth required)
	api.HandleFunc("/auth/login", adapter.GinHandlerAdapter(cfg.AuthHandler.Login)).Methods("POST")

	// Protected API routes (require authentication)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.MuxAuthMiddleware(cfg.JWTSecret, cfg.Logger))

	// Auth routes
	protected.HandleFunc("/auth/me", adapter.GinHandlerAdapter(cfg.AuthHandler.GetCurrentUser)).Methods("GET")
	protected.HandleFunc("/auth/logout", adapter.GinHandlerAdapter(cfg.AuthHandler.Logout)).Methods("POST")

	// User routes
	userRoutes := protected.PathPrefix("/users").Subrouter()
	userRoutes.HandleFunc("", adapter.GinHandlerAdapter(cfg.UserHandler.List)).Methods("GET")
	userRoutes.HandleFunc("", adapter.GinHandlerAdapter(cfg.UserHandler.Create)).Methods("POST")
	userRoutes.HandleFunc("/{id}", adapter.GinHandlerAdapter(cfg.UserHandler.Get)).Methods("GET")
	userRoutes.HandleFunc("/{id}", adapter.GinHandlerAdapter(cfg.UserHandler.Update)).Methods("PUT")
	userRoutes.HandleFunc("/{id}", adapter.GinHandlerAdapter(cfg.UserHandler.Delete)).Methods("DELETE")
	userRoutes.HandleFunc("/password/change", adapter.GinHandlerAdapter(cfg.UserHandler.ChangePassword)).Methods("POST")
	userRoutes.HandleFunc("/{id}/password/reset", adapter.GinHandlerAdapter(cfg.UserHandler.ResetPassword)).Methods("POST")

	// Environment routes
	envRoutes := protected.PathPrefix("/environments").Subrouter()
	envRoutes.HandleFunc("", cfg.EnvironmentHandler.List).Methods("GET")
	envRoutes.HandleFunc("", cfg.EnvironmentHandler.Create).Methods("POST")
	envRoutes.HandleFunc("/{id}", cfg.EnvironmentHandler.Get).Methods("GET")
	envRoutes.HandleFunc("/{id}", cfg.EnvironmentHandler.Update).Methods("PUT")
	envRoutes.HandleFunc("/{id}", cfg.EnvironmentHandler.Delete).Methods("DELETE")
	
	// Environment operations
	envRoutes.HandleFunc("/{id}/restart", cfg.EnvironmentHandler.Restart).Methods("POST")
	envRoutes.HandleFunc("/{id}/check-health", cfg.EnvironmentHandler.CheckHealth).Methods("POST")
	envRoutes.HandleFunc("/{id}/versions", cfg.EnvironmentHandler.GetVersions).Methods("GET")
	envRoutes.HandleFunc("/{id}/upgrade", cfg.EnvironmentHandler.Upgrade).Methods("POST")
	envRoutes.HandleFunc("/{id}/logs", adapter.GinHandlerAdapter(cfg.LogHandler.GetEnvironmentLogs)).Methods("GET")

	// Log routes
	logRoutes := protected.PathPrefix("/logs").Subrouter()
	logRoutes.HandleFunc("", adapter.GinHandlerAdapter(cfg.LogHandler.List)).Methods("GET")
	logRoutes.HandleFunc("/count", adapter.GinHandlerAdapter(cfg.LogHandler.Count)).Methods("GET")
	logRoutes.HandleFunc("/{id}", adapter.GinHandlerAdapter(cfg.LogHandler.GetByID)).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           86400,
	})

	return c.Handler(r)
}

// HealthCheck handles health check requests
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","version":"1.0.0"}`))
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(wsHub *hub.Hub, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP connection to WebSocket
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper origin checking
				return true
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.WithError(err).Error("Failed to upgrade connection")
			return
		}

		// Create client
		clientID := generateClientID()
		client := hub.NewClient(clientID, conn, wsHub, logger)

		// Register client
		wsHub.RegisterClient(client)

		// Start client pumps
		go client.WritePump()
		go client.ReadPump()
	}
}

func generateClientID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}

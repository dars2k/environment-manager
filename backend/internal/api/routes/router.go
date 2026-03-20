package routes

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"app-env-manager/internal/api/adapter"
	"app-env-manager/internal/api/handlers"
	"app-env-manager/internal/api/middleware"
	"app-env-manager/internal/websocket/hub"
	"github.com/golang-jwt/jwt/v5"
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

	// Login rate limiter: 10 attempts per minute per IP
	loginRateLimiter := middleware.NewRateLimiter(10, time.Minute)

	// WebSocket endpoint – registered before the API sub-router so the
	// gorilla/mux middleware chain does not interfere with the upgrade.
	r.HandleFunc("/ws", HandleWebSocket(cfg.WebSocketHub, cfg.Logger, cfg.JWTSecret, cfg.AllowedOrigins)).Methods("GET")

	// Setup middleware for all API routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.SecurityHeadersMiddleware())
	apiRouter.Use(middleware.LoggingMiddleware(cfg.Logger))
	apiRouter.Use(middleware.RecoveryMiddleware(cfg.Logger))

	// API routes
	api := apiRouter.PathPrefix("/v1").Subrouter()

	// Health check (no auth, no rate limit)
	api.HandleFunc("/health", HealthCheck).Methods("GET")

	// Login – rate-limited, no auth required
	api.Handle("/auth/login",
		middleware.RateLimitMiddleware(loginRateLimiter)(
			http.HandlerFunc(adapter.GinHandlerAdapter(cfg.AuthHandler.Login)),
		),
	).Methods("POST")

	// Protected API routes (require authentication)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.MuxAuthMiddleware(cfg.JWTSecret, cfg.Logger))

	// Auth routes
	protected.HandleFunc("/auth/me", adapter.GinHandlerAdapter(cfg.AuthHandler.GetCurrentUser)).Methods("GET")
	protected.HandleFunc("/auth/logout", adapter.GinHandlerAdapter(cfg.AuthHandler.Logout)).Methods("POST")

	// User routes — all require admin
	userRoutes := protected.PathPrefix("/users").Subrouter()
	userRoutes.Use(middleware.RequireAdmin)
	userRoutes.HandleFunc("", adapter.GinHandlerAdapter(cfg.UserHandler.List)).Methods("GET")
	userRoutes.HandleFunc("", adapter.GinHandlerAdapter(cfg.UserHandler.Create)).Methods("POST")
	userRoutes.HandleFunc("/{id}", adapter.GinHandlerAdapter(cfg.UserHandler.Get)).Methods("GET")
	userRoutes.HandleFunc("/{id}", adapter.GinHandlerAdapter(cfg.UserHandler.Update)).Methods("PUT")
	userRoutes.HandleFunc("/{id}", adapter.GinHandlerAdapter(cfg.UserHandler.Delete)).Methods("DELETE")
	userRoutes.HandleFunc("/{id}/password/reset", adapter.GinHandlerAdapter(cfg.UserHandler.ResetPassword)).Methods("POST")

	// Password change is self-service — any authenticated user may call it
	protected.HandleFunc("/users/password/change", adapter.GinHandlerAdapter(cfg.UserHandler.ChangePassword)).Methods("POST")

	// Environment routes
	envRoutes := protected.PathPrefix("/environments").Subrouter()

	// Read-only: any authenticated user
	envRoutes.HandleFunc("", cfg.EnvironmentHandler.List).Methods("GET")
	envRoutes.HandleFunc("/{id}", cfg.EnvironmentHandler.Get).Methods("GET")
	envRoutes.HandleFunc("/{id}/versions", cfg.EnvironmentHandler.GetVersions).Methods("GET")
	envRoutes.HandleFunc("/{id}/logs", adapter.GinHandlerAdapter(cfg.LogHandler.GetEnvironmentLogs)).Methods("GET")

	// Operator actions: any authenticated user
	envRoutes.HandleFunc("/{id}/restart", cfg.EnvironmentHandler.Restart).Methods("POST")
	envRoutes.HandleFunc("/{id}/upgrade", cfg.EnvironmentHandler.Upgrade).Methods("POST")
	envRoutes.HandleFunc("/{id}/check-health", cfg.EnvironmentHandler.CheckHealth).Methods("POST")

	// Mutating CRUD: admin only
	envRoutes.Handle("", middleware.RequireAdmin(http.HandlerFunc(cfg.EnvironmentHandler.Create))).Methods("POST")
	envRoutes.Handle("/{id}", middleware.RequireAdmin(http.HandlerFunc(cfg.EnvironmentHandler.Update))).Methods("PUT")
	envRoutes.Handle("/{id}", middleware.RequireAdmin(http.HandlerFunc(cfg.EnvironmentHandler.Delete))).Methods("DELETE")

	// Log routes
	logRoutes := protected.PathPrefix("/logs").Subrouter()
	logRoutes.HandleFunc("", adapter.GinHandlerAdapter(cfg.LogHandler.List)).Methods("GET")
	logRoutes.HandleFunc("/count", adapter.GinHandlerAdapter(cfg.LogHandler.Count)).Methods("GET")
	logRoutes.HandleFunc("/{id}", adapter.GinHandlerAdapter(cfg.LogHandler.GetByID)).Methods("GET")

	// Setup CORS with explicit allowed headers instead of wildcard
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept", "X-Requested-With"},
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

// HandleWebSocket handles WebSocket upgrade requests.
// It validates the JWT token supplied as the ?token= query parameter and
// verifies the request Origin against the configured allowed origins before
// upgrading the connection.
func HandleWebSocket(wsHub *hub.Hub, logger *logrus.Logger, jwtSecret string, allowedOrigins []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authenticate via token query parameter (browser WebSocket API does
		// not support custom request headers).
		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return isOriginAllowed(r, allowedOrigins)
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.WithError(err).Error("Failed to upgrade WebSocket connection")
			return
		}

		clientID := generateClientID()
		client := hub.NewClient(clientID, conn, wsHub, logger)
		wsHub.RegisterClient(client)
		go client.WritePump()
		go client.ReadPump()
	}
}

// isOriginAllowed returns true when the request Origin header matches one of
// the configured allowed origins. Connections without an Origin header
// (e.g. server-to-server) are permitted.
func isOriginAllowed(r *http.Request, allowedOrigins []string) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	for _, allowed := range allowedOrigins {
		if allowed == origin {
			return true
		}
	}
	return false
}

// generateClientID returns a cryptographically random hex client identifier.
func generateClientID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback (should never happen in practice)
		return fmt.Sprintf("client-%d", time.Now().UnixNano())
	}
	return "client-" + hex.EncodeToString(b)
}

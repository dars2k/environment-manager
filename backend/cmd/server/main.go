package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app-env-manager/internal/api/handlers"
	"app-env-manager/internal/api/routes"
	"app-env-manager/internal/infrastructure/config"
	"app-env-manager/internal/infrastructure/database"
	"app-env-manager/internal/repository/mongodb"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/auth"
	"app-env-manager/internal/service/environment"
	"app-env-manager/internal/service/health"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/ssh"
	"app-env-manager/internal/service/user"
	"app-env-manager/internal/websocket/hub"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Connect to MongoDB
	mongoDB, err := database.NewMongoDB(
		cfg.Database.URI,
		cfg.Database.Database,
		cfg.Database.MaxConnections,
		cfg.Database.Timeout,
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to MongoDB")
	}
	defer mongoDB.Close(context.Background())

	// Create indexes
	if err := mongoDB.CreateIndexes(context.Background()); err != nil {
		logger.WithError(err).Error("Failed to create indexes")
	}

	// Initialize repositories
	envRepo := mongodb.NewEnvironmentRepository(mongoDB.Database())
	auditRepo := mongodb.NewAuditLogRepository(mongoDB.Database())
	logRepo := mongodb.NewLogRepository(mongoDB.Database())
	userRepo := mongodb.NewUserRepository(mongoDB.Database())

	// Initialize services
	sshManager := ssh.NewManager(ssh.Config{
		ConnectionTimeout: cfg.SSH.ConnectionTimeout,
		CommandTimeout:    cfg.SSH.CommandTimeout,
		MaxConnections:    cfg.SSH.MaxConnections,
	})
	defer sshManager.Close()

	healthChecker := health.NewChecker(cfg.Health.Timeout)
	
	logService := log.NewService(logRepo)
	
	authService := auth.NewService(
		userRepo,
		logService,
		cfg.Security.JWTSecret,
		24*time.Hour, // JWT expiry
	)
	
	userService := user.NewService(userRepo, logService)
	
	// Create initial admin user
	if err := authService.CreateInitialAdmin(context.Background()); err != nil {
		logger.WithError(err).Error("Failed to create initial admin user")
	}

	envService := environment.NewService(
		envRepo,
		auditRepo,
		sshManager,
		healthChecker,
		logService,
	)

	// Initialize WebSocket hub
	wsHub := hub.NewHub(logger)
	go wsHub.Run()

	// Initialize handlers
	envHandler := handlers.NewEnvironmentHandler(envService, wsHub, logger)
	logHandler := handlers.NewLogHandler(logService, logger)
	authHandler := handlers.NewAuthHandler(authService, logger)
	userHandler := handlers.NewUserHandler(userService, logger)

	// Setup routes
	router := routes.NewRouter(routes.Config{
		EnvironmentHandler: envHandler,
		LogHandler:        logHandler,
		AuthHandler:       authHandler,
		UserHandler:       userHandler,
		AuthService:       authService,
		UserService:       userService,
		WebSocketHub:      wsHub,
		Logger:            logger,
		JWTSecret:         cfg.Security.JWTSecret,
		AllowedOrigins:    cfg.Security.AllowedOrigins,
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start health check scheduler
	go startHealthCheckScheduler(envService, cfg.Health.CheckInterval, logger)

	// Start server
	go func() {
		logger.WithField("addr", srv.Addr).Info("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exiting")
}

// startHealthCheckScheduler runs periodic health checks on all environments
func startHealthCheckScheduler(service *environment.Service, interval time.Duration, logger *logrus.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			
			// Get all environments
			envs, err := service.ListEnvironments(ctx, interfaces.ListFilter{})
			if err != nil {
				logger.WithError(err).Error("Failed to list environments for health check")
				cancel()
				continue
			}

			// Check health for each environment
			for _, env := range envs {
				go func(envID string) {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					
					if err := service.CheckHealth(ctx, envID); err != nil {
						logger.WithFields(logrus.Fields{
							"environmentId": envID,
							"error":         err,
						}).Error("Health check failed")
					}
				}(env.ID.Hex())
			}
			
			cancel()
		}
	}
}

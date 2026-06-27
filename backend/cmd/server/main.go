package main

import (
	"context"
	"errors"
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
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/repository/mongodb"
	"app-env-manager/internal/service/auth"
	"app-env-manager/internal/service/environment"
	"app-env-manager/internal/service/health"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/ssh"
	"app-env-manager/internal/service/user"
	"app-env-manager/internal/websocket/hub"
	"github.com/sirupsen/logrus"
)

type mongoConnectFn func(uri, database string, maxConn int, timeout time.Duration) (*database.MongoDB, error)

func main() {
	logger := newLogger()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	if err := run(logger, "config/config.yaml", database.NewMongoDB, quit); err != nil {
		logger.WithError(err).Fatal("Server failed")
	}
}

func newLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	return logger
}

func run(logger *logrus.Logger, configPath string, connectMongo mongoConnectFn, quit <-chan os.Signal) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	mongoDB, err := connectMongo(cfg.Database.URI, cfg.Database.Database, cfg.Database.MaxConnections, cfg.Database.Timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer mongoDB.Close(context.Background())

	if err := mongoDB.CreateIndexes(context.Background()); err != nil {
		logger.WithError(err).Error("Failed to create indexes")
	}

	envRepo := mongodb.NewEnvironmentRepository(mongoDB.Database())
	auditRepo := mongodb.NewAuditLogRepository(mongoDB.Database())
	logRepo := mongodb.NewLogRepository(mongoDB.Database())
	userRepo := mongodb.NewUserRepository(mongoDB.Database())

	srv, wsHub, envService, sshMgr := buildServerApp(cfg, envRepo, auditRepo, logRepo, userRepo, logger)
	defer sshMgr.Close()

	go wsHub.Run()
	go startHealthCheckScheduler(envService, cfg.Health.CheckInterval, logger)

	runServer(srv, quit, logger)
	return nil
}

func buildServerApp(
	cfg *config.Config,
	envRepo interfaces.EnvironmentRepository,
	auditRepo interfaces.AuditLogRepository,
	logRepo interfaces.LogRepository,
	userRepo interfaces.UserRepository,
	logger *logrus.Logger,
) (*http.Server, *hub.Hub, *environment.Service, *ssh.Manager) {
	sshManager := ssh.NewManager(ssh.Config{
		ConnectionTimeout: cfg.SSH.ConnectionTimeout,
		CommandTimeout:    cfg.SSH.CommandTimeout,
		MaxConnections:    cfg.SSH.MaxConnections,
	})

	healthChecker := health.NewChecker(cfg.Health.Timeout)
	logService := log.NewService(logRepo)
	authService := auth.NewService(userRepo, logService, cfg.Security.JWTSecret, 24*time.Hour)
	userService := user.NewService(userRepo, logService)

	if err := authService.CreateInitialAdmin(context.Background()); err != nil {
		logger.WithError(err).Error("Failed to create initial admin user")
	}

	envService := environment.NewService(envRepo, auditRepo, sshManager, healthChecker, logService, cfg.Security.AllowedHosts)
	wsHub := hub.NewHub(logger)

	envHandler := handlers.NewEnvironmentHandler(envService, wsHub, logger)
	logHandler := handlers.NewLogHandler(logService, logger)
	authHandler := handlers.NewAuthHandler(authService, logger)
	userHandler := handlers.NewUserHandler(userService, logger)

	router := routes.NewRouter(routes.Config{
		EnvironmentHandler: envHandler,
		LogHandler:         logHandler,
		AuthHandler:        authHandler,
		UserHandler:        userHandler,
		AuthService:        authService,
		UserService:        userService,
		WebSocketHub:       wsHub,
		Logger:             logger,
		JWTSecret:          cfg.Security.JWTSecret,
		AllowedOrigins:     cfg.Security.AllowedOrigins,
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return srv, wsHub, envService, sshManager
}

func runServer(srv *http.Server, quit <-chan os.Signal, logger *logrus.Logger) {
	go func() {
		logger.WithField("addr", srv.Addr).Info("Starting server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exiting")
}

// startHealthCheckScheduler runs periodic health checks on all environments,
// respecting each environment's individual health check interval setting.
func startHealthCheckScheduler(service *environment.Service, defaultInterval time.Duration, logger *logrus.Logger) {
	// Poll frequently; actual per-environment cadence is controlled by lastChecked tracking.
	pollInterval := 5 * time.Second
	if defaultInterval < pollInterval {
		pollInterval = defaultInterval
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	lastChecked := make(map[string]time.Time)

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		envs, err := service.ListEnvironments(ctx, interfaces.ListFilter{})
		if err != nil {
			logger.WithError(err).Error("Failed to list environments for health check")
			cancel()
			continue
		}

		now := time.Now()
		for _, env := range envs {
			if !env.HealthCheck.Enabled {
				continue
			}

			// Determine this environment's interval (fall back to global default).
			interval := defaultInterval
			if env.HealthCheck.Interval > 0 {
				interval = time.Duration(env.HealthCheck.Interval) * time.Second
			}

			envID := env.ID.Hex()
			if t, ok := lastChecked[envID]; ok && now.Sub(t) < interval {
				continue // not due yet
			}
			lastChecked[envID] = now

			go func(id string) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				if err := service.CheckHealth(ctx, id); err != nil {
					logger.WithFields(logrus.Fields{
						"environmentId": id,
						"error":         err,
					}).Error("Health check failed")
				}
			}(envID)
		}

		cancel()
	}
}

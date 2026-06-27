package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"app-env-manager/internal/infrastructure/config"
	"app-env-manager/internal/infrastructure/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	logger := newLogger()
	require.NotNil(t, logger)
	_, ok := logger.Formatter.(*logrus.JSONFormatter)
	assert.True(t, ok, "expected JSON formatter")
	assert.Equal(t, logrus.InfoLevel, logger.Level)
}

func TestRun_ConfigError(t *testing.T) {
	// Empty JWT_SECRET and SSH_KEY_ENCRYPTION_KEY cause config.Validate() to fail
	t.Setenv("JWT_SECRET", "")
	t.Setenv("SSH_KEY_ENCRYPTION_KEY", "")

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	quit := make(chan os.Signal, 1)

	// Use empty config path so config falls back to defaults which will fail validation
	err := run(logger, "", func(uri, db string, maxConn int, timeout time.Duration) (*database.MongoDB, error) {
		return nil, errors.New("should not be called")
	}, quit)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

func TestRun_MongoConnectError(t *testing.T) {
	// Set valid env vars so config.Validate passes
	t.Setenv("JWT_SECRET", "testsecretkey")
	t.Setenv("SSH_KEY_ENCRYPTION_KEY", "12345678901234567890123456789012") // exactly 32 chars

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	quit := make(chan os.Signal, 1)

	mongoErr := fmt.Errorf("connection refused")
	err := run(logger, "", func(uri, db string, maxConn int, timeout time.Duration) (*database.MongoDB, error) {
		return nil, mongoErr
	}, quit)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to MongoDB")
}

func TestBuildServerApp(t *testing.T) {
	cfg := &config.Config{
		Server:   config.ServerConfig{Host: "127.0.0.1", Port: 0, ReadTimeout: time.Second, WriteTimeout: time.Second, IdleTimeout: time.Second},
		SSH:      config.SSHConfig{ConnectionTimeout: time.Second, CommandTimeout: time.Second, MaxConnections: 1},
		Health:   config.HealthConfig{Timeout: time.Second, CheckInterval: time.Second},
		Security: config.SecurityConfig{JWTSecret: "test", AllowedOrigins: []string{"http://localhost"}},
		WebSocket: config.WebSocketConfig{PingInterval: time.Second, PongTimeout: time.Second, WriteTimeout: time.Second, MaxMessageSize: 512},
	}

	envRepo := new(mockEnvRepoMain)
	auditRepo := new(mockAuditRepoMain)
	logRepo := new(mockLogRepoMain)
	userRepo := new(mockUserRepoMain)

	// Count returns 1 → users already exist, skip creation
	userRepo.On("Count", mock.Anything).Return(int64(1), nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	srv, wsHub, envService, sshMgr := buildServerApp(cfg, envRepo, auditRepo, logRepo, userRepo, logger)

	assert.NotNil(t, srv)
	assert.NotNil(t, wsHub)
	assert.NotNil(t, envService)
	assert.NotNil(t, sshMgr)

	sshMgr.Close()
}

func TestRunServer_Shutdown(t *testing.T) {
	srv := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: http.NewServeMux(),
	}

	quit := make(chan os.Signal, 1)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	done := make(chan struct{})
	go func() {
		defer close(done)
		runServer(srv, quit, logger)
	}()

	// Give the server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Send shutdown signal
	quit <- syscall.SIGTERM

	// Verify it shuts down within 5 seconds
	select {
	case <-done:
		// success
	case <-time.After(5 * time.Second):
		t.Fatal("runServer did not shut down within 5 seconds")
	}
}

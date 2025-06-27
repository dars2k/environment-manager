package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"app-env-manager/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultConfig(t *testing.T) {
	// Set required environment variables
	os.Setenv("JWT_SECRET", "test-jwt-secret-for-testing-only")
	os.Setenv("SSH_KEY_ENCRYPTION_KEY", "12345678901234567890123456789012")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SSH_KEY_ENCRYPTION_KEY")
	}()

	// Load config without file
	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check default values
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 120*time.Second, cfg.Server.IdleTimeout)

	assert.Equal(t, "mongodb://localhost:27017", cfg.Database.URI)
	assert.Equal(t, "app-env-manager", cfg.Database.Database)
	assert.Equal(t, 100, cfg.Database.MaxConnections)
	assert.Equal(t, 10*time.Second, cfg.Database.Timeout)

	assert.Equal(t, 30*time.Second, cfg.SSH.ConnectionTimeout)
	assert.Equal(t, 300*time.Second, cfg.SSH.CommandTimeout)
	assert.Equal(t, 50, cfg.SSH.MaxConnections)
	assert.Equal(t, "12345678901234567890123456789012", cfg.SSH.EncryptionKey)

	assert.Equal(t, 30*time.Second, cfg.Health.CheckInterval)
	assert.Equal(t, 5*time.Second, cfg.Health.Timeout)
	assert.Equal(t, 3, cfg.Health.MaxRetries)
	assert.Equal(t, 10, cfg.Health.ConcurrentChecks)

	assert.Equal(t, "test-jwt-secret-for-testing-only", cfg.Security.JWTSecret)
	assert.Equal(t, 24*time.Hour, cfg.Security.TokenExpiration)
	assert.Equal(t, 10, cfg.Security.BCryptCost)
	assert.Equal(t, []string{"http://localhost:3000"}, cfg.Security.AllowedOrigins)
	assert.Equal(t, 100, cfg.Security.RateLimitRequests)
	assert.Equal(t, 1*time.Minute, cfg.Security.RateLimitWindow)

	assert.Equal(t, 30*time.Second, cfg.WebSocket.PingInterval)
	assert.Equal(t, 60*time.Second, cfg.WebSocket.PongTimeout)
	assert.Equal(t, 10*time.Second, cfg.WebSocket.WriteTimeout)
	assert.Equal(t, int64(512*1024), cfg.WebSocket.MaxMessageSize)
}

func TestLoad_FromYAMLFile(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `
server:
  host: "127.0.0.1"
  port: 9090
  readTimeout: 60s
  writeTimeout: 60s
  idleTimeout: 180s

database:
  uri: "mongodb://test:27017"
  database: "test-db"
  maxConnections: 50
  timeout: 20s

ssh:
  connectionTimeout: 60s
  commandTimeout: 600s
  maxConnections: 100

health:
  checkInterval: 60s
  timeout: 10s
  maxRetries: 5
  concurrentChecks: 20

security:
  tokenExpiration: 48h
  bcryptCost: 12
  allowedOrigins:
    - "http://localhost:4000"
    - "https://example.com"
  rateLimitRequests: 200
  rateLimitWindow: 2m

websocket:
  pingInterval: 45s
  pongTimeout: 90s
  writeTimeout: 15s
  maxMessageSize: 1048576
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set required environment variables
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("SSH_KEY_ENCRYPTION_KEY", "12345678901234567890123456789012")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SSH_KEY_ENCRYPTION_KEY")
	}()

	// Load config from file
	cfg, err := config.Load(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check loaded values
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, 60*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 60*time.Second, cfg.Server.WriteTimeout)
	assert.Equal(t, 180*time.Second, cfg.Server.IdleTimeout)

	assert.Equal(t, "mongodb://test:27017", cfg.Database.URI)
	assert.Equal(t, "test-db", cfg.Database.Database)
	assert.Equal(t, 50, cfg.Database.MaxConnections)
	assert.Equal(t, 20*time.Second, cfg.Database.Timeout)

	assert.Equal(t, 60*time.Second, cfg.SSH.ConnectionTimeout)
	assert.Equal(t, 600*time.Second, cfg.SSH.CommandTimeout)
	assert.Equal(t, 100, cfg.SSH.MaxConnections)

	assert.Equal(t, 60*time.Second, cfg.Health.CheckInterval)
	assert.Equal(t, 10*time.Second, cfg.Health.Timeout)
	assert.Equal(t, 5, cfg.Health.MaxRetries)
	assert.Equal(t, 20, cfg.Health.ConcurrentChecks)

	assert.Equal(t, 48*time.Hour, cfg.Security.TokenExpiration)
	assert.Equal(t, 12, cfg.Security.BCryptCost)
	assert.Equal(t, []string{"http://localhost:4000", "https://example.com"}, cfg.Security.AllowedOrigins)
	assert.Equal(t, 200, cfg.Security.RateLimitRequests)
	assert.Equal(t, 2*time.Minute, cfg.Security.RateLimitWindow)

	assert.Equal(t, 45*time.Second, cfg.WebSocket.PingInterval)
	assert.Equal(t, 90*time.Second, cfg.WebSocket.PongTimeout)
	assert.Equal(t, 15*time.Second, cfg.WebSocket.WriteTimeout)
	assert.Equal(t, int64(1048576), cfg.WebSocket.MaxMessageSize)
}

func TestLoad_EnvironmentOverrides(t *testing.T) {
	// Set all environment variables
	envVars := map[string]string{
		"SERVER_HOST":            "192.168.1.1",
		"MONGODB_URI":            "mongodb://prod:27017",
		"MONGODB_DATABASE":       "prod-db",
		"JWT_SECRET":             "prod-jwt-secret",
		"SSH_KEY_ENCRYPTION_KEY": "12345678901234567890123456789012",
		"ALLOWED_ORIGINS":        "https://prod.example.com",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	// Load config
	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check environment overrides
	assert.Equal(t, "192.168.1.1", cfg.Server.Host)
	assert.Equal(t, "mongodb://prod:27017", cfg.Database.URI)
	assert.Equal(t, "prod-db", cfg.Database.Database)
	assert.Equal(t, "prod-jwt-secret", cfg.Security.JWTSecret)
	assert.Equal(t, "12345678901234567890123456789012", cfg.SSH.EncryptionKey)
	assert.Equal(t, []string{"https://prod.example.com"}, cfg.Security.AllowedOrigins)
}

func TestLoad_InvalidYAML(t *testing.T) {
	// Create temporary config file with invalid YAML
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	invalidContent := `
server:
  host: "127.0.0.1"
  port: invalid-port
  [invalid yaml content
`
	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	// Set required environment variables
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("SSH_KEY_ENCRYPTION_KEY", "12345678901234567890123456789012")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SSH_KEY_ENCRYPTION_KEY")
	}()

	// Try to load config
	cfg, err := config.Load(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid configuration",
			setupConfig: func() *config.Config {
				cfg := &config.Config{
					Server: config.ServerConfig{Port: 8080},
					Security: config.SecurityConfig{
						JWTSecret: "valid-jwt-secret",
					},
					SSH: config.SSHConfig{
						EncryptionKey: "12345678901234567890123456789012",
					},
				}
				return cfg
			},
			expectError: false,
		},
		{
			name: "Missing JWT secret",
			setupConfig: func() *config.Config {
				cfg := &config.Config{
					Server: config.ServerConfig{Port: 8080},
					SSH: config.SSHConfig{
						EncryptionKey: "12345678901234567890123456789012",
					},
				}
				return cfg
			},
			expectError: true,
			errorMsg:    "JWT_SECRET is required",
		},
		{
			name: "Missing SSH encryption key",
			setupConfig: func() *config.Config {
				cfg := &config.Config{
					Server: config.ServerConfig{Port: 8080},
					Security: config.SecurityConfig{
						JWTSecret: "valid-jwt-secret",
					},
				}
				return cfg
			},
			expectError: true,
			errorMsg:    "SSH_KEY_ENCRYPTION_KEY is required",
		},
		{
			name: "Invalid SSH encryption key length",
			setupConfig: func() *config.Config {
				cfg := &config.Config{
					Server: config.ServerConfig{Port: 8080},
					Security: config.SecurityConfig{
						JWTSecret: "valid-jwt-secret",
					},
					SSH: config.SSHConfig{
						EncryptionKey: "short-key",
					},
				}
				return cfg
			},
			expectError: true,
			errorMsg:    "SSH_KEY_ENCRYPTION_KEY must be 32 bytes",
		},
		{
			name: "Invalid port - too low",
			setupConfig: func() *config.Config {
				cfg := &config.Config{
					Server: config.ServerConfig{Port: 0},
					Security: config.SecurityConfig{
						JWTSecret: "valid-jwt-secret",
					},
					SSH: config.SSHConfig{
						EncryptionKey: "12345678901234567890123456789012",
					},
				}
				return cfg
			},
			expectError: true,
			errorMsg:    "invalid server port: 0",
		},
		{
			name: "Invalid port - too high",
			setupConfig: func() *config.Config {
				cfg := &config.Config{
					Server: config.ServerConfig{Port: 70000},
					Security: config.SecurityConfig{
						JWTSecret: "valid-jwt-secret",
					},
					SSH: config.SSHConfig{
						EncryptionKey: "12345678901234567890123456789012",
					},
				}
				return cfg
			},
			expectError: true,
			errorMsg:    "invalid server port: 70000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setupConfig()
			err := cfg.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoad_NoEnvFile(t *testing.T) {
	// Ensure no .env file exists in test directory
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Set required environment variables
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("SSH_KEY_ENCRYPTION_KEY", "12345678901234567890123456789012")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SSH_KEY_ENCRYPTION_KEY")
	}()

	// Load config - should not error even without .env file
	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestLoad_WithEnvFile(t *testing.T) {
	// Create temporary .env file
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	envContent := `JWT_SECRET=env-file-jwt-secret
SSH_KEY_ENCRYPTION_KEY=12345678901234567890123456789012
SERVER_HOST=env-file-host
MONGODB_URI=mongodb://env-file:27017
`
	err := os.WriteFile(".env", []byte(envContent), 0644)
	require.NoError(t, err)

	// Load config
	cfg, err := config.Load("")
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check values from .env file
	assert.Equal(t, "env-file-jwt-secret", cfg.Security.JWTSecret)
	assert.Equal(t, "env-file-host", cfg.Server.Host)
	assert.Equal(t, "mongodb://env-file:27017", cfg.Database.URI)
}

func TestConfig_CompleteIntegration(t *testing.T) {
	// Create temporary directory and change to it to avoid .env file interference
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)
	
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config file
	configContent := `
server:
  host: "file-host"
  port: 8888

database:
  uri: "mongodb://file:27017"
  database: "file-db"

security:
  allowedOrigins:
    - "http://file.example.com"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables to override some values
	os.Setenv("SERVER_HOST", "env-host")
	os.Setenv("JWT_SECRET", "env-jwt-secret")
	os.Setenv("SSH_KEY_ENCRYPTION_KEY", "12345678901234567890123456789012")
	// Make sure MONGODB_URI is not set
	os.Unsetenv("MONGODB_URI")
	defer func() {
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SSH_KEY_ENCRYPTION_KEY")
	}()

	// Load config
	cfg, err := config.Load(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check that env vars override file values
	assert.Equal(t, "env-host", cfg.Server.Host) // Overridden by env
	assert.Equal(t, 8888, cfg.Server.Port)       // From file
	assert.Equal(t, "mongodb://file:27017", cfg.Database.URI)
	assert.Equal(t, "file-db", cfg.Database.Database)
	assert.Equal(t, "env-jwt-secret", cfg.Security.JWTSecret)
	assert.Equal(t, []string{"http://file.example.com"}, cfg.Security.AllowedOrigins)
}

// Benchmark tests
func BenchmarkLoad(b *testing.B) {
	// Set required environment variables
	os.Setenv("JWT_SECRET", "bench-jwt-secret")
	os.Setenv("SSH_KEY_ENCRYPTION_KEY", "12345678901234567890123456789012")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("SSH_KEY_ENCRYPTION_KEY")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = config.Load("")
	}
}

func BenchmarkValidate(b *testing.B) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		Security: config.SecurityConfig{
			JWTSecret: "valid-jwt-secret",
		},
		SSH: config.SSHConfig{
			EncryptionKey: "12345678901234567890123456789012",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}

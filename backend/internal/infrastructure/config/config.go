package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	SSH      SSHConfig      `yaml:"ssh"`
	Health   HealthConfig   `yaml:"health"`
	Security SecurityConfig `yaml:"security"`
	WebSocket WebSocketConfig `yaml:"websocket"`
}

// ServerConfig contains server settings
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
}

// DatabaseConfig contains database settings
type DatabaseConfig struct {
	URI            string        `yaml:"uri"`
	Database       string        `yaml:"database"`
	MaxConnections int           `yaml:"maxConnections"`
	Timeout        time.Duration `yaml:"timeout"`
}

// SSHConfig contains SSH settings
type SSHConfig struct {
	ConnectionTimeout time.Duration `yaml:"connectionTimeout"`
	CommandTimeout    time.Duration `yaml:"commandTimeout"`
	MaxConnections    int           `yaml:"maxConnections"`
	EncryptionKey     string        `yaml:"-"` // Not in YAML, from env
}

// HealthConfig contains health check settings
type HealthConfig struct {
	CheckInterval    time.Duration `yaml:"checkInterval"`
	Timeout          time.Duration `yaml:"timeout"`
	MaxRetries       int           `yaml:"maxRetries"`
	ConcurrentChecks int           `yaml:"concurrentChecks"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	JWTSecret         string        `yaml:"-"` // From env
	TokenExpiration   time.Duration `yaml:"tokenExpiration"`
	BCryptCost        int           `yaml:"bcryptCost"`
	AllowedOrigins    []string      `yaml:"allowedOrigins"`
	RateLimitRequests int           `yaml:"rateLimitRequests"`
	RateLimitWindow   time.Duration `yaml:"rateLimitWindow"`
}

// WebSocketConfig contains WebSocket settings
type WebSocketConfig struct {
	PingInterval   time.Duration `yaml:"pingInterval"`
	PongTimeout    time.Duration `yaml:"pongTimeout"`
	WriteTimeout   time.Duration `yaml:"writeTimeout"`
	MaxMessageSize int64         `yaml:"maxMessageSize"`
}

// Load loads configuration from file and environment
func Load(path string) (*Config, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist
		fmt.Println("No .env file found")
	}

	// Load default configuration
	cfg := defaultConfig()

	// Load from YAML file if exists
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variables
	cfg.applyEnvOverrides()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// defaultConfig returns the default configuration
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Database: DatabaseConfig{
			URI:            "mongodb://localhost:27017",
			Database:       "app-env-manager",
			MaxConnections: 100,
			Timeout:        10 * time.Second,
		},
		SSH: SSHConfig{
			ConnectionTimeout: 30 * time.Second,
			CommandTimeout:    300 * time.Second,
			MaxConnections:    50,
		},
		Health: HealthConfig{
			CheckInterval:    30 * time.Second,
			Timeout:          5 * time.Second,
			MaxRetries:       3,
			ConcurrentChecks: 10,
		},
		Security: SecurityConfig{
			TokenExpiration:   24 * time.Hour,
			BCryptCost:        10,
			AllowedOrigins:    []string{"http://localhost:3000"},
			RateLimitRequests: 100,
			RateLimitWindow:   1 * time.Minute,
		},
		WebSocket: WebSocketConfig{
			PingInterval:   30 * time.Second,
			PongTimeout:    60 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxMessageSize: 512 * 1024, // 512KB
		},
	}
}

// applyEnvOverrides applies environment variable overrides
func (c *Config) applyEnvOverrides() {
	// Server
	if host := os.Getenv("SERVER_HOST"); host != "" {
		c.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		var p int
		if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
			c.Server.Port = p
		}
	}

	// Database
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		c.Database.URI = uri
	}
	if db := os.Getenv("MONGODB_DATABASE"); db != "" {
		c.Database.Database = db
	}

	// Security
	c.Security.JWTSecret = os.Getenv("JWT_SECRET")
	c.SSH.EncryptionKey = os.Getenv("SSH_KEY_ENCRYPTION_KEY")

	// CORS
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		c.Security.AllowedOrigins = []string{origins}
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Security.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.SSH.EncryptionKey == "" {
		return fmt.Errorf("SSH_KEY_ENCRYPTION_KEY is required")
	}
	if len(c.SSH.EncryptionKey) != 32 {
		return fmt.Errorf("SSH_KEY_ENCRYPTION_KEY must be 32 bytes")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	return nil
}

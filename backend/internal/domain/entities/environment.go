package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Environment represents an application environment
type Environment struct {
	ID             primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name           string                 `bson:"name" json:"name"`
	Description    string                 `bson:"description" json:"description"`
	EnvironmentURL string                 `bson:"environmentURL" json:"environmentURL"` // URL to access the environment
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

// Target represents the connection target
type Target struct {
	Host   string `bson:"host" json:"host"`
	Port   int    `bson:"port" json:"port"`
	Domain string `bson:"domain,omitempty" json:"domain,omitempty"`
}

// CredentialRef references the credentials
type CredentialRef struct {
	Type     string             `bson:"type" json:"type"` // "key" or "password"
	Username string             `bson:"username" json:"username"`
	KeyID    primitive.ObjectID `bson:"keyId,omitempty" json:"keyId,omitempty"`
}

// HealthCheckConfig defines health check settings
type HealthCheckConfig struct {
	Enabled    bool                   `bson:"enabled" json:"enabled"`
	Endpoint   string                 `bson:"endpoint" json:"endpoint"`
	Method     string                 `bson:"method" json:"method"`
	Interval   int                    `bson:"interval" json:"interval"` // seconds
	Timeout    int                    `bson:"timeout" json:"timeout"`   // seconds
	Validation ValidationConfig       `bson:"validation" json:"validation"`
	Headers    map[string]string      `bson:"headers,omitempty" json:"headers,omitempty"`
}

// ValidationConfig defines how to validate health check responses
type ValidationConfig struct {
	Type  string      `bson:"type" json:"type"`   // "statusCode" or "jsonRegex"
	Value interface{} `bson:"value" json:"value"` // Expected status code or regex pattern
}

// Status represents the current environment status
type Status struct {
	Health       HealthStatus `bson:"health" json:"health"`
	LastCheck    time.Time    `bson:"lastCheck" json:"lastCheck"`
	Message      string       `bson:"message" json:"message"`
	ResponseTime int64        `bson:"responseTime" json:"responseTime"` // milliseconds
}

// HealthStatus enum
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// SystemInfo contains system information
type SystemInfo struct {
	OSVersion   string    `bson:"osVersion" json:"osVersion"`
	AppVersion  string    `bson:"appVersion" json:"appVersion"`
	LastUpdated time.Time `bson:"lastUpdated" json:"lastUpdated"`
}

// Timestamps tracks important dates
type Timestamps struct {
	CreatedAt    time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time  `bson:"updatedAt" json:"updatedAt"`
	LastRestartAt *time.Time `bson:"lastRestartAt,omitempty" json:"lastRestartAt,omitempty"`
	LastUpgradeAt *time.Time `bson:"lastUpgradeAt,omitempty" json:"lastUpgradeAt,omitempty"`
	LastHealthyAt *time.Time `bson:"lastHealthyAt,omitempty" json:"lastHealthyAt,omitempty"`
}

// CommandConfig defines custom commands for environment operations
type CommandConfig struct {
	Type    CommandType   `bson:"type" json:"type"`        // "ssh" or "http"
	Restart RestartConfig `bson:"restart" json:"restart"`
}

// CommandType enum
type CommandType string

const (
	CommandTypeSSH  CommandType = "ssh"
	CommandTypeHTTP CommandType = "http"
)

// RestartConfig defines restart command configuration
type RestartConfig struct {
	Enabled bool                   `bson:"enabled" json:"enabled"`
	Command string                 `bson:"command,omitempty" json:"command,omitempty"` // For SSH
	URL     string                 `bson:"url,omitempty" json:"url,omitempty"`         // For HTTP
	Method  string                 `bson:"method,omitempty" json:"method,omitempty"`   // For HTTP
	Headers map[string]string      `bson:"headers,omitempty" json:"headers,omitempty"` // For HTTP
	Body    map[string]interface{} `bson:"body,omitempty" json:"body,omitempty"`       // For HTTP
}

// CommandDetails defines specific command details
type CommandDetails struct {
	Command string                 `bson:"command,omitempty" json:"command,omitempty"` // For SSH
	URL     string                 `bson:"url,omitempty" json:"url,omitempty"`         // For HTTP
	Method  string                 `bson:"method,omitempty" json:"method,omitempty"`   // For HTTP
	Headers map[string]string      `bson:"headers,omitempty" json:"headers,omitempty"` // For HTTP
	Body    map[string]interface{} `bson:"body,omitempty" json:"body,omitempty"`       // For HTTP
}

// UpgradeConfig defines configuration for version upgrades
type UpgradeConfig struct {
	Enabled             bool                   `bson:"enabled" json:"enabled"`
	Type                CommandType            `bson:"type" json:"type"`                                                 // "ssh" or "http" for upgrade command
	VersionListURL      string                 `bson:"versionListURL" json:"versionListURL"`                             // URL to fetch available versions
	VersionListMethod   string                 `bson:"versionListMethod,omitempty" json:"versionListMethod,omitempty"`   // HTTP method for version list request
	VersionListHeaders  map[string]string      `bson:"versionListHeaders,omitempty" json:"versionListHeaders,omitempty"` // Headers for version list request
	VersionListBody     string                 `bson:"versionListBody,omitempty" json:"versionListBody,omitempty"`       // Body for version list request
	JSONPathResponse    string                 `bson:"jsonPathResponse" json:"jsonPathResponse"`                         // JSONPath to extract version list from response
	UpgradeCommand      CommandDetails         `bson:"upgradeCommand" json:"upgradeCommand"`                             // SSH command or HTTP details for upgrade
}

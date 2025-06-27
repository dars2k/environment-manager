package dto

import (
	"app-env-manager/internal/domain/entities"
)

// Request DTOs

// RestartRequest represents a restart operation request
type RestartRequest struct {
	Force           bool `json:"force"`
	GracefulTimeout int  `json:"gracefulTimeout,omitempty"`
}


// UpgradeRequest represents an upgrade operation request
type UpgradeRequest struct {
	Version           string `json:"version"`
	BackupFirst       bool   `json:"backupFirst"`
	RollbackOnFailure bool   `json:"rollbackOnFailure"`
}

// Response DTOs

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success  bool               `json:"success"`
	Data     interface{}        `json:"data"`
	Metadata ResponseMetadata   `json:"metadata"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success  bool               `json:"success"`
	Error    ErrorInfo          `json:"error"`
	Metadata ResponseMetadata   `json:"metadata"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ResponseMetadata contains response metadata
type ResponseMetadata struct {
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// ListEnvironmentsResponse represents a list of environments response
type ListEnvironmentsResponse struct {
	Environments []*entities.Environment `json:"environments"`
	Pagination   PaginationResponse      `json:"pagination"`
}

// EnvironmentResponse represents a single environment response
type EnvironmentResponse struct {
	Environment *entities.Environment `json:"environment"`
}

// PaginationResponse contains pagination information
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int   `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// OperationResponse represents an async operation response
type OperationResponse struct {
	OperationID string `json:"operationId"`
	Status      string `json:"status"`
	StartedAt   string `json:"startedAt,omitempty"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}

// AuditLogResponse represents audit log entries response
type AuditLogResponse struct {
	Logs       []*entities.AuditLog `json:"logs"`
	Pagination PaginationResponse   `json:"pagination"`
}

// MetricsResponse represents environment metrics
type MetricsResponse struct {
	Period     string                   `json:"period"`
	Resolution string                   `json:"resolution"`
	Data       []MetricDataPoint        `json:"data"`
	Summary    MetricsSummary           `json:"summary"`
}

// MetricDataPoint represents a single metric data point
type MetricDataPoint struct {
	Timestamp    string  `json:"timestamp"`
	ResponseTime int64   `json:"responseTime"`
	Availability float64 `json:"availability"`
	ErrorRate    float64 `json:"errorRate"`
}

// MetricsSummary represents metrics summary
type MetricsSummary struct {
	AvgResponseTime int64   `json:"avgResponseTime"`
	MaxResponseTime int64   `json:"maxResponseTime"`
	MinResponseTime int64   `json:"minResponseTime"`
	Availability    float64 `json:"availability"`
	TotalChecks     int64   `json:"totalChecks"`
}

// VersionsResponse represents available versions for upgrade
type VersionsResponse struct {
	CurrentVersion    string   `json:"currentVersion"`
	AvailableVersions []string `json:"availableVersions"`
}

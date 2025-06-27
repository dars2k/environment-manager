package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLog represents an event in the audit trail
type AuditLog struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Timestamp      time.Time          `bson:"timestamp" json:"timestamp"`
	EnvironmentID  primitive.ObjectID `bson:"environmentId" json:"environmentId"`
	EnvironmentName string            `bson:"environmentName" json:"environmentName"`
	Type           EventType          `bson:"type" json:"type"`
	Severity       Severity           `bson:"severity" json:"severity"`
	Actor          Actor              `bson:"actor" json:"actor"`
	Action         Action             `bson:"action" json:"action"`
	Payload        Payload            `bson:"payload" json:"payload"`
	Tags           []string           `bson:"tags,omitempty" json:"tags,omitempty"`
}

// EventType represents the type of event
type EventType string

const (
	EventTypeHealthChange      EventType = "health_change"
	EventTypeRestart           EventType = "restart"
	EventTypeUpgrade           EventType = "upgrade"
	EventTypeShutdown          EventType = "shutdown"
	EventTypeConfigUpdate      EventType = "config_update"
	EventTypeCredentialUpdate  EventType = "credential_update"
	EventTypeConnectionFailed  EventType = "connection_failed"
	EventTypeCommandExecuted   EventType = "command_executed"
)

// Severity represents the severity level
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Actor represents who performed the action
type Actor struct {
	Type string `bson:"type" json:"type"` // "user", "system", "healthCheck"
	ID   string `bson:"id" json:"id"`
	Name string `bson:"name" json:"name"`
	IP   string `bson:"ip,omitempty" json:"ip,omitempty"`
}

// Action represents the action performed
type Action struct {
	Operation string  `bson:"operation" json:"operation"`
	Status    string  `bson:"status" json:"status"` // "started", "completed", "failed"
	Duration  int64   `bson:"duration,omitempty" json:"duration,omitempty"` // milliseconds
	Error     string  `bson:"error,omitempty" json:"error,omitempty"`
}

// Payload contains the event details
type Payload struct {
	Before   map[string]interface{} `bson:"before,omitempty" json:"before,omitempty"`
	After    map[string]interface{} `bson:"after,omitempty" json:"after,omitempty"`
	Metadata map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

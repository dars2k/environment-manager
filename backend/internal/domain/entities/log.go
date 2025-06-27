package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LogType represents the type of log entry
type LogType string

const (
	LogTypeHealthCheck LogType = "health_check"
	LogTypeAction      LogType = "action"
	LogTypeSystem      LogType = "system"
	LogTypeError       LogType = "error"
	LogTypeAuth        LogType = "auth"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelSuccess LogLevel = "success"
)

// ActionType represents the type of action performed
type ActionType string

const (
	ActionTypeCreate   ActionType = "create"
	ActionTypeUpdate   ActionType = "update"
	ActionTypeDelete   ActionType = "delete"
	ActionTypeRestart  ActionType = "restart"
	ActionTypeShutdown ActionType = "shutdown"
	ActionTypeUpgrade  ActionType = "upgrade"
	ActionTypeLogin    ActionType = "login"
	ActionTypeLogout   ActionType = "logout"
)

// Log represents a system log entry
type Log struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Timestamp       time.Time              `bson:"timestamp" json:"timestamp"`
	EnvironmentID   *primitive.ObjectID    `bson:"environmentId,omitempty" json:"environmentId,omitempty"`
	EnvironmentName string                 `bson:"environmentName,omitempty" json:"environmentName,omitempty"`
	UserID          *primitive.ObjectID    `bson:"userId,omitempty" json:"userId,omitempty"`
	Username        string                 `bson:"username,omitempty" json:"username,omitempty"`
	Type            LogType                `bson:"type" json:"type"`
	Level           LogLevel               `bson:"level" json:"level"`
	Action          ActionType             `bson:"action,omitempty" json:"action,omitempty"`
	Message         string                 `bson:"message" json:"message"`
	Details         map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
}

// NewLog creates a new log entry
func NewLog(logType LogType, level LogLevel, message string) *Log {
	return &Log{
		ID:        primitive.NewObjectID(),
		Timestamp: time.Now(),
		Type:      logType,
		Level:     level,
		Message:   message,
		Details:   make(map[string]interface{}),
	}
}

// WithEnvironment adds environment information to the log
func (l *Log) WithEnvironment(envID primitive.ObjectID, envName string) *Log {
	l.EnvironmentID = &envID
	l.EnvironmentName = envName
	return l
}

// WithUser adds user information to the log
func (l *Log) WithUser(userID primitive.ObjectID, username string) *Log {
	l.UserID = &userID
	l.Username = username
	return l
}

// WithAction adds action information to the log
func (l *Log) WithAction(action ActionType) *Log {
	l.Action = action
	return l
}

// WithDetails adds additional details to the log
func (l *Log) WithDetails(details map[string]interface{}) *Log {
	if l.Details == nil {
		l.Details = make(map[string]interface{})
	}
	for k, v := range details {
		l.Details[k] = v
	}
	return l
}

package entities_test

import (
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLogTestSuite tests for AuditLog entity
type AuditLogTestSuite struct {
	suite.Suite
}

func (suite *AuditLogTestSuite) TestEventType() {
	// Test EventType constants
	assert.Equal(suite.T(), entities.EventType("health_change"), entities.EventTypeHealthChange)
	assert.Equal(suite.T(), entities.EventType("restart"), entities.EventTypeRestart)
	assert.Equal(suite.T(), entities.EventType("upgrade"), entities.EventTypeUpgrade)
	assert.Equal(suite.T(), entities.EventType("shutdown"), entities.EventTypeShutdown)
	assert.Equal(suite.T(), entities.EventType("config_update"), entities.EventTypeConfigUpdate)
	assert.Equal(suite.T(), entities.EventType("credential_update"), entities.EventTypeCredentialUpdate)
	assert.Equal(suite.T(), entities.EventType("connection_failed"), entities.EventTypeConnectionFailed)
	assert.Equal(suite.T(), entities.EventType("command_executed"), entities.EventTypeCommandExecuted)

	// Test all event types
	eventTypes := []entities.EventType{
		entities.EventTypeHealthChange,
		entities.EventTypeRestart,
		entities.EventTypeUpgrade,
		entities.EventTypeShutdown,
		entities.EventTypeConfigUpdate,
		entities.EventTypeCredentialUpdate,
		entities.EventTypeConnectionFailed,
		entities.EventTypeCommandExecuted,
	}

	for _, eventType := range eventTypes {
		assert.NotEmpty(suite.T(), eventType)
	}
}

func (suite *AuditLogTestSuite) TestSeverity() {
	// Test Severity constants
	assert.Equal(suite.T(), entities.Severity("info"), entities.SeverityInfo)
	assert.Equal(suite.T(), entities.Severity("warning"), entities.SeverityWarning)
	assert.Equal(suite.T(), entities.Severity("error"), entities.SeverityError)
	assert.Equal(suite.T(), entities.Severity("critical"), entities.SeverityCritical)

	// Test all severity levels
	severities := []entities.Severity{
		entities.SeverityInfo,
		entities.SeverityWarning,
		entities.SeverityError,
		entities.SeverityCritical,
	}

	for _, severity := range severities {
		assert.NotEmpty(suite.T(), severity)
	}
}

func (suite *AuditLogTestSuite) TestActor() {
	// Test Actor with all fields
	actor := entities.Actor{
		Type: "user",
		ID:   "user-123",
		Name: "John Doe",
		IP:   "192.168.1.100",
	}

	assert.Equal(suite.T(), "user", actor.Type)
	assert.Equal(suite.T(), "user-123", actor.ID)
	assert.Equal(suite.T(), "John Doe", actor.Name)
	assert.Equal(suite.T(), "192.168.1.100", actor.IP)

	// Test Actor without optional IP
	systemActor := entities.Actor{
		Type: "system",
		ID:   "system",
		Name: "Automated System",
	}

	assert.Equal(suite.T(), "system", systemActor.Type)
	assert.Equal(suite.T(), "system", systemActor.ID)
	assert.Equal(suite.T(), "Automated System", systemActor.Name)
	assert.Empty(suite.T(), systemActor.IP)

	// Test health check actor
	healthCheckActor := entities.Actor{
		Type: "healthCheck",
		ID:   "health-monitor",
		Name: "Health Monitor Service",
	}

	assert.Equal(suite.T(), "healthCheck", healthCheckActor.Type)
	assert.Equal(suite.T(), "health-monitor", healthCheckActor.ID)
}

func (suite *AuditLogTestSuite) TestAction() {
	// Test Action with all fields
	action := entities.Action{
		Operation: "restart_environment",
		Status:    "completed",
		Duration:  1500,
		Error:     "",
	}

	assert.Equal(suite.T(), "restart_environment", action.Operation)
	assert.Equal(suite.T(), "completed", action.Status)
	assert.Equal(suite.T(), int64(1500), action.Duration)
	assert.Empty(suite.T(), action.Error)

	// Test failed action
	failedAction := entities.Action{
		Operation: "upgrade_environment",
		Status:    "failed",
		Duration:  30000,
		Error:     "Connection timeout",
	}

	assert.Equal(suite.T(), "upgrade_environment", failedAction.Operation)
	assert.Equal(suite.T(), "failed", failedAction.Status)
	assert.Equal(suite.T(), int64(30000), failedAction.Duration)
	assert.Equal(suite.T(), "Connection timeout", failedAction.Error)

	// Test action without duration
	startedAction := entities.Action{
		Operation: "shutdown_environment",
		Status:    "started",
	}

	assert.Equal(suite.T(), "shutdown_environment", startedAction.Operation)
	assert.Equal(suite.T(), "started", startedAction.Status)
	assert.Zero(suite.T(), startedAction.Duration)
	assert.Empty(suite.T(), startedAction.Error)
}

func (suite *AuditLogTestSuite) TestPayload() {
	// Test Payload with all fields
	payload := entities.Payload{
		Before: map[string]interface{}{
			"status": "healthy",
			"version": "1.0.0",
		},
		After: map[string]interface{}{
			"status": "upgrading",
			"version": "2.0.0",
		},
		Metadata: map[string]interface{}{
			"trigger": "manual",
			"approved_by": "admin",
		},
	}

	assert.NotNil(suite.T(), payload.Before)
	assert.Equal(suite.T(), "healthy", payload.Before["status"])
	assert.Equal(suite.T(), "1.0.0", payload.Before["version"])

	assert.NotNil(suite.T(), payload.After)
	assert.Equal(suite.T(), "upgrading", payload.After["status"])
	assert.Equal(suite.T(), "2.0.0", payload.After["version"])

	assert.NotNil(suite.T(), payload.Metadata)
	assert.Equal(suite.T(), "manual", payload.Metadata["trigger"])
	assert.Equal(suite.T(), "admin", payload.Metadata["approved_by"])

	// Test Payload with only metadata
	metadataOnlyPayload := entities.Payload{
		Metadata: map[string]interface{}{
			"command": "systemctl restart nginx",
			"exit_code": 0,
		},
	}

	assert.Nil(suite.T(), metadataOnlyPayload.Before)
	assert.Nil(suite.T(), metadataOnlyPayload.After)
	assert.NotNil(suite.T(), metadataOnlyPayload.Metadata)
	assert.Equal(suite.T(), 0, metadataOnlyPayload.Metadata["exit_code"])

	// Test empty Payload
	emptyPayload := entities.Payload{}
	assert.Nil(suite.T(), emptyPayload.Before)
	assert.Nil(suite.T(), emptyPayload.After)
	assert.Nil(suite.T(), emptyPayload.Metadata)
}

func (suite *AuditLogTestSuite) TestAuditLog_BasicFields() {
	// Test AuditLog creation with basic fields
	now := time.Now()
	envID := primitive.NewObjectID()
	
	auditLog := &entities.AuditLog{
		ID:             primitive.NewObjectID(),
		Timestamp:      now,
		EnvironmentID:  envID,
		EnvironmentName: "production-api",
		Type:           entities.EventTypeRestart,
		Severity:       entities.SeverityInfo,
	}

	assert.NotNil(suite.T(), auditLog)
	assert.False(suite.T(), auditLog.ID.IsZero())
	assert.Equal(suite.T(), now, auditLog.Timestamp)
	assert.Equal(suite.T(), envID, auditLog.EnvironmentID)
	assert.Equal(suite.T(), "production-api", auditLog.EnvironmentName)
	assert.Equal(suite.T(), entities.EventTypeRestart, auditLog.Type)
	assert.Equal(suite.T(), entities.SeverityInfo, auditLog.Severity)
}

func (suite *AuditLogTestSuite) TestAuditLog_CompleteEntity() {
	// Test complete AuditLog with all fields
	now := time.Now()
	envID := primitive.NewObjectID()
	
	auditLog := &entities.AuditLog{
		ID:             primitive.NewObjectID(),
		Timestamp:      now,
		EnvironmentID:  envID,
		EnvironmentName: "staging-backend",
		Type:           entities.EventTypeUpgrade,
		Severity:       entities.SeverityWarning,
		Actor: entities.Actor{
			Type: "user",
			ID:   "user-456",
			Name: "Jane Smith",
			IP:   "10.0.0.50",
		},
		Action: entities.Action{
			Operation: "upgrade_application",
			Status:    "completed",
			Duration:  120000,
		},
		Payload: entities.Payload{
			Before: map[string]interface{}{
				"version": "2.3.0",
				"memory": "4GB",
			},
			After: map[string]interface{}{
				"version": "3.0.0",
				"memory": "8GB",
			},
			Metadata: map[string]interface{}{
				"deployment_strategy": "blue-green",
				"rollback_enabled": true,
			},
		},
		Tags: []string{"upgrade", "production", "critical"},
	}

	// Verify all fields
	assert.NotNil(suite.T(), auditLog)
	assert.False(suite.T(), auditLog.ID.IsZero())
	assert.Equal(suite.T(), now, auditLog.Timestamp)
	assert.Equal(suite.T(), envID, auditLog.EnvironmentID)
	assert.Equal(suite.T(), "staging-backend", auditLog.EnvironmentName)
	assert.Equal(suite.T(), entities.EventTypeUpgrade, auditLog.Type)
	assert.Equal(suite.T(), entities.SeverityWarning, auditLog.Severity)
	
	// Verify Actor
	assert.Equal(suite.T(), "user", auditLog.Actor.Type)
	assert.Equal(suite.T(), "user-456", auditLog.Actor.ID)
	assert.Equal(suite.T(), "Jane Smith", auditLog.Actor.Name)
	assert.Equal(suite.T(), "10.0.0.50", auditLog.Actor.IP)
	
	// Verify Action
	assert.Equal(suite.T(), "upgrade_application", auditLog.Action.Operation)
	assert.Equal(suite.T(), "completed", auditLog.Action.Status)
	assert.Equal(suite.T(), int64(120000), auditLog.Action.Duration)
	
	// Verify Payload
	assert.Equal(suite.T(), "2.3.0", auditLog.Payload.Before["version"])
	assert.Equal(suite.T(), "3.0.0", auditLog.Payload.After["version"])
	assert.Equal(suite.T(), "blue-green", auditLog.Payload.Metadata["deployment_strategy"])
	
	// Verify Tags
	assert.Len(suite.T(), auditLog.Tags, 3)
	assert.Contains(suite.T(), auditLog.Tags, "upgrade")
	assert.Contains(suite.T(), auditLog.Tags, "production")
	assert.Contains(suite.T(), auditLog.Tags, "critical")
}

func (suite *AuditLogTestSuite) TestAuditLog_Scenarios() {
	// Scenario 1: Health change event
	healthChangeLog := &entities.AuditLog{
		ID:             primitive.NewObjectID(),
		Timestamp:      time.Now(),
		EnvironmentID:  primitive.NewObjectID(),
		EnvironmentName: "api-server",
		Type:           entities.EventTypeHealthChange,
		Severity:       entities.SeverityError,
		Actor: entities.Actor{
			Type: "healthCheck",
			ID:   "health-monitor",
			Name: "Health Check Service",
		},
		Action: entities.Action{
			Operation: "health_check",
			Status:    "completed",
			Duration:  500,
		},
		Payload: entities.Payload{
			Before: map[string]interface{}{
				"status": "healthy",
			},
			After: map[string]interface{}{
				"status": "unhealthy",
			},
			Metadata: map[string]interface{}{
				"error": "Connection refused",
				"retry_count": 3,
			},
		},
		Tags: []string{"health", "alert"},
	}

	assert.Equal(suite.T(), entities.EventTypeHealthChange, healthChangeLog.Type)
	assert.Equal(suite.T(), entities.SeverityError, healthChangeLog.Severity)
	assert.Equal(suite.T(), "healthCheck", healthChangeLog.Actor.Type)
	assert.Equal(suite.T(), "unhealthy", healthChangeLog.Payload.After["status"])

	// Scenario 2: Configuration update
	configUpdateLog := &entities.AuditLog{
		ID:             primitive.NewObjectID(),
		Timestamp:      time.Now(),
		EnvironmentID:  primitive.NewObjectID(),
		EnvironmentName: "database-cluster",
		Type:           entities.EventTypeConfigUpdate,
		Severity:       entities.SeverityInfo,
		Actor: entities.Actor{
			Type: "user",
			ID:   "admin-123",
			Name: "System Admin",
			IP:   "192.168.1.10",
		},
		Action: entities.Action{
			Operation: "update_configuration",
			Status:    "completed",
			Duration:  200,
		},
		Payload: entities.Payload{
			Before: map[string]interface{}{
				"max_connections": 100,
				"timeout": 30,
			},
			After: map[string]interface{}{
				"max_connections": 200,
				"timeout": 60,
			},
		},
	}

	assert.Equal(suite.T(), entities.EventTypeConfigUpdate, configUpdateLog.Type)
	assert.Equal(suite.T(), entities.SeverityInfo, configUpdateLog.Severity)
	assert.Equal(suite.T(), 100, configUpdateLog.Payload.Before["max_connections"])
	assert.Equal(suite.T(), 200, configUpdateLog.Payload.After["max_connections"])

	// Scenario 3: Failed command execution
	failedCommandLog := &entities.AuditLog{
		ID:             primitive.NewObjectID(),
		Timestamp:      time.Now(),
		EnvironmentID:  primitive.NewObjectID(),
		EnvironmentName: "worker-node-01",
		Type:           entities.EventTypeCommandExecuted,
		Severity:       entities.SeverityCritical,
		Actor: entities.Actor{
			Type: "system",
			ID:   "automation",
			Name: "Automation Service",
		},
		Action: entities.Action{
			Operation: "execute_maintenance_script",
			Status:    "failed",
			Duration:  5000,
			Error:     "Script exited with code 1: Disk space insufficient",
		},
		Payload: entities.Payload{
			Metadata: map[string]interface{}{
				"script": "/opt/scripts/cleanup.sh",
				"exit_code": 1,
				"stderr": "Error: Disk space insufficient",
			},
		},
		Tags: []string{"maintenance", "failure", "disk-space"},
	}

	assert.Equal(suite.T(), entities.EventTypeCommandExecuted, failedCommandLog.Type)
	assert.Equal(suite.T(), entities.SeverityCritical, failedCommandLog.Severity)
	assert.Equal(suite.T(), "failed", failedCommandLog.Action.Status)
	assert.NotEmpty(suite.T(), failedCommandLog.Action.Error)
	assert.Equal(suite.T(), 1, failedCommandLog.Payload.Metadata["exit_code"])

	// Scenario 4: Connection failure
	connectionFailureLog := &entities.AuditLog{
		ID:             primitive.NewObjectID(),
		Timestamp:      time.Now(),
		EnvironmentID:  primitive.NewObjectID(),
		EnvironmentName: "remote-service",
		Type:           entities.EventTypeConnectionFailed,
		Severity:       entities.SeverityError,
		Actor: entities.Actor{
			Type: "system",
			ID:   "ssh-manager",
			Name: "SSH Manager",
		},
		Action: entities.Action{
			Operation: "establish_ssh_connection",
			Status:    "failed",
			Error:     "Connection timeout after 30s",
		},
		Payload: entities.Payload{
			Metadata: map[string]interface{}{
				"host": "10.0.0.100",
				"port": 22,
				"timeout": 30,
				"retry_attempts": 3,
			},
		},
	}

	assert.Equal(suite.T(), entities.EventTypeConnectionFailed, connectionFailureLog.Type)
	assert.Equal(suite.T(), entities.SeverityError, connectionFailureLog.Severity)
	assert.Equal(suite.T(), "Connection timeout after 30s", connectionFailureLog.Action.Error)
	assert.Equal(suite.T(), 3, connectionFailureLog.Payload.Metadata["retry_attempts"])
}

func (suite *AuditLogTestSuite) TestAuditLog_EdgeCases() {
	// Test with minimal fields
	minimalLog := &entities.AuditLog{
		ID:             primitive.NewObjectID(),
		Timestamp:      time.Now(),
		EnvironmentID:  primitive.NewObjectID(),
		EnvironmentName: "test",
		Type:           entities.EventTypeRestart,
		Severity:       entities.SeverityInfo,
		Actor:          entities.Actor{},
		Action:         entities.Action{},
		Payload:        entities.Payload{},
	}

	assert.NotNil(suite.T(), minimalLog)
	assert.Empty(suite.T(), minimalLog.Actor.Type)
	assert.Empty(suite.T(), minimalLog.Action.Operation)
	assert.Nil(suite.T(), minimalLog.Tags)

	// Test with empty tags
	emptyTagsLog := &entities.AuditLog{
		ID:             primitive.NewObjectID(),
		Timestamp:      time.Now(),
		EnvironmentID:  primitive.NewObjectID(),
		EnvironmentName: "test",
		Type:           entities.EventTypeShutdown,
		Severity:       entities.SeverityWarning,
		Tags:           []string{},
	}

	assert.NotNil(suite.T(), emptyTagsLog.Tags)
	assert.Empty(suite.T(), emptyTagsLog.Tags)

	// Test with nil maps in payload
	nilMapsPayload := entities.Payload{
		Before:   nil,
		After:    nil,
		Metadata: nil,
	}

	assert.Nil(suite.T(), nilMapsPayload.Before)
	assert.Nil(suite.T(), nilMapsPayload.After)
	assert.Nil(suite.T(), nilMapsPayload.Metadata)
}

// Run the test suite
func TestAuditLogTestSuite(t *testing.T) {
	suite.Run(t, new(AuditLogTestSuite))
}

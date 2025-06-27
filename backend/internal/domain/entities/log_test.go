package entities_test

import (
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LogTestSuite tests for Log entity
type LogTestSuite struct {
	suite.Suite
}

func (suite *LogTestSuite) TestLogType() {
	// Test LogType constants
	assert.Equal(suite.T(), entities.LogType("health_check"), entities.LogTypeHealthCheck)
	assert.Equal(suite.T(), entities.LogType("action"), entities.LogTypeAction)
	assert.Equal(suite.T(), entities.LogType("system"), entities.LogTypeSystem)
	assert.Equal(suite.T(), entities.LogType("error"), entities.LogTypeError)
	assert.Equal(suite.T(), entities.LogType("auth"), entities.LogTypeAuth)

	// Test all log types
	logTypes := []entities.LogType{
		entities.LogTypeHealthCheck,
		entities.LogTypeAction,
		entities.LogTypeSystem,
		entities.LogTypeError,
		entities.LogTypeAuth,
	}

	for _, logType := range logTypes {
		assert.NotEmpty(suite.T(), logType)
	}
}

func (suite *LogTestSuite) TestLogLevel() {
	// Test LogLevel constants
	assert.Equal(suite.T(), entities.LogLevel("info"), entities.LogLevelInfo)
	assert.Equal(suite.T(), entities.LogLevel("warning"), entities.LogLevelWarning)
	assert.Equal(suite.T(), entities.LogLevel("error"), entities.LogLevelError)
	assert.Equal(suite.T(), entities.LogLevel("success"), entities.LogLevelSuccess)

	// Test all log levels
	logLevels := []entities.LogLevel{
		entities.LogLevelInfo,
		entities.LogLevelWarning,
		entities.LogLevelError,
		entities.LogLevelSuccess,
	}

	for _, level := range logLevels {
		assert.NotEmpty(suite.T(), level)
	}
}

func (suite *LogTestSuite) TestActionType() {
	// Test ActionType constants
	assert.Equal(suite.T(), entities.ActionType("create"), entities.ActionTypeCreate)
	assert.Equal(suite.T(), entities.ActionType("update"), entities.ActionTypeUpdate)
	assert.Equal(suite.T(), entities.ActionType("delete"), entities.ActionTypeDelete)
	assert.Equal(suite.T(), entities.ActionType("restart"), entities.ActionTypeRestart)
	assert.Equal(suite.T(), entities.ActionType("shutdown"), entities.ActionTypeShutdown)
	assert.Equal(suite.T(), entities.ActionType("upgrade"), entities.ActionTypeUpgrade)
	assert.Equal(suite.T(), entities.ActionType("login"), entities.ActionTypeLogin)
	assert.Equal(suite.T(), entities.ActionType("logout"), entities.ActionTypeLogout)

	// Test all action types
	actionTypes := []entities.ActionType{
		entities.ActionTypeCreate,
		entities.ActionTypeUpdate,
		entities.ActionTypeDelete,
		entities.ActionTypeRestart,
		entities.ActionTypeShutdown,
		entities.ActionTypeUpgrade,
		entities.ActionTypeLogin,
		entities.ActionTypeLogout,
	}

	for _, action := range actionTypes {
		assert.NotEmpty(suite.T(), action)
	}
}

func (suite *LogTestSuite) TestNewLog() {
	// Test creating a new log
	beforeCreate := time.Now()
	log := entities.NewLog(entities.LogTypeSystem, entities.LogLevelInfo, "Test log message")
	afterCreate := time.Now()

	assert.NotNil(suite.T(), log)
	assert.False(suite.T(), log.ID.IsZero())
	assert.Equal(suite.T(), entities.LogTypeSystem, log.Type)
	assert.Equal(suite.T(), entities.LogLevelInfo, log.Level)
	assert.Equal(suite.T(), "Test log message", log.Message)
	assert.NotNil(suite.T(), log.Details)
	assert.Empty(suite.T(), log.Details)
	
	// Check timestamp is within expected range
	assert.True(suite.T(), log.Timestamp.After(beforeCreate) || log.Timestamp.Equal(beforeCreate))
	assert.True(suite.T(), log.Timestamp.Before(afterCreate) || log.Timestamp.Equal(afterCreate))
	
	// Check optional fields are nil
	assert.Nil(suite.T(), log.EnvironmentID)
	assert.Empty(suite.T(), log.EnvironmentName)
	assert.Nil(suite.T(), log.UserID)
	assert.Empty(suite.T(), log.Username)
	assert.Empty(suite.T(), log.Action)
}

func (suite *LogTestSuite) TestNewLog_DifferentTypes() {
	// Test creating logs with different types and levels
	testCases := []struct {
		logType  entities.LogType
		level    entities.LogLevel
		message  string
	}{
		{entities.LogTypeHealthCheck, entities.LogLevelSuccess, "Health check passed"},
		{entities.LogTypeAction, entities.LogLevelInfo, "Environment restarted"},
		{entities.LogTypeError, entities.LogLevelError, "Connection failed"},
		{entities.LogTypeAuth, entities.LogLevelWarning, "Invalid login attempt"},
		{entities.LogTypeSystem, entities.LogLevelInfo, "System initialized"},
	}

	for _, tc := range testCases {
		log := entities.NewLog(tc.logType, tc.level, tc.message)
		assert.Equal(suite.T(), tc.logType, log.Type)
		assert.Equal(suite.T(), tc.level, log.Level)
		assert.Equal(suite.T(), tc.message, log.Message)
	}
}

func (suite *LogTestSuite) TestWithEnvironment() {
	// Test adding environment information
	log := entities.NewLog(entities.LogTypeAction, entities.LogLevelInfo, "Environment action")
	envID := primitive.NewObjectID()
	envName := "production-api"

	result := log.WithEnvironment(envID, envName)

	// Check that it returns the same log instance (fluent interface)
	assert.Equal(suite.T(), log, result)
	assert.NotNil(suite.T(), log.EnvironmentID)
	assert.Equal(suite.T(), envID, *log.EnvironmentID)
	assert.Equal(suite.T(), envName, log.EnvironmentName)
}

func (suite *LogTestSuite) TestWithUser() {
	// Test adding user information
	log := entities.NewLog(entities.LogTypeAuth, entities.LogLevelInfo, "User logged in")
	userID := primitive.NewObjectID()
	username := "testuser"

	result := log.WithUser(userID, username)

	// Check that it returns the same log instance (fluent interface)
	assert.Equal(suite.T(), log, result)
	assert.NotNil(suite.T(), log.UserID)
	assert.Equal(suite.T(), userID, *log.UserID)
	assert.Equal(suite.T(), username, log.Username)
}

func (suite *LogTestSuite) TestWithAction() {
	// Test adding action information
	log := entities.NewLog(entities.LogTypeAction, entities.LogLevelInfo, "Resource created")

	result := log.WithAction(entities.ActionTypeCreate)

	// Check that it returns the same log instance (fluent interface)
	assert.Equal(suite.T(), log, result)
	assert.Equal(suite.T(), entities.ActionTypeCreate, log.Action)
}

func (suite *LogTestSuite) TestWithDetails() {
	// Test adding details
	log := entities.NewLog(entities.LogTypeSystem, entities.LogLevelError, "Database error")
	
	details := map[string]interface{}{
		"error_code": "DB001",
		"retry_count": 3,
		"connection_string": "mongodb://localhost:27017",
	}

	result := log.WithDetails(details)

	// Check that it returns the same log instance (fluent interface)
	assert.Equal(suite.T(), log, result)
	assert.NotNil(suite.T(), log.Details)
	assert.Len(suite.T(), log.Details, 3)
	assert.Equal(suite.T(), "DB001", log.Details["error_code"])
	assert.Equal(suite.T(), 3, log.Details["retry_count"])
	assert.Equal(suite.T(), "mongodb://localhost:27017", log.Details["connection_string"])
}

func (suite *LogTestSuite) TestWithDetails_MultipleCallsAppend() {
	// Test that calling WithDetails multiple times appends details
	log := entities.NewLog(entities.LogTypeSystem, entities.LogLevelInfo, "System event")
	
	// First call
	log.WithDetails(map[string]interface{}{
		"key1": "value1",
		"key2": 2,
	})

	// Second call
	log.WithDetails(map[string]interface{}{
		"key3": true,
		"key4": []string{"a", "b", "c"},
	})

	// Check all details are present
	assert.Len(suite.T(), log.Details, 4)
	assert.Equal(suite.T(), "value1", log.Details["key1"])
	assert.Equal(suite.T(), 2, log.Details["key2"])
	assert.Equal(suite.T(), true, log.Details["key3"])
	assert.Equal(suite.T(), []string{"a", "b", "c"}, log.Details["key4"])
}

func (suite *LogTestSuite) TestWithDetails_OverwriteExisting() {
	// Test that existing keys are overwritten
	log := entities.NewLog(entities.LogTypeSystem, entities.LogLevelInfo, "System event")
	
	// First call
	log.WithDetails(map[string]interface{}{
		"status": "pending",
		"count": 1,
	})

	// Second call with same key
	log.WithDetails(map[string]interface{}{
		"status": "completed",
		"count": 5,
	})

	// Check values are overwritten
	assert.Len(suite.T(), log.Details, 2)
	assert.Equal(suite.T(), "completed", log.Details["status"])
	assert.Equal(suite.T(), 5, log.Details["count"])
}

func (suite *LogTestSuite) TestFluentInterface() {
	// Test fluent interface chaining
	envID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	
	log := entities.NewLog(entities.LogTypeAction, entities.LogLevelSuccess, "Environment upgraded").
		WithEnvironment(envID, "production").
		WithUser(userID, "admin").
		WithAction(entities.ActionTypeUpgrade).
		WithDetails(map[string]interface{}{
			"from_version": "1.0.0",
			"to_version": "2.0.0",
			"duration": "5m",
		})

	// Verify all fields are set correctly
	assert.Equal(suite.T(), entities.LogTypeAction, log.Type)
	assert.Equal(suite.T(), entities.LogLevelSuccess, log.Level)
	assert.Equal(suite.T(), "Environment upgraded", log.Message)
	assert.Equal(suite.T(), envID, *log.EnvironmentID)
	assert.Equal(suite.T(), "production", log.EnvironmentName)
	assert.Equal(suite.T(), userID, *log.UserID)
	assert.Equal(suite.T(), "admin", log.Username)
	assert.Equal(suite.T(), entities.ActionTypeUpgrade, log.Action)
	assert.Equal(suite.T(), "1.0.0", log.Details["from_version"])
	assert.Equal(suite.T(), "2.0.0", log.Details["to_version"])
	assert.Equal(suite.T(), "5m", log.Details["duration"])
}

func (suite *LogTestSuite) TestLog_CompleteScenarios() {
	// Test complete log scenarios

	// Scenario 1: Health check log
	healthLog := entities.NewLog(entities.LogTypeHealthCheck, entities.LogLevelSuccess, "Health check passed").
		WithEnvironment(primitive.NewObjectID(), "staging-api").
		WithDetails(map[string]interface{}{
			"response_time": 45,
			"status_code": 200,
		})

	assert.Equal(suite.T(), entities.LogTypeHealthCheck, healthLog.Type)
	assert.Equal(suite.T(), entities.LogLevelSuccess, healthLog.Level)
	assert.NotNil(suite.T(), healthLog.EnvironmentID)
	assert.Equal(suite.T(), 45, healthLog.Details["response_time"])

	// Scenario 2: Auth log
	authLog := entities.NewLog(entities.LogTypeAuth, entities.LogLevelWarning, "Failed login attempt").
		WithUser(primitive.NewObjectID(), "john.doe").
		WithAction(entities.ActionTypeLogin).
		WithDetails(map[string]interface{}{
			"ip_address": "192.168.1.100",
			"attempt_count": 3,
			"reason": "invalid_password",
		})

	assert.Equal(suite.T(), entities.LogTypeAuth, authLog.Type)
	assert.Equal(suite.T(), entities.LogLevelWarning, authLog.Level)
	assert.NotNil(suite.T(), authLog.UserID)
	assert.Equal(suite.T(), entities.ActionTypeLogin, authLog.Action)
	assert.Equal(suite.T(), "192.168.1.100", authLog.Details["ip_address"])

	// Scenario 3: System error log
	errorLog := entities.NewLog(entities.LogTypeError, entities.LogLevelError, "Database connection failed").
		WithDetails(map[string]interface{}{
			"error": "connection timeout",
			"host": "db.example.com",
			"port": 5432,
			"retry_attempts": 5,
		})

	assert.Equal(suite.T(), entities.LogTypeError, errorLog.Type)
	assert.Equal(suite.T(), entities.LogLevelError, errorLog.Level)
	assert.Nil(suite.T(), errorLog.EnvironmentID)
	assert.Nil(suite.T(), errorLog.UserID)
	assert.Equal(suite.T(), "connection timeout", errorLog.Details["error"])

	// Scenario 4: Action log with all fields
	actionLog := entities.NewLog(entities.LogTypeAction, entities.LogLevelInfo, "Environment configuration updated").
		WithEnvironment(primitive.NewObjectID(), "production").
		WithUser(primitive.NewObjectID(), "ops-admin").
		WithAction(entities.ActionTypeUpdate).
		WithDetails(map[string]interface{}{
			"changes": map[string]interface{}{
				"memory": "4GB -> 8GB",
				"cpu": "2 -> 4",
			},
			"approved_by": "manager",
			"ticket_id": "OPS-1234",
		})

	assert.Equal(suite.T(), entities.LogTypeAction, actionLog.Type)
	assert.Equal(suite.T(), entities.LogLevelInfo, actionLog.Level)
	assert.NotNil(suite.T(), actionLog.EnvironmentID)
	assert.NotNil(suite.T(), actionLog.UserID)
	assert.Equal(suite.T(), entities.ActionTypeUpdate, actionLog.Action)
	assert.NotNil(suite.T(), actionLog.Details["changes"])
}

func (suite *LogTestSuite) TestLog_EdgeCases() {
	// Test edge cases

	// Empty message
	emptyMessageLog := entities.NewLog(entities.LogTypeSystem, entities.LogLevelInfo, "")
	assert.Equal(suite.T(), "", emptyMessageLog.Message)

	// WithDetails on nil details map
	log := &entities.Log{
		ID:        primitive.NewObjectID(),
		Timestamp: time.Now(),
		Type:      entities.LogTypeSystem,
		Level:     entities.LogLevelInfo,
		Message:   "Test",
		Details:   nil, // Explicitly nil
	}
	
	log.WithDetails(map[string]interface{}{"key": "value"})
	assert.NotNil(suite.T(), log.Details)
	assert.Equal(suite.T(), "value", log.Details["key"])

	// Empty details
	emptyDetailsLog := entities.NewLog(entities.LogTypeSystem, entities.LogLevelInfo, "Test").
		WithDetails(map[string]interface{}{})
	assert.NotNil(suite.T(), emptyDetailsLog.Details)
	assert.Empty(suite.T(), emptyDetailsLog.Details)

	// Nil values in details
	nilDetailsLog := entities.NewLog(entities.LogTypeSystem, entities.LogLevelInfo, "Test").
		WithDetails(map[string]interface{}{
			"nil_value": nil,
			"string_value": "test",
		})
	assert.Nil(suite.T(), nilDetailsLog.Details["nil_value"])
	assert.Equal(suite.T(), "test", nilDetailsLog.Details["string_value"])
}

// Run the test suite
func TestLogTestSuite(t *testing.T) {
	suite.Run(t, new(LogTestSuite))
}

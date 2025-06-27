package entities_test

import (
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EnvironmentTestSuite tests for Environment entity
type EnvironmentTestSuite struct {
	suite.Suite
}

func (suite *EnvironmentTestSuite) TestEnvironment_BasicFields() {
	// Test environment creation with basic fields
	env := &entities.Environment{
		ID:             primitive.NewObjectID(),
		Name:           "test-env",
		Description:    "Test Environment",
		EnvironmentURL: "https://test.example.com",
	}

	assert.NotNil(suite.T(), env)
	assert.Equal(suite.T(), "test-env", env.Name)
	assert.Equal(suite.T(), "Test Environment", env.Description)
	assert.Equal(suite.T(), "https://test.example.com", env.EnvironmentURL)
	assert.False(suite.T(), env.ID.IsZero())
}

func (suite *EnvironmentTestSuite) TestTarget() {
	// Test Target struct
	target := entities.Target{
		Host:   "example.com",
		Port:   22,
		Domain: "test.local",
	}

	assert.Equal(suite.T(), "example.com", target.Host)
	assert.Equal(suite.T(), 22, target.Port)
	assert.Equal(suite.T(), "test.local", target.Domain)
}

func (suite *EnvironmentTestSuite) TestCredentialRef() {
	// Test CredentialRef with key type
	keyID := primitive.NewObjectID()
	cred := entities.CredentialRef{
		Type:     "key",
		Username: "testuser",
		KeyID:    keyID,
	}

	assert.Equal(suite.T(), "key", cred.Type)
	assert.Equal(suite.T(), "testuser", cred.Username)
	assert.Equal(suite.T(), keyID, cred.KeyID)

	// Test CredentialRef with password type
	passwordCred := entities.CredentialRef{
		Type:     "password",
		Username: "testuser2",
	}

	assert.Equal(suite.T(), "password", passwordCred.Type)
	assert.Equal(suite.T(), "testuser2", passwordCred.Username)
	assert.True(suite.T(), passwordCred.KeyID.IsZero())
}

func (suite *EnvironmentTestSuite) TestHealthCheckConfig() {
	// Test HealthCheckConfig
	healthCheck := entities.HealthCheckConfig{
		Enabled:  true,
		Endpoint: "/health",
		Method:   "GET",
		Interval: 60,
		Timeout:  10,
		Validation: entities.ValidationConfig{
			Type:  "statusCode",
			Value: 200,
		},
		Headers: map[string]string{
			"Authorization": "Bearer token",
			"Accept":        "application/json",
		},
	}

	assert.True(suite.T(), healthCheck.Enabled)
	assert.Equal(suite.T(), "/health", healthCheck.Endpoint)
	assert.Equal(suite.T(), "GET", healthCheck.Method)
	assert.Equal(suite.T(), 60, healthCheck.Interval)
	assert.Equal(suite.T(), 10, healthCheck.Timeout)
	assert.Equal(suite.T(), "statusCode", healthCheck.Validation.Type)
	assert.Equal(suite.T(), 200, healthCheck.Validation.Value)
	assert.Len(suite.T(), healthCheck.Headers, 2)

	// Test disabled health check
	disabledCheck := entities.HealthCheckConfig{
		Enabled: false,
	}
	assert.False(suite.T(), disabledCheck.Enabled)
}

func (suite *EnvironmentTestSuite) TestValidationConfig() {
	// Test statusCode validation
	statusValidation := entities.ValidationConfig{
		Type:  "statusCode",
		Value: 200,
	}
	assert.Equal(suite.T(), "statusCode", statusValidation.Type)
	assert.Equal(suite.T(), 200, statusValidation.Value)

	// Test jsonRegex validation
	regexValidation := entities.ValidationConfig{
		Type:  "jsonRegex",
		Value: `"status":\s*"ok"`,
	}
	assert.Equal(suite.T(), "jsonRegex", regexValidation.Type)
	assert.Equal(suite.T(), `"status":\s*"ok"`, regexValidation.Value)
}

func (suite *EnvironmentTestSuite) TestStatus() {
	// Test Status with all fields
	now := time.Now()
	status := entities.Status{
		Health:       entities.HealthStatusHealthy,
		LastCheck:    now,
		Message:      "All systems operational",
		ResponseTime: 150,
	}

	assert.Equal(suite.T(), entities.HealthStatusHealthy, status.Health)
	assert.Equal(suite.T(), now, status.LastCheck)
	assert.Equal(suite.T(), "All systems operational", status.Message)
	assert.Equal(suite.T(), int64(150), status.ResponseTime)
}

func (suite *EnvironmentTestSuite) TestHealthStatus() {
	// Test HealthStatus constants
	assert.Equal(suite.T(), entities.HealthStatus("healthy"), entities.HealthStatusHealthy)
	assert.Equal(suite.T(), entities.HealthStatus("unhealthy"), entities.HealthStatusUnhealthy)
	assert.Equal(suite.T(), entities.HealthStatus("unknown"), entities.HealthStatusUnknown)

	// Test different status values
	statuses := []entities.HealthStatus{
		entities.HealthStatusHealthy,
		entities.HealthStatusUnhealthy,
		entities.HealthStatusUnknown,
	}

	for _, status := range statuses {
		assert.NotEmpty(suite.T(), status)
	}
}

func (suite *EnvironmentTestSuite) TestSystemInfo() {
	// Test SystemInfo
	now := time.Now()
	sysInfo := entities.SystemInfo{
		OSVersion:   "Ubuntu 20.04",
		AppVersion:  "1.2.3",
		LastUpdated: now,
	}

	assert.Equal(suite.T(), "Ubuntu 20.04", sysInfo.OSVersion)
	assert.Equal(suite.T(), "1.2.3", sysInfo.AppVersion)
	assert.Equal(suite.T(), now, sysInfo.LastUpdated)
}

func (suite *EnvironmentTestSuite) TestTimestamps() {
	// Test Timestamps with all fields
	now := time.Now()
	restartTime := now.Add(-1 * time.Hour)
	upgradeTime := now.Add(-2 * time.Hour)
	healthyTime := now.Add(-30 * time.Minute)

	timestamps := entities.Timestamps{
		CreatedAt:     now.Add(-24 * time.Hour),
		UpdatedAt:     now,
		LastRestartAt: &restartTime,
		LastUpgradeAt: &upgradeTime,
		LastHealthyAt: &healthyTime,
	}

	assert.NotNil(suite.T(), timestamps.CreatedAt)
	assert.NotNil(suite.T(), timestamps.UpdatedAt)
	assert.NotNil(suite.T(), timestamps.LastRestartAt)
	assert.NotNil(suite.T(), timestamps.LastUpgradeAt)
	assert.NotNil(suite.T(), timestamps.LastHealthyAt)

	// Test Timestamps with optional fields nil
	minimalTimestamps := entities.Timestamps{
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Nil(suite.T(), minimalTimestamps.LastRestartAt)
	assert.Nil(suite.T(), minimalTimestamps.LastUpgradeAt)
	assert.Nil(suite.T(), minimalTimestamps.LastHealthyAt)
}

func (suite *EnvironmentTestSuite) TestCommandConfig() {
	// Test SSH command config
	sshConfig := entities.CommandConfig{
		Type: entities.CommandTypeSSH,
		Restart: entities.RestartConfig{
			Enabled: true,
			Command: "sudo systemctl restart myapp",
		},
	}

	assert.Equal(suite.T(), entities.CommandTypeSSH, sshConfig.Type)
	assert.True(suite.T(), sshConfig.Restart.Enabled)
	assert.Equal(suite.T(), "sudo systemctl restart myapp", sshConfig.Restart.Command)

	// Test HTTP command config
	httpConfig := entities.CommandConfig{
		Type: entities.CommandTypeHTTP,
		Restart: entities.RestartConfig{
			Enabled: true,
			URL:     "https://api.example.com/restart",
			Method:  "POST",
			Headers: map[string]string{
				"Authorization": "Bearer token",
			},
			Body: map[string]interface{}{
				"force": true,
			},
		},
	}

	assert.Equal(suite.T(), entities.CommandTypeHTTP, httpConfig.Type)
	assert.True(suite.T(), httpConfig.Restart.Enabled)
	assert.Equal(suite.T(), "https://api.example.com/restart", httpConfig.Restart.URL)
	assert.Equal(suite.T(), "POST", httpConfig.Restart.Method)
	assert.NotNil(suite.T(), httpConfig.Restart.Headers)
	assert.NotNil(suite.T(), httpConfig.Restart.Body)
}

func (suite *EnvironmentTestSuite) TestCommandType() {
	// Test CommandType constants
	assert.Equal(suite.T(), entities.CommandType("ssh"), entities.CommandTypeSSH)
	assert.Equal(suite.T(), entities.CommandType("http"), entities.CommandTypeHTTP)
}

func (suite *EnvironmentTestSuite) TestRestartConfig() {
	// Test disabled restart
	disabledRestart := entities.RestartConfig{
		Enabled: false,
	}
	assert.False(suite.T(), disabledRestart.Enabled)

	// Test SSH restart config
	sshRestart := entities.RestartConfig{
		Enabled: true,
		Command: "systemctl restart app",
	}
	assert.True(suite.T(), sshRestart.Enabled)
	assert.Equal(suite.T(), "systemctl restart app", sshRestart.Command)
	assert.Empty(suite.T(), sshRestart.URL)

	// Test HTTP restart config
	httpRestart := entities.RestartConfig{
		Enabled: true,
		URL:     "http://localhost:8080/api/restart",
		Method:  "PUT",
		Headers: map[string]string{"X-API-Key": "secret"},
		Body: map[string]interface{}{
			"graceful": true,
			"timeout":  30,
		},
	}
	assert.True(suite.T(), httpRestart.Enabled)
	assert.Equal(suite.T(), "http://localhost:8080/api/restart", httpRestart.URL)
	assert.Equal(suite.T(), "PUT", httpRestart.Method)
	assert.Equal(suite.T(), "secret", httpRestart.Headers["X-API-Key"])
	assert.Equal(suite.T(), true, httpRestart.Body["graceful"])
}

func (suite *EnvironmentTestSuite) TestCommandDetails() {
	// Test SSH command details
	sshCmd := entities.CommandDetails{
		Command: "apt-get update && apt-get install -y myapp",
	}
	assert.Equal(suite.T(), "apt-get update && apt-get install -y myapp", sshCmd.Command)

	// Test HTTP command details
	httpCmd := entities.CommandDetails{
		URL:    "https://api.example.com/deploy",
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Deploy-Key": "deploy123",
		},
		Body: map[string]interface{}{
			"version": "1.2.3",
			"force":   false,
		},
	}
	assert.Equal(suite.T(), "https://api.example.com/deploy", httpCmd.URL)
	assert.Equal(suite.T(), "POST", httpCmd.Method)
	assert.Len(suite.T(), httpCmd.Headers, 2)
	assert.Equal(suite.T(), "1.2.3", httpCmd.Body["version"])
}

func (suite *EnvironmentTestSuite) TestUpgradeConfig() {
	// Test basic upgrade config
	upgradeConfig := entities.UpgradeConfig{
		Enabled:          true,
		Type:             entities.CommandTypeHTTP,
		VersionListURL:   "https://api.example.com/versions",
		JSONPathResponse: "$.versions[*].tag",
	}

	assert.True(suite.T(), upgradeConfig.Enabled)
	assert.Equal(suite.T(), entities.CommandTypeHTTP, upgradeConfig.Type)
	assert.Equal(suite.T(), "https://api.example.com/versions", upgradeConfig.VersionListURL)
	assert.Equal(suite.T(), "$.versions[*].tag", upgradeConfig.JSONPathResponse)

	// Test complete upgrade config with all fields
	fullUpgradeConfig := entities.UpgradeConfig{
		Enabled:           true,
		Type:              entities.CommandTypeSSH,
		VersionListURL:    "https://registry.example.com/api/versions",
		VersionListMethod: "POST",
		VersionListHeaders: map[string]string{
			"Authorization": "Bearer token",
		},
		VersionListBody:  `{"app": "myapp", "channel": "stable"}`,
		JSONPathResponse: "$.data.versions",
		UpgradeCommand: entities.CommandDetails{
			Command: "upgrade-script.sh --version={VERSION}",
		},
	}

	assert.True(suite.T(), fullUpgradeConfig.Enabled)
	assert.Equal(suite.T(), entities.CommandTypeSSH, fullUpgradeConfig.Type)
	assert.Equal(suite.T(), "POST", fullUpgradeConfig.VersionListMethod)
	assert.NotEmpty(suite.T(), fullUpgradeConfig.VersionListHeaders)
	assert.NotEmpty(suite.T(), fullUpgradeConfig.VersionListBody)
	assert.Equal(suite.T(), "upgrade-script.sh --version={VERSION}", fullUpgradeConfig.UpgradeCommand.Command)

	// Test disabled upgrade config
	disabledUpgrade := entities.UpgradeConfig{
		Enabled: false,
	}
	assert.False(suite.T(), disabledUpgrade.Enabled)
}

func (suite *EnvironmentTestSuite) TestEnvironment_CompleteEntity() {
	// Test complete environment entity with all fields
	now := time.Now()
	restartTime := now.Add(-1 * time.Hour)
	
	env := &entities.Environment{
		ID:             primitive.NewObjectID(),
		Name:           "production-api",
		Description:    "Production API Server",
		EnvironmentURL: "https://api.production.example.com",
		Target: entities.Target{
			Host:   "prod-server-01.example.com",
			Port:   22,
			Domain: "production.local",
		},
		Credentials: entities.CredentialRef{
			Type:     "key",
			Username: "deploy",
			KeyID:    primitive.NewObjectID(),
		},
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Endpoint: "/api/health",
			Method:   "GET",
			Interval: 30,
			Timeout:  5,
			Validation: entities.ValidationConfig{
				Type:  "statusCode",
				Value: 200,
			},
			Headers: map[string]string{
				"X-Health-Check": "true",
			},
		},
		Status: entities.Status{
			Health:       entities.HealthStatusHealthy,
			LastCheck:    now,
			Message:      "Service is healthy",
			ResponseTime: 45,
		},
		SystemInfo: entities.SystemInfo{
			OSVersion:   "Ubuntu 20.04.3 LTS",
			AppVersion:  "2.1.0",
			LastUpdated: now.Add(-24 * time.Hour),
		},
		Timestamps: entities.Timestamps{
			CreatedAt:     now.Add(-30 * 24 * time.Hour),
			UpdatedAt:     now,
			LastRestartAt: &restartTime,
		},
		Commands: entities.CommandConfig{
			Type: entities.CommandTypeSSH,
			Restart: entities.RestartConfig{
				Enabled: true,
				Command: "sudo systemctl restart api-service",
			},
		},
		UpgradeConfig: entities.UpgradeConfig{
			Enabled:          true,
			Type:             entities.CommandTypeSSH,
			VersionListURL:   "https://releases.example.com/api/versions",
			JSONPathResponse: "$.versions[*]",
			UpgradeCommand: entities.CommandDetails{
				Command: "sudo upgrade-api.sh --version={VERSION}",
			},
		},
		Metadata: map[string]interface{}{
			"team":        "backend",
			"cost-center": "engineering",
			"tags":        []string{"api", "production", "critical"},
		},
	}

	// Verify all fields
	assert.NotNil(suite.T(), env)
	assert.False(suite.T(), env.ID.IsZero())
	assert.Equal(suite.T(), "production-api", env.Name)
	assert.Equal(suite.T(), "Production API Server", env.Description)
	assert.Equal(suite.T(), "https://api.production.example.com", env.EnvironmentURL)
	
	// Verify nested structures
	assert.Equal(suite.T(), "prod-server-01.example.com", env.Target.Host)
	assert.Equal(suite.T(), 22, env.Target.Port)
	assert.Equal(suite.T(), "key", env.Credentials.Type)
	assert.True(suite.T(), env.HealthCheck.Enabled)
	assert.Equal(suite.T(), entities.HealthStatusHealthy, env.Status.Health)
	assert.Equal(suite.T(), "2.1.0", env.SystemInfo.AppVersion)
	assert.NotNil(suite.T(), env.Timestamps.LastRestartAt)
	assert.Equal(suite.T(), entities.CommandTypeSSH, env.Commands.Type)
	assert.True(suite.T(), env.UpgradeConfig.Enabled)
	assert.NotNil(suite.T(), env.Metadata)
	assert.Equal(suite.T(), "backend", env.Metadata["team"])
}

func (suite *EnvironmentTestSuite) TestEnvironment_MinimalEntity() {
	// Test minimal environment entity with only required fields
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "test-minimal",
	}

	assert.NotNil(suite.T(), env)
	assert.False(suite.T(), env.ID.IsZero())
	assert.Equal(suite.T(), "test-minimal", env.Name)
	assert.Empty(suite.T(), env.Description)
	assert.Empty(suite.T(), env.EnvironmentURL)
	assert.Empty(suite.T(), env.Target.Host)
	assert.Zero(suite.T(), env.Target.Port)
	assert.Empty(suite.T(), env.Credentials.Type)
	assert.False(suite.T(), env.HealthCheck.Enabled)
	assert.Empty(suite.T(), env.Status.Health)
	assert.Nil(suite.T(), env.Metadata)
}

// Run the test suite
func TestEnvironmentTestSuite(t *testing.T) {
	suite.Run(t, new(EnvironmentTestSuite))
}

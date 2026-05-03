package environment_test

// Tests that cover SSH restart/upgrade paths that require buildSSHTarget to succeed.
// With valid credentials but no real SSH server, Execute() returns a dial error,
// covering the "execute failed" branch without needing actual infrastructure.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"app-env-manager/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// envWithPasswordCreds builds an environment with password credentials so
// buildSSHTarget can succeed (credentials present) even though no SSH server runs.
func envWithPasswordCreds(id primitive.ObjectID) *entities.Environment {
	env := newSampleEnv(id)
	env.Target = entities.Target{Host: "127.0.0.1", Port: 22}
	env.Credentials = entities.CredentialRef{
		Type:     "password",
		Username: "testuser",
	}
	env.Metadata = map[string]interface{}{
		"password": "testpass",
	}
	return env
}

// TestService_RestartEnvironment_SSH_ForceTrue exercises the SSH command-type branch
// where force=true causes the default command to include "--force".
// buildSSHTarget succeeds; Execute fails (no real SSH) → covers the SSH error path.
func TestService_RestartEnvironment_SSH_ForceTrue(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := envWithPasswordCreds(id)
	env.Commands = entities.CommandConfig{
		Type: entities.CommandTypeSSH,
		Restart: entities.RestartConfig{
			Enabled: true,
			Command: "", // empty → uses default + force flag
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), true)
	assert.Error(t, err) // SSH dial fails in test environment
}

// TestService_RestartEnvironment_SSH_ForceFalse exercises the SSH command-type branch
// with force=false (default command, no "--force").
func TestService_RestartEnvironment_SSH_ForceFalse(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := envWithPasswordCreds(id)
	env.Commands = entities.CommandConfig{
		Type: entities.CommandTypeSSH,
		Restart: entities.RestartConfig{
			Enabled: true,
			Command: "", // empty → default command
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), false)
	assert.Error(t, err) // SSH dial fails in test environment
}

// TestService_RestartEnvironment_SSH_CustomCommand exercises the SSH branch where
// a non-empty Restart.Command is provided (skips the `if command == ""` block).
func TestService_RestartEnvironment_SSH_CustomCommand(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := envWithPasswordCreds(id)
	env.Commands = entities.CommandConfig{
		Type: entities.CommandTypeSSH,
		Restart: entities.RestartConfig{
			Enabled: true,
			Command: "echo custom-restart", // non-empty command
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), false)
	assert.Error(t, err) // SSH dial fails
}

// TestService_RestartEnvironment_Default_ForceTrue exercises the default command-type
// branch where force=true (buildSSHTarget succeeds, Execute fails).
func TestService_RestartEnvironment_Default_ForceTrue(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := envWithPasswordCreds(id)
	env.Commands = entities.CommandConfig{
		Type: entities.CommandType(""), // empty → default branch
		Restart: entities.RestartConfig{
			Enabled: true,
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), true)
	assert.Error(t, err) // SSH dial fails
}

// TestService_UpgradeEnvironment_SSH_WithCredentials exercises the SSH upgrade branch
// with valid credentials so buildSSHTarget succeeds. Execute fails (no SSH server).
func TestService_UpgradeEnvironment_SSH_WithCredentials(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := envWithPasswordCreds(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled: true,
		Type:    entities.CommandTypeSSH,
		UpgradeCommand: entities.CommandDetails{
			Command: "echo upgrade {VERSION}",
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v2.0")
	assert.Error(t, err) // SSH dial fails
}

// TestService_UpgradeEnvironment_SSH_EmptyCommand exercises the SSH upgrade branch
// where Command is empty (generates a default command).
func TestService_UpgradeEnvironment_SSH_EmptyCommand(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := envWithPasswordCreds(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled: true,
		Type:    entities.CommandTypeSSH,
		UpgradeCommand: entities.CommandDetails{
			Command: "", // empty → default generated
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v1.5")
	assert.Error(t, err) // SSH dial fails
}

// TestService_UpgradeEnvironment_Body_NonStringValues exercises the body-replacement
// branch where body values are non-string (the `else` branch copies them as-is).
func TestService_UpgradeEnvironment_Body_NonStringValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled: true,
		Type:    entities.CommandTypeHTTP,
		UpgradeCommand: entities.CommandDetails{
			URL:    srv.URL,
			Method: "POST",
			Body: map[string]interface{}{
				"version": "{VERSION}", // string: gets replaced
				"count":   42,          // int: falls to else branch (copied as-is)
				"enabled": true,        // bool: falls to else branch
			},
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v3.0")
	assert.NoError(t, err)
}

// TestService_UpgradeEnvironment_URL_Replacement exercises the URL replacement branch
// where the UpgradeCommand URL contains {VERSION} placeholder.
func TestService_UpgradeEnvironment_URL_Replacement(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled: true,
		Type:    entities.CommandTypeHTTP,
		UpgradeCommand: entities.CommandDetails{
			URL:    srv.URL + "/{VERSION}/upgrade",
			Method: "POST",
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v4.0")
	assert.NoError(t, err)
}

// TestService_GetAvailableVersions_FloatVersions tests that float64 values in a JSON
// array are converted to string versions (covers the float64 case in the type switch).
func TestService_GetAvailableVersions_FloatVersions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// JSON numbers are decoded as float64 by Go's json package
		w.Write([]byte(`{"releases":[1.0, 2.0, 3.0]}`))
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled:          true,
		VersionListURL:   srv.URL,
		JSONPathResponse: "$.releases",
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	versions, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.NoError(t, err)
	assert.Len(t, versions, 3)
}

// TestService_GetAvailableVersions_JSONPath_NotArray exercises the error path where
// the JSONPath result is not an array.
func TestService_GetAvailableVersions_JSONPath_NotArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"releases":"single-string"}`))
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled:          true,
		VersionListURL:   srv.URL,
		JSONPathResponse: "$.releases",
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
}

// TestService_UpgradeEnvironment_SSH_MultiLine exercises multi-line SSH commands
// where some lines are empty (gets skipped by the trim+empty check).
func TestService_UpgradeEnvironment_SSH_MultiLine(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := envWithPasswordCreds(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled: true,
		Type:    entities.CommandTypeSSH,
		UpgradeCommand: entities.CommandDetails{
			Command: "echo step1\n\necho step2", // multi-line with blank line
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v2.0")
	assert.Error(t, err) // SSH dial fails in test
}

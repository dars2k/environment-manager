package environment_test

// Additional tests to push environment service coverage above 90%.

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/domain/errors"
	"app-env-manager/internal/service/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestService_CreateEnvironment_WithHealthCheckEnabled triggers the async goroutine
// that calls CheckHealth after creation when healthCheck.Enabled is true.
func TestService_CreateEnvironment_WithHealthCheckEnabled(t *testing.T) {
	// Run a local HTTP server to respond to health check
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, nil)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)

	repo.On("GetByName", mock.Anything, "asyncenv").Return(nil, errors.ErrEnvironmentNotFound)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Environment")).Run(func(args mock.Arguments) {
		// Set ID to a known value for subsequent mocks
		e := args.Get(1).(*entities.Environment)
		e.ID = id
	}).Return(nil)
	repo.On("GetByID", mock.Anything, mock.Anything).Return(env, nil).Maybe()
	repo.On("UpdateStatus", mock.Anything, mock.Anything, mock.AnythingOfType("entities.Status")).Return(nil).Maybe()
	repo.On("Update", mock.Anything, mock.Anything, mock.AnythingOfType("*entities.Environment")).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	req := environment.CreateEnvironmentRequest{
		Name: "asyncenv",
		Target: entities.Target{
			Host: "localhost",
			Port: 22,
		},
		Credentials: entities.CredentialRef{Type: "password"},
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Endpoint: srv.URL,
			Method:   "GET",
		},
	}

	result, err := svc.CreateEnvironment(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Allow async goroutine to finish
	time.Sleep(100 * time.Millisecond)
}

// TestService_CheckHealth_StatusChangedToHealthy tests the branch where status changes
// from unhealthy to healthy and Update is called to set LastHealthyAt.
func TestService_CheckHealth_StatusChangedToHealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, nil)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.HealthCheck = entities.HealthCheckConfig{
		Enabled:  true,
		Method:   "GET",
		Endpoint: srv.URL,
	}
	// Start with an unhealthy status — so when the check comes back healthy, status changes
	env.Status = entities.Status{
		Health:    entities.HealthStatusUnhealthy,
		LastCheck: time.Now(),
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(nil)
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.CheckHealth(context.Background(), id.Hex())
	assert.NoError(t, err)
}

// TestService_RestartEnvironment_SSHType tests the SSH command type branch of RestartEnvironment.
func TestService_RestartEnvironment_SSHType(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newRestartEnv(id, entities.CommandTypeSSH)
	// No credentials set, so buildSSHTarget will fail → success = false → error returned
	env.Credentials = entities.CredentialRef{Type: "unsupported"}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), true)
	assert.Error(t, err)
}

// TestService_RestartEnvironment_SSHCommand_NoCommand tests the SSH branch with empty command
// (defaults to "sudo systemctl restart app").
func TestService_RestartEnvironment_SSHCommand_NoCommand(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	// SSH type, but command empty → default command used, but no credentials → error
	env.Commands = entities.CommandConfig{
		Type: entities.CommandTypeSSH,
		Restart: entities.RestartConfig{
			Enabled: true,
			Command: "", // empty command → default "sudo systemctl restart app"
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), false)
	assert.Error(t, err)
}

// TestService_UpgradeEnvironment_WithVersionBody tests the version placeholder substitution
// path in UpgradeEnvironment for non-string values in body.
func TestService_UpgradeEnvironment_HTTPSuccess_WithVersionInBody(t *testing.T) {
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
				"version": "{VERSION}",
				"count":   42, // non-string value
			},
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v3.0")
	assert.NoError(t, err)
}

// TestService_UpgradeEnvironment_URLVersionReplacement tests the URL version placeholder.
func TestService_UpgradeEnvironment_HTTPSuccess_URLVersion(t *testing.T) {
	var capturedURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
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
			URL:    srv.URL + "/upgrade/{VERSION}",
			Method: "GET",
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v5.0")
	assert.NoError(t, err)
	assert.Contains(t, capturedURL, "v5.0")
}

// TestService_GetAvailableVersions_JSONPathArrayWildcard tests the array wildcard JSONPath path.
func TestService_GetAvailableVersions_JSONPathArrayWildcard(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items":[{"name":"v1.0"},{"name":"v2.0"}]}`))
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
		JSONPathResponse: "$.items[*].name",
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	versions, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.NoError(t, err)
	assert.Equal(t, []string{"v1.0", "v2.0"}, versions)
}

// TestService_GetAvailableVersions_InvalidURLSSRF tests the SSRF URL validation path.
func TestService_GetAvailableVersions_InvalidURL(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled:        true,
		VersionListURL: "ftp://invalid-scheme.example.com",
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
}

// TestService_GetAvailableVersions_BadJSON tests invalid JSON body path.
func TestService_GetAvailableVersions_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled:        true,
		VersionListURL: srv.URL,
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
}

// TestService_GetAvailableVersions_PlainJSONArray tests no-JSONPath path (plain JSON array).
func TestService_GetAvailableVersions_PlainJSONArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["v1.0","v2.0","v3.0"]`))
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
		JSONPathResponse: "", // no JSONPath → plain array parse
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	versions, current, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.NoError(t, err)
	assert.Equal(t, []string{"v1.0", "v2.0", "v3.0"}, versions)
	_ = current
}

// TestService_GetAvailableVersions_WithBodyAndHeaders tests body + headers paths.
func TestService_GetAvailableVersions_WithBodyAndHeaders(t *testing.T) {
	var capturedHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Token")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["v1.0"]`))
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled:            true,
		VersionListURL:     srv.URL,
		VersionListMethod:  "POST",
		VersionListBody:    `{"filter":"all"}`,
		VersionListHeaders: map[string]string{"X-Token": "mytoken"},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	versions, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.NoError(t, err)
	assert.Equal(t, []string{"v1.0"}, versions)
	assert.Equal(t, "mytoken", capturedHeader)
}

// TestService_GetAvailableVersions_Non200Status tests non-2xx response path.
func TestService_GetAvailableVersions_Non200Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled:        true,
		VersionListURL: srv.URL,
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

// TestService_GetAvailableVersions_NoVersionListURL tests the missing URL path.
func TestService_GetAvailableVersions_NoVersionListURL(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled:        true,
		VersionListURL: "", // empty
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

// TestService_CheckHealth_UpdateStatusError2 tests UpdateStatus failure path (duplicate guard).
func TestService_CheckHealth_UpdateStatusError2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.HealthCheck = entities.HealthCheckConfig{
		Enabled:  true,
		Method:   "GET",
		Endpoint: srv.URL,
	}
	env.Status = entities.Status{Health: entities.HealthStatusHealthy}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(fmt.Errorf("update status failed"))
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.CheckHealth(context.Background(), id.Hex())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update status")
}

// TestService_UpgradeEnvironment_SSHSuccess tests the SSH upgrade success path
// with a multi-line command.
func TestService_UpgradeEnvironment_SSHMultilineCommand(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled: true,
		Type:    entities.CommandTypeSSH,
		UpgradeCommand: entities.CommandDetails{
			// Multi-line command; but no credentials → buildSSHTarget fails
			Command: "echo step1\necho step2",
		},
	}
	// No credentials set → buildSSHTarget will fail
	env.Credentials = entities.CredentialRef{Type: "unsupported"}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v2.0")
	assert.Error(t, err)
}

// TestService_UpgradeEnvironment_SSHEmptyCommand tests SSH upgrade with default command.
func TestService_UpgradeEnvironment_SSHEmptyCommand(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled: true,
		Type:    entities.CommandTypeSSH,
		UpgradeCommand: entities.CommandDetails{
			Command: "", // empty → uses default command "sudo app-upgrade --version=..."
		},
	}
	env.Credentials = entities.CredentialRef{Type: "unsupported"}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v3.0")
	assert.Error(t, err)
}

// TestService_GetAvailableVersions_JSONPathNumericValues tests float64 version values.
func TestService_GetAvailableVersions_JSONPathNumericValues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return numeric versions in array
		w.Write([]byte(`{"releases":[{"ver":1},{"ver":2}]}`))
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
		JSONPathResponse: "$.releases[*].ver",
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	versions, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.NoError(t, err)
	assert.Len(t, versions, 2)
}

// TestService_RestartEnvironment_SSH_ForceTrue tests SSH restart with force=true
// and empty command — exercises the force branch in the SSH case.
func TestService_RestartEnvironment_SSH_ForceTrue_NoCmd(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.Commands = entities.CommandConfig{
		Type: entities.CommandTypeSSH,
		Restart: entities.RestartConfig{
			Enabled: true,
			Command: "", // empty → force branch picks "sudo systemctl restart app --force"
		},
	}
	env.Credentials = entities.CredentialRef{Type: "unsupported"}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), true) // force=true
	assert.Error(t, err) // fails because buildSSHTarget can't build target
}

// TestService_RestartEnvironment_DefaultType_ForceTrue exercises the default branch
// with force=true.
func TestService_RestartEnvironment_DefaultType_ForceTrue(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.Commands = entities.CommandConfig{
		Type: entities.CommandType(""), // default case
		Restart: entities.RestartConfig{
			Enabled: true,
		},
	}
	env.Credentials = entities.CredentialRef{Type: "unsupported"}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), true) // force=true
	assert.Error(t, err)
}

// TestService_GetAvailableVersions_FetchError tests fetch failure (unreachable host).
func TestService_GetAvailableVersions_FetchError(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	// Use a closed server — request will fail immediately
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srvURL := srv.URL
	srv.Close() // Close it so the request fails

	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled:        true,
		VersionListURL: srvURL,
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
}

// TestService_UpgradeEnvironment_SSHCommandWithVersion tests SSH upgrade where the command
// contains the {VERSION} placeholder (exercises the version substitution in upgrade).
func TestService_UpgradeEnvironment_SSHWithVersionPlaceholder(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled: true,
		Type:    entities.CommandTypeSSH,
		UpgradeCommand: entities.CommandDetails{
			Command: "echo upgrade-{VERSION}", // has placeholder
		},
	}
	env.Credentials = entities.CredentialRef{Type: "unsupported"}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v9.0")
	assert.Error(t, err) // buildSSHTarget fails → upgrade fails
}

// TestService_CheckHealth_StatusChangedToUnhealthy tests status change from healthy to unhealthy.
// This covers the branch where oldStatus != newStatus but newStatus is NOT healthy.
func TestService_CheckHealth_StatusChangedToUnhealthy(t *testing.T) {
	// Use a server that returns 500 → unhealthy
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.HealthCheck = entities.HealthCheckConfig{
		Enabled:  true,
		Method:   "GET",
		Endpoint: srv.URL,
	}
	// Start with healthy status — so check returns unhealthy → status changes
	env.Status = entities.Status{Health: entities.HealthStatusHealthy}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.CheckHealth(context.Background(), id.Hex())
	assert.NoError(t, err)
}

// TestService_CheckHealth_StatusNotChanged verifies no Update call when status is same.
func TestService_CheckHealth_StatusNotChanged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // returns healthy
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.HealthCheck = entities.HealthCheckConfig{
		Enabled:  true,
		Method:   "GET",
		Endpoint: srv.URL,
	}
	// Status already healthy — so after health check, status won't change
	env.Status = entities.Status{Health: entities.HealthStatusHealthy}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(nil)
	// Note: Update should NOT be called because status didn't change from healthy to healthy
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.CheckHealth(context.Background(), id.Hex())
	assert.NoError(t, err)
	// Verify Update was NOT called (status didn't change)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
}

// TestService_GetAvailableVersions_EmptyJSONPathResult tests when JSONPath returns
// a type that is not []interface{} or []string.
func TestService_GetAvailableVersions_JSONPathStringResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"version":"v1.0"}`))
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
		JSONPathResponse: "$.version", // returns a string, not an array
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	// Returns error because JSONPath returned a string not an array
	assert.Error(t, err)
}

// TestService_UpgradeEnvironment_HTTPFailure tests HTTP upgrade failure path.
func TestService_UpgradeEnvironment_HTTPFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
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
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v2.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upgrade failed")
}

// TestService_CheckHealth_LastHealthyAt_Updated verifies that LastHealthyAt is set when
// the status transitions to healthy. This requires a valid statusCode validation config
// so the validator returns true.
func TestService_CheckHealth_LastHealthyAt_Updated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, nil)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.HealthCheck = entities.HealthCheckConfig{
		Enabled:  true,
		Method:   "GET",
		Endpoint: srv.URL,
		Validation: entities.ValidationConfig{
			Type:  "statusCode",
			Value: float64(200), // validator checks exact status code match
		},
	}
	// Start unhealthy so status change triggers the LastHealthyAt branch
	env.Status = entities.Status{
		Health:    entities.HealthStatusUnhealthy,
		LastCheck: time.Now(),
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(nil)
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.CheckHealth(context.Background(), id.Hex())
	assert.NoError(t, err)
	repo.AssertCalled(t, "Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment"))
}

// TestService_GetAvailableVersions_InvalidJSONWithJSONPath verifies the error path when
// the response is not valid JSON but a JSONPath is configured.
func TestService_GetAvailableVersions_InvalidJSONWithJSONPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json at all`))
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
		JSONPathResponse: "$.versions[*]",
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON response")
}

// TestService_UpdateEnvironmentPartial_NameChange covers the name-change tracking branch
// (changes["name"] = ...) when the new name differs from the current name.
func TestService_UpdateEnvironmentPartial_NameChange(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.Name = "old-name"

	newName := "new-name"
	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	// GetByName returns not-found so no conflict
	repo.On("GetByName", mock.Anything, newName).Return(nil, fmt.Errorf("not found"))
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	req := environment.UpdateEnvironmentRequest{
		Name: &newName,
	}
	result, err := svc.UpdateEnvironmentPartial(context.Background(), id.Hex(), req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new-name", result.Name)
}

package environment_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/domain/errors"
	"app-env-manager/internal/service/environment"
	"app-env-manager/internal/service/health"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// newTestServiceWithAllowedHosts builds a service with specific allowed hosts (for HTTP command tests).
func newTestServiceWithAllowedHosts(repo *MockEnvironmentRepository, logRepo *MockLogRepository, allowed []string) *environment.Service {
	sshMgr := ssh.NewManager(ssh.Config{
		ConnectionTimeout: time.Second,
		CommandTimeout:    time.Second,
		MaxConnections:    1,
	})
	checker := health.NewChecker(time.Second)
	auditRepo := &MockAuditLogRepository{}
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logSvc := log.NewService(logRepo)
	return environment.NewService(repo, auditRepo, sshMgr, checker, logSvc, allowed)
}

func newRestartEnv(id primitive.ObjectID, cmdType entities.CommandType) *entities.Environment {
	env := newSampleEnv(id)
	env.Commands = entities.CommandConfig{
		Type: cmdType,
		Restart: entities.RestartConfig{
			Enabled: true,
			Command: "echo restart",
		},
	}
	return env
}

// ---- RestartEnvironment ----

func TestService_RestartEnvironment_NotFound(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	repo.On("GetByID", mock.Anything, "nope").Return(nil, errors.ErrEnvironmentNotFound)

	err := svc.RestartEnvironment(context.Background(), "nope", false)
	assert.Error(t, err)
}

func TestService_RestartEnvironment_NotEnabled(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.Commands = entities.CommandConfig{
		Restart: entities.RestartConfig{Enabled: false},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enabled")
}

func TestService_RestartEnvironment_HTTPSuccess(t *testing.T) {
	// Set up an HTTP test server that returns 200
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	// Use allowed hosts so SSRF check passes
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.Commands = entities.CommandConfig{
		Type: entities.CommandTypeHTTP,
		Restart: entities.RestartConfig{
			Enabled: true,
			URL:     srv.URL,
			Method:  "POST",
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), false)
	assert.NoError(t, err)
}

func TestService_RestartEnvironment_HTTPFailure(t *testing.T) {
	// Return 500 from HTTP server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.Commands = entities.CommandConfig{
		Type: entities.CommandTypeHTTP,
		Restart: entities.RestartConfig{
			Enabled: true,
			URL:     srv.URL,
			Method:  "POST",
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "restart failed")
}

func TestService_RestartEnvironment_DefaultCommandType(t *testing.T) {
	// Default command type with no SSH credentials → buildSSHTarget will fail
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.Commands = entities.CommandConfig{
		Type: entities.CommandType(""), // empty → default branch
		Restart: entities.RestartConfig{
			Enabled: true,
		},
	}
	// No credentials set, so buildSSHTarget should fail

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), false)
	assert.Error(t, err)
}

// ---- UpgradeEnvironment ----

func TestService_UpgradeEnvironment_NotFound(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	repo.On("GetByID", mock.Anything, "bad-id").Return(nil, errors.ErrEnvironmentNotFound)

	err := svc.UpgradeEnvironment(context.Background(), "bad-id", "v2.0")
	assert.Error(t, err)
}

func TestService_UpgradeEnvironment_NotEnabled(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{Enabled: false}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v2.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enabled")
}

func TestService_UpgradeEnvironment_DefaultCommandType_Fails(t *testing.T) {
	// Unknown command type → "No command type specified"
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled: true,
		Type:    entities.CommandType("unknown"),
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v2.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upgrade failed")
}

func TestService_UpgradeEnvironment_HTTPSuccess(t *testing.T) {
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
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v2.0")
	assert.NoError(t, err)
}

func TestService_UpgradeEnvironment_SSHNoCredentials(t *testing.T) {
	// SSH with no credentials → buildSSHTarget fails
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
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
	assert.Error(t, err)
}

// ---- CheckHealth (success path needs a real HTTP endpoint) ----

func TestService_CheckHealth_Success(t *testing.T) {
	// HTTP server to act as healthy endpoint
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
	env.Status = entities.Status{
		Health:    entities.HealthStatusUnknown,
		LastCheck: time.Now(),
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.CheckHealth(context.Background(), id.Hex())
	assert.NoError(t, err)
}

// ---- GetAvailableVersions (HTTP fetch) ----

func TestService_GetAvailableVersions_HTTPSuccess_JSONArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["v1.0","v1.1","v2.0"]`))
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
	env.SystemInfo.AppVersion = "v1.0"

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	versions, current, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.NoError(t, err)
	assert.Equal(t, "v1.0", current)
	assert.Equal(t, []string{"v1.0", "v1.1", "v2.0"}, versions)
}

func TestService_GetAvailableVersions_HTTPSuccess_JSONPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"releases":["v3.0","v3.1"]}`))
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
	assert.Equal(t, []string{"v3.0", "v3.1"}, versions)
}

func TestService_GetAvailableVersions_HTTP_NonSuccess(t *testing.T) {
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
		Enabled:        true,
		VersionListURL: srv.URL,
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
}

func TestService_GetAvailableVersions_WithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["v4.0"]`))
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
		VersionListBody:    `{"filter":"stable"}`,
		VersionListHeaders: map[string]string{"X-Custom": "test"},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	versions, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.NoError(t, err)
	assert.Equal(t, []string{"v4.0"}, versions)
}

func TestService_CheckHealth_UpdateStatusError(t *testing.T) {
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

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(assert.AnError)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.CheckHealth(context.Background(), id.Hex())
	assert.Error(t, err)
}

package environment_test

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

func TestService_GetAvailableVersions_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
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

func TestService_GetAvailableVersions_InvalidJSONPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"releases":["v1"]}`))
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
		JSONPathResponse: "invalid json path",
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
}

func TestService_GetAvailableVersions_SSRF(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"example.com"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled:        true,
		VersionListURL: "http://127.0.0.1/versions",
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestService_RestartEnvironment_InvalidURL(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.Commands = entities.CommandConfig{
		Type: entities.CommandTypeHTTP,
		Restart: entities.RestartConfig{
			Enabled: true,
			URL:     "://invalid-url",
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.RestartEnvironment(context.Background(), id.Hex(), false)
	assert.Error(t, err)
}

func TestService_UpgradeEnvironment_InvalidURL(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestServiceWithAllowedHosts(repo, logRepo, []string{"127.0.0.1"})

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{
		Enabled: true,
		Type:    entities.CommandTypeHTTP,
		UpgradeCommand: entities.CommandDetails{
			URL: "://invalid",
		},
	}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.UpgradeEnvironment(context.Background(), id.Hex(), "v2")
	assert.Error(t, err)
}

package environment_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/domain/errors"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/environment"
	"app-env-manager/internal/service/health"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// helper to build a service with real SSH manager and health checker (not used in these tests)
func newTestService(repo *MockEnvironmentRepository, logRepo *MockLogRepository) *environment.Service {
	sshMgr := ssh.NewManager(ssh.Config{
		ConnectionTimeout: time.Second,
		CommandTimeout:    time.Second,
		MaxConnections:    1,
	})
	checker := health.NewChecker(time.Second)
	auditRepo := &MockAuditLogRepository{}
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logSvc := log.NewService(logRepo)
	return environment.NewService(repo, auditRepo, sshMgr, checker, logSvc, nil)
}

func newSampleEnv(id primitive.ObjectID) *entities.Environment {
	return &entities.Environment{
		ID:          id,
		Name:        "myenv",
		Description: "test env",
		Target: entities.Target{
			Host: "localhost",
			Port: 22,
		},
		Status: entities.Status{
			Health:    entities.HealthStatusUnknown,
			LastCheck: time.Now(),
		},
	}
}

// ---- GetEnvironment ----

func TestService_GetEnvironment_Success(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	expected := newSampleEnv(id)

	repo.On("GetByID", mock.Anything, id.Hex()).Return(expected, nil)

	result, err := svc.GetEnvironment(context.Background(), id.Hex())

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestService_GetEnvironment_NotFound(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	repo.On("GetByID", mock.Anything, "nonexistent").Return(nil, errors.ErrEnvironmentNotFound)

	result, err := svc.GetEnvironment(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

// ---- ListEnvironments ----

func TestService_ListEnvironments_Success(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	envs := []*entities.Environment{
		newSampleEnv(primitive.NewObjectID()),
		newSampleEnv(primitive.NewObjectID()),
	}

	filter := interfaces.ListFilter{Pagination: &interfaces.Pagination{Page: 1, Limit: 10}}
	repo.On("List", mock.Anything, filter).Return(envs, nil)

	result, err := svc.ListEnvironments(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

func TestService_ListEnvironments_Error(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	filter := interfaces.ListFilter{}
	repo.On("List", mock.Anything, filter).Return(nil, fmt.Errorf("db error"))

	result, err := svc.ListEnvironments(context.Background(), filter)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ---- CreateEnvironment ----

func TestService_CreateEnvironment_Success(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	req := environment.CreateEnvironmentRequest{
		Name:        "newenv",
		Description: "desc",
		Target: entities.Target{
			Host: "host.local",
			Port: 22,
		},
		Credentials: entities.CredentialRef{
			Type: "password",
		},
		HealthCheck: entities.HealthCheckConfig{
			Enabled: false, // Disabled so no async health check goroutine runs
		},
	}

	// GetByName returns nil (no duplicate)
	repo.On("GetByName", mock.Anything, "newenv").Return(nil, errors.ErrEnvironmentNotFound)
	// Create succeeds
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Environment")).Return(nil)
	// Log create call
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	result, err := svc.CreateEnvironment(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "newenv", result.Name)
	repo.AssertExpectations(t)
}

func TestService_CreateEnvironment_DuplicateName(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	existing := newSampleEnv(primitive.NewObjectID())
	existing.Name = "taken"

	repo.On("GetByName", mock.Anything, "taken").Return(existing, nil)

	req := environment.CreateEnvironmentRequest{Name: "taken"}
	result, err := svc.CreateEnvironment(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, errors.ErrEnvironmentAlreadyExists, err)
}

func TestService_CreateEnvironment_RepositoryError(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	repo.On("GetByName", mock.Anything, "newenv2").Return(nil, errors.ErrEnvironmentNotFound)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Environment")).Return(fmt.Errorf("db error"))

	req := environment.CreateEnvironmentRequest{Name: "newenv2"}
	result, err := svc.CreateEnvironment(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ---- DeleteEnvironment ----

func TestService_DeleteEnvironment_Success(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("Delete", mock.Anything, id.Hex()).Return(nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := svc.DeleteEnvironment(context.Background(), id.Hex())

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestService_DeleteEnvironment_NotFound(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	repo.On("GetByID", mock.Anything, "ghost").Return(nil, errors.ErrEnvironmentNotFound)

	err := svc.DeleteEnvironment(context.Background(), "ghost")

	assert.Error(t, err)
}

func TestService_DeleteEnvironment_RepositoryError(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("Delete", mock.Anything, id.Hex()).Return(fmt.Errorf("delete failed"))

	err := svc.DeleteEnvironment(context.Background(), id.Hex())

	assert.Error(t, err)
}

// ---- CheckHealth ----

func TestService_CheckHealth_NotFound(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	repo.On("GetByID", mock.Anything, "bad-id").Return(nil, errors.ErrEnvironmentNotFound)

	err := svc.CheckHealth(context.Background(), "bad-id")
	assert.Error(t, err)
}

// ---- UpdateEnvironment ----

func TestService_UpdateEnvironment_Success(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("GetByName", mock.Anything, "newname").Return(nil, errors.ErrEnvironmentNotFound)
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	req := environment.CreateEnvironmentRequest{
		Name:   "newname",
		Target: entities.Target{Host: "h", Port: 22},
		Credentials: entities.CredentialRef{Type: "password"},
	}

	result, err := svc.UpdateEnvironment(context.Background(), id.Hex(), req)

	assert.NoError(t, err)
	assert.Equal(t, "newname", result.Name)
}

func TestService_UpdateEnvironment_NotFound(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	repo.On("GetByID", mock.Anything, "nope").Return(nil, errors.ErrEnvironmentNotFound)

	req := environment.CreateEnvironmentRequest{Name: "x"}
	_, err := svc.UpdateEnvironment(context.Background(), "nope", req)

	assert.Error(t, err)
}

func TestService_UpdateEnvironment_NameConflict(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	other := primitive.NewObjectID()
	env := newSampleEnv(id)
	conflicting := newSampleEnv(other)
	conflicting.Name = "taken"

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("GetByName", mock.Anything, "taken").Return(conflicting, nil)

	req := environment.CreateEnvironmentRequest{Name: "taken"}
	_, err := svc.UpdateEnvironment(context.Background(), id.Hex(), req)

	assert.Error(t, err)
	assert.Equal(t, errors.ErrEnvironmentAlreadyExists, err)
}

// ---- GetAvailableVersions ----

func TestService_GetAvailableVersions_NotEnabled(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{Enabled: false}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enabled")
}

func TestService_GetAvailableVersions_NoURL(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.UpgradeConfig = entities.UpgradeConfig{Enabled: true, VersionListURL: ""}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	_, _, err := svc.GetAvailableVersions(context.Background(), id.Hex())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestService_GetAvailableVersions_NotFound(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	repo.On("GetByID", mock.Anything, "nope").Return(nil, errors.ErrEnvironmentNotFound)

	_, _, err := svc.GetAvailableVersions(context.Background(), "nope")
	assert.Error(t, err)
}

// ---- UpdateEnvironmentPartial ----

func TestService_UpdateEnvironmentPartial_DescriptionOnly(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	newDesc := "updated description"
	req := environment.UpdateEnvironmentRequest{Description: &newDesc}

	result, err := svc.UpdateEnvironmentPartial(context.Background(), id.Hex(), req)

	assert.NoError(t, err)
	assert.Equal(t, "updated description", result.Description)
}

func TestService_UpdateEnvironmentPartial_NameConflict(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	otherId := primitive.NewObjectID()
	env := newSampleEnv(id)
	conflicting := newSampleEnv(otherId)
	conflicting.Name = "taken"

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("GetByName", mock.Anything, "taken").Return(conflicting, nil)

	takenName := "taken"
	req := environment.UpdateEnvironmentRequest{Name: &takenName}

	result, err := svc.UpdateEnvironmentPartial(context.Background(), id.Hex(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestService_UpdateEnvironmentPartial_AllFields(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	newURL := "http://newenv.example.com"
	newTarget := entities.Target{Host: "newhost", Port: 443}
	newCreds := entities.CredentialRef{Type: "key", Username: "deploy"}
	newCommands := entities.CommandConfig{Type: entities.CommandTypeHTTP}
	newUpgrade := entities.UpgradeConfig{Enabled: true}
	newMeta := map[string]interface{}{"env": "staging"}

	req := environment.UpdateEnvironmentRequest{
		EnvironmentURL: &newURL,
		Target:         &newTarget,
		Credentials:    &newCreds,
		Commands:       &newCommands,
		UpgradeConfig:  &newUpgrade,
		Metadata:       newMeta,
	}

	result, err := svc.UpdateEnvironmentPartial(context.Background(), id.Hex(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "http://newenv.example.com", result.EnvironmentURL)
	assert.Equal(t, "newhost", result.Target.Host)
}

func TestService_UpdateEnvironmentPartial_UpdateError(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(fmt.Errorf("db error"))

	newDesc := "new"
	req := environment.UpdateEnvironmentRequest{Description: &newDesc}
	_, err := svc.UpdateEnvironmentPartial(context.Background(), id.Hex(), req)
	assert.Error(t, err)
}

func TestService_UpdateEnvironmentPartial_HealthCheckDisabled(t *testing.T) {
	repo := new(MockEnvironmentRepository)
	logRepo := new(MockLogRepository)
	svc := newTestService(repo, logRepo)

	id := primitive.NewObjectID()
	env := newSampleEnv(id)
	env.HealthCheck = entities.HealthCheckConfig{Enabled: true}

	repo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	repo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(nil)
	repo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	disabled := entities.HealthCheckConfig{Enabled: false}
	req := environment.UpdateEnvironmentRequest{HealthCheck: &disabled}

	result, err := svc.UpdateEnvironmentPartial(context.Background(), id.Hex(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.HealthCheck.Enabled)
}

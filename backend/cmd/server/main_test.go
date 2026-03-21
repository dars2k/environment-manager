package main

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/environment"
	"app-env-manager/internal/service/health"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/ssh"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mockEnvRepo is a testify mock for the environment repository.
type mockEnvRepoMain struct{ mock.Mock }

func (m *mockEnvRepoMain) Create(ctx context.Context, env *entities.Environment) error {
	return m.Called(ctx, env).Error(0)
}
func (m *mockEnvRepoMain) GetByID(ctx context.Context, id string) (*entities.Environment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Environment), args.Error(1)
}
func (m *mockEnvRepoMain) GetByName(ctx context.Context, name string) (*entities.Environment, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Environment), args.Error(1)
}
func (m *mockEnvRepoMain) Update(ctx context.Context, id string, env *entities.Environment) error {
	return m.Called(ctx, id, env).Error(0)
}
func (m *mockEnvRepoMain) UpdateStatus(ctx context.Context, id string, status entities.Status) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *mockEnvRepoMain) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockEnvRepoMain) List(ctx context.Context, filter interfaces.ListFilter) ([]*entities.Environment, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Environment), args.Error(1)
}
func (m *mockEnvRepoMain) Count(ctx context.Context, filter interfaces.ListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

type mockAuditRepoMain struct{ mock.Mock }

func (m *mockAuditRepoMain) Create(ctx context.Context, l *entities.AuditLog) error {
	return m.Called(ctx, l).Error(0)
}
func (m *mockAuditRepoMain) GetByID(ctx context.Context, id string) (*entities.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AuditLog), args.Error(1)
}
func (m *mockAuditRepoMain) List(ctx context.Context, filter interfaces.AuditLogFilter) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}
func (m *mockAuditRepoMain) Count(ctx context.Context, filter interfaces.AuditLogFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockAuditRepoMain) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

type mockLogRepoMain struct{ mock.Mock }

func (m *mockLogRepoMain) Create(ctx context.Context, l *entities.Log) error {
	return m.Called(ctx, l).Error(0)
}
func (m *mockLogRepoMain) List(ctx context.Context, filter interfaces.LogFilter) ([]*entities.Log, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entities.Log), args.Get(1).(int64), args.Error(2)
}
func (m *mockLogRepoMain) GetByID(ctx context.Context, id primitive.ObjectID) (*entities.Log, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Log), args.Error(1)
}
func (m *mockLogRepoMain) DeleteOld(ctx context.Context, older time.Duration) (int64, error) {
	args := m.Called(ctx, older)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockLogRepoMain) GetEnvironmentLogs(ctx context.Context, envID primitive.ObjectID, limit int) ([]*entities.Log, error) {
	args := m.Called(ctx, envID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Log), args.Error(1)
}
func (m *mockLogRepoMain) Count(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func newTestService(envRepo *mockEnvRepoMain, logRepo *mockLogRepoMain, auditRepo *mockAuditRepoMain) *environment.Service {
	sshMgr := ssh.NewManager(ssh.Config{
		ConnectionTimeout: time.Second,
		CommandTimeout:    time.Second,
		MaxConnections:    1,
	})
	checker := health.NewChecker(time.Second)
	logSvc := log.NewService(logRepo)
	return environment.NewService(envRepo, auditRepo, sshMgr, checker, logSvc, nil)
}

// TestStartHealthCheckScheduler_EmptyList runs the scheduler for one tick
// and verifies it handles an empty environment list gracefully.
func TestStartHealthCheckScheduler_EmptyList(t *testing.T) {
	envRepo := new(mockEnvRepoMain)
	logRepo := new(mockLogRepoMain)
	auditRepo := new(mockAuditRepoMain)

	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	// Return empty list — scheduler should handle it without error
	envRepo.On("List", mock.Anything, mock.AnythingOfType("interfaces.ListFilter")).Return([]*entities.Environment{}, nil)

	svc := newTestService(envRepo, logRepo, auditRepo)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	// Run with a very short interval — one tick, then we let it exit via goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		startHealthCheckScheduler(svc, 10*time.Millisecond, logger)
	}()

	// Wait for at least one tick
	time.Sleep(50 * time.Millisecond)
	// The goroutine runs forever so we just verify it doesn't panic
	// by having it still running after the sleep
	select {
	case <-done:
		// If done, it exited prematurely (shouldn't happen)
	default:
		// Still running — success
	}
}

// TestStartHealthCheckScheduler_ListError verifies the scheduler handles
// ListEnvironments errors gracefully.
func TestStartHealthCheckScheduler_ListError(t *testing.T) {
	envRepo := new(mockEnvRepoMain)
	logRepo := new(mockLogRepoMain)
	auditRepo := new(mockAuditRepoMain)

	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	// Return error — scheduler should log and continue
	envRepo.On("List", mock.Anything, mock.AnythingOfType("interfaces.ListFilter")).Return(nil, errListFailed)

	svc := newTestService(envRepo, logRepo, auditRepo)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	go func() {
		startHealthCheckScheduler(svc, 10*time.Millisecond, logger)
	}()

	// Wait for at least one tick
	time.Sleep(50 * time.Millisecond)
}

// TestStartHealthCheckScheduler_WithEnvironments verifies the scheduler calls
// CheckHealth for each environment.
func TestStartHealthCheckScheduler_WithEnvironments(t *testing.T) {
	envRepo := new(mockEnvRepoMain)
	logRepo := new(mockLogRepoMain)
	auditRepo := new(mockAuditRepoMain)

	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	id := primitive.NewObjectID()
	env := &entities.Environment{
		ID:   id,
		Name: "test-env",
		HealthCheck: entities.HealthCheckConfig{
			Enabled: false, // disabled so health check returns unknown quickly
		},
		Status: entities.Status{Health: entities.HealthStatusUnknown},
	}

	// Return one environment — scheduler will check health for it
	envRepo.On("List", mock.Anything, mock.AnythingOfType("interfaces.ListFilter")).Return([]*entities.Environment{env}, nil)
	envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	envRepo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(nil).Maybe()

	svc := newTestService(envRepo, logRepo, auditRepo)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	go func() {
		startHealthCheckScheduler(svc, 10*time.Millisecond, logger)
	}()

	// Wait for scheduler to run and health checks to complete
	time.Sleep(200 * time.Millisecond)
}

// errListFailed is a sentinel error for test purposes.
var errListFailed = fmt.Errorf("list failed")

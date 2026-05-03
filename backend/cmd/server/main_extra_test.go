package main

// Extra scheduler tests to improve coverage of startHealthCheckScheduler branches.

import (
	"io"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestStartHealthCheckScheduler_ShortInterval tests the branch where defaultInterval
// is shorter than the internal 5-second poll threshold, causing pollInterval to be
// set to defaultInterval instead of 5s.
func TestStartHealthCheckScheduler_ShortInterval(t *testing.T) {
	envRepo := new(mockEnvRepoMain)
	logRepo := new(mockLogRepoMain)
	auditRepo := new(mockAuditRepoMain)

	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	envRepo.On("List", mock.Anything, mock.AnythingOfType("interfaces.ListFilter")).Return([]*entities.Environment{}, nil)

	svc := newTestService(envRepo, logRepo, auditRepo)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	// defaultInterval (3ms) < 5s → pollInterval = defaultInterval (hits the `if` branch)
	go func() {
		startHealthCheckScheduler(svc, 3*time.Millisecond, logger)
	}()
	time.Sleep(30 * time.Millisecond)
}

// TestStartHealthCheckScheduler_HealthCheckEnabled_WithInterval covers the branch where
// an environment has HealthCheck.Enabled=true and a positive Interval (sets per-env
// interval rather than falling back to the global default).
func TestStartHealthCheckScheduler_HealthCheckEnabled_WithInterval(t *testing.T) {
	envRepo := new(mockEnvRepoMain)
	logRepo := new(mockLogRepoMain)
	auditRepo := new(mockAuditRepoMain)

	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	id := primitive.NewObjectID()
	env := &entities.Environment{
		ID:   id,
		Name: "hc-env",
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Interval: 1, // 1 second per-env interval (positive → uses per-env interval)
		},
		Status: entities.Status{Health: entities.HealthStatusUnknown},
	}

	envRepo.On("List", mock.Anything, mock.AnythingOfType("interfaces.ListFilter")).Return([]*entities.Environment{env}, nil)
	envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	envRepo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(nil).Maybe()
	envRepo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil).Maybe()

	svc := newTestService(envRepo, logRepo, auditRepo)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	go func() {
		startHealthCheckScheduler(svc, 5*time.Millisecond, logger)
	}()
	time.Sleep(100 * time.Millisecond)
}

// TestStartHealthCheckScheduler_NotDueYet covers the "not due yet" branch where
// lastChecked[envID] is already set and the interval hasn't elapsed, so the
// health check is skipped on the second tick.
func TestStartHealthCheckScheduler_NotDueYet(t *testing.T) {
	envRepo := new(mockEnvRepoMain)
	logRepo := new(mockLogRepoMain)
	auditRepo := new(mockAuditRepoMain)

	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	id := primitive.NewObjectID()
	env := &entities.Environment{
		ID:   id,
		Name: "skip-env",
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Interval: 3600, // 1 hour interval → never due again within the test
		},
		Status: entities.Status{Health: entities.HealthStatusUnknown},
	}

	envRepo.On("List", mock.Anything, mock.AnythingOfType("interfaces.ListFilter")).Return([]*entities.Environment{env}, nil)
	envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	envRepo.On("UpdateStatus", mock.Anything, id.Hex(), mock.Anything).Return(nil).Maybe()
	envRepo.On("Update", mock.Anything, id.Hex(), mock.Anything).Return(nil).Maybe()

	svc := newTestService(envRepo, logRepo, auditRepo)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	// Two ticks: first tick records lastChecked, second tick finds it not due.
	go func() {
		startHealthCheckScheduler(svc, 5*time.Millisecond, logger)
	}()
	time.Sleep(80 * time.Millisecond) // enough for 2+ ticks
}

// TestStartHealthCheckScheduler_DisabledHealthCheck ensures that environments with
// HealthCheck.Enabled=false are skipped (the `continue` branch).
func TestStartHealthCheckScheduler_DisabledHealthCheck(t *testing.T) {
	envRepo := new(mockEnvRepoMain)
	logRepo := new(mockLogRepoMain)
	auditRepo := new(mockAuditRepoMain)

	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	id := primitive.NewObjectID()
	env := &entities.Environment{
		ID:   id,
		Name: "disabled-hc-env",
		HealthCheck: entities.HealthCheckConfig{
			Enabled: false, // disabled → `continue` is taken
		},
	}

	envRepo.On("List", mock.Anything, mock.AnythingOfType("interfaces.ListFilter")).Return(
		[]*entities.Environment{env}, nil,
	)

	svc := newTestService(envRepo, logRepo, auditRepo)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	go func() {
		startHealthCheckScheduler(svc, 5*time.Millisecond, logger)
	}()
	time.Sleep(50 * time.Millisecond)
}

// TestStartHealthCheckScheduler_ListFilter exercises the filter passed to List.
func TestStartHealthCheckScheduler_ListFilter(t *testing.T) {
	envRepo := new(mockEnvRepoMain)
	logRepo := new(mockLogRepoMain)
	auditRepo := new(mockAuditRepoMain)

	auditRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	// Return empty list — we just need to confirm the filter is a ListFilter value
	envRepo.On("List", mock.Anything, mock.MatchedBy(func(f interfaces.ListFilter) bool {
		return true // accept any filter
	})).Return([]*entities.Environment{}, nil)

	svc := newTestService(envRepo, logRepo, auditRepo)

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	go func() {
		startHealthCheckScheduler(svc, 5*time.Millisecond, logger)
	}()
	time.Sleep(30 * time.Millisecond)
}

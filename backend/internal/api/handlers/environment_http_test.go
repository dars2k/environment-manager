package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/api/dto"
	"app-env-manager/internal/api/handlers"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/domain/errors"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/environment"
	"app-env-manager/internal/service/health"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/ssh"
	"app-env-manager/internal/websocket/hub"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// envTestMockEnvRepo satisfies interfaces.EnvironmentRepository (correct Count signature).
type envTestMockEnvRepo struct{ mock.Mock }

func (m *envTestMockEnvRepo) Create(ctx context.Context, env *entities.Environment) error {
	return m.Called(ctx, env).Error(0)
}
func (m *envTestMockEnvRepo) GetByID(ctx context.Context, id string) (*entities.Environment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Environment), args.Error(1)
}
func (m *envTestMockEnvRepo) GetByName(ctx context.Context, name string) (*entities.Environment, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Environment), args.Error(1)
}
func (m *envTestMockEnvRepo) Update(ctx context.Context, id string, env *entities.Environment) error {
	return m.Called(ctx, id, env).Error(0)
}
func (m *envTestMockEnvRepo) UpdateStatus(ctx context.Context, id string, status entities.Status) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *envTestMockEnvRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *envTestMockEnvRepo) List(ctx context.Context, filter interfaces.ListFilter) ([]*entities.Environment, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Environment), args.Error(1)
}
func (m *envTestMockEnvRepo) Count(ctx context.Context, filter interfaces.ListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// envTestMockAuditRepo satisfies interfaces.AuditLogRepository for handler tests.
type envTestMockAuditRepo struct{ mock.Mock }

func (m *envTestMockAuditRepo) Create(ctx context.Context, log *entities.AuditLog) error {
	return m.Called(ctx, log).Error(0)
}
func (m *envTestMockAuditRepo) GetByID(ctx context.Context, id string) (*entities.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AuditLog), args.Error(1)
}
func (m *envTestMockAuditRepo) List(ctx context.Context, filter interfaces.AuditLogFilter) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}
func (m *envTestMockAuditRepo) Count(ctx context.Context, filter interfaces.AuditLogFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}
func (m *envTestMockAuditRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

// handlerSetup builds an EnvironmentHandler with mocked repos.
type handlerSetup struct {
	envRepo   *envTestMockEnvRepo
	logRepo   *MockLogRepository
	auditRepo *envTestMockAuditRepo
	hub       *hub.Hub
	handler   *handlers.EnvironmentHandler
}

func newHandlerSetup(t *testing.T) *handlerSetup {
	t.Helper()
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	envRepo := new(envTestMockEnvRepo)
	logRepo := new(MockLogRepository)
	auditRepo := new(envTestMockAuditRepo)
	logSvc := log.NewService(logRepo)

	sshMgr := ssh.NewManager(ssh.Config{
		ConnectionTimeout: time.Second,
		CommandTimeout:    time.Second,
		MaxConnections:    1,
	})
	checker := health.NewChecker(time.Second)

	svc := environment.NewService(envRepo, auditRepo, sshMgr, checker, logSvc, nil)

	h := hub.NewHub(logger)
	go h.Run()

	envHandler := handlers.NewEnvironmentHandler(svc, h, logger)

	return &handlerSetup{
		envRepo:   envRepo,
		logRepo:   logRepo,
		auditRepo: auditRepo,
		hub:       h,
		handler:   envHandler,
	}
}

func sampleEnvForHandler(id primitive.ObjectID) *entities.Environment {
	return &entities.Environment{
		ID:          id,
		Name:        "testenv",
		Description: "test environment",
		Target: entities.Target{
			Host: "localhost",
			Port: 22,
		},
		Status: entities.Status{
			Health:    entities.HealthStatusHealthy,
			LastCheck: time.Now(),
		},
	}
}

// ---- List ----

func TestEnvironmentHandler_List_Success(t *testing.T) {
	s := newHandlerSetup(t)

	id := primitive.NewObjectID()
	envs := []*entities.Environment{sampleEnvForHandler(id)}

	s.envRepo.On("List", mock.Anything, mock.AnythingOfType("interfaces.ListFilter")).Return(envs, nil)

	req := httptest.NewRequest("GET", "/api/environments", nil)
	w := httptest.NewRecorder()

	s.handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.SuccessResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

func TestEnvironmentHandler_List_Error(t *testing.T) {
	s := newHandlerSetup(t)

	s.envRepo.On("List", mock.Anything, mock.AnythingOfType("interfaces.ListFilter")).Return(nil, errors.NewInternalError(fmt.Errorf("db error")))

	req := httptest.NewRequest("GET", "/api/environments", nil)
	w := httptest.NewRecorder()

	s.handler.List(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- Get ----

func TestEnvironmentHandler_Get_Success(t *testing.T) {
	s := newHandlerSetup(t)

	id := primitive.NewObjectID()
	env := sampleEnvForHandler(id)
	s.envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	req := httptest.NewRequest("GET", "/api/environments/"+id.Hex(), nil)
	req = muxSetVar(req, "id", id.Hex())
	w := httptest.NewRecorder()

	s.handler.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.SuccessResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

func TestEnvironmentHandler_Get_NotFound(t *testing.T) {
	s := newHandlerSetup(t)

	s.envRepo.On("GetByID", mock.Anything, "abc123").Return(nil, errors.ErrEnvironmentNotFound)

	req := httptest.NewRequest("GET", "/api/environments/abc123", nil)
	req = muxSetVar(req, "id", "abc123")
	w := httptest.NewRecorder()

	s.handler.Get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- Create ----

func TestEnvironmentHandler_Create_Success(t *testing.T) {
	s := newHandlerSetup(t)

	body := environment.CreateEnvironmentRequest{
		Name: "newenv",
		Target: entities.Target{
			Host: "host.local",
			Port: 22,
		},
		Credentials: entities.CredentialRef{Type: "password"},
		HealthCheck: entities.HealthCheckConfig{Enabled: false},
	}
	bodyBytes, _ := json.Marshal(body)

	s.envRepo.On("GetByName", mock.Anything, "newenv").Return(nil, errors.ErrEnvironmentNotFound)
	s.envRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Environment")).Return(nil)
	s.logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	req := httptest.NewRequest("POST", "/api/environments", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestEnvironmentHandler_Create_InvalidJSON(t *testing.T) {
	s := newHandlerSetup(t)

	req := httptest.NewRequest("POST", "/api/environments", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnvironmentHandler_Create_DuplicateName(t *testing.T) {
	s := newHandlerSetup(t)

	existing := sampleEnvForHandler(primitive.NewObjectID())
	s.envRepo.On("GetByName", mock.Anything, "newenv").Return(existing, nil)

	body := environment.CreateEnvironmentRequest{
		Name: "newenv",
		Target: entities.Target{Host: "h", Port: 22},
		Credentials: entities.CredentialRef{Type: "password"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/environments", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handler.Create(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// ---- Update ----

func TestEnvironmentHandler_Update_Success(t *testing.T) {
	s := newHandlerSetup(t)

	id := primitive.NewObjectID()
	env := sampleEnvForHandler(id)

	s.envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	s.envRepo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(nil)
	s.logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	newDesc := "updated"
	body := environment.UpdateEnvironmentRequest{Description: &newDesc}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/api/environments/"+id.Hex(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = muxSetVar(req, "id", id.Hex())
	w := httptest.NewRecorder()

	s.handler.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEnvironmentHandler_Update_InvalidJSON(t *testing.T) {
	s := newHandlerSetup(t)

	req := httptest.NewRequest("PUT", "/api/environments/x", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	req = muxSetVar(req, "id", "x")
	w := httptest.NewRecorder()

	s.handler.Update(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---- Delete ----

func TestEnvironmentHandler_Delete_Success(t *testing.T) {
	s := newHandlerSetup(t)

	id := primitive.NewObjectID()
	env := sampleEnvForHandler(id)

	s.envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	s.envRepo.On("Delete", mock.Anything, id.Hex()).Return(nil)
	s.logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	req := httptest.NewRequest("DELETE", "/api/environments/"+id.Hex(), nil)
	req = muxSetVar(req, "id", id.Hex())
	w := httptest.NewRecorder()

	s.handler.Delete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEnvironmentHandler_Delete_NotFound(t *testing.T) {
	s := newHandlerSetup(t)

	s.envRepo.On("GetByID", mock.Anything, "missing").Return(nil, errors.ErrEnvironmentNotFound)

	req := httptest.NewRequest("DELETE", "/api/environments/missing", nil)
	req = muxSetVar(req, "id", "missing")
	w := httptest.NewRecorder()

	s.handler.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- Restart ----

func TestEnvironmentHandler_Restart_AcceptsRequest(t *testing.T) {
	s := newHandlerSetup(t)

	body := dto.RestartRequest{Force: false}
	bodyBytes, _ := json.Marshal(body)

	// The background goroutine will call GetByID; allow it
	s.envRepo.On("GetByID", mock.Anything, "env1").Return(nil, errors.ErrEnvironmentNotFound).Maybe()

	req := httptest.NewRequest("POST", "/api/environments/env1/restart", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = muxSetVar(req, "id", "env1")
	w := httptest.NewRecorder()

	// The handler returns 202 immediately; the actual restart runs in background
	s.handler.Restart(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	var resp dto.SuccessResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	// Give background goroutine a moment to complete (avoids mock panic after test)
	time.Sleep(10 * time.Millisecond)
}

func TestEnvironmentHandler_Restart_NoBody(t *testing.T) {
	s := newHandlerSetup(t)

	s.envRepo.On("GetByID", mock.Anything, "env1").Return(nil, errors.ErrEnvironmentNotFound).Maybe()

	req := httptest.NewRequest("POST", "/api/environments/env1/restart", http.NoBody)
	req = muxSetVar(req, "id", "env1")
	w := httptest.NewRecorder()

	s.handler.Restart(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	time.Sleep(10 * time.Millisecond)
}

// ---- CheckHealth ----

func TestEnvironmentHandler_CheckHealth_NotFound(t *testing.T) {
	s := newHandlerSetup(t)

	s.envRepo.On("GetByID", mock.Anything, "missing").Return(nil, errors.ErrEnvironmentNotFound)

	req := httptest.NewRequest("POST", "/api/environments/missing/check-health", nil)
	req = muxSetVar(req, "id", "missing")
	w := httptest.NewRecorder()

	s.handler.CheckHealth(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- GetVersions ----

func TestEnvironmentHandler_GetVersions_NotFound(t *testing.T) {
	s := newHandlerSetup(t)

	s.envRepo.On("GetByID", mock.Anything, "noenv").Return(nil, errors.ErrEnvironmentNotFound)

	req := httptest.NewRequest("GET", "/api/environments/noenv/versions", nil)
	req = muxSetVar(req, "id", "noenv")
	w := httptest.NewRecorder()

	s.handler.GetVersions(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ---- Upgrade ----

func TestEnvironmentHandler_Upgrade_InvalidJSON(t *testing.T) {
	s := newHandlerSetup(t)

	req := httptest.NewRequest("POST", "/api/environments/e1/upgrade", bytes.NewReader([]byte("bad json")))
	req.Header.Set("Content-Type", "application/json")
	req = muxSetVar(req, "id", "e1")
	w := httptest.NewRecorder()

	s.handler.Upgrade(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnvironmentHandler_Upgrade_EmptyVersion(t *testing.T) {
	s := newHandlerSetup(t)

	body := dto.UpgradeRequest{Version: ""}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/environments/e1/upgrade", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = muxSetVar(req, "id", "e1")
	w := httptest.NewRecorder()

	s.handler.Upgrade(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnvironmentHandler_Upgrade_ValidVersion(t *testing.T) {
	s := newHandlerSetup(t)

	// Background goroutine will call GetByID
	s.envRepo.On("GetByID", mock.Anything, "e1").Return(nil, errors.ErrEnvironmentNotFound).Maybe()

	body := dto.UpgradeRequest{Version: "1.2.3"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/environments/e1/upgrade", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = muxSetVar(req, "id", "e1")
	w := httptest.NewRecorder()

	s.handler.Upgrade(w, req)

	// Returns 202 immediately; upgrade happens in background
	assert.Equal(t, http.StatusAccepted, w.Code)
	time.Sleep(10 * time.Millisecond)
}

// muxSetVar injects a Gorilla Mux variable into the request context.
func muxSetVar(r *http.Request, key, value string) *http.Request {
	vars := map[string]string{key: value}
	return mux.SetURLVars(r, vars)
}

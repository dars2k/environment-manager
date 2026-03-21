package handlers_test

// Additional tests to push handler coverage above 90%.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/api/dto"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/domain/errors"
	envservice "app-env-manager/internal/service/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---- Update handler: service error and not-found ----

func TestEnvironmentHandler_Update_NotFound(t *testing.T) {
	s := newHandlerSetup(t)

	s.envRepo.On("GetByID", mock.Anything, "ghost").Return(nil, errors.ErrEnvironmentNotFound)

	newDesc := "updated"
	body := envservice.UpdateEnvironmentRequest{Description: &newDesc}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/api/environments/ghost", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = muxSetVar(req, "id", "ghost")
	w := httptest.NewRecorder()

	s.handler.Update(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEnvironmentHandler_Update_ServiceError(t *testing.T) {
	s := newHandlerSetup(t)

	id := primitive.NewObjectID()
	env := sampleEnvForHandler(id)

	// The partial update calls GetByID first; return the env
	s.envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	// Then Update fails
	s.envRepo.On("Update", mock.Anything, id.Hex(), mock.AnythingOfType("*entities.Environment")).Return(fmt.Errorf("update failed"))

	newDesc := "new desc"
	body := envservice.UpdateEnvironmentRequest{Description: &newDesc}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/api/environments/"+id.Hex(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = muxSetVar(req, "id", id.Hex())
	w := httptest.NewRecorder()

	s.handler.Update(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- Create handler: validation error ----

func TestEnvironmentHandler_Create_ValidationError(t *testing.T) {
	s := newHandlerSetup(t)

	// Name too short (< 3 chars) — triggers validator.Struct error
	body := envservice.CreateEnvironmentRequest{
		Name: "ab", // less than min=3
		Target: entities.Target{Host: "h", Port: 22},
		Credentials: entities.CredentialRef{Type: "password"},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/environments", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ---- CheckHealth handler: success case ----

func TestEnvironmentHandler_CheckHealth_Success(t *testing.T) {
	s := newHandlerSetup(t)

	id := primitive.NewObjectID()
	env := sampleEnvForHandler(id)
	env.HealthCheck = entities.HealthCheckConfig{Enabled: false}

	s.envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	s.envRepo.On("UpdateStatus", mock.Anything, id.Hex(), mock.AnythingOfType("entities.Status")).Return(nil).Maybe()
	s.logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	req := httptest.NewRequest("POST", "/api/environments/"+id.Hex()+"/check-health", nil)
	req = muxSetVar(req, "id", id.Hex())
	w := httptest.NewRecorder()

	s.handler.CheckHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp dto.SuccessResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
}

// ---- GetVersions handler: upgrade disabled ----

func TestEnvironmentHandler_GetVersions_UpgradeNotEnabled(t *testing.T) {
	s := newHandlerSetup(t)

	id := primitive.NewObjectID()
	env := sampleEnvForHandler(id)
	env.UpgradeConfig = entities.UpgradeConfig{Enabled: false}

	s.envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)

	req := httptest.NewRequest("GET", "/api/environments/"+id.Hex()+"/versions", nil)
	req = muxSetVar(req, "id", id.Hex())
	w := httptest.NewRecorder()

	s.handler.GetVersions(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- Restart: invalid JSON body still returns 202 ----

func TestEnvironmentHandler_Restart_InvalidJSON_Still202(t *testing.T) {
	s := newHandlerSetup(t)

	s.envRepo.On("GetByID", mock.Anything, "e2").Return(nil, errors.ErrEnvironmentNotFound).Maybe()

	req := httptest.NewRequest("POST", "/api/environments/e2/restart", bytes.NewReader([]byte("bad json")))
	req.Header.Set("Content-Type", "application/json")
	req = muxSetVar(req, "id", "e2")
	w := httptest.NewRecorder()

	s.handler.Restart(w, req)

	// Even with bad JSON body the handler uses default values (force=false) and returns 202
	assert.Equal(t, http.StatusAccepted, w.Code)
	time.Sleep(10 * time.Millisecond)
}

// ---- Upgrade: background failure is logged but not surfaced ----

func TestEnvironmentHandler_Upgrade_ServiceError(t *testing.T) {
	s := newHandlerSetup(t)

	id := primitive.NewObjectID()
	env := sampleEnvForHandler(id)
	env.UpgradeConfig = entities.UpgradeConfig{Enabled: false}

	// Background goroutine will call GetByID
	s.envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil).Maybe()
	s.logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	body := dto.UpgradeRequest{Version: "2.0.0"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/environments/"+id.Hex()+"/upgrade", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = muxSetVar(req, "id", id.Hex())
	w := httptest.NewRecorder()

	s.handler.Upgrade(w, req)

	// Returns 202 immediately; upgrade fails in background but that's not surfaced here
	assert.Equal(t, http.StatusAccepted, w.Code)
	time.Sleep(20 * time.Millisecond)
}

// ---- Delete: service error (not not-found) ----

func TestEnvironmentHandler_Delete_ServiceError(t *testing.T) {
	s := newHandlerSetup(t)

	id := primitive.NewObjectID()
	env := sampleEnvForHandler(id)

	s.envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	s.envRepo.On("Delete", mock.Anything, id.Hex()).Return(fmt.Errorf("delete failed"))
	s.logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	req := httptest.NewRequest("DELETE", "/api/environments/"+id.Hex(), nil)
	req = muxSetVar(req, "id", id.Hex())
	w := httptest.NewRecorder()

	s.handler.Delete(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- List with query params ----

func TestEnvironmentHandler_List_WithQueryParams(t *testing.T) {
	s := newHandlerSetup(t)

	envs := []*entities.Environment{sampleEnvForHandler(primitive.NewObjectID())}
	s.envRepo.On("List", mock.Anything, mock.AnythingOfType("interfaces.ListFilter")).Return(envs, nil)

	req := httptest.NewRequest("GET", "/api/environments?page=2&limit=20&status=healthy", nil)
	w := httptest.NewRecorder()

	s.handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

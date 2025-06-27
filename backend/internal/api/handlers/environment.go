package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"app-env-manager/internal/api/dto"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/domain/errors"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/environment"
	"app-env-manager/internal/websocket/hub"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Request and Response types used by handlers
type (
	RestartRequest struct {
		Force bool `json:"force"`
	}

	ShutdownRequest struct {
		GracefulTimeout int `json:"gracefulTimeout"`
	}

	OperationResponse struct {
		OperationID string `json:"operationId"`
		Status      string `json:"status"`
	}

	MessageResponse struct {
		Message string `json:"message"`
	}

	SuccessResponse struct {
		Success  bool             `json:"success"`
		Data     interface{}      `json:"data"`
		Metadata ResponseMetadata `json:"metadata"`
	}

	ErrorResponse struct {
		Success  bool             `json:"success"`
		Error    ErrorInfo        `json:"error"`
		Metadata ResponseMetadata `json:"metadata"`
	}

	ErrorInfo struct {
		Code    string                 `json:"code"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details,omitempty"`
	}

	ResponseMetadata struct {
		Timestamp string `json:"timestamp"`
		Version   string `json:"version"`
	}
)

// EnvironmentHandler handles environment-related HTTP requests
type EnvironmentHandler struct {
	service   *environment.Service
	validator *validator.Validate
	hub       *hub.Hub
	logger    *logrus.Logger
}

// NewEnvironmentHandler creates a new environment handler
func NewEnvironmentHandler(
	service *environment.Service,
	hub *hub.Hub,
	logger *logrus.Logger,
) *EnvironmentHandler {
	return &EnvironmentHandler{
		service:   service,
		validator: validator.New(),
		hub:       hub,
		logger:    logger,
	}
}

// List handles GET /environments
func (h *EnvironmentHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	filter := parseListFilter(r)

	// Get environments
	envs, err := h.service.ListEnvironments(ctx, filter)
	if err != nil {
		h.respondError(w, err)
		return
	}

	// Redact sensitive fields before response
	redactedEnvs := redactSensitiveFieldsList(envs)

	// Convert to response format
	response := dto.ListEnvironmentsResponse{
		Environments: redactedEnvs,
		Pagination: dto.PaginationResponse{
			Page:  filter.Pagination.Page,
			Limit: filter.Pagination.GetLimit(),
			Total: len(envs), // TODO: Get total count from service
		},
	}

	h.respondJSON(w, http.StatusOK, response)
}

// Get handles GET /environments/{id}
func (h *EnvironmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	env, err := h.service.GetEnvironment(ctx, id)
	if err != nil {
		h.respondError(w, err)
		return
	}

	// Redact sensitive fields before response
	redactedEnv := redactSensitiveFields(env)

	h.respondJSON(w, http.StatusOK, dto.EnvironmentResponse{Environment: redactedEnv})
}

// Create handles POST /environments
func (h *EnvironmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req environment.CreateEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, errors.NewValidationError("body", "invalid JSON"))
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.respondError(w, errors.NewValidationError("validation", err.Error()))
		return
	}

	// Create environment
	env, err := h.service.CreateEnvironment(ctx, req)
	if err != nil {
		h.respondError(w, err)
		return
	}

	// Redact sensitive fields before response
	redactedEnv := redactSensitiveFields(env)

	// Broadcast creation
	h.hub.BroadcastEnvironmentUpdate(env.ID.Hex(), map[string]interface{}{
		"action": "created",
		"environment": redactedEnv,
	})

	h.respondJSON(w, http.StatusCreated, redactedEnv)
}

// Update handles PUT /environments/{id}
func (h *EnvironmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	var req environment.UpdateEnvironmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, errors.NewValidationError("body", "invalid JSON"))
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		h.respondError(w, errors.NewValidationError("validation", err.Error()))
		return
	}

	// Update environment with partial update
	env, err := h.service.UpdateEnvironmentPartial(ctx, id, req)
	if err != nil {
		h.respondError(w, err)
		return
	}

	// Redact sensitive fields before response
	redactedEnv := redactSensitiveFields(env)

	// Broadcast update
	h.hub.BroadcastEnvironmentUpdate(env.ID.Hex(), map[string]interface{}{
		"action": "updated",
		"environment": redactedEnv,
	})

	h.respondJSON(w, http.StatusOK, redactedEnv)
}

// Delete handles DELETE /environments/{id}
func (h *EnvironmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	h.logger.WithFields(logrus.Fields{
		"environmentId": id,
	}).Info("Starting delete environment operation")

	if err := h.service.DeleteEnvironment(ctx, id); err != nil {
		h.logger.WithError(err).WithField("environmentId", id).Error("Delete environment operation failed")
		h.respondError(w, err)
		return
	}

	h.logger.WithField("environmentId", id).Info("Delete environment operation completed successfully")

	// Broadcast deletion
	h.hub.BroadcastEnvironmentUpdate(id, map[string]interface{}{
		"action": "deleted",
	})

	h.respondJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "Environment deleted successfully",
	})
}

// Restart handles POST /environments/{id}/restart
func (h *EnvironmentHandler) Restart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req dto.RestartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use default values if no body
		req = dto.RestartRequest{Force: false}
	}

	operationID := generateOperationID()

	// Start operation asynchronously with a background context
	go func() {
		// Create a new context with timeout for the async operation
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		h.logger.WithFields(logrus.Fields{
			"operationId": operationID,
			"environmentId": id,
			"force": req.Force,
		}).Info("Starting restart operation")

		if err := h.service.RestartEnvironment(bgCtx, id, req.Force); err != nil {
			h.logger.WithError(err).WithField("operationId", operationID).Error("Restart operation failed")
			h.hub.BroadcastOperationUpdate(operationID, map[string]interface{}{
				"status": "failed",
				"error":  err.Error(),
			})
		} else {
			h.logger.WithField("operationId", operationID).Info("Restart operation completed")
			h.hub.BroadcastOperationUpdate(operationID, map[string]interface{}{
				"status": "completed",
			})
		}
	}()

	h.respondJSON(w, http.StatusAccepted, dto.OperationResponse{
		OperationID: operationID,
		Status:      "in_progress",
	})
}


// CheckHealth handles POST /environments/{id}/check-health
func (h *EnvironmentHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.CheckHealth(ctx, id); err != nil {
		h.respondError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "Health check initiated",
	})
}

// GetVersions handles GET /environments/{id}/versions
func (h *EnvironmentHandler) GetVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	availableVersions, currentVersion, err := h.service.GetAvailableVersions(ctx, id)
	if err != nil {
		h.respondError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.VersionsResponse{
		CurrentVersion:    currentVersion,
		AvailableVersions: availableVersions,
	})
}

// Upgrade handles POST /environments/{id}/upgrade
func (h *EnvironmentHandler) Upgrade(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req dto.UpgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, errors.NewValidationError("body", "invalid JSON"))
		return
	}

	if req.Version == "" {
		h.respondError(w, errors.NewValidationError("version", "version is required"))
		return
	}

	operationID := generateOperationID()

	// Start operation asynchronously with a background context
	go func() {
		// Create a new context with timeout for the async operation
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		h.logger.WithFields(logrus.Fields{
			"operationId": operationID,
			"environmentId": id,
			"version": req.Version,
		}).Info("Starting upgrade operation")

		if err := h.service.UpgradeEnvironment(bgCtx, id, req.Version); err != nil {
			h.logger.WithError(err).WithField("operationId", operationID).Error("Upgrade operation failed")
			h.hub.BroadcastOperationUpdate(operationID, map[string]interface{}{
				"status": "failed",
				"error":  err.Error(),
			})
		} else {
			h.logger.WithField("operationId", operationID).Info("Upgrade operation completed")
			h.hub.BroadcastOperationUpdate(operationID, map[string]interface{}{
				"status": "completed",
			})
		}
	}()

	h.respondJSON(w, http.StatusAccepted, dto.OperationResponse{
		OperationID: operationID,
		Status:      "in_progress",
	})
}

// Helper functions

func (h *EnvironmentHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	response := dto.SuccessResponse{
		Success: true,
		Data:    data,
		Metadata: dto.ResponseMetadata{
			Timestamp: currentTimestamp(),
			Version:   "1.0.0",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (h *EnvironmentHandler) respondError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	errorResponse := dto.ErrorInfo{
		Code:    "INTERNAL_ERROR",
		Message: err.Error(),
	}

	// Map domain errors to HTTP status codes
	if domainErr, ok := err.(errors.DomainError); ok {
		errorResponse.Code = domainErr.Code
		errorResponse.Details = domainErr.Details

		switch domainErr.Code {
		case "ENV_NOT_FOUND":
			status = http.StatusNotFound
		case "ENV_DUPLICATE":
			status = http.StatusConflict
		case "VALIDATION_ERROR":
			status = http.StatusBadRequest
		case "AUTH_INVALID", "AUTH_UNAUTHORIZED":
			status = http.StatusUnauthorized
		}
	}

	response := dto.ErrorResponse{
		Success: false,
		Error:   errorResponse,
		Metadata: dto.ResponseMetadata{
			Timestamp: currentTimestamp(),
			Version:   "1.0.0",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func parseListFilter(r *http.Request) interfaces.ListFilter {
	filter := interfaces.ListFilter{
		Pagination: &interfaces.Pagination{
			Page:  1,
			Limit: 200,
		},
	}

	query := r.URL.Query()

	// Parse status filter
	if status := query.Get("status"); status != "" {
		healthStatus := entities.HealthStatus(status)
		filter.Status = &healthStatus
	}

	// Parse pagination
	if page := query.Get("page"); page != "" {
		var p int
		fmt.Sscanf(page, "%d", &p)
		if p > 0 {
			filter.Pagination.Page = p
		}
	}

	if limit := query.Get("limit"); limit != "" {
		var l int
		fmt.Sscanf(limit, "%d", &l)
		if l > 0 && l <= 100 {
			filter.Pagination.Limit = l
		}
	}

	return filter
}

func generateOperationID() string {
	return fmt.Sprintf("op-%s", time.Now().Format("20060102150405"))
}

func currentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// redactSensitiveFields creates a copy of the environment with sensitive fields redacted
func redactSensitiveFields(env *entities.Environment) *entities.Environment {
	if env == nil {
		return nil
	}

	// Create a deep copy to avoid modifying the original
	redacted := *env

	// Redact credentials - never expose passwords or keys
	redacted.Credentials.KeyID = primitive.NilObjectID
	
	// Clear any sensitive metadata
	if redacted.Metadata != nil {
		redactedMeta := make(map[string]interface{})
		for k, v := range redacted.Metadata {
			// Skip known sensitive fields
			if k == "password" || k == "privateKey" || k == "key" || k == "secret" {
				continue
			}
			redactedMeta[k] = v
		}
		redacted.Metadata = redactedMeta
	}

	// Clear sensitive command data
	if redacted.Commands.Restart.Headers != nil {
		redactedHeaders := make(map[string]string)
		for k, v := range redacted.Commands.Restart.Headers {
			// Redact authorization headers
			if k == "Authorization" || k == "X-API-Key" || k == "X-Auth-Token" {
				redactedHeaders[k] = "[REDACTED]"
			} else {
				redactedHeaders[k] = v
			}
		}
		redacted.Commands.Restart.Headers = redactedHeaders
	}

	// Redact upgrade command headers
	if redacted.UpgradeConfig.UpgradeCommand.Headers != nil {
		redactedHeaders := make(map[string]string)
		for k, v := range redacted.UpgradeConfig.UpgradeCommand.Headers {
			if k == "Authorization" || k == "X-API-Key" || k == "X-Auth-Token" {
				redactedHeaders[k] = "[REDACTED]"
			} else {
				redactedHeaders[k] = v
			}
		}
		redacted.UpgradeConfig.UpgradeCommand.Headers = redactedHeaders
	}

	// Redact health check headers
	if redacted.HealthCheck.Headers != nil {
		redactedHeaders := make(map[string]string)
		for k, v := range redacted.HealthCheck.Headers {
			if k == "Authorization" || k == "X-API-Key" || k == "X-Auth-Token" {
				redactedHeaders[k] = "[REDACTED]"
			} else {
				redactedHeaders[k] = v
			}
		}
		redacted.HealthCheck.Headers = redactedHeaders
	}

	return &redacted
}

// redactSensitiveFieldsList redacts sensitive fields from a list of environments
func redactSensitiveFieldsList(envs []*entities.Environment) []*entities.Environment {
	redacted := make([]*entities.Environment, len(envs))
	for i, env := range envs {
		redacted[i] = redactSensitiveFields(env)
	}
	return redacted
}

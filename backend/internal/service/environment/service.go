package environment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/domain/errors"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/health"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/ssh"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles environment business logic
type Service struct {
	repo         interfaces.EnvironmentRepository
	auditRepo    interfaces.AuditLogRepository
	sshManager   *ssh.Manager
	healthChecker *health.Checker
	logService   *log.Service
}

// NewService creates a new environment service
func NewService(
	repo interfaces.EnvironmentRepository,
	auditRepo interfaces.AuditLogRepository,
	sshManager *ssh.Manager,
	healthChecker *health.Checker,
	logService *log.Service,
) *Service {
	return &Service{
		repo:         repo,
		auditRepo:    auditRepo,
		sshManager:   sshManager,
		healthChecker: healthChecker,
		logService:   logService,
	}
}

// CreateEnvironmentRequest represents a request to create an environment
type CreateEnvironmentRequest struct {
	Name           string                      `json:"name" validate:"required,alphanum,min=3,max=50"`
	Description    string                      `json:"description"`
	EnvironmentURL string                      `json:"environmentURL"`
	Target         entities.Target             `json:"target" validate:"required"`
	Credentials    entities.CredentialRef      `json:"credentials" validate:"required"`
	HealthCheck    entities.HealthCheckConfig  `json:"healthCheck"`
	Commands       entities.CommandConfig      `json:"commands"`
	UpgradeConfig  entities.UpgradeConfig      `json:"upgradeConfig"`
	Metadata       map[string]interface{}      `json:"metadata,omitempty"`
}

// UpdateEnvironmentRequest represents a request to update an environment
type UpdateEnvironmentRequest struct {
	Name           *string                      `json:"name,omitempty" validate:"omitempty,alphanum,min=3,max=50"`
	Description    *string                      `json:"description,omitempty"`
	EnvironmentURL *string                      `json:"environmentURL,omitempty"`
	Target         *entities.Target             `json:"target,omitempty"`
	Credentials    *entities.CredentialRef      `json:"credentials,omitempty"`
	HealthCheck    *entities.HealthCheckConfig  `json:"healthCheck,omitempty"`
	Commands       *entities.CommandConfig      `json:"commands,omitempty"`
	UpgradeConfig  *entities.UpgradeConfig      `json:"upgradeConfig,omitempty"`
	Metadata       map[string]interface{}       `json:"metadata,omitempty"`
}

// CreateEnvironment creates a new environment
func (s *Service) CreateEnvironment(ctx context.Context, req CreateEnvironmentRequest) (*entities.Environment, error) {
	// Check for duplicates
	existing, _ := s.repo.GetByName(ctx, req.Name)
	if existing != nil {
		return nil, errors.ErrEnvironmentAlreadyExists
	}

	// Create environment entity
	env := &entities.Environment{
		Name:           req.Name,
		Description:    req.Description,
		EnvironmentURL: req.EnvironmentURL,
		Target:         req.Target,
		Credentials:    req.Credentials,
		HealthCheck:    req.HealthCheck,
		Commands:       req.Commands,
		UpgradeConfig:  req.UpgradeConfig,
		Status: entities.Status{
			Health:    entities.HealthStatusUnknown,
			LastCheck: time.Now(),
			Message:   "Environment created",
		},
		SystemInfo: entities.SystemInfo{
			LastUpdated: time.Now(),
		},
		Metadata: req.Metadata,
	}

	// Store in repository
	if err := s.repo.Create(ctx, env); err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Create log entry
	_ = s.logService.LogEnvironmentAction(ctx, env, entities.ActionTypeCreate, "Environment created", map[string]interface{}{
		"name": env.Name,
		"target": fmt.Sprintf("%s:%d", env.Target.Host, env.Target.Port),
	})

	// Trigger initial health check asynchronously only if health check is enabled
	if env.HealthCheck.Enabled {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = s.CheckHealth(ctx, env.ID.Hex())
		}()
	}
	// If health check is disabled, keep status as Unknown

	return env, nil
}

// GetEnvironment retrieves an environment by ID
func (s *Service) GetEnvironment(ctx context.Context, id string) (*entities.Environment, error) {
	return s.repo.GetByID(ctx, id)
}

// ListEnvironments lists all environments
func (s *Service) ListEnvironments(ctx context.Context, filter interfaces.ListFilter) ([]*entities.Environment, error) {
	return s.repo.List(ctx, filter)
}

// UpdateEnvironment updates an environment
func (s *Service) UpdateEnvironment(ctx context.Context, id string, req CreateEnvironmentRequest) (*entities.Environment, error) {
	// Get existing environment
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check for name conflicts if name is changing
	if env.Name != req.Name {
		existing, _ := s.repo.GetByName(ctx, req.Name)
		if existing != nil && existing.ID.Hex() != id {
			return nil, errors.ErrEnvironmentAlreadyExists
		}
	}

	// Update fields
	oldEnv := *env
	env.Name = req.Name
	env.Description = req.Description
	env.EnvironmentURL = req.EnvironmentURL
	env.Target = req.Target
	env.Credentials = req.Credentials
	env.HealthCheck = req.HealthCheck
	env.Commands = req.Commands
	env.UpgradeConfig = req.UpgradeConfig
	env.Metadata = req.Metadata

	// Update in repository
	if err := s.repo.Update(ctx, id, env); err != nil {
		return nil, err
	}

	// Log the update
	_ = s.logService.LogEnvironmentAction(ctx, env, entities.ActionTypeUpdate, "Environment updated", map[string]interface{}{
		"changes": map[string]interface{}{
			"name": map[string]string{"from": oldEnv.Name, "to": env.Name},
			"description": map[string]string{"from": oldEnv.Description, "to": env.Description},
		},
	})

	return env, nil
}

// UpdateEnvironmentPartial updates only provided fields of an environment
func (s *Service) UpdateEnvironmentPartial(ctx context.Context, id string, req UpdateEnvironmentRequest) (*entities.Environment, error) {
	// Get existing environment
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Track changes for audit log
	changes := make(map[string]interface{})

	// Check for name conflicts if name is changing
	if req.Name != nil && env.Name != *req.Name {
		existing, _ := s.repo.GetByName(ctx, *req.Name)
		if existing != nil && existing.ID.Hex() != id {
			return nil, errors.ErrEnvironmentAlreadyExists
		}
		changes["name"] = map[string]string{"from": env.Name, "to": *req.Name}
		env.Name = *req.Name
	}

	// Update only provided fields
	if req.Description != nil {
		changes["description"] = map[string]string{"from": env.Description, "to": *req.Description}
		env.Description = *req.Description
	}

	if req.EnvironmentURL != nil {
		changes["environmentURL"] = map[string]string{"from": env.EnvironmentURL, "to": *req.EnvironmentURL}
		env.EnvironmentURL = *req.EnvironmentURL
	}

	if req.Target != nil {
		changes["target"] = map[string]interface{}{
			"from": fmt.Sprintf("%s:%d", env.Target.Host, env.Target.Port),
			"to": fmt.Sprintf("%s:%d", req.Target.Host, req.Target.Port),
		}
		env.Target = *req.Target
	}

	if req.Credentials != nil {
		// Don't log credential changes for security
		env.Credentials = *req.Credentials
		changes["credentials"] = "updated"
	}

	if req.HealthCheck != nil {
		changes["healthCheck"] = "updated"
		env.HealthCheck = *req.HealthCheck
		
		// If health check was disabled, update status to unknown
		if !env.HealthCheck.Enabled {
			_ = s.repo.UpdateStatus(ctx, env.ID.Hex(), entities.Status{
				Health:    entities.HealthStatusUnknown,
				LastCheck: time.Now(),
				Message:   "Health check disabled",
			})
		}
	}

	if req.Commands != nil {
		changes["commands"] = "updated"
		env.Commands = *req.Commands
	}

	if req.UpgradeConfig != nil {
		changes["upgradeConfig"] = "updated"
		env.UpgradeConfig = *req.UpgradeConfig
	}

	// Handle metadata separately - merge instead of replace
	if req.Metadata != nil {
		if env.Metadata == nil {
			env.Metadata = make(map[string]interface{})
		}
		// Merge metadata including sensitive fields
		// TODO: In production, these should be stored in a secure credential store
		for k, v := range req.Metadata {
			env.Metadata[k] = v
		}
	}

	// Update in repository
	if err := s.repo.Update(ctx, id, env); err != nil {
		return nil, err
	}

	// Log the update if there were changes
	if len(changes) > 0 {
		_ = s.logService.LogEnvironmentAction(ctx, env, entities.ActionTypeUpdate, "Environment updated", map[string]interface{}{
			"changes": changes,
		})
	}

	return env, nil
}

// DeleteEnvironment deletes an environment
func (s *Service) DeleteEnvironment(ctx context.Context, id string) error {
	// Get environment for audit log
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from repository
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Log deletion
	_ = s.logService.LogEnvironmentAction(ctx, env, entities.ActionTypeDelete, "Environment deleted", map[string]interface{}{
		"name": env.Name,
	})

	return nil
}

// CheckHealth performs a health check on an environment
func (s *Service) CheckHealth(ctx context.Context, id string) error {
	// Get environment
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Perform health check
	result, err := s.healthChecker.CheckHealth(ctx, env)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Update status
	oldStatus := env.Status
	newStatus := entities.Status{
		Health:       result.Status,
		LastCheck:    time.Now(),
		Message:      result.Message,
		ResponseTime: result.ResponseTime,
	}

	if err := s.repo.UpdateStatus(ctx, id, newStatus); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Only log health check if status changed
	if oldStatus.Health != newStatus.Health {
		_ = s.logService.LogHealthCheck(ctx, env, newStatus.Health, result.Message, map[string]interface{}{
			"statusCode": result.StatusCode,
			"responseTime": result.ResponseTime,
			"previousStatus": string(oldStatus.Health),
			"currentStatus": string(newStatus.Health),
			"statusChanged": true,
		})
		
		// Update last healthy timestamp if now healthy
		if newStatus.Health == entities.HealthStatusHealthy {
			now := time.Now()
			env.Timestamps.LastHealthyAt = &now
			_ = s.repo.Update(ctx, id, env)
		}
	}

	return nil
}

// RestartEnvironment restarts an environment
func (s *Service) RestartEnvironment(ctx context.Context, id string, force bool) error {
	// Get environment
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if restart is enabled
	if !env.Commands.Restart.Enabled {
		return fmt.Errorf("restart is not enabled for this environment")
	}

	// Log start of operation
	operationID := primitive.NewObjectID()
	
	// Add to logs screen
	_ = s.logService.LogEnvironmentAction(ctx, env, entities.ActionTypeRestart, "Restart operation initiated", map[string]interface{}{
		"operationId": operationID.Hex(),
		"force": force,
		"commandType": env.Commands.Type,
	})
	
	// Also log to audit
	s.logEvent(ctx, env, entities.EventTypeRestart, entities.SeverityInfo, "restart", "Restart initiated", 
		map[string]interface{}{
			"operationId": operationID.Hex(),
			"force": force,
			"commandType": env.Commands.Type,
			"restartConfig": env.Commands.Restart,
		})

	start := time.Now()
	var errorMsg string
	var success bool

	// Log command type being used
	fmt.Printf("Restart command type: %s\n", env.Commands.Type)
	fmt.Printf("Restart config: %+v\n", env.Commands.Restart)

	// Execute restart based on command type
	switch env.Commands.Type {
	case entities.CommandTypeHTTP:
		// Execute HTTP command
		fmt.Println("Executing HTTP restart command")
		// Convert RestartConfig to CommandDetails
		cmdDetails := entities.CommandDetails{
			URL:     env.Commands.Restart.URL,
			Method:  env.Commands.Restart.Method,
			Headers: env.Commands.Restart.Headers,
			Body:    env.Commands.Restart.Body,
		}
		errorMsg, success = s.executeHTTPCommand(ctx, cmdDetails)
		fmt.Printf("HTTP command result - success: %v, error: %s\n", success, errorMsg)
	case entities.CommandTypeSSH:
		// Execute SSH command
		command := env.Commands.Restart.Command
		if command == "" {
			// Default command if not specified
			command = "sudo systemctl restart app"
			if force {
				command = "sudo systemctl restart app --force"
			}
		}
		
		target, err := s.buildSSHTarget(env)
		if err != nil {
			errorMsg = err.Error()
			success = false
		} else {
			result, err := s.sshManager.Execute(ctx, *target, command)
			if err != nil || (result != nil && result.ExitCode != 0) {
				success = false
				if err != nil {
					errorMsg = err.Error()
				} else if result != nil {
					errorMsg = result.Output
				}
			} else {
				success = true
			}
		}
	default:
		// Default to SSH with standard command
		target, err := s.buildSSHTarget(env)
		if err != nil {
			errorMsg = err.Error()
			success = false
		} else {
			command := "sudo systemctl restart app"
			if force {
				command = "sudo systemctl restart app --force"
			}
			result, err := s.sshManager.Execute(ctx, *target, command)
			if err != nil || (result != nil && result.ExitCode != 0) {
				success = false
				if err != nil {
					errorMsg = err.Error()
				} else if result != nil {
					errorMsg = result.Output
				}
			} else {
				success = true
			}
		}
	}

	duration := time.Since(start).Milliseconds()

	if !success {
		// Add to logs screen
		_ = s.logService.LogEnvironmentAction(ctx, env, entities.ActionTypeRestart, fmt.Sprintf("Restart operation failed: %s", errorMsg), map[string]interface{}{
			"operationId": operationID.Hex(),
			"duration": duration,
			"error": errorMsg,
		})
		
		// Also log to audit
		s.logEvent(ctx, env, entities.EventTypeRestart, entities.SeverityError, "restart", 
			fmt.Sprintf("Restart failed: %s", errorMsg), 
			map[string]interface{}{
				"operationId": operationID.Hex(),
				"duration": duration,
			})
		
		return fmt.Errorf("restart failed: %s", errorMsg)
	}

	// Update timestamps
	now := time.Now()
	env.Timestamps.LastRestartAt = &now
	_ = s.repo.Update(ctx, id, env)

	// Log success
	// Add to logs screen
	_ = s.logService.LogEnvironmentAction(ctx, env, entities.ActionTypeRestart, "Restart operation completed successfully", map[string]interface{}{
		"operationId": operationID.Hex(),
		"duration": duration,
	})
	
	// Also log to audit
	s.logEvent(ctx, env, entities.EventTypeRestart, entities.SeverityInfo, "restart", "Restart completed successfully", 
		map[string]interface{}{
			"operationId": operationID.Hex(),
			"duration": duration,
		})

	// Trigger health check after restart
	go func() {
		time.Sleep(10 * time.Second) // Wait for service to start
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = s.CheckHealth(ctx, id)
	}()

	return nil
}


// UpgradeEnvironment upgrades an environment to a new version
func (s *Service) UpgradeEnvironment(ctx context.Context, id string, version string) error {
	// Get environment
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if upgrade is enabled
	if !env.UpgradeConfig.Enabled {
		return fmt.Errorf("upgrade is not enabled for this environment")
	}

	// Log start of operation
	operationID := primitive.NewObjectID()
	
	// Add to logs screen
	_ = s.logService.LogEnvironmentAction(ctx, env, entities.ActionTypeUpgrade, "Upgrade operation initiated", map[string]interface{}{
		"operationId": operationID.Hex(),
		"targetVersion": version,
		"currentVersion": env.SystemInfo.AppVersion,
	})
	
	// Also log to audit
	s.logEvent(ctx, env, entities.EventTypeUpgrade, entities.SeverityInfo, "upgrade", "Upgrade initiated", 
		map[string]interface{}{
			"operationId": operationID.Hex(),
			"targetVersion": version,
			"currentVersion": env.SystemInfo.AppVersion,
		})

	start := time.Now()
	var errorMsg string
	var success bool

	// Execute upgrade command
	upgradeCmd := env.UpgradeConfig.UpgradeCommand
	// Replace version placeholder in command
	if upgradeCmd.Command != "" {
		upgradeCmd.Command = strings.ReplaceAll(upgradeCmd.Command, "{VERSION}", version)
	}
	if upgradeCmd.URL != "" {
		upgradeCmd.URL = strings.ReplaceAll(upgradeCmd.URL, "{VERSION}", version)
	}
	// Replace version placeholder in body
	if upgradeCmd.Body != nil && len(upgradeCmd.Body) > 0 {
		// Create a new map to avoid modifying the original
		newBody := make(map[string]interface{})
		for k, v := range upgradeCmd.Body {
			// Replace {VERSION} in string values
			if strVal, ok := v.(string); ok {
				newBody[k] = strings.ReplaceAll(strVal, "{VERSION}", version)
			} else {
				// For non-string values, keep as is
				newBody[k] = v
			}
		}
		upgradeCmd.Body = newBody
	}

	switch env.UpgradeConfig.Type {
	case entities.CommandTypeHTTP:
		// Execute HTTP command
		errorMsg, success = s.executeHTTPCommand(ctx, upgradeCmd)
	case entities.CommandTypeSSH:
		// Execute SSH command - support multi-line commands
		if upgradeCmd.Command == "" {
			upgradeCmd.Command = fmt.Sprintf("sudo app-upgrade --version=%s", version)
		}
		
		target, err := s.buildSSHTarget(env)
		if err != nil {
			errorMsg = err.Error()
			success = false
		} else {
			// Split multi-line commands and execute them sequentially
			commands := strings.Split(upgradeCmd.Command, "\n")
			for _, cmd := range commands {
				cmd = strings.TrimSpace(cmd)
				if cmd == "" {
					continue
				}
				result, err := s.sshManager.Execute(ctx, *target, cmd)
				if err != nil || (result != nil && result.ExitCode != 0) {
					success = false
					if err != nil {
						errorMsg = fmt.Sprintf("Command failed: %s - Error: %v", cmd, err)
					} else if result != nil {
						errorMsg = fmt.Sprintf("Command failed: %s - Output: %s", cmd, result.Output)
					}
					break
				}
			}
			if errorMsg == "" {
				success = true
			}
		}
	default:
		errorMsg = "No command type specified for upgrade"
		success = false
	}

	duration := time.Since(start).Milliseconds()

	if !success {
		// Add to logs screen
		_ = s.logService.LogEnvironmentAction(ctx, env, entities.ActionTypeUpgrade, fmt.Sprintf("Upgrade operation failed: %s", errorMsg), map[string]interface{}{
			"operationId": operationID.Hex(),
			"duration": duration,
			"targetVersion": version,
			"error": errorMsg,
		})
		
		// Also log to audit
		s.logEvent(ctx, env, entities.EventTypeUpgrade, entities.SeverityError, "upgrade", 
			fmt.Sprintf("Upgrade failed: %s", errorMsg), 
			map[string]interface{}{
				"operationId": operationID.Hex(),
				"duration": duration,
				"targetVersion": version,
			})
		
		return fmt.Errorf("upgrade failed: %s", errorMsg)
	}

	// Update timestamps and version
	now := time.Now()
	env.Timestamps.LastUpgradeAt = &now
	env.SystemInfo.AppVersion = version
	env.SystemInfo.LastUpdated = now
	_ = s.repo.Update(ctx, id, env)

	// Log success
	// Add to logs screen
	_ = s.logService.LogEnvironmentAction(ctx, env, entities.ActionTypeUpgrade, "Upgrade operation completed successfully", map[string]interface{}{
		"operationId": operationID.Hex(),
		"duration": duration,
		"newVersion": version,
		"previousVersion": env.SystemInfo.AppVersion,
	})
	
	// Also log to audit
	s.logEvent(ctx, env, entities.EventTypeUpgrade, entities.SeverityInfo, "upgrade", "Upgrade completed successfully", 
		map[string]interface{}{
			"operationId": operationID.Hex(),
			"duration": duration,
			"newVersion": version,
		})

	// Trigger health check after upgrade
	go func() {
		time.Sleep(30 * time.Second) // Wait longer for upgrade to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = s.CheckHealth(ctx, id)
	}()

	return nil
}

// GetAvailableVersions fetches available versions for upgrade
func (s *Service) GetAvailableVersions(ctx context.Context, id string) ([]string, string, error) {
	// Get environment
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	// Check if upgrade is enabled
	if !env.UpgradeConfig.Enabled {
		return nil, "", fmt.Errorf("upgrade is not enabled for this environment")
	}

	// Check if version list URL is configured
	if env.UpgradeConfig.VersionListURL == "" {
		return nil, "", fmt.Errorf("version list URL is not configured")
	}

	// Fetch versions from endpoint
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Determine method (default to GET if not specified)
	method := env.UpgradeConfig.VersionListMethod
	if method == "" {
		method = "GET"
	}

	// Prepare body if provided
	var body io.Reader
	if env.UpgradeConfig.VersionListBody != "" {
		body = strings.NewReader(env.UpgradeConfig.VersionListBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, env.UpgradeConfig.VersionListURL, body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	if env.UpgradeConfig.VersionListHeaders != nil {
		for key, value := range env.UpgradeConfig.VersionListHeaders {
			req.Header.Set(key, value)
		}
	}
	
	// Set Content-Type if body is provided
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("version endpoint returned status %d", resp.StatusCode)
	}

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response based on JSONPath if provided
	var availableVersions []string
	currentVersion := env.SystemInfo.AppVersion

	if env.UpgradeConfig.JSONPathResponse != "" {
		// Parse as JSON
		var responseData interface{}
		if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
			return nil, "", fmt.Errorf("failed to parse JSON response: %w", err)
		}

		// Use JSONPath to extract versions
		versions, err := extractJSONPath(responseData, env.UpgradeConfig.JSONPathResponse)
		if err != nil {
			return nil, "", fmt.Errorf("failed to extract versions using JSONPath: %w", err)
		}

		// Convert to string array
		switch v := versions.(type) {
		case []interface{}:
			for _, version := range v {
				switch val := version.(type) {
				case string:
					availableVersions = append(availableVersions, val)
				case float64:
					// Handle numeric values (JSON numbers are float64)
					availableVersions = append(availableVersions, fmt.Sprintf("%.0f", val))
				case int:
					availableVersions = append(availableVersions, fmt.Sprintf("%d", val))
				default:
					// Try to convert any other type to string
					availableVersions = append(availableVersions, fmt.Sprintf("%v", val))
				}
			}
		case []string:
			availableVersions = v
		default:
			return nil, "", fmt.Errorf("JSONPath did not return an array of versions")
		}
	} else {
		// Try to parse as a simple JSON array
		if err := json.Unmarshal(bodyBytes, &availableVersions); err != nil {
			return nil, "", fmt.Errorf("failed to parse response as version array: %w", err)
		}
	}

	return availableVersions, currentVersion, nil
}

// extractJSONPath extracts value from JSON using JSONPath syntax
func extractJSONPath(data interface{}, jsonPath string) (interface{}, error) {
	// Enhanced JSONPath implementation that handles array element property access
	// Remove leading $. if present
	path := strings.TrimPrefix(jsonPath, "$.")
	
	// Check if we have array element property access like products[*]["id"] or products[*].id
	if strings.Contains(path, "[*]") {
		// Split into array path and property path
		parts := strings.SplitN(path, "[*]", 2)
		arrayPath := parts[0]
		propertyPath := ""
		
		if len(parts) > 1 {
			propertyPath = parts[1]
			// Clean up property path - remove leading dot or [""] notation
			propertyPath = strings.TrimPrefix(propertyPath, ".")
			propertyPath = strings.TrimPrefix(propertyPath, "[\"")
			propertyPath = strings.TrimSuffix(propertyPath, "\"]")
			propertyPath = strings.Trim(propertyPath, "\"")
		}
		
		// Navigate to the array
		current := data
		if arrayPath != "" {
			arrayParts := strings.Split(arrayPath, ".")
			for _, part := range arrayParts {
				if part == "" {
					continue
				}
				
				switch v := current.(type) {
				case map[string]interface{}:
					val, ok := v[part]
					if !ok {
						return nil, fmt.Errorf("path '%s' not found", part)
					}
					current = val
				default:
					return nil, fmt.Errorf("cannot navigate path '%s' in non-object", part)
				}
			}
		}
		
		// Now current should be an array
		switch arr := current.(type) {
		case []interface{}:
			// If no property path, return the array as is
			if propertyPath == "" {
				return arr, nil
			}
			
			// Extract property from each array element
			result := make([]interface{}, 0)
			for _, elem := range arr {
				if objElem, ok := elem.(map[string]interface{}); ok {
					if val, exists := objElem[propertyPath]; exists {
						result = append(result, val)
					}
				}
			}
			return result, nil
			
		default:
			return nil, fmt.Errorf("path '%s' is not an array", arrayPath)
		}
	}
	
	// Handle simple paths without array notation
	parts := strings.Split(path, ".")
	current := data
	
	for _, part := range parts {
		if part == "" {
			continue
		}
		
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("path '%s' not found", part)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot navigate path '%s' in non-object", part)
		}
	}
	
	return current, nil
}

// getJSONPath extracts value from JSON using dot notation path
func getJSONPath(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := interface{}(data)
	
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}
	
	return current, true
}

// validateURL validates a URL to prevent SSRF attacks
func validateURL(rawURL string) error {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme - only allow HTTP and HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS schemes are allowed")
	}

	// Extract hostname
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	// Check for local/private addresses
	// Resolve the hostname to IP addresses
	ips, err := net.LookupIP(hostname)
	if err != nil {
		// If we can't resolve, it might be a non-existent host
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	for _, ip := range ips {
		// Check for localhost
		if ip.IsLoopback() {
			return fmt.Errorf("localhost addresses are not allowed")
		}

		// Check for private networks
		if ip.IsPrivate() {
			return fmt.Errorf("private network addresses are not allowed")
		}

		// Check for link-local addresses
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("link-local addresses are not allowed")
		}

		// Check for multicast
		if ip.IsMulticast() {
			return fmt.Errorf("multicast addresses are not allowed")
		}

		// Check for unspecified addresses (0.0.0.0 or ::)
		if ip.IsUnspecified() {
			return fmt.Errorf("unspecified addresses are not allowed")
		}
	}

	// Additional checks for common internal hostnames
	lowerHostname := strings.ToLower(hostname)
	blockedHostnames := []string{
		"localhost", "127.0.0.1", "0.0.0.0", "::1",
		"metadata", "metadata.google.internal", // GCP metadata
		"169.254.169.254", // AWS/Azure metadata
	}

	for _, blocked := range blockedHostnames {
		if lowerHostname == blocked {
			return fmt.Errorf("hostname '%s' is not allowed", hostname)
		}
	}

	return nil
}

// executeHTTPCommand executes an HTTP command
func (s *Service) executeHTTPCommand(ctx context.Context, cmd entities.CommandDetails) (string, bool) {
	if cmd.URL == "" {
		return "HTTP command URL is required", false
	}

	// Validate URL to prevent SSRF
	if err := validateURL(cmd.URL); err != nil {
		return fmt.Sprintf("URL validation failed: %v", err), false
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
		// Disable following redirects to prevent redirect-based SSRF
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Validate each redirect URL
			if err := validateURL(req.URL.String()); err != nil {
				return fmt.Errorf("redirect URL validation failed: %w", err)
			}
			// Limit redirect chain
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Build request
	method := cmd.Method
	if method == "" {
		method = "POST"
	}

	var body io.Reader
	if cmd.Body != nil && len(cmd.Body) > 0 {
		// Marshal the body map to JSON
		jsonBody, err := json.Marshal(cmd.Body)
		if err != nil {
			return fmt.Sprintf("Failed to marshal request body: %v", err), false
		}
		body = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, cmd.URL, body)
	if err != nil {
		return fmt.Sprintf("Failed to create request: %v", err), false
	}

	// Add headers
	for key, value := range cmd.Headers {
		req.Header.Set(key, value)
	}
	
	// Set default Content-Type if not provided and body exists
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Request failed: %v", err), false
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Failed to read response: %v", err), false
	}

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return string(responseBody), true
	}

	return fmt.Sprintf("Request failed with status %d: %s", resp.StatusCode, string(responseBody)), false
}

// buildSSHTarget builds an SSH target from environment
func (s *Service) buildSSHTarget(env *entities.Environment) (*ssh.Target, error) {
	target := &ssh.Target{
		Host:     env.Target.Host,
		Port:     env.Target.Port,
		Username: env.Credentials.Username,
	}

	// Load credentials based on type
	switch env.Credentials.Type {
	case "password":
		// Look for password in metadata (temporary solution)
		if env.Metadata != nil {
			if password, ok := env.Metadata["password"].(string); ok && password != "" {
				target.Password = password
			} else {
				return nil, fmt.Errorf("SSH password not found in environment configuration")
			}
		} else {
			return nil, fmt.Errorf("no SSH password configured for environment")
		}
		
	case "key":
		// Look for private key in metadata (temporary solution)
		if env.Metadata != nil {
			if privateKey, ok := env.Metadata["privateKey"].(string); ok && privateKey != "" {
				target.PrivateKey = []byte(privateKey)
			} else {
				return nil, fmt.Errorf("SSH private key not found in environment configuration")
			}
		} else {
			return nil, fmt.Errorf("no SSH private key configured for environment")
		}
		
	default:
		return nil, fmt.Errorf("unsupported credential type: %s", env.Credentials.Type)
	}

	return target, nil
}

// logEvent creates an audit log entry
func (s *Service) logEvent(ctx context.Context, env *entities.Environment, eventType entities.EventType, 
	severity entities.Severity, operation, message string, metadata map[string]interface{}) {
	
	s.logEventWithPayload(ctx, env, eventType, severity, operation, message, map[string]interface{}{
		"metadata": metadata,
	})
}

// logEventWithPayload creates an audit log entry with payload
func (s *Service) logEventWithPayload(ctx context.Context, env *entities.Environment, eventType entities.EventType,
	severity entities.Severity, operation, message string, payload map[string]interface{}) {
	
	log := &entities.AuditLog{
		Timestamp:      time.Now(),
		EnvironmentID:  env.ID,
		EnvironmentName: env.Name,
		Type:           eventType,
		Severity:       severity,
		Actor: entities.Actor{
			Type: "system", // TODO: Get from context
			ID:   "system",
			Name: "System",
		},
		Action: entities.Action{
			Operation: operation,
			Status:    "completed",
		},
		Payload: entities.Payload{
			Metadata: payload,
		},
	}

	// Create audit log asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.auditRepo.Create(ctx, log)
	}()
}

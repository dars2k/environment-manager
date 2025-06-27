package dto_test

import (
	"testing"
	"time"

	"app-env-manager/internal/api/dto"
	"app-env-manager/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSuccessResponse(t *testing.T) {
	data := map[string]string{"key": "value"}
	response := dto.SuccessResponse{
		Success: true,
		Data:    data,
		Metadata: dto.ResponseMetadata{
			Timestamp: "2024-01-01T00:00:00Z",
			Version:   "1.0.0",
		},
	}

	assert.True(t, response.Success)
	assert.Equal(t, data, response.Data)
	assert.Equal(t, "2024-01-01T00:00:00Z", response.Metadata.Timestamp)
	assert.Equal(t, "1.0.0", response.Metadata.Version)
}

func TestErrorResponse(t *testing.T) {
	response := dto.ErrorResponse{
		Success: false,
		Error: dto.ErrorInfo{
			Code:    "TEST_ERROR",
			Message: "Test error message",
			Details: map[string]interface{}{
				"field": "test",
			},
		},
		Metadata: dto.ResponseMetadata{
			Timestamp: "2024-01-01T00:00:00Z",
			Version:   "1.0.0",
		},
	}

	assert.False(t, response.Success)
	assert.Equal(t, "TEST_ERROR", response.Error.Code)
	assert.Equal(t, "Test error message", response.Error.Message)
	assert.Equal(t, "test", response.Error.Details["field"])
}

func TestEnvironmentResponse(t *testing.T) {
	env := &entities.Environment{
		ID:             primitive.NewObjectID(),
		Name:           "Test Environment",
		Description:    "Test Description",
		EnvironmentURL: "https://test.example.com",
		Status: entities.Status{
			Health:       entities.HealthStatusHealthy,
			Message:      "All systems operational",
			ResponseTime: 150,
		},
	}

	response := dto.EnvironmentResponse{
		Environment: env,
	}

	assert.Equal(t, env, response.Environment)
}

func TestListEnvironmentsResponse(t *testing.T) {
	envs := []*entities.Environment{
		{
			ID:   primitive.NewObjectID(),
			Name: "Env 1",
		},
		{
			ID:   primitive.NewObjectID(),
			Name: "Env 2",
		},
	}

	response := dto.ListEnvironmentsResponse{
		Environments: envs,
		Pagination: dto.PaginationResponse{
			Page:  1,
			Limit: 10,
			Total: 2,
		},
	}

	assert.Len(t, response.Environments, 2)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 10, response.Pagination.Limit)
	assert.Equal(t, 2, response.Pagination.Total)
}

func TestMessageResponse(t *testing.T) {
	response := dto.MessageResponse{
		Message: "Operation successful",
	}

	assert.Equal(t, "Operation successful", response.Message)
}

func TestOperationResponse(t *testing.T) {
	response := dto.OperationResponse{
		OperationID: "op-12345",
		Status:      "in_progress",
	}

	assert.Equal(t, "op-12345", response.OperationID)
	assert.Equal(t, "in_progress", response.Status)
}

func TestVersionsResponse(t *testing.T) {
	response := dto.VersionsResponse{
		CurrentVersion:    "1.0.0",
		AvailableVersions: []string{"1.0.0", "1.1.0", "1.2.0"},
	}

	assert.Equal(t, "1.0.0", response.CurrentVersion)
	assert.Len(t, response.AvailableVersions, 3)
}

func TestRestartRequest(t *testing.T) {
	req := dto.RestartRequest{
		Force: true,
	}

	assert.True(t, req.Force)
}

func TestUpgradeRequest(t *testing.T) {
	req := dto.UpgradeRequest{
		Version: "1.2.0",
	}

	assert.Equal(t, "1.2.0", req.Version)
}

func TestUserDTOs(t *testing.T) {
	// Test UserResponse
	now := time.Now()
	userResp := dto.UserResponse{
		ID:       primitive.NewObjectID().Hex(),
		Username: "testuser",
		Role:     entities.UserRoleAdmin,
		Active:   true,
		CreatedAt: now,
		UpdatedAt: now,
		LastLoginAt: &now,
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	assert.NotEmpty(t, userResp.ID)
	assert.Equal(t, "testuser", userResp.Username)
	assert.Equal(t, entities.UserRoleAdmin, userResp.Role)
	assert.True(t, userResp.Active)
	assert.NotNil(t, userResp.LastLoginAt)
	assert.Equal(t, "value", userResp.Metadata["key"])

	// Test ToUserResponse
	user := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
		Role:     entities.UserRoleAdmin,
		Active:   true,
		CreatedAt: now,
		UpdatedAt: now,
		LastLoginAt: &now,
	}
	
	converted := dto.ToUserResponse(user)
	assert.NotNil(t, converted)
	assert.Equal(t, user.ID.Hex(), converted.ID)
	assert.Equal(t, user.Username, converted.Username)
	
	// Test ToUserResponse with nil
	nilConverted := dto.ToUserResponse(nil)
	assert.Nil(t, nilConverted)
	
	// Test ToUserResponses
	users := []*entities.User{user}
	convertedList := dto.ToUserResponses(users)
	assert.Len(t, convertedList, 1)
	assert.Equal(t, user.ID.Hex(), convertedList[0].ID)
	
	// Test ListUsersResponse
	listResp := dto.ListUsersResponse{
		Users: []*dto.UserResponse{&userResp},
	}
	assert.Len(t, listResp.Users, 1)
	
	// Test SingleUserResponse
	singleResp := dto.SingleUserResponse{
		User: &userResp,
	}
	assert.NotNil(t, singleResp.User)
}

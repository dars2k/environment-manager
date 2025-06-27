package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"app-env-manager/internal/api/handlers"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


// UserHandler Test Suite
type UserHandlerTestSuite struct {
	suite.Suite
	logger *logrus.Logger
}

func (suite *UserHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	
	suite.logger = logrus.New()
	suite.logger.SetOutput(io.Discard) // Disable logging during tests
}

func (suite *UserHandlerTestSuite) TestList_Success() {
	// Arrange
	users := []*entities.User{
		{
			ID:       primitive.NewObjectID(),
			Username: "user1",
			Role:     entities.UserRoleUser,
			Active:   true,
		},
		{
			ID:       primitive.NewObjectID(),
			Username: "user2",
			Role:     entities.UserRoleAdmin,
			Active:   true,
		},
	}
	
	filter := interfaces.ListFilter{
		Page:  1,
		Limit: 100,
	}
	
	// Create mock repositories and services
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("List", mock.Anything, filter).Return(users, nil)

	// Act
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/users", handler.List)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	usersResponse := data["users"].([]interface{})
	assert.Len(suite.T(), usersResponse, 2)

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestList_WithPagination() {
	// Arrange
	filter := interfaces.ListFilter{
		Page:  2,
		Limit: 50,
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("List", mock.Anything, filter).Return([]*entities.User{}, nil)

	// Act
	req := httptest.NewRequest("GET", "/api/users?page=2&limit=50", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/users", handler.List)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestList_InvalidPage() {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	// Should use default page 1
	mockUserRepo.On("List", mock.Anything, mock.MatchedBy(func(f interfaces.ListFilter) bool {
		return f.Page == 1
	})).Return([]*entities.User{}, nil)

	// Act
	req := httptest.NewRequest("GET", "/api/users?page=-1", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/users", handler.List)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestList_LimitExceedsMax() {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	// Should use default limit 100
	mockUserRepo.On("List", mock.Anything, mock.MatchedBy(func(f interfaces.ListFilter) bool {
		return f.Limit == 100
	})).Return([]*entities.User{}, nil)

	// Act
	req := httptest.NewRequest("GET", "/api/users?limit=200", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/users", handler.List)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestList_Error() {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("List", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("database error"))

	// Act
	req := httptest.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/users", handler.List)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Failed to retrieve users", response["error"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestGet_Success() {
	// Arrange
	userID := primitive.NewObjectID()
	testUser := &entities.User{
		ID:       userID,
		Username: "testuser",
		Role:     entities.UserRoleUser,
		Active:   true,
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(testUser, nil)

	// Act
	req := httptest.NewRequest("GET", "/api/users/"+userID.Hex(), nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/users/:id", handler.Get)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	userResponse := data["user"].(map[string]interface{})
	assert.Equal(suite.T(), testUser.Username, userResponse["username"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestGet_NotFound() {
	// Arrange
	userID := primitive.NewObjectID()
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(nil, interfaces.ErrNotFound)

	// Act
	req := httptest.NewRequest("GET", "/api/users/"+userID.Hex(), nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/users/:id", handler.Get)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User not found", response["error"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestGet_Error() {
	// Arrange
	userID := primitive.NewObjectID()
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(nil, fmt.Errorf("database error"))

	// Act
	req := httptest.NewRequest("GET", "/api/users/"+userID.Hex(), nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/users/:id", handler.Get)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Failed to retrieve user", response["error"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestCreate_Success() {
	// Arrange
	createReq := entities.CreateUserRequest{
		Username: "newuser",
		Password: "password123",
	}
	
	currentUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	// Mock that username doesn't exist
	mockUserRepo.On("GetByUsername", mock.Anything, createReq.Username).Return(nil, interfaces.ErrUserNotFound)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)
	mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Setup router with context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", currentUser)
		c.Next()
	})
	router.POST("/api/users", handler.Create)

	// Act
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	userResponse := data["user"].(map[string]interface{})
	assert.Equal(suite.T(), createReq.Username, userResponse["username"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestCreate_InvalidJSON() {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)

	// Act
	req := httptest.NewRequest("POST", "/api/users", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.POST("/api/users", handler.Create)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid request format", response["error"])
}

func (suite *UserHandlerTestSuite) TestCreate_Error() {
	// Arrange
	createReq := entities.CreateUserRequest{
		Username: "newuser",
		Password: "password123",
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	// Mock that username already exists
	existingUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: createReq.Username,
	}
	mockUserRepo.On("GetByUsername", mock.Anything, createReq.Username).Return(existingUser, nil)

	// Act
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.POST("/api/users", handler.Create)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "username already exists", response["error"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestUpdate_Success() {
	// Arrange
	userID := primitive.NewObjectID()
	role := entities.UserRoleUser
	updateReq := entities.UpdateUserRequest{
		Role: &role,
	}
	
	currentUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}
	
	existingUser := &entities.User{
		ID:       userID,
		Username: "testuser",
		Role:     entities.UserRoleAdmin,
		Active:   true,
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(existingUser, nil)
	mockUserRepo.On("Update", mock.Anything, userID.Hex(), mock.AnythingOfType("*entities.User")).Return(nil)
	mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Setup router with context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", currentUser)
		c.Next()
	})
	router.PUT("/api/users/:id", handler.Update)

	// Act
	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/api/users/"+userID.Hex(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	userResponse := data["user"].(map[string]interface{})
	assert.Equal(suite.T(), string(role), userResponse["role"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestUpdate_NotFound() {
	// Arrange
	userID := primitive.NewObjectID()
	role := entities.UserRoleAdmin
	updateReq := entities.UpdateUserRequest{
		Role: &role,
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(nil, interfaces.ErrNotFound)

	// Act
	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/api/users/"+userID.Hex(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.PUT("/api/users/:id", handler.Update)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User not found", response["error"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestUpdate_InvalidJSON() {
	// Arrange
	userID := primitive.NewObjectID()
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)

	// Act
	req := httptest.NewRequest("PUT", "/api/users/"+userID.Hex(), bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.PUT("/api/users/:id", handler.Update)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid request format", response["error"])
}

func (suite *UserHandlerTestSuite) TestDelete_Success() {
	// Arrange
	userID := primitive.NewObjectID()
	currentUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}
	
	userToDelete := &entities.User{
		ID:       userID,
		Username: "testuser",
		Role:     entities.UserRoleUser,
		Active:   true,
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(userToDelete, nil)
	mockUserRepo.On("Delete", mock.Anything, userID.Hex()).Return(nil)
	mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Setup router with context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", currentUser)
		c.Next()
	})
	router.DELETE("/api/users/:id", handler.Delete)

	// Act
	req := httptest.NewRequest("DELETE", "/api/users/"+userID.Hex(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "User deleted successfully", data["message"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestDelete_NotFound() {
	// Arrange
	userID := primitive.NewObjectID()
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(nil, interfaces.ErrNotFound)

	// Act
	req := httptest.NewRequest("DELETE", "/api/users/"+userID.Hex(), nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.DELETE("/api/users/:id", handler.Delete)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User not found", response["error"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestDelete_Error() {
	// Arrange
	userID := primitive.NewObjectID()
	currentUser := &entities.User{
		ID:   userID, // Same ID - trying to delete self
		Role: entities.UserRoleAdmin,
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	// No mock expectations - the service will check permissions before calling GetByID

	// Setup router with context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", currentUser)
		c.Next()
	})
	router.DELETE("/api/users/:id", handler.Delete)

	// Act
	req := httptest.NewRequest("DELETE", "/api/users/"+userID.Hex(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "cannot delete your own account", response["error"])
}

func (suite *UserHandlerTestSuite) TestChangePassword_Success() {
	// Arrange
	userID := primitive.NewObjectID()
	changePasswordReq := entities.ChangePasswordRequest{
		CurrentPassword: "oldpass",
		NewPassword:     "newpass",
	}
	
	currentUser := &entities.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: "", // Will be set by CheckPassword
		Role:         entities.UserRoleUser,
		Active:       true,
	}
	// Set password so CheckPassword works
	passwdErr := currentUser.UpdatePassword("oldpass")
	assert.NoError(suite.T(), passwdErr)
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(currentUser, nil)
	mockUserRepo.On("Update", mock.Anything, userID.Hex(), mock.AnythingOfType("*entities.User")).Return(nil)
	mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Setup router with context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID.Hex())
		c.Next()
	})
	router.PUT("/api/users/change-password", handler.ChangePassword)

	// Act
	body, _ := json.Marshal(changePasswordReq)
	req := httptest.NewRequest("PUT", "/api/users/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Password changed successfully", data["message"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestChangePassword_InvalidJSON() {
	// Arrange
	userID := primitive.NewObjectID()
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)

	// Setup router with context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID.Hex())
		c.Next()
	})
	router.PUT("/api/users/change-password", handler.ChangePassword)

	// Act
	req := httptest.NewRequest("PUT", "/api/users/change-password", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid request format", response["error"])
}

func (suite *UserHandlerTestSuite) TestChangePassword_Error() {
	// Arrange
	userID := primitive.NewObjectID()
	changePasswordReq := entities.ChangePasswordRequest{
		CurrentPassword: "wrongpass",
		NewPassword:     "newpass",
	}
	
	currentUser := &entities.User{
		ID:           userID,
		Username:     "testuser",
		PasswordHash: "", // Will be set by SetPassword
		Role:         entities.UserRoleUser,
		Active:       true,
	}
	// Set password different from request
	passwdErr := currentUser.UpdatePassword("correctpass")
	assert.NoError(suite.T(), passwdErr)
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(currentUser, nil)

	// Setup router with context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID.Hex())
		c.Next()
	})
	router.PUT("/api/users/change-password", handler.ChangePassword)

	// Act
	body, _ := json.Marshal(changePasswordReq)
	req := httptest.NewRequest("PUT", "/api/users/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "current password is incorrect", response["error"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestResetPassword_Success() {
	// Arrange
	userID := primitive.NewObjectID()
	resetPasswordReq := entities.ResetPasswordRequest{
		NewPassword: "newpassword123",
	}
	
	currentUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}
	
	targetUser := &entities.User{
		ID:       userID,
		Username: "testuser",
		Role:     entities.UserRoleUser,
		Active:   true,
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(targetUser, nil)
	mockUserRepo.On("Update", mock.Anything, userID.Hex(), mock.AnythingOfType("*entities.User")).Return(nil)
	mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Setup router with context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", currentUser)
		c.Next()
	})
	router.PUT("/api/users/:id/reset-password", handler.ResetPassword)

	// Act
	body, _ := json.Marshal(resetPasswordReq)
	req := httptest.NewRequest("PUT", "/api/users/"+userID.Hex()+"/reset-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Password reset successfully", data["message"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestResetPassword_NotFound() {
	// Arrange
	userID := primitive.NewObjectID()
	resetPasswordReq := entities.ResetPasswordRequest{
		NewPassword: "newpassword123",
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(nil, interfaces.ErrNotFound)

	// Act
	body, _ := json.Marshal(resetPasswordReq)
	req := httptest.NewRequest("PUT", "/api/users/"+userID.Hex()+"/reset-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.PUT("/api/users/:id/reset-password", handler.ResetPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User not found", response["error"])

	mockUserRepo.AssertExpectations(suite.T())
}

func (suite *UserHandlerTestSuite) TestResetPassword_InvalidJSON() {
	// Arrange
	userID := primitive.NewObjectID()
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)

	// Act
	req := httptest.NewRequest("PUT", "/api/users/"+userID.Hex()+"/reset-password", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.PUT("/api/users/:id/reset-password", handler.ResetPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid request format", response["error"])
}

func (suite *UserHandlerTestSuite) TestResetPassword_Error() {
	// Arrange
	userID := primitive.NewObjectID()
	resetPasswordReq := entities.ResetPasswordRequest{
		NewPassword: "newpassword123",
	}
	
	// Non-admin trying to reset password
	currentUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleUser, // Not admin
	}
	
	mockUserRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	userService := user.NewService(mockUserRepo, logService)
	handler := handlers.NewUserHandler(userService, suite.logger)
	
	// No mock expectations - the service will check permissions before calling GetByID

	// Setup router with context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", currentUser)
		c.Next()
	})
	router.PUT("/api/users/:id/reset-password", handler.ResetPassword)

	// Act
	body, _ := json.Marshal(resetPasswordReq)
	req := httptest.NewRequest("PUT", "/api/users/"+userID.Hex()+"/reset-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "unauthorized: insufficient permissions", response["error"])
}

// Run the test suite
func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/api/handlers"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/auth"
	"app-env-manager/internal/service/log"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


// AuthHandler Test Suite
type AuthHandlerTestSuite struct {
	suite.Suite
	handler         *handlers.AuthHandler
	mockUserRepo    *MockUserRepository
	mockLogRepo     *MockLogRepository
	authService     *auth.Service
	logService      *log.Service
	logger          *logrus.Logger
	router          *gin.Engine
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	
	suite.mockUserRepo = new(MockUserRepository)
	suite.mockLogRepo = new(MockLogRepository)
	suite.logger = logrus.New()
	suite.logger.SetOutput(io.Discard) // Disable logging during tests

	// Create log service with mock repository
	suite.logService = log.NewService(suite.mockLogRepo)

	// Create real auth service with mocks including log service
	suite.authService = auth.NewService(
		suite.mockUserRepo,
		suite.logService,
		"test-secret",
		24*time.Hour,
	)

	suite.handler = handlers.NewAuthHandler(suite.authService, suite.logger)

	suite.router = gin.New()
	suite.router.POST("/api/auth/login", suite.handler.Login)
	suite.router.POST("/api/auth/logout", suite.handler.Logout)
	suite.router.GET("/api/auth/me", suite.handler.GetCurrentUser)
}

func (suite *AuthHandlerTestSuite) TestLogin_Success() {
	// Arrange
	loginReq := entities.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	user, _ := entities.NewUser("testuser", "password123")
	user.ID = primitive.NewObjectID()
	user.Active = true

	suite.mockUserRepo.On("GetByUsername", mock.Anything, loginReq.Username).Return(user, nil)
	suite.mockUserRepo.On("UpdateLastLogin", mock.Anything, user.ID).Return(nil)
	suite.mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(suite.T(), data["token"])
	assert.NotNil(suite.T(), data["user"])

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestLogin_InvalidCredentials() {
	// Arrange
	loginReq := entities.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	user, _ := entities.NewUser("testuser", "correctpassword")
	user.ID = primitive.NewObjectID()
	user.Active = true

	suite.mockUserRepo.On("GetByUsername", mock.Anything, loginReq.Username).Return(user, nil)
	suite.mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid credentials", response["error"])

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestLogin_UserNotFound() {
	// Arrange
	loginReq := entities.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	suite.mockUserRepo.On("GetByUsername", mock.Anything, loginReq.Username).Return(nil, interfaces.ErrUserNotFound)
	suite.mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "invalid credentials", response["error"])

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestLogin_InactiveUser() {
	// Arrange
	loginReq := entities.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	user, _ := entities.NewUser("testuser", "password123")
	user.ID = primitive.NewObjectID()
	user.Active = false // Inactive user

	suite.mockUserRepo.On("GetByUsername", mock.Anything, loginReq.Username).Return(user, nil)
	suite.mockLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "account is inactive", response["error"])

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestLogin_InvalidJSON() {
	// Act
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response["error"], "Invalid request format")
}

func (suite *AuthHandlerTestSuite) TestLogout_Success() {
	// Act
	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), "Logged out successfully", data["message"])
}

func (suite *AuthHandlerTestSuite) TestGetCurrentUser_Success() {
	// Arrange
	userID := primitive.NewObjectID()
	user, _ := entities.NewUser("testuser", "password123")
	user.ID = userID

	// Setup Gin context with userID
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID.Hex())
		c.Next()
	})
	router.GET("/api/auth/me", suite.handler.GetCurrentUser)

	suite.mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(user, nil)

	// Act
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	
	data := response["data"].(map[string]interface{})
	userResponse := data["user"].(map[string]interface{})
	assert.Equal(suite.T(), user.Username, userResponse["username"])

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthHandlerTestSuite) TestGetCurrentUser_NoAuth() {
	// Act
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User not authenticated", response["error"])
}

func (suite *AuthHandlerTestSuite) TestGetCurrentUser_Error() {
	// Arrange
	userID := primitive.NewObjectID()

	// Setup Gin context with userID
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", userID.Hex())
		c.Next()
	})
	router.GET("/api/auth/me", suite.handler.GetCurrentUser)

	suite.mockUserRepo.On("GetByID", mock.Anything, userID.Hex()).Return(nil, interfaces.ErrNotFound)

	// Act
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Failed to get user information", response["error"])

	suite.mockUserRepo.AssertExpectations(suite.T())
}

// Run the test suite
func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

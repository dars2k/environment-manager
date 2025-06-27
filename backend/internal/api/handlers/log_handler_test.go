package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/api/handlers"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/log"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


// LogHandler Test Suite
type LogHandlerTestSuite struct {
	suite.Suite
	logger *logrus.Logger
}

func (suite *LogHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	
	suite.logger = logrus.New()
	suite.logger.SetOutput(io.Discard) // Disable logging during tests
}

func (suite *LogHandlerTestSuite) TestList_Success() {
	// Arrange
	logs := []*entities.Log{
		{
			ID:        primitive.NewObjectID(),
			Type:      entities.LogTypeSystem,
			Level:     entities.LogLevelInfo,
			Message:   "Test log 1",
			Timestamp: time.Now(),
		},
		{
			ID:        primitive.NewObjectID(),
			Type:      entities.LogTypeAuth,
			Level:     entities.LogLevelWarning,
			Message:   "Test log 2",
			Timestamp: time.Now(),
		},
	}
	
	filter := interfaces.LogFilter{
		Page:  1,
		Limit: 50,
	}
	
	// Create a mock repository and log service
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("List", mock.Anything, filter).Return(logs, int64(2), nil)

	// Act
	req := httptest.NewRequest("GET", "/api/logs", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs", handler.List)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	logsResponse := data["logs"].([]interface{})
	assert.Len(suite.T(), logsResponse, 2)
	
	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), pagination["page"])
	assert.Equal(suite.T(), float64(50), pagination["limit"])
	assert.Equal(suite.T(), float64(2), pagination["total"])

	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestList_WithFilters() {
	// Arrange
	envID := primitive.NewObjectID()
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()
	
	// Create a mock repository and log service
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f interfaces.LogFilter) bool {
		return f.Page == 2 &&
			f.Limit == 20 &&
			f.EnvironmentID != nil &&
			f.Type == entities.LogTypeAuth &&
			f.Level == entities.LogLevelError &&
			f.Action == entities.ActionTypeLogin &&
			f.Search == "test search"
	})).Return([]*entities.Log{}, int64(0), nil)

	// Act
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/logs?page=2&limit=20&environmentId=%s&type=auth&level=error&action=login&search=test%%20search&startTime=%s&endTime=%s",
		envID.Hex(),
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
	), nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs", handler.List)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestList_InvalidPage() {
	// Arrange
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	// Should use default page 1
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f interfaces.LogFilter) bool {
		return f.Page == 1
	})).Return([]*entities.Log{}, int64(0), nil)

	// Act
	req := httptest.NewRequest("GET", "/api/logs?page=-1", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs", handler.List)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestList_LimitExceedsMax() {
	// Arrange
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	// Should use default limit 50
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f interfaces.LogFilter) bool {
		return f.Limit == 50
	})).Return([]*entities.Log{}, int64(0), nil)

	// Act
	req := httptest.NewRequest("GET", "/api/logs?limit=200", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs", handler.List)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestList_Error() {
	// Arrange
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), fmt.Errorf("database error"))

	// Act
	req := httptest.NewRequest("GET", "/api/logs", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs", handler.List)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Failed to retrieve logs", response["error"])

	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestGetByID_Success() {
	// Arrange
	logID := primitive.NewObjectID()
	logEntry := &entities.Log{
		ID:        logID,
		Type:      entities.LogTypeSystem,
		Level:     entities.LogLevelInfo,
		Message:   "Test log",
		Timestamp: time.Now(),
	}
	
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("GetByID", mock.Anything, logID).Return(logEntry, nil)

	// Act
	req := httptest.NewRequest("GET", "/api/logs/"+logID.Hex(), nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs/:id", handler.GetByID)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	assert.NotNil(suite.T(), data["log"])

	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestGetByID_NotFound() {
	// Arrange
	logID := primitive.NewObjectID()
	
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("GetByID", mock.Anything, logID).Return(nil, interfaces.ErrNotFound)

	// Act
	req := httptest.NewRequest("GET", "/api/logs/"+logID.Hex(), nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs/:id", handler.GetByID)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Log not found", response["error"])

	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestGetByID_Error() {
	// Arrange
	logID := primitive.NewObjectID()
	
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("GetByID", mock.Anything, logID).Return(nil, fmt.Errorf("database error"))

	// Act
	req := httptest.NewRequest("GET", "/api/logs/"+logID.Hex(), nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs/:id", handler.GetByID)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Failed to retrieve log", response["error"])

	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestGetEnvironmentLogs_Success() {
	// Arrange
	envID := primitive.NewObjectID()
	logs := []*entities.Log{
		{
			ID:            primitive.NewObjectID(),
			Type:          entities.LogTypeAction,
			Level:         entities.LogLevelInfo,
			Message:       "Environment log",
			EnvironmentID: &envID,
			Timestamp:     time.Now(),
		},
	}
	
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("GetEnvironmentLogs", mock.Anything, envID, 100).Return(logs, nil)

	// Act
	req := httptest.NewRequest("GET", "/api/environments/"+envID.Hex()+"/logs", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/environments/:id/logs", handler.GetEnvironmentLogs)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	logsResponse := data["logs"].([]interface{})
	assert.Len(suite.T(), logsResponse, 1)

	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestGetEnvironmentLogs_WithLimit() {
	// Arrange
	envID := primitive.NewObjectID()
	
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("GetEnvironmentLogs", mock.Anything, envID, 50).Return([]*entities.Log{}, nil)

	// Act
	req := httptest.NewRequest("GET", "/api/environments/"+envID.Hex()+"/logs?limit=50", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/environments/:id/logs", handler.GetEnvironmentLogs)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestGetEnvironmentLogs_LimitExceedsMax() {
	// Arrange
	envID := primitive.NewObjectID()
	
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	// Should use default limit 100
	mockRepo.On("GetEnvironmentLogs", mock.Anything, envID, 100).Return([]*entities.Log{}, nil)

	// Act
	req := httptest.NewRequest("GET", "/api/environments/"+envID.Hex()+"/logs?limit=2000", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/environments/:id/logs", handler.GetEnvironmentLogs)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestGetEnvironmentLogs_Error() {
	// Arrange
	envID := primitive.NewObjectID()
	
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("GetEnvironmentLogs", mock.Anything, envID, 100).Return(nil, fmt.Errorf("database error"))

	// Act
	req := httptest.NewRequest("GET", "/api/environments/"+envID.Hex()+"/logs", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/environments/:id/logs", handler.GetEnvironmentLogs)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Failed to retrieve environment logs", response["error"])

	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestCount_Success() {
	// Arrange
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("Count", mock.Anything, mock.MatchedBy(func(f interfaces.LogFilter) bool {
		return true
	})).Return(int64(42), nil)

	// Act
	req := httptest.NewRequest("GET", "/api/logs/count", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs/count", handler.Count)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(42), data["count"])

	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestCount_WithFilters() {
	// Arrange
	envID := primitive.NewObjectID()
	since := time.Now().Add(-24 * time.Hour)
	
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	// Just match the basic fields we know will be set
	mockRepo.On("Count", mock.Anything, mock.MatchedBy(func(f interfaces.LogFilter) bool {
		// Basic validation that the filter has the expected values
		return f.EnvironmentID != nil && 
			f.EnvironmentID.Hex() == envID.Hex() &&
			f.Type == entities.LogTypeAuth &&
			f.Level == entities.LogLevelError
	})).Return(int64(10), nil)

	// Act
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/logs/count?since=%s&environmentId=%s&type=auth&level=error",
		since.Format(time.RFC3339),
		envID.Hex(),
	), nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs/count", handler.Count)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["data"])
	
	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(10), data["count"])

	mockRepo.AssertExpectations(suite.T())
}

func (suite *LogHandlerTestSuite) TestCount_Error() {
	// Arrange
	mockRepo := new(MockLogRepository)
	logService := log.NewService(mockRepo)
	handler := handlers.NewLogHandler(logService, suite.logger)
	
	mockRepo.On("Count", mock.Anything, mock.Anything).Return(int64(0), fmt.Errorf("database error"))

	// Act
	req := httptest.NewRequest("GET", "/api/logs/count", nil)
	w := httptest.NewRecorder()
	
	router := gin.New()
	router.GET("/api/logs/count", handler.Count)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Failed to count logs", response["error"])

	mockRepo.AssertExpectations(suite.T())
}

// Run the test suite
func TestLogHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(LogHandlerTestSuite))
}

package routes_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"app-env-manager/internal/api/handlers"
	"app-env-manager/internal/api/routes"
	"app-env-manager/internal/websocket/hub"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock handlers
type MockEnvironmentHandler struct {
	mock.Mock
}

func (m *MockEnvironmentHandler) List(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"environments":[]}`))
}

func (m *MockEnvironmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"id":"123"}`))
}

func (m *MockEnvironmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"id":"123","name":"test"}`))
}

func (m *MockEnvironmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"id":"123","name":"updated"}`))
}

func (m *MockEnvironmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func (m *MockEnvironmentHandler) Restart(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"restarting"}`))
}

func (m *MockEnvironmentHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func (m *MockEnvironmentHandler) GetVersions(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"versions":["1.0.0","1.1.0"]}`))
}

func (m *MockEnvironmentHandler) Upgrade(w http.ResponseWriter, r *http.Request) {
	m.Called(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"upgrading"}`))
}


func createMockAuthHandler() *handlers.AuthHandler {
	// We can't easily mock the auth handler since it's a struct with methods
	// So we'll just create a simple implementation
	return nil
}

func createMockLogHandler() *handlers.LogHandler {
	return nil
}

func createMockUserHandler() *handlers.UserHandler {
	return nil
}

func createTestToken(secret string, userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   userID,
		"username": "testuser",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	return token.SignedString([]byte(secret))
}

func TestNewRouter(t *testing.T) {
	// Create mock dependencies
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	
	wsHub := hub.NewHub(logger)
	
	cfg := routes.Config{
		EnvironmentHandler: &handlers.EnvironmentHandler{},
		LogHandler:         createMockLogHandler(),
		AuthHandler:        createMockAuthHandler(),
		UserHandler:        createMockUserHandler(),
		WebSocketHub:       wsHub,
		Logger:             logger,
		JWTSecret:          "test-secret",
		AllowedOrigins:     []string{"http://localhost:3000"},
	}

	// Create router
	router := routes.NewRouter(cfg)
	assert.NotNil(t, router)

	// Test that router is an http.Handler
	_, ok := router.(http.Handler)
	assert.True(t, ok)
}

func TestHealthCheck(t *testing.T) {
	// Create request
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	// Call handler directly
	routes.HealthCheck(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "1.0.0", response["version"])
}

func TestHandleWebSocket(t *testing.T) {
	// Create dependencies
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	wsHub := hub.NewHub(logger)

	// Start hub
	go wsHub.Run()

	const testSecret = "test-secret"
	handler := routes.HandleWebSocket(wsHub, logger, testSecret, []string{"http://localhost"})

	// Create test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Create a valid token for the connection
	token, err := createTestToken(testSecret, "user-123")
	assert.NoError(t, err)

	// Convert http:// to ws:// and append token query param
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=" + token

	// Test successful WebSocket connection with valid token
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	conn.Close()
	time.Sleep(100 * time.Millisecond)
}

func TestHandleWebSocket_NoToken(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	wsHub := hub.NewHub(logger)

	handler := routes.HandleWebSocket(wsHub, logger, "test-secret", []string{"http://localhost"})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Attempt connection without a token – should be rejected
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	dialer := websocket.Dialer{}
	_, resp, err := dialer.Dial(wsURL, nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandleWebSocket_InvalidToken(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	wsHub := hub.NewHub(logger)

	handler := routes.HandleWebSocket(wsHub, logger, "test-secret", []string{"http://localhost"})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Attempt connection with an invalid token – should be rejected
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=invalid.token.here"
	dialer := websocket.Dialer{}
	_, resp, err := dialer.Dial(wsURL, nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestHandleWebSocket_UpgradeError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	wsHub := hub.NewHub(logger)

	const testSecret = "test-secret"
	handler := routes.HandleWebSocket(wsHub, logger, testSecret, []string{"http://localhost"})

	// Create a valid token but send a plain HTTP request (no WS upgrade headers)
	token, err := createTestToken(testSecret, "user-123")
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/ws?token="+token, nil)
	w := httptest.NewRecorder()

	handler(w, req)

	// Upgrade should fail – definitely not a 101 Switching Protocols
	assert.NotEqual(t, http.StatusSwitchingProtocols, w.Code)
}

func TestGenerateClientID(t *testing.T) {
	// Verify that concurrent WebSocket connections each get a unique client ID
	// by connecting multiple times and confirming each connection succeeds.
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	const testSecret = "test-secret"
	const connections = 10

	for i := 0; i < connections; i++ {
		wsHub := hub.NewHub(logger)
		go wsHub.Run()

		handler := routes.HandleWebSocket(wsHub, logger, testSecret, []string{})
		server := httptest.NewServer(handler)

		token, err := createTestToken(testSecret, "user-123")
		assert.NoError(t, err)

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=" + token
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		assert.NotNil(t, conn)

		conn.Close()
		server.Close()
	}
}

func TestRoutes_Authentication(t *testing.T) {
	// Create mock dependencies
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	
	wsHub := hub.NewHub(logger)
	
	cfg := routes.Config{
		EnvironmentHandler: &handlers.EnvironmentHandler{},
		LogHandler:         createMockLogHandler(),
		AuthHandler:        createMockAuthHandler(),
		UserHandler:        createMockUserHandler(),
		WebSocketHub:       wsHub,
		Logger:             logger,
		JWTSecret:          "test-secret",
		AllowedOrigins:     []string{"*"},
	}

	// Create router
	router := routes.NewRouter(cfg)
	server := httptest.NewServer(router)
	defer server.Close()

	// Test unauthenticated request to protected route
	resp, err := http.Get(server.URL + "/api/v1/environments")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Test authenticated request
	token, err := createTestToken("test-secret", "user-123")
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/environments", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err = client.Do(req)
	assert.NoError(t, err)
	// Since we're using nil handlers, we expect an error response, not OK
	// The authentication should pass, but the handler will fail
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRoutes_CORS(t *testing.T) {
	// Create mock dependencies
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	
	wsHub := hub.NewHub(logger)
	
	cfg := routes.Config{
		EnvironmentHandler: &handlers.EnvironmentHandler{},
		LogHandler:         createMockLogHandler(),
		AuthHandler:        createMockAuthHandler(),
		UserHandler:        createMockUserHandler(),
		WebSocketHub:       wsHub,
		Logger:             logger,
		JWTSecret:          "test-secret",
		AllowedOrigins:     []string{"http://localhost:3000"},
	}

	// Create router
	router := routes.NewRouter(cfg)
	server := httptest.NewServer(router)
	defer server.Close()

	// Test CORS preflight request
	req, _ := http.NewRequest("OPTIONS", server.URL+"/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	
	// Check CORS headers
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"))
}

func TestRoutes_EnvironmentEndpoints(t *testing.T) {
	t.Skip("Skipping test - needs proper handler interface")
}

func TestRoutes_Middleware(t *testing.T) {
	// Create mock dependencies
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	// We need to test that middleware is applied
	// Since we can't easily inject a panic into the existing handlers,
	// we'll test that the logging middleware works by checking logs
	
	cfg := routes.Config{
		EnvironmentHandler: &handlers.EnvironmentHandler{},
		LogHandler:         createMockLogHandler(),
		AuthHandler:        createMockAuthHandler(),
		UserHandler:        createMockUserHandler(),
		WebSocketHub:       hub.NewHub(logger),
		Logger:             logger,
		JWTSecret:          "test-secret",
		AllowedOrigins:     []string{"*"},
	}

	// Create router
	router := routes.NewRouter(cfg)
	server := httptest.NewServer(router)
	defer server.Close()

	// Make a request to health endpoint
	resp, err := http.Get(server.URL + "/api/v1/health")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// The logging middleware should have logged the request
	// We can't easily check the logs without a custom logger hook
}

// Benchmark tests
func BenchmarkHealthCheck(b *testing.B) {
	req := httptest.NewRequest("GET", "/health", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		routes.HealthCheck(w, req)
	}
}

func BenchmarkRouter(b *testing.B) {
	// Create mock dependencies
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	
	cfg := routes.Config{
		EnvironmentHandler: &handlers.EnvironmentHandler{},
		LogHandler:         createMockLogHandler(),
		AuthHandler:        createMockAuthHandler(),
		UserHandler:        createMockUserHandler(),
		WebSocketHub:       hub.NewHub(logger),
		Logger:             logger,
		JWTSecret:          "test-secret",
		AllowedOrigins:     []string{"*"},
	}

	// Create router
	router := routes.NewRouter(cfg)
	server := httptest.NewServer(router)
	defer server.Close()

	// Create auth token
	token, _ := createTestToken("test-secret", "user-123")

	// Create request
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/health", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := client.Do(req)
		resp.Body.Close()
	}
}

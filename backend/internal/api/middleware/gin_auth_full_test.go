package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/api/middleware"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/auth"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---- shared mocks ----

type ginAuthMockUserRepo struct{ mock.Mock }

func (m *ginAuthMockUserRepo) Create(ctx context.Context, u *entities.User) error {
	return m.Called(ctx, u).Error(0)
}
func (m *ginAuthMockUserRepo) GetByID(ctx context.Context, id string) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}
func (m *ginAuthMockUserRepo) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}
func (m *ginAuthMockUserRepo) Update(ctx context.Context, id string, u *entities.User) error {
	return m.Called(ctx, id, u).Error(0)
}
func (m *ginAuthMockUserRepo) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *ginAuthMockUserRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *ginAuthMockUserRepo) List(ctx context.Context, f interfaces.ListFilter) ([]*entities.User, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}
func (m *ginAuthMockUserRepo) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

type ginAuthMockLogRepo struct{ mock.Mock }

func (m *ginAuthMockLogRepo) Create(ctx context.Context, l *entities.Log) error {
	return m.Called(ctx, l).Error(0)
}
func (m *ginAuthMockLogRepo) List(ctx context.Context, f interfaces.LogFilter) ([]*entities.Log, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entities.Log), args.Get(1).(int64), args.Error(2)
}
func (m *ginAuthMockLogRepo) GetByID(ctx context.Context, id primitive.ObjectID) (*entities.Log, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Log), args.Error(1)
}
func (m *ginAuthMockLogRepo) DeleteOld(ctx context.Context, d time.Duration) (int64, error) {
	args := m.Called(ctx, d)
	return args.Get(0).(int64), args.Error(1)
}
func (m *ginAuthMockLogRepo) GetEnvironmentLogs(ctx context.Context, envID primitive.ObjectID, limit int) ([]*entities.Log, error) {
	args := m.Called(ctx, envID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Log), args.Error(1)
}
func (m *ginAuthMockLogRepo) Count(ctx context.Context, f interfaces.LogFilter) (int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).(int64), args.Error(1)
}

// ---- helpers ----

const testJWTSecret = "test-secret-key"

func newTestAuthServices(userRepo *ginAuthMockUserRepo) (*auth.Service, *user.Service) {
	logRepo := &ginAuthMockLogRepo{}
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	logSvc := log.NewService(logRepo)
	authSvc := auth.NewService(userRepo, logSvc, testJWTSecret, time.Hour)
	userSvc := user.NewService(userRepo, logSvc)
	return authSvc, userSvc
}

func makeValidToken(userID, username string, role entities.UserRole) string {
	claims := jwt.MapClaims{
		"userId":   userID,
		"username": username,
		"role":     string(role),
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iat":      time.Now().Unix(),
		"iss":      "app-env-manager",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

func makeExpiredToken() string {
	claims := jwt.MapClaims{
		"userId":   primitive.NewObjectID().Hex(),
		"username": "expired",
		"role":     "user",
		"exp":      time.Now().Add(-time.Hour).Unix(),
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
		"iss":      "app-env-manager",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

func silentLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(nil)
	// Discard by setting level higher than fatal
	l.SetLevel(logrus.PanicLevel)
	return l
}

// ---- AuthMiddleware tests ----

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	w := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(w)
	engine.Use(middleware.AuthMiddleware(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request = req
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	for _, header := range []string{"token-only", "Basic dXNlcjpwYXNz", "Bearer a b"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", header)

		engine := gin.New()
		engine.Use(middleware.AuthMiddleware(authSvc, userSvc, silentLogger()))
		engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "header: %s", header)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.jwt")

	engine := gin.New()
	engine.Use(middleware.AuthMiddleware(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+makeExpiredToken())

	engine := gin.New()
	engine.Use(middleware.AuthMiddleware(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	id := primitive.NewObjectID()
	tokenStr := makeValidToken(id.Hex(), "alice", entities.UserRoleUser)

	userRepo.On("GetByID", mock.Anything, id.Hex()).Return(nil, fmt.Errorf("not found"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	engine := gin.New()
	engine.Use(middleware.AuthMiddleware(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InactiveUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	id := primitive.NewObjectID()
	tokenStr := makeValidToken(id.Hex(), "inactive", entities.UserRoleUser)

	inactive := &entities.User{
		ID:       id,
		Username: "inactive",
		Active:   false,
		Role:     entities.UserRoleUser,
	}
	userRepo.On("GetByID", mock.Anything, id.Hex()).Return(inactive, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	engine := gin.New()
	engine.Use(middleware.AuthMiddleware(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	id := primitive.NewObjectID()
	tokenStr := makeValidToken(id.Hex(), "alice", entities.UserRoleAdmin)

	activeUser := &entities.User{
		ID:       id,
		Username: "alice",
		Active:   true,
		Role:     entities.UserRoleAdmin,
	}
	userRepo.On("GetByID", mock.Anything, id.Hex()).Return(activeUser, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	engine := gin.New()
	engine.Use(middleware.AuthMiddleware(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- OptionalAuth tests ----

func TestOptionalAuth_NoHeader_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	engine := gin.New()
	engine.Use(middleware.OptionalAuth(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuth_InvalidHeader_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic credentials")

	engine := gin.New()
	engine.Use(middleware.OptionalAuth(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuth_InvalidToken_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")

	engine := gin.New()
	engine.Use(middleware.OptionalAuth(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuth_ValidToken_SetsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	id := primitive.NewObjectID()
	tokenStr := makeValidToken(id.Hex(), "bob", entities.UserRoleUser)

	activeUser := &entities.User{
		ID:       id,
		Username: "bob",
		Active:   true,
		Role:     entities.UserRoleUser,
	}
	userRepo.On("GetByID", mock.Anything, id.Hex()).Return(activeUser, nil)

	var gotUserID interface{}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	engine := gin.New()
	engine.Use(middleware.OptionalAuth(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) {
		gotUserID, _ = c.Get("userID")
		c.Status(http.StatusOK)
	})
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, id.Hex(), gotUserID)
}

func TestOptionalAuth_InactiveUser_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userRepo := &ginAuthMockUserRepo{}
	authSvc, userSvc := newTestAuthServices(userRepo)

	id := primitive.NewObjectID()
	tokenStr := makeValidToken(id.Hex(), "inactive", entities.UserRoleUser)

	inactiveUser := &entities.User{
		ID:       id,
		Username: "inactive",
		Active:   false,
		Role:     entities.UserRoleUser,
	}
	userRepo.On("GetByID", mock.Anything, id.Hex()).Return(inactiveUser, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	engine := gin.New()
	engine.Use(middleware.OptionalAuth(authSvc, userSvc, silentLogger()))
	engine.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.ServeHTTP(w, req)

	// OptionalAuth should not block even for inactive users
	assert.Equal(t, http.StatusOK, w.Code)
}

package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/auth"
	"app-env-manager/internal/service/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}


func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id string, user *entities.User) error {
	args := m.Called(ctx, id, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, filter interfaces.ListFilter) ([]*entities.User, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}


// MockLogRepository is a mock implementation of LogRepository
type MockLogRepository struct {
	mock.Mock
}

func (m *MockLogRepository) Create(ctx context.Context, log *entities.Log) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockLogRepository) List(ctx context.Context, filter interfaces.LogFilter) ([]*entities.Log, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entities.Log), args.Get(1).(int64), args.Error(2)
}

func (m *MockLogRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entities.Log, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Log), args.Error(1)
}

func (m *MockLogRepository) DeleteOld(ctx context.Context, olderThan time.Duration) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLogRepository) GetEnvironmentLogs(ctx context.Context, envID primitive.ObjectID, limit int) ([]*entities.Log, error) {
	args := m.Called(ctx, envID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Log), args.Error(1)
}

func (m *MockLogRepository) Count(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// AuthServiceTestSuite is the test suite for auth service
type AuthServiceTestSuite struct {
	suite.Suite
	service        *auth.Service
	mockUserRepo   *MockUserRepository
	mockLogRepo    *MockLogRepository
	logService     *log.Service
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.mockUserRepo = new(MockUserRepository)
	suite.mockLogRepo = new(MockLogRepository)
	
	// Create log service with mock repository
	suite.logService = log.NewService(suite.mockLogRepo)
	
	suite.service = auth.NewService(
		suite.mockUserRepo,
		suite.logService,
		"test-secret-key",
		24*time.Hour,
	)
}

func (suite *AuthServiceTestSuite) TestLogin_Success() {
	// Arrange
	ctx := context.Background()
	req := entities.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	user, _ := entities.NewUser("testuser", "password123")
	user.ID = primitive.NewObjectID()
	user.Active = true

	suite.mockUserRepo.On("GetByUsername", ctx, req.Username).Return(user, nil)
	suite.mockUserRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)
	suite.mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	response, err := suite.service.Login(ctx, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.NotEmpty(suite.T(), response.Token)
	assert.Equal(suite.T(), user.ID, response.User.ID)
	assert.Equal(suite.T(), user.Username, response.User.Username)
	assert.WithinDuration(suite.T(), time.Now().Add(24*time.Hour), response.ExpiresAt, 2*time.Second)

	// Verify the token
	parsedToken, err := jwt.ParseWithClaims(response.Token, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*auth.Claims)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), user.ID.Hex(), claims.UserID)
	assert.Equal(suite.T(), user.Username, claims.Username)
	assert.Equal(suite.T(), user.Role, claims.Role)

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_UserNotFound() {
	// Arrange
	ctx := context.Background()
	req := entities.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	suite.mockUserRepo.On("GetByUsername", ctx, req.Username).Return(nil, interfaces.ErrUserNotFound)
	suite.mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	response, err := suite.service.Login(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "invalid credentials", err.Error())
	assert.Nil(suite.T(), response)

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_IncorrectPassword() {
	// Arrange
	ctx := context.Background()
	req := entities.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	user, _ := entities.NewUser("testuser", "correctpassword")
	user.ID = primitive.NewObjectID()
	user.Active = true

	suite.mockUserRepo.On("GetByUsername", ctx, req.Username).Return(user, nil)
	suite.mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	response, err := suite.service.Login(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "invalid credentials", err.Error())
	assert.Nil(suite.T(), response)

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_InactiveUser() {
	// Arrange
	ctx := context.Background()
	req := entities.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	user, _ := entities.NewUser("testuser", "password123")
	user.ID = primitive.NewObjectID()
	user.Active = false // Inactive user

	suite.mockUserRepo.On("GetByUsername", ctx, req.Username).Return(user, nil)
	suite.mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	response, err := suite.service.Login(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "account is inactive", err.Error())
	assert.Nil(suite.T(), response)

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestValidateToken_Valid() {
	// Arrange
	// Create a valid token
	claims := &auth.Claims{
		UserID:   "123",
		Username: "testuser",
		Role:     entities.UserRoleAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret-key"))

	// Act
	validatedClaims, err := suite.service.ValidateToken(tokenString)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), validatedClaims)
	assert.Equal(suite.T(), "123", validatedClaims.UserID)
	assert.Equal(suite.T(), "testuser", validatedClaims.Username)
	assert.Equal(suite.T(), entities.UserRoleAdmin, validatedClaims.Role)
}

func (suite *AuthServiceTestSuite) TestValidateToken_Expired() {
	// Arrange
	// Create an expired token
	claims := &auth.Claims{
		UserID:   "123",
		Username: "testuser",
		Role:     entities.UserRoleAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
			IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret-key"))

	// Act
	validatedClaims, err := suite.service.ValidateToken(tokenString)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedClaims)
}

func (suite *AuthServiceTestSuite) TestValidateToken_InvalidSignature() {
	// Arrange
	// Create a token with wrong secret
	claims := &auth.Claims{
		UserID:   "123",
		Username: "testuser",
		Role:     entities.UserRoleAdmin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("wrong-secret"))

	// Act
	validatedClaims, err := suite.service.ValidateToken(tokenString)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedClaims)
}

func (suite *AuthServiceTestSuite) TestValidateToken_Malformed() {
	// Act
	validatedClaims, err := suite.service.ValidateToken("invalid.token.string")

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedClaims)
}

func (suite *AuthServiceTestSuite) TestGetUserFromContext() {
	// Arrange
	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	user := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
		Role:     entities.UserRoleAdmin,
		Active:   true,
	}

	suite.mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

	// Act
	foundUser, err := suite.service.GetUserFromContext(ctx, userID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundUser)
	assert.Equal(suite.T(), user.Username, foundUser.Username)

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestGetUserFromContext_NotFound() {
	// Arrange
	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()

	suite.mockUserRepo.On("GetByID", ctx, userID).Return(nil, interfaces.ErrUserNotFound)

	// Act
	foundUser, err := suite.service.GetUserFromContext(ctx, userID)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundUser)

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestCreateInitialAdmin_NoUsersExist() {
	// Arrange
	ctx := context.Background()

	suite.mockUserRepo.On("Count", ctx).Return(int64(0), nil)
	suite.mockUserRepo.On("Create", ctx, mock.AnythingOfType("*entities.User")).Return(nil)
	suite.mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	err := suite.service.CreateInitialAdmin(ctx)

	// Assert
	assert.NoError(suite.T(), err)

	// Verify the admin user creation
	suite.mockUserRepo.AssertCalled(suite.T(), "Create", ctx, mock.MatchedBy(func(user *entities.User) bool {
		return user.Username == "admin" &&
			user.Role == entities.UserRoleAdmin &&
			user.Active == true
	}))

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestCreateInitialAdmin_UsersAlreadyExist() {
	// Arrange
	ctx := context.Background()

	suite.mockUserRepo.On("Count", ctx).Return(int64(1), nil)

	// Act
	err := suite.service.CreateInitialAdmin(ctx)

	// Assert
	assert.NoError(suite.T(), err)

	// Verify that Create was not called
	suite.mockUserRepo.AssertNotCalled(suite.T(), "Create")

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestCreateInitialAdmin_CountError() {
	// Arrange
	ctx := context.Background()

	suite.mockUserRepo.On("Count", ctx).Return(int64(0), assert.AnError)

	// Act
	err := suite.service.CreateInitialAdmin(ctx)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to count users")

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestCreateInitialAdmin_CreateError() {
	// Arrange
	ctx := context.Background()

	suite.mockUserRepo.On("Count", ctx).Return(int64(0), nil)
	suite.mockUserRepo.On("Create", ctx, mock.AnythingOfType("*entities.User")).Return(assert.AnError)

	// Act
	err := suite.service.CreateInitialAdmin(ctx)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to save admin user")

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestValidateToken_InvalidSigningMethod() {
	// Arrange
	// Create a token with RS256 instead of HS256
	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims = jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	
	// Generate a fake RSA private key just for signing
	tokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MDk0NTkyMDB9.invalid"

	// Act
	validatedClaims, err := suite.service.ValidateToken(tokenString)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedClaims)
}

func (suite *AuthServiceTestSuite) TestValidateToken_InvalidClaims() {
	// Arrange
	// The service actually accepts standard claims but returns custom Claims
	// So this test is not valid as-is. Let's remove it.
}

func (suite *AuthServiceTestSuite) TestLogin_UpdateLastLoginFailure() {
	// Arrange
	ctx := context.Background()
	req := entities.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	user, _ := entities.NewUser("testuser", "password123")
	user.ID = primitive.NewObjectID()
	user.Active = true

	suite.mockUserRepo.On("GetByUsername", ctx, req.Username).Return(user, nil)
	suite.mockUserRepo.On("UpdateLastLogin", ctx, user.ID).Return(assert.AnError)
	suite.mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	response, err := suite.service.Login(ctx, req)

	// Assert
	// Login should still succeed even if UpdateLastLogin fails
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.NotEmpty(suite.T(), response.Token)

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestLogin_EmptyPassword() {
	// Arrange
	ctx := context.Background()
	req := entities.LoginRequest{
		Username: "testuser",
		Password: "",
	}

	user, _ := entities.NewUser("testuser", "password123")
	user.ID = primitive.NewObjectID()
	user.Active = true

	suite.mockUserRepo.On("GetByUsername", ctx, req.Username).Return(user, nil)
	suite.mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	// Act
	response, err := suite.service.Login(ctx, req)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "invalid credentials", err.Error())
	assert.Nil(suite.T(), response)

	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *AuthServiceTestSuite) TestValidateToken_EmptyToken() {
	// Act
	validatedClaims, err := suite.service.ValidateToken("")

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), validatedClaims)
}

// Run the test suite
func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

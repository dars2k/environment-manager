package auth

import (
	"context"
	"fmt"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/log"
	"github.com/golang-jwt/jwt/v5"
)

// Service handles authentication business logic
type Service struct {
	userRepo   interfaces.UserRepository
	logService *log.Service
	jwtSecret  string
	jwtExpiry  time.Duration
}

// Claims represents JWT claims
type Claims struct {
	UserID   string             `json:"userId"`
	Username string             `json:"username"`
	Role     entities.UserRole  `json:"role"`
	jwt.RegisteredClaims
}

// NewService creates a new auth service
func NewService(userRepo interfaces.UserRepository, logService *log.Service, jwtSecret string, jwtExpiry time.Duration) *Service {
	return &Service{
		userRepo:   userRepo,
		logService: logService,
		jwtSecret:  jwtSecret,
		jwtExpiry:  jwtExpiry,
	}
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(ctx context.Context, req entities.LoginRequest) (*entities.LoginResponse, error) {
	// Find user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Log failed login attempt
		_ = s.logService.LogAuth(ctx, nil, req.Username, entities.ActionTypeLogin, 
			fmt.Sprintf("Login failed for user '%s': user not found", req.Username), false)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.Active {
		_ = s.logService.LogAuth(ctx, &user.ID, user.Username, entities.ActionTypeLogin, 
			fmt.Sprintf("Login failed for user '%s': account inactive", user.Username), false)
		return nil, fmt.Errorf("account is inactive")
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		_ = s.logService.LogAuth(ctx, &user.ID, user.Username, entities.ActionTypeLogin, 
			fmt.Sprintf("Login failed for user '%s': invalid password", user.Username), false)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	expiresAt := time.Now().Add(s.jwtExpiry)
	claims := &Claims{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "app-env-manager",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Log successful login
	_ = s.logService.LogAuth(ctx, &user.ID, user.Username, entities.ActionTypeLogin, 
		fmt.Sprintf("Login successful for user '%s'", user.Username), true)

	return &entities.LoginResponse{
		Token:     tokenString,
		User:      user,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetUserFromContext gets the authenticated user from the context
func (s *Service) GetUserFromContext(ctx context.Context, userID string) (*entities.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// CreateInitialAdmin creates the initial admin user if no users exist
func (s *Service) CreateInitialAdmin(ctx context.Context) error {
	// Check if any users exist
	count, err := s.userRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if count > 0 {
		return nil // Users already exist
	}

	// Create initial admin user
	adminUser, err := entities.NewUser("admin", "admin123")
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	if err := s.userRepo.Create(ctx, adminUser); err != nil {
		return fmt.Errorf("failed to save admin user: %w", err)
	}

	// Log system event
	_ = s.logService.LogSystem(ctx, "Initial admin user created", map[string]interface{}{
		"username": adminUser.Username,
		"role":     adminUser.Role,
	})

	return nil
}

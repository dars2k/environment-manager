# Testing Guide

This document describes the testing infrastructure and strategies for the Application Environment Manager.

## Overview

The project uses comprehensive testing strategies for both frontend and backend components:

- **Backend**: Go testing with testify framework
- **Frontend**: Vitest with React Testing Library

## Backend Testing

### Test Structure

Backend tests are organized alongside the code they test:

```
backend/
├── internal/
│   ├── api/
│   │   └── handlers/
│   │       ├── auth_handler.go
│   │       ├── auth_handler_test.go
│   │       ├── environment.go
│   │       └── environment_handler_test.go
│   └── service/
│       ├── auth/
│       │   ├── service.go
│       │   └── service_test.go
│       └── environment/
│           ├── service.go
│           └── service_test.go
```

### Running Backend Tests

```bash
# Run all tests
cd backend
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/service/auth
go test ./internal/api/handlers

# Run tests with verbose output
go test -v ./...
```

### Writing Backend Tests

Backend tests use the [testify](https://github.com/stretchr/testify) framework for assertions and mocking:

```go
package auth

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestAuthService_Login(t *testing.T) {
    // Setup
    mockRepo := new(MockUserRepository)
    service := NewService(mockRepo, "secret", 24*time.Hour)
    
    // Test successful login
    t.Run("successful login", func(t *testing.T) {
        // Arrange
        mockRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
        
        // Act
        token, err := service.Login(ctx, "test@example.com", "password")
        
        // Assert
        assert.NoError(t, err)
        assert.NotEmpty(t, token)
        mockRepo.AssertExpectations(t)
    })
}
```

### Mocking Dependencies

Use testify/mock for creating mock objects:

```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entities.User), args.Error(1)
}
```

## Frontend Testing

### Test Structure

Frontend tests are organized in `__tests__` directories:

```
frontend/src/
├── components/
│   └── environments/
│       ├── EnvironmentCard.tsx
│       └── __tests__/
│           └── EnvironmentCard.test.tsx
├── test/
│   ├── setup.ts
│   └── test-utils.tsx
└── vitest.config.ts
```

### Running Frontend Tests

```bash
# Run all tests
cd frontend
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with UI
npm run test:ui

# Run tests with coverage
npm run test:coverage

# Run specific test file
npm test EnvironmentCard
```

### Writing Frontend Tests

Frontend tests use [Vitest](https://vitest.dev/) and [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/):

```tsx
import { describe, it, expect } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '@/test/test-utils';
import { EnvironmentCard } from '../EnvironmentCard';

describe('EnvironmentCard', () => {
  it('should render environment information correctly', () => {
    render(<EnvironmentCard environment={mockEnvironment} />);
    
    expect(screen.getByText(mockEnvironment.name)).toBeInTheDocument();
    expect(screen.getByText(/healthy/i)).toBeInTheDocument();
  });

  it('should handle user interactions', async () => {
    const user = userEvent.setup();
    render(<EnvironmentCard environment={mockEnvironment} />);
    
    const button = screen.getByRole('button', { name: /more/i });
    await user.click(button);
    
    await waitFor(() => {
      expect(screen.getByText(/edit/i)).toBeInTheDocument();
    });
  });
});
```

### Test Utilities

The `test-utils.tsx` file provides a custom render function with all necessary providers:

```tsx
import { render } from '@/test/test-utils';

// This automatically wraps components with:
// - Redux Provider
// - React Router
// - Material-UI Theme Provider
// - WebSocket Provider
```

### Testing Best Practices

1. **Test User Behavior**: Focus on testing what users see and do, not implementation details
2. **Use Semantic Queries**: Prefer `getByRole`, `getByLabelText`, and `getByText` over test IDs
3. **Async Testing**: Use `waitFor` for async operations and state updates
4. **Mock External Dependencies**: Mock API calls and WebSocket connections
5. **Test Accessibility**: Ensure components are accessible by testing with semantic queries

## Integration Testing

### API Integration Tests

Create integration tests that test the full API flow:

```go
func TestEnvironmentAPI_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(db)
    
    // Create test server
    router := setupRouter(db)
    
    // Test environment creation
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/v1/environments", body)
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 201, w.Code)
}
```

### E2E Testing

For end-to-end testing, consider using:
- **Playwright** or **Cypress** for browser automation
- **Docker Compose** for spinning up the full stack

## Continuous Integration

### GitHub Actions Workflow

```yaml
name: Tests

on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run tests
        run: |
          cd backend
          go test -v -cover ./...

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install dependencies
        run: |
          cd frontend
          npm ci
      - name: Run tests
        run: |
          cd frontend
          npm test -- --coverage
```

## Test Coverage Goals

- **Backend**: Aim for >80% coverage for business logic
- **Frontend**: Aim for >70% coverage for components and hooks
- **Critical Paths**: 100% coverage for authentication, authorization, and data mutations

## Performance Testing

### Load Testing

Use tools like [k6](https://k6.io/) or [Apache JMeter](https://jmeter.apache.org/):

```javascript
// k6 load test example
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 0 },
  ],
};

export default function() {
  let response = http.get('http://localhost:8080/api/v1/environments');
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
}
```

## Security Testing

1. **Dependency Scanning**: Use `npm audit` and `go mod audit`
2. **SAST**: Integrate tools like SonarQube or Snyk
3. **Penetration Testing**: Regular security assessments

## Debugging Tests

### Backend Debugging

```bash
# Run specific test with debugging
go test -v -run TestAuthService_Login ./internal/service/auth

# Use delve debugger
dlv test ./internal/service/auth
```

### Frontend Debugging

```bash
# Run tests in debug mode
npm test -- --inspect

# Run specific test file in UI mode
npm run test:ui -- EnvironmentCard
```

## Test Data Management

### Fixtures

Store test data in dedicated files:

```
backend/test/fixtures/
├── environments.json
├── users.json
└── credentials.json
```

### Test Database

Use a separate test database:

```go
func setupTestDB(t *testing.T) *mongo.Database {
    client, err := mongo.Connect(context.Background(), 
        options.Client().ApplyURI("mongodb://localhost:27017"))
    require.NoError(t, err)
    
    db := client.Database("test_" + t.Name())
    t.Cleanup(func() {
        db.Drop(context.Background())
    })
    
    return db
}
```

## Troubleshooting

### Common Issues

1. **Flaky Tests**: Use proper waiting strategies and avoid time-based assertions
2. **Test Isolation**: Ensure tests don't affect each other by proper cleanup
3. **Mock Leaks**: Reset mocks between tests
4. **Slow Tests**: Use parallel testing where appropriate

### Tips

- Keep tests focused and independent
- Use descriptive test names
- Follow the AAA pattern (Arrange, Act, Assert)
- Mock external dependencies
- Test edge cases and error scenarios

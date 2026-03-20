# Testing Guide

This document describes the testing infrastructure and strategies for the Environment Manager.

## Overview

| Layer | Framework |
|-------|-----------|
| Backend | Go `testing` package + [testify](https://github.com/stretchr/testify) |
| Frontend | [Vitest](https://vitest.dev/) + [React Testing Library](https://testing-library.com/) |

## Running Tests

### All tests

```bash
make test
```

### Backend only

```bash
cd backend
go test ./...

# With coverage
go test -cover ./...

# Detailed HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Verbose output
go test -v ./...

# Specific package
go test ./internal/service/auth/...
go test ./internal/api/handlers/...
```

### Frontend only

```bash
cd frontend
npm test                  # Run once
npm test -- --watch       # Watch mode
npm run test:ui           # Run with Vitest UI
npm run test:coverage     # With coverage report
npm test EnvironmentCard  # Specific test file
```

## Backend Testing

### Test Structure

Tests live alongside the code they test:

```
backend/internal/
├── api/
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── auth_handler_test.go
│   │   ├── environment.go
│   │   └── environment_handler_test.go
│   └── middleware/
│       ├── auth.go
│       ├── auth_test.go
│       ├── mux_auth.go
│       └── mux_auth_test.go
└── service/
    ├── auth/
    │   ├── service.go
    │   └── service_test.go
    └── environment/
        ├── service.go
        └── service_test.go
```

### Writing Backend Tests

Use testify for assertions and mocking:

```go
package auth

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestAuthService_Login(t *testing.T) {
    mockRepo := new(MockUserRepository)
    service := NewService(mockRepo, "secret", 24*time.Hour)

    t.Run("successful login", func(t *testing.T) {
        mockRepo.On("GetByUsername", mock.Anything, "admin").Return(user, nil)

        resp, err := service.Login(context.Background(), "admin", "correct-password")

        assert.NoError(t, err)
        assert.NotEmpty(t, resp.Token)
        mockRepo.AssertExpectations(t)
    })

    t.Run("invalid credentials", func(t *testing.T) {
        mockRepo.On("GetByUsername", mock.Anything, "admin").Return(user, nil)

        _, err := service.Login(context.Background(), "admin", "wrong-password")

        assert.ErrorIs(t, err, ErrInvalidCredentials)
    })
}
```

### Mocking Dependencies

```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
    args := m.Called(ctx, username)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entities.User), args.Error(1)
}
```

### Debugging Backend Tests

```bash
# Run a specific test with verbose output
go test -v -run TestAuthService_Login ./internal/service/auth/...

# Use the delve debugger
dlv test ./internal/service/auth/...
```

## Frontend Testing

### Test Structure

```
frontend/src/
├── components/
│   └── environments/
│       ├── EnvironmentCard.tsx
│       └── __tests__/
│           └── EnvironmentCard.test.tsx
└── test/
    ├── setup.ts
    └── test-utils.tsx
```

### Writing Frontend Tests

```tsx
import { describe, it, expect } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from '@/test/test-utils';
import { EnvironmentCard } from '../EnvironmentCard';

describe('EnvironmentCard', () => {
  it('renders environment information', () => {
    render(<EnvironmentCard environment={mockEnvironment} />);

    expect(screen.getByText(mockEnvironment.name)).toBeInTheDocument();
    expect(screen.getByText(/healthy/i)).toBeInTheDocument();
  });

  it('handles restart action', async () => {
    const user = userEvent.setup();
    render(<EnvironmentCard environment={mockEnvironment} />);

    await user.click(screen.getByRole('button', { name: /restart/i }));

    await waitFor(() => {
      expect(screen.getByText(/restarting/i)).toBeInTheDocument();
    });
  });
});
```

### Test Utilities

`test-utils.tsx` provides a custom render that wraps components with all required providers:

```tsx
import { render } from '@/test/test-utils';

// Automatically wraps with:
// - Redux Provider
// - React Router
// - Material-UI Theme Provider
// - WebSocket Provider
```

### Debugging Frontend Tests

```bash
# Run in debug mode
npm test -- --inspect

# Run specific file with UI
npm run test:ui -- EnvironmentCard
```

## Integration Testing

The CI pipeline builds and runs the full Docker Compose stack on every PR via `docker-compose.test.yml`.

For local integration testing:

```bash
# Build and start everything
docker compose build
docker compose up -d

# Check health
make health

# View logs if something is wrong
make logs
```

## Continuous Integration

GitHub Actions runs the following on every push and pull request:

```yaml
jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - name: Run tests
        run: cd backend && go test -v -cover ./...

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
      - run: cd frontend && npm ci
      - run: cd frontend && npm test -- --run

  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build and test with Docker Compose
        run: docker compose build && docker compose up -d
```

## Coverage Goals

| Area | Target |
|------|--------|
| Business logic (services) | > 80% |
| UI components and hooks | > 70% |
| Authentication and authorization | 100% |
| Data mutations | 100% |

## Best Practices

1. **Test user behavior** — focus on what users see and do, not implementation details
2. **Semantic queries** — prefer `getByRole`, `getByLabelText`, `getByText` over test IDs
3. **Async testing** — use `waitFor` for async state updates
4. **Mock at boundaries** — mock API calls and WebSocket connections, not internal logic
5. **AAA pattern** — Arrange, Act, Assert
6. **Isolated tests** — each test should set up and tear down its own state
7. **Descriptive names** — test names should read like sentences

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Flaky tests | Use proper `waitFor` instead of `setTimeout` |
| Test pollution | Reset mocks and state between tests |
| Slow tests | Run packages in parallel with `go test -parallel 4 ./...` |
| Import errors in frontend | Check path aliases in `vite.config.ts` and `tsconfig.json` |

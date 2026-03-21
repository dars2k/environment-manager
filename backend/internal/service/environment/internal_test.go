package environment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"app-env-manager/internal/ctxutil"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/health"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// -- minimal mocks for internal tests --

type mockEnvRepo struct{ mock.Mock }

func (m *mockEnvRepo) Create(ctx context.Context, env *entities.Environment) error {
	return m.Called(ctx, env).Error(0)
}
func (m *mockEnvRepo) GetByID(ctx context.Context, id string) (*entities.Environment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Environment), args.Error(1)
}
func (m *mockEnvRepo) GetByName(ctx context.Context, name string) (*entities.Environment, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Environment), args.Error(1)
}
func (m *mockEnvRepo) Update(ctx context.Context, id string, env *entities.Environment) error {
	return m.Called(ctx, id, env).Error(0)
}
func (m *mockEnvRepo) UpdateStatus(ctx context.Context, id string, status entities.Status) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *mockEnvRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockEnvRepo) List(ctx context.Context, filter interfaces.ListFilter) ([]*entities.Environment, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Environment), args.Error(1)
}
func (m *mockEnvRepo) Count(ctx context.Context, filter interfaces.ListFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

type mockAuditRepo struct{ mock.Mock }

func (m *mockAuditRepo) Create(ctx context.Context, log *entities.AuditLog) error {
	return m.Called(ctx, log).Error(0)
}
func (m *mockAuditRepo) GetByID(ctx context.Context, id string) (*entities.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.AuditLog), args.Error(1)
}
func (m *mockAuditRepo) List(ctx context.Context, filter interfaces.AuditLogFilter) ([]*entities.AuditLog, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.AuditLog), args.Error(1)
}
func (m *mockAuditRepo) Count(ctx context.Context, filter interfaces.AuditLogFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockAuditRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

type mockLogRepo struct{ mock.Mock }

func (m *mockLogRepo) Create(ctx context.Context, l *entities.Log) error {
	return m.Called(ctx, l).Error(0)
}
func (m *mockLogRepo) List(ctx context.Context, filter interfaces.LogFilter) ([]*entities.Log, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entities.Log), args.Get(1).(int64), args.Error(2)
}
func (m *mockLogRepo) GetByID(ctx context.Context, id primitive.ObjectID) (*entities.Log, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Log), args.Error(1)
}
func (m *mockLogRepo) DeleteOld(ctx context.Context, older time.Duration) (int64, error) {
	args := m.Called(ctx, older)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockLogRepo) GetEnvironmentLogs(ctx context.Context, envID primitive.ObjectID, limit int) ([]*entities.Log, error) {
	args := m.Called(ctx, envID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Log), args.Error(1)
}
func (m *mockLogRepo) Count(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func newInternalService(envRepo *mockEnvRepo, logRepo *mockLogRepo, auditRepo *mockAuditRepo) *Service {
	sshMgr := ssh.NewManager(ssh.Config{
		ConnectionTimeout: time.Second,
		CommandTimeout:    time.Second,
		MaxConnections:    1,
	})
	checker := health.NewChecker(time.Second)
	logSvc := log.NewService(logRepo)
	return NewService(envRepo, auditRepo, sshMgr, checker, logSvc, nil)
}

// These tests are in the 'environment' package (not _test) so they can access
// private functions directly.

func TestExtractJSONPath_SimpleKey(t *testing.T) {
	data := map[string]interface{}{
		"version": "1.2.3",
	}
	result, err := extractJSONPath(data, "$.version")
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3", result)
}

func TestExtractJSONPath_NestedKey(t *testing.T) {
	data := map[string]interface{}{
		"info": map[string]interface{}{
			"version": "2.0.0",
		},
	}
	result, err := extractJSONPath(data, "$.info.version")
	assert.NoError(t, err)
	assert.Equal(t, "2.0.0", result)
}

func TestExtractJSONPath_MissingKey(t *testing.T) {
	data := map[string]interface{}{
		"other": "value",
	}
	_, err := extractJSONPath(data, "$.missing")
	assert.Error(t, err)
}

func TestExtractJSONPath_ArrayWildcard(t *testing.T) {
	data := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"id": "v1"},
			map[string]interface{}{"id": "v2"},
		},
	}
	result, err := extractJSONPath(data, "$.items[*].id")
	assert.NoError(t, err)
	items, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{"v1", "v2"}, items)
}

func TestExtractJSONPath_ArrayWildcardNoProperty(t *testing.T) {
	data := map[string]interface{}{
		"items": []interface{}{"a", "b"},
	}
	result, err := extractJSONPath(data, "$.items[*]")
	assert.NoError(t, err)
	items, ok := result.([]interface{})
	assert.True(t, ok)
	assert.Len(t, items, 2)
}

func TestExtractJSONPath_ArrayWildcardNotArray(t *testing.T) {
	data := map[string]interface{}{
		"items": "not-an-array",
	}
	_, err := extractJSONPath(data, "$.items[*].id")
	assert.Error(t, err)
}

func TestExtractJSONPath_ArrayPathNotFound(t *testing.T) {
	data := map[string]interface{}{}
	_, err := extractJSONPath(data, "$.missing[*].id")
	assert.Error(t, err)
}

func TestExtractJSONPath_NonObjectNavigation(t *testing.T) {
	data := map[string]interface{}{
		"scalar": "value",
	}
	_, err := extractJSONPath(data, "$.scalar.nested")
	assert.Error(t, err)
}

func TestGetJSONPath_SimpleKey(t *testing.T) {
	data := map[string]interface{}{"key": "val"}
	result, ok := getJSONPath(data, "key")
	assert.True(t, ok)
	assert.Equal(t, "val", result)
}

func TestGetJSONPath_NestedKey(t *testing.T) {
	data := map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": 42,
		},
	}
	result, ok := getJSONPath(data, "outer.inner")
	assert.True(t, ok)
	assert.Equal(t, 42, result)
}

func TestGetJSONPath_Missing(t *testing.T) {
	data := map[string]interface{}{}
	_, ok := getJSONPath(data, "missing")
	assert.False(t, ok)
}

func TestGetJSONPath_NonObjectNavigation(t *testing.T) {
	data := map[string]interface{}{"a": "string"}
	_, ok := getJSONPath(data, "a.b")
	assert.False(t, ok)
}

func TestValidateURLStrict_ValidHTTPS(t *testing.T) {
	err := validateURLStrict("https://example.com/path")
	// May succeed or fail depending on DNS resolution; we just check it doesn't panic
	_ = err
}

func TestValidateURLStrict_BadScheme(t *testing.T) {
	err := validateURLStrict("ftp://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP and HTTPS")
}

func TestValidateURLStrict_NoHostname(t *testing.T) {
	err := validateURLStrict("http://")
	assert.Error(t, err)
}

func TestValidateURL_AllowedHost(t *testing.T) {
	// Build a minimal service with an allowed host
	svc := &Service{allowedHosts: []string{"mycontainer"}}

	// http is valid scheme for allowed host
	err := svc.validateURL("http://mycontainer/api")
	assert.NoError(t, err)

	// ftp scheme should still fail
	err = svc.validateURL("ftp://mycontainer/api")
	assert.Error(t, err)
}

func TestValidateURL_NotAllowedFallsThrough(t *testing.T) {
	svc := &Service{allowedHosts: []string{"trusted"}}

	// An unknown host will fall through to validateURLStrict which may fail DNS
	err := svc.validateURL("ftp://unknown-host/")
	assert.Error(t, err)
}

// ---- logEvent / logEventWithPayload ----

func TestLogEvent_SystemActor(t *testing.T) {
	envRepo := new(mockEnvRepo)
	logRepo := new(mockLogRepo)
	auditRepo := new(mockAuditRepo)
	svc := newInternalService(envRepo, logRepo, auditRepo)

	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "testenv",
	}

	auditRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil).Maybe()

	// logEvent is called with background context (no user)
	svc.logEvent(context.Background(), env, entities.EventTypeRestart, entities.SeverityInfo,
		"create", "environment created", map[string]interface{}{"key": "val"})

	// Give background goroutine time to complete
	time.Sleep(20 * time.Millisecond)
}

func TestLogEventWithPayload_UserActor(t *testing.T) {
	envRepo := new(mockEnvRepo)
	logRepo := new(mockLogRepo)
	auditRepo := new(mockAuditRepo)
	svc := newInternalService(envRepo, logRepo, auditRepo)

	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "testenv",
	}

	auditRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.AuditLog")).Return(nil).Maybe()

	// Context with user ID triggers the user actor branch
	ctx := ctxutil.WithUser(context.Background(), "user-123", "alice")
	svc.logEventWithPayload(ctx, env, entities.EventTypeConfigUpdate, entities.SeverityInfo,
		"update", "environment updated", map[string]interface{}{"change": "name"})

	time.Sleep(20 * time.Millisecond)
}

// ---- buildSSHTarget ----

func TestBuildSSHTarget_PasswordCredentials(t *testing.T) {
	svc := &Service{}

	env := &entities.Environment{
		Target: entities.Target{Host: "host.local", Port: 22},
		Credentials: entities.CredentialRef{
			Type:     "password",
			Username: "deploy",
		},
		Metadata: map[string]interface{}{
			"password": "s3cr3t",
		},
	}

	target, err := svc.buildSSHTarget(env)
	assert.NoError(t, err)
	assert.Equal(t, "host.local", target.Host)
	assert.Equal(t, 22, target.Port)
	assert.Equal(t, "deploy", target.Username)
	assert.Equal(t, "s3cr3t", target.Password)
}

func TestBuildSSHTarget_MissingPassword(t *testing.T) {
	svc := &Service{}

	env := &entities.Environment{
		Credentials: entities.CredentialRef{Type: "password"},
		Metadata:    map[string]interface{}{},
	}

	_, err := svc.buildSSHTarget(env)
	assert.Error(t, err)
}

func TestBuildSSHTarget_NoMetadata(t *testing.T) {
	svc := &Service{}

	env := &entities.Environment{
		Credentials: entities.CredentialRef{Type: "password"},
		Metadata:    nil,
	}

	_, err := svc.buildSSHTarget(env)
	assert.Error(t, err)
}

func TestBuildSSHTarget_KeyCredentials(t *testing.T) {
	svc := &Service{}

	env := &entities.Environment{
		Target: entities.Target{Host: "host.local", Port: 22},
		Credentials: entities.CredentialRef{
			Type:     "key",
			Username: "admin",
		},
		Metadata: map[string]interface{}{
			"privateKey": "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
		},
	}

	target, err := svc.buildSSHTarget(env)
	assert.NoError(t, err)
	assert.NotEmpty(t, target.PrivateKey)
}

func TestBuildSSHTarget_MissingKey(t *testing.T) {
	svc := &Service{}

	env := &entities.Environment{
		Credentials: entities.CredentialRef{Type: "key"},
		Metadata:    map[string]interface{}{},
	}

	_, err := svc.buildSSHTarget(env)
	assert.Error(t, err)
}

func TestBuildSSHTarget_UnsupportedType(t *testing.T) {
	svc := &Service{}

	env := &entities.Environment{
		Credentials: entities.CredentialRef{Type: "unknown"},
	}

	_, err := svc.buildSSHTarget(env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestBuildSSHTarget_InsecureSkipVerification(t *testing.T) {
	svc := &Service{}

	env := &entities.Environment{
		Credentials: entities.CredentialRef{Type: "password", Username: "u"},
		Metadata: map[string]interface{}{
			"password":                        "pass",
			"insecureSkipHostKeyVerification": true,
		},
	}

	target, err := svc.buildSSHTarget(env)
	assert.NoError(t, err)
	assert.True(t, target.InsecureSkipHostKeyVerify)
}

// ---- executeHTTPCommand ----

func TestExecuteHTTPCommand_EmptyURL(t *testing.T) {
	svc := &Service{allowedHosts: nil}
	msg, ok := svc.executeHTTPCommand(context.Background(), entities.CommandDetails{URL: ""})
	assert.False(t, ok)
	assert.Contains(t, msg, "URL is required")
}

func TestExecuteHTTPCommand_InvalidURL(t *testing.T) {
	svc := &Service{allowedHosts: nil}
	msg, ok := svc.executeHTTPCommand(context.Background(), entities.CommandDetails{URL: "ftp://bad.scheme"})
	assert.False(t, ok)
	assert.Contains(t, msg, "URL validation failed")
}

func TestExecuteHTTPCommand_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	// Allow the test server host
	svc := &Service{allowedHosts: []string{extractHost(srv.URL)}}
	cmd := entities.CommandDetails{
		URL:    srv.URL + "/restart",
		Method: "POST",
	}

	msg, ok := svc.executeHTTPCommand(context.Background(), cmd)
	assert.True(t, ok)
	assert.Contains(t, msg, "ok")
}

func TestExecuteHTTPCommand_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer srv.Close()

	svc := &Service{allowedHosts: []string{extractHost(srv.URL)}}
	cmd := entities.CommandDetails{URL: srv.URL, Method: "POST"}

	msg, ok := svc.executeHTTPCommand(context.Background(), cmd)
	assert.False(t, ok)
	assert.Contains(t, msg, "500")
}

func TestExecuteHTTPCommand_WithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	svc := &Service{allowedHosts: []string{extractHost(srv.URL)}}
	cmd := entities.CommandDetails{
		URL:    srv.URL,
		Method: "POST",
		Body:   map[string]interface{}{"force": true},
	}

	_, ok := svc.executeHTTPCommand(context.Background(), cmd)
	assert.True(t, ok)
}

// extractHost returns just the hostname (no port) from a URL like "http://127.0.0.1:PORT"
func extractHost(rawURL string) string {
	// The validateURL allowedHosts check uses parsedURL.Hostname() which strips port
	// httptest servers bind to 127.0.0.1
	return "127.0.0.1"
}

func TestBuildSSHTarget_HostKey(t *testing.T) {
	svc := &Service{}

	env := &entities.Environment{
		Credentials: entities.CredentialRef{Type: "password", Username: "u"},
		Metadata: map[string]interface{}{
			"password": "pass",
			"hostKey":  "ssh-rsa AAAA...",
		},
	}

	target, err := svc.buildSSHTarget(env)
	assert.NoError(t, err)
	assert.NotEmpty(t, target.HostKey)
}

func TestExecuteHTTPCommand_WithHeaders(t *testing.T) {
	var capturedHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc := &Service{allowedHosts: []string{extractHost(srv.URL)}}
	cmd := entities.CommandDetails{
		URL:     srv.URL,
		Method:  "POST",
		Headers: map[string]string{"X-API-Key": "secret123"},
	}

	_, ok := svc.executeHTTPCommand(context.Background(), cmd)
	assert.True(t, ok)
	assert.Equal(t, "secret123", capturedHeader)
}

func TestExecuteHTTPCommand_DefaultMethod(t *testing.T) {
	var capturedMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc := &Service{allowedHosts: []string{extractHost(srv.URL)}}
	// Empty method → defaults to POST
	cmd := entities.CommandDetails{URL: srv.URL, Method: ""}

	_, ok := svc.executeHTTPCommand(context.Background(), cmd)
	assert.True(t, ok)
	assert.Equal(t, "POST", capturedMethod)
}

func TestExecuteHTTPCommand_WithBodyAndContentType(t *testing.T) {
	var capturedContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	svc := &Service{allowedHosts: []string{extractHost(srv.URL)}}
	cmd := entities.CommandDetails{
		URL:    srv.URL,
		Method: "POST",
		Body:   map[string]interface{}{"key": "value"},
	}

	_, ok := svc.executeHTTPCommand(context.Background(), cmd)
	assert.True(t, ok)
	assert.Equal(t, "application/json", capturedContentType)
}

func TestValidateURLStrict_LoopbackBlocked(t *testing.T) {
	err := validateURLStrict("http://localhost/path")
	assert.Error(t, err)
}

func TestValidateURLStrict_EmptyScheme(t *testing.T) {
	err := validateURLStrict("://no-scheme")
	// Invalid URL
	assert.Error(t, err)
}

func TestValidateURLStrict_InvalidScheme(t *testing.T) {
	err := validateURLStrict("ftp://example.com/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only HTTP and HTTPS schemes are allowed")
}

func TestValidateURLStrict_EmptyHostname(t *testing.T) {
	err := validateURLStrict("http:///path")
	assert.Error(t, err)
}

func TestValidateURLStrict_PrivateIP(t *testing.T) {
	// 192.168.1.1 is a private network address
	err := validateURLStrict("http://192.168.1.1/path")
	assert.Error(t, err)
}

func TestValidateURLStrict_LoopbackIP(t *testing.T) {
	// Direct IP loopback
	err := validateURLStrict("http://127.0.0.1/path")
	assert.Error(t, err)
}

func TestValidateURLStrict_BlockedHostname(t *testing.T) {
	// metadata.google.internal is in the blocked list
	err := validateURLStrict("http://metadata.google.internal/path")
	// Will fail at DNS lookup or at blocked hostname check
	assert.Error(t, err)
}

func TestValidateURLStrict_AWSMetadata(t *testing.T) {
	// AWS metadata IP
	err := validateURLStrict("http://169.254.169.254/latest/meta-data/")
	assert.Error(t, err)
}

func TestExecuteHTTPCommand_InvalidMethod(t *testing.T) {
	// An invalid HTTP method (with space) causes http.NewRequestWithContext to fail
	svc := &Service{allowedHosts: []string{"127.0.0.1"}}
	cmd := entities.CommandDetails{
		URL:    "http://127.0.0.1/path", // passes SSRF check (allowed host)
		Method: "INVALID METHOD",        // invalid → request creation fails
	}

	msg, ok := svc.executeHTTPCommand(context.Background(), cmd)
	assert.False(t, ok)
	assert.Contains(t, msg, "Failed to create request")
}

func TestExecuteHTTPCommand_URLEmpty(t *testing.T) {
	svc := &Service{}
	cmd := entities.CommandDetails{URL: ""}

	msg, ok := svc.executeHTTPCommand(context.Background(), cmd)
	assert.False(t, ok)
	assert.Contains(t, msg, "URL is required")
}

func TestCheckHealth_HealthCheckerFails(t *testing.T) {
	// Use an invalid HTTP method to force buildRequest (and hence healthChecker.CheckHealth) to return an error.
	// http.NewRequestWithContext fails when the method contains invalid characters.
	envRepo := new(mockEnvRepo)
	logRepo := new(mockLogRepo)
	auditRepo := new(mockAuditRepo)
	svc := newInternalService(envRepo, logRepo, auditRepo)

	id := primitive.NewObjectID()
	env := &entities.Environment{
		ID:   id,
		Name: "fail-env",
		HealthCheck: entities.HealthCheckConfig{
			Enabled:  true,
			Endpoint: "http://example.com/health",
			Method:   "INVALID METHOD WITH SPACE", // causes http.NewRequestWithContext to fail
		},
		Status: entities.Status{
			Health: entities.HealthStatusUnknown,
		},
	}

	envRepo.On("GetByID", mock.Anything, id.Hex()).Return(env, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()

	// The call should fail because buildRequest returns an error
	err := svc.CheckHealth(context.Background(), id.Hex())
	// The health checker returns error → CheckHealth wraps it and returns error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}

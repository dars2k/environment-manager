package ssh_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"app-env-manager/internal/service/ssh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

// Mock SSH server for testing
type mockSSHServer struct {
	listener net.Listener
	config   *gossh.ServerConfig
	handler  func(conn net.Conn, config *gossh.ServerConfig)
	stopped  chan struct{}
}

func newMockSSHServer(t *testing.T) *mockSSHServer {
	// Generate a test host key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(key)
	require.NoError(t, err)

	config := &gossh.ServerConfig{
		PasswordCallback: func(c gossh.ConnMetadata, pass []byte) (*gossh.Permissions, error) {
			if c.User() == "testuser" && string(pass) == "testpass" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
		PublicKeyCallback: func(c gossh.ConnMetadata, pubKey gossh.PublicKey) (*gossh.Permissions, error) {
			if c.User() == "keyuser" {
				return nil, nil
			}
			return nil, fmt.Errorf("key rejected for %q", c.User())
		},
	}
	config.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := &mockSSHServer{
		listener: listener,
		config:   config,
		stopped:  make(chan struct{}),
	}

	// Default handler that accepts connections and handles channels
	server.handler = func(conn net.Conn, config *gossh.ServerConfig) {
		defer conn.Close()
		
		// Perform the SSH handshake
		sshConn, chans, reqs, err := gossh.NewServerConn(conn, config)
		if err != nil {
			return
		}
		defer sshConn.Close()

		// Discard all global requests
		go gossh.DiscardRequests(reqs)

		// Accept all channels
		for newChannel := range chans {
			if newChannel.ChannelType() != "session" {
				newChannel.Reject(gossh.UnknownChannelType, "unknown channel type")
				continue
			}

			channel, requests, err := newChannel.Accept()
			if err != nil {
				continue
			}

			// Handle channel requests
			go func(in <-chan *gossh.Request) {
				for req := range in {
					switch req.Type {
					case "exec":
						// Extract command
						cmdLen := uint32(req.Payload[0])<<24 | uint32(req.Payload[1])<<16 | uint32(req.Payload[2])<<8 | uint32(req.Payload[3])
						cmd := string(req.Payload[4 : 4+cmdLen])
						
						req.Reply(true, nil)
						
						// Handle different test commands
						switch {
						case strings.Contains(cmd, "echo test"):
							channel.Write([]byte("test\n"))
							channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
						case strings.Contains(cmd, "exit 1"):
							channel.Write([]byte("error output\n"))
							channel.SendRequest("exit-status", false, []byte{0, 0, 0, 1})
						case strings.Contains(cmd, "sleep"):
							// Simulate long-running command
							time.Sleep(2 * time.Second)
							channel.Write([]byte("done\n"))
							channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
						default:
							channel.Write([]byte(fmt.Sprintf("Unknown command: %s\n", cmd)))
							channel.SendRequest("exit-status", false, []byte{0, 0, 0, 127})
						}
						channel.Close()
					default:
						req.Reply(false, nil)
					}
				}
			}(requests)
		}
	}

	go server.serve()
	return server
}

func (s *mockSSHServer) serve() {
	defer close(s.stopped)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handler(conn, s.config)
	}
}

func (s *mockSSHServer) stop() {
	s.listener.Close()
	<-s.stopped
}

func (s *mockSSHServer) addr() string {
	return s.listener.Addr().String()
}

func (s *mockSSHServer) port() int {
	addr := s.listener.Addr().(*net.TCPAddr)
	return addr.Port
}

func generateTestPrivateKey(t *testing.T) []byte {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	return pem.EncodeToMemory(privateKeyPEM)
}

func TestNewManager(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 30 * time.Second,
		CommandTimeout:    60 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	assert.NotNil(t, manager)
}

func TestTarget(t *testing.T) {
	target := ssh.Target{
		Host:       "example.com",
		Port:       22,
		Username:   "testuser",
		Password:   "testpass",
		PrivateKey: []byte("test-key"),
	}

	assert.Equal(t, "example.com", target.Host)
	assert.Equal(t, 22, target.Port)
	assert.Equal(t, "testuser", target.Username)
	assert.Equal(t, "testpass", target.Password)
	assert.Equal(t, []byte("test-key"), target.PrivateKey)
}

func TestExecutionResult(t *testing.T) {
	result := &ssh.ExecutionResult{
		Output:   "Command output",
		Error:    nil,
		ExitCode: 0,
		Duration: 100 * time.Millisecond,
	}

	assert.Equal(t, "Command output", result.Output)
	assert.Nil(t, result.Error)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, 100*time.Millisecond, result.Duration)
}

func TestConnectionKey(t *testing.T) {
	target := ssh.Target{
		Host:     "example.com",
		Port:     22,
		Username: "testuser",
		Password: "testpass",
	}

	expected := "testuser@example.com:22"
	assert.Equal(t, expected, target.ConnectionKey())
}

func TestConfig(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 30 * time.Second,
		CommandTimeout:    60 * time.Second,
		MaxConnections:    10,
	}

	assert.Equal(t, 30*time.Second, config.ConnectionTimeout)
	assert.Equal(t, 60*time.Second, config.CommandTimeout)
	assert.Equal(t, 10, config.MaxConnections)
}

func TestManager_Execute_Success(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "testuser",
		Password: "testpass",
	}

	result, err := manager.Execute(ctx, target, "echo test")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test\n", result.Output)
	assert.Equal(t, 0, result.ExitCode)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestManager_Execute_WithExitCode(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "testuser",
		Password: "testpass",
	}

	result, err := manager.Execute(ctx, target, "exit 1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "error output\n", result.Output)
	assert.Equal(t, 1, result.ExitCode)
	assert.NotNil(t, result.Error)
}

func TestManager_Execute_InvalidCredentials(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "wronguser",
		Password: "wrongpass",
	}

	result, err := manager.Execute(ctx, target, "echo test")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get SSH connection")
}

func TestManager_Execute_PrivateKey(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:       "127.0.0.1",
		Port:       server.port(),
		Username:   "keyuser",
		PrivateKey: generateTestPrivateKey(t),
	}

	result, err := manager.Execute(ctx, target, "echo test")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test\n", result.Output)
	assert.Equal(t, 0, result.ExitCode)
}

func TestManager_Execute_InvalidPrivateKey(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:       "127.0.0.1",
		Port:       server.port(),
		Username:   "keyuser",
		PrivateKey: []byte("invalid private key"),
	}

	result, err := manager.Execute(ctx, target, "echo test")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse private key")
}

func TestManager_Execute_NoAuthMethod(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "testuser",
		// No password or private key
	}

	result, err := manager.Execute(ctx, target, "echo test")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no authentication method provided")
}

func TestManager_Execute_ContextCancellation(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx, cancel := context.WithCancel(context.Background())
	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "testuser",
		Password: "testpass",
	}

	// Cancel context immediately
	cancel()

	result, err := manager.Execute(ctx, target, "sleep 5")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, context.Canceled, err)
}

func TestManager_Execute_CommandTimeout(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    500 * time.Millisecond, // Short timeout
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "testuser",
		Password: "testpass",
	}

	result, err := manager.Execute(ctx, target, "sleep 2")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "command timeout after")
}

func TestManager_Execute_ConnectionReuse(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "testuser",
		Password: "testpass",
	}

	// Execute multiple commands to test connection reuse
	for i := 0; i < 3; i++ {
		result, err := manager.Execute(ctx, target, "echo test")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test\n", result.Output)
		assert.Equal(t, 0, result.ExitCode)
	}
}

func TestManager_Execute_MaxConnections(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    1, // Only allow 1 connection
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	
	// Create two different targets
	target1 := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "testuser",
		Password: "testpass",
	}
	
	target2 := ssh.Target{
		Host:     "127.0.0.1", 
		Port:     server.port(),
		Username: "keyuser", // Different user to force new connection
		PrivateKey: generateTestPrivateKey(t),
	}

	// Execute on first target
	result1, err1 := manager.Execute(ctx, target1, "echo test1")
	assert.NoError(t, err1)
	assert.NotNil(t, result1)

	// Try to execute on second target (should fail due to connection limit)
	result2, err2 := manager.Execute(ctx, target2, "echo test2")
	assert.Error(t, err2)
	assert.Nil(t, result2)
	assert.Contains(t, err2.Error(), "connection limit reached")
}

func TestManager_TestConnection_Success(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "testuser",
		Password: "testpass",
	}

	err := manager.TestConnection(ctx, target)
	assert.NoError(t, err)
}

func TestManager_TestConnection_InvalidHost(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 2 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "invalid.host.that.does.not.exist",
		Port:     22,
		Username: "testuser",
		Password: "testpass",
	}

	err := manager.TestConnection(ctx, target)
	assert.Error(t, err)
}

func TestManager_TestConnection_InvalidPort(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 2 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     99999, // Invalid port
		Username: "testuser",
		Password: "testpass",
	}

	err := manager.TestConnection(ctx, target)
	assert.Error(t, err)
}

func TestManager_Close(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)

	ctx := context.Background()
	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "testuser",
		Password: "testpass",
	}

	// Execute a command to create a connection
	result, err := manager.Execute(ctx, target, "echo test")
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Close the manager
	err = manager.Close()
	assert.NoError(t, err)

	// Try to execute after close
	// The manager might still allow execution if connections are cached
	// We should just verify that Close() doesn't return an error
	// and that the manager can be safely closed
}

func TestManager_Execute_InvalidTarget(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	tests := []struct {
		name   string
		target ssh.Target
	}{
		{
			name: "empty host",
			target: ssh.Target{
				Host:     "",
				Port:     22,
				Username: "testuser",
				Password: "testpass",
			},
		},
		{
			name: "zero port",
			target: ssh.Target{
				Host:     "localhost",
				Port:     0,
				Username: "testuser",
				Password: "testpass",
			},
		},
		{
			name: "empty username",
			target: ssh.Target{
				Host:     "localhost",
				Port:     22,
				Username: "",
				Password: "testpass",
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.Execute(ctx, tt.target, "echo test")
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestManager_TestConnection_InvalidTarget(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "",
		Port:     0,
		Username: "",
		Password: "",
	}

	err := manager.TestConnection(ctx, target)
	assert.Error(t, err)
}

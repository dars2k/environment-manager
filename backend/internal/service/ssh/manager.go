package ssh

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Manager manages SSH connections with pooling
type Manager struct {
	connections map[string]*Connection
	mu          sync.RWMutex
	config      Config
}

// Connection represents an SSH connection
type Connection struct {
	client   *ssh.Client
	lastUsed time.Time
	refCount int
}

// Config contains SSH manager configuration
type Config struct {
	ConnectionTimeout time.Duration
	CommandTimeout    time.Duration
	MaxConnections    int
}

// Target represents an SSH connection target
type Target struct {
	Host     string
	Port     int
	Username string
	Password string
	PrivateKey []byte
}

// ExecutionResult contains the result of an SSH command
type ExecutionResult struct {
	Output   string
	ExitCode int
	Duration time.Duration
	Error    error
}

// NewManager creates a new SSH manager
func NewManager(config Config) *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
		config:      config,
	}
}

// Execute executes a command on a remote host
func (m *Manager) Execute(ctx context.Context, target Target, command string) (*ExecutionResult, error) {
	start := time.Now()
	
	// Get or create connection
	conn, err := m.getConnection(target)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH connection: %w", err)
	}
	defer m.releaseConnection(target)

	// Create session
	session, err := conn.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	// Execute command with timeout
	done := make(chan error, 1)
	var output []byte
	
	go func() {
		output, err = session.CombinedOutput(command)
		done <- err
	}()

	select {
	case err := <-done:
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				exitCode = exitErr.ExitStatus()
			} else {
				exitCode = -1
			}
		}
		
		return &ExecutionResult{
			Output:   string(output),
			ExitCode: exitCode,
			Duration: time.Since(start),
			Error:    err,
		}, nil
		
	case <-ctx.Done():
		return nil, ctx.Err()
		
	case <-time.After(m.config.CommandTimeout):
		return nil, fmt.Errorf("command timeout after %v", m.config.CommandTimeout)
	}
}

// TestConnection tests if an SSH connection can be established
func (m *Manager) TestConnection(ctx context.Context, target Target) error {
	// Try to establish connection
	client, err := m.createSSHClient(target)
	if err != nil {
		return err
	}
	defer client.Close()
	
	// Create a session to verify the connection works
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()
	
	// Run a simple command
	if err := session.Run("echo test"); err != nil {
		return fmt.Errorf("failed to run test command: %w", err)
	}
	
	return nil
}

// getConnection gets or creates a connection
func (m *Manager) getConnection(target Target) (*Connection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := target.ConnectionKey()
	
	// Check existing connection
	if conn, exists := m.connections[key]; exists {
		if m.isConnectionAlive(conn) {
			conn.refCount++
			conn.lastUsed = time.Now()
			return conn, nil
		}
		// Remove dead connection
		delete(m.connections, key)
	}

	// Check connection limit
	if len(m.connections) >= m.config.MaxConnections {
		// Try to clean up old connections
		m.cleanupConnections()
		if len(m.connections) >= m.config.MaxConnections {
			return nil, fmt.Errorf("connection limit reached")
		}
	}

	// Create new connection
	client, err := m.createSSHClient(target)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		client:   client,
		lastUsed: time.Now(),
		refCount: 1,
	}
	
	m.connections[key] = conn
	return conn, nil
}

// releaseConnection releases a connection
func (m *Manager) releaseConnection(target Target) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := target.ConnectionKey()
	if conn, exists := m.connections[key]; exists {
		conn.refCount--
		if conn.refCount <= 0 {
			// Keep connection alive for reuse
			conn.lastUsed = time.Now()
		}
	}
}

// createSSHClient creates a new SSH client
func (m *Manager) createSSHClient(target Target) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            target.Username,
		Timeout:         m.config.ConnectionTimeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
	}

	// Configure authentication
	if target.PrivateKey != nil {
		signer, err := ssh.ParsePrivateKey(target.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if target.Password != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(target.Password)}
	} else {
		return nil, fmt.Errorf("no authentication method provided")
	}

	// Connect
	addr := fmt.Sprintf("%s:%d", target.Host, target.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return client, nil
}

// isConnectionAlive checks if a connection is still alive
func (m *Manager) isConnectionAlive(conn *Connection) bool {
	// Create a session to test the connection
	session, err := conn.client.NewSession()
	if err != nil {
		return false
	}
	session.Close()
	return true
}

// cleanupConnections removes old unused connections
func (m *Manager) cleanupConnections() {
	idleTimeout := 5 * time.Minute
	now := time.Now()

	for key, conn := range m.connections {
		if conn.refCount == 0 && now.Sub(conn.lastUsed) > idleTimeout {
			conn.client.Close()
			delete(m.connections, key)
		}
	}
}

// Close closes all connections
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.connections {
		conn.client.Close()
	}
	m.connections = make(map[string]*Connection)
	
	return nil
}

// ConnectionKey returns a unique key for the connection
func (t Target) ConnectionKey() string {
	return fmt.Sprintf("%s@%s:%d", t.Username, t.Host, t.Port)
}

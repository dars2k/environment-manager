package ssh_test

import (
	"context"
	"testing"
	"time"

	"app-env-manager/internal/service/ssh"
	"github.com/stretchr/testify/assert"
	gossh "golang.org/x/crypto/ssh"
)

func TestManager_Execute_InvalidCommand(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "localhost",
		Port:     22,
		Username: "test",
		Password: "password",
	}

	result, err := manager.Execute(ctx, target, "")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "command cannot be empty")
}

func TestManager_Execute_HostKeyMismatch(t *testing.T) {
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

	// Use a valid but different host key
	otherKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC8N+76z5... test@test.com"

	target := ssh.Target{
		Host:     "127.0.0.1",
		Port:     server.port(),
		Username: "testuser",
		Password: "testpass",
		HostKey:  []byte(otherKey),
	}

	result, err := manager.Execute(ctx, target, "echo test")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestManager_CleanupConnections(t *testing.T) {
	server := newMockSSHServer(t)
	defer server.stop()

	// Get the host key from the server
	hostKey := server.hostKey.PublicKey()
	hostKeyBytes := gossh.MarshalAuthorizedKey(hostKey)

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
		HostKey:  hostKeyBytes,
	}

	// Create a connection
	_, err := manager.Execute(ctx, target, "echo test")
	assert.NoError(t, err)

	// Close the server to make the connection dead
	server.stop()

	err = manager.Close()
	assert.NoError(t, err)
}

func TestManager_Execute_InsecureSkipHostKeyVerify(t *testing.T) {
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
        InsecureSkipHostKeyVerify: true,
    }

    result, err := manager.Execute(ctx, target, "echo test")
    assert.NoError(t, err)
    assert.Equal(t, "test\n", result.Output)
}

func TestManager_Execute_NullByteCommand(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "localhost",
		Port:     22,
		Username: "test",
		Password: "password",
	}

	result, err := manager.Execute(ctx, target, "echo \x00 test")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "null bytes")
}

func TestManager_Execute_NewlineCommand(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "localhost",
		Port:     22,
		Username: "test",
		Password: "password",
	}

	result, err := manager.Execute(ctx, target, "echo test\nrm -rf /")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "newline characters")
}

func TestManager_Execute_DangerousCharCommand(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "localhost",
		Port:     22,
		Username: "test",
		Password: "password",
	}

	result, err := manager.Execute(ctx, target, "echo test; rm -rf /")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "potentially dangerous character")
}

func TestManager_Execute_ParenthesesCommand(t *testing.T) {
	config := ssh.Config{
		ConnectionTimeout: 5 * time.Second,
		CommandTimeout:    10 * time.Second,
		MaxConnections:    10,
	}
	manager := ssh.NewManager(config)
	defer manager.Close()

	ctx := context.Background()
	target := ssh.Target{
		Host:     "localhost",
		Port:     22,
		Username: "test",
		Password: "password",
	}

	result, err := manager.Execute(ctx, target, "echo (test)")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "potentially dangerous character")
}

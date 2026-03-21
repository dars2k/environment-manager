package ssh

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCommand_Valid(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"simple command", "ls -la"},
		{"command with flags", "systemctl status app"},
		{"command with path", "/usr/bin/myapp start"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, validateCommand(tt.command))
		})
	}
}

func TestValidateCommand_Empty(t *testing.T) {
	assert.Error(t, validateCommand(""))
	assert.Error(t, validateCommand("   "))
}

func TestValidateCommand_NullByte(t *testing.T) {
	err := validateCommand("cmd\x00injection")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "null bytes")
}

func TestValidateCommand_Newline(t *testing.T) {
	assert.Error(t, validateCommand("cmd\ninjection"))
	assert.Error(t, validateCommand("cmd\rinjection"))
}

func TestValidateCommand_DangerousChars(t *testing.T) {
	dangerous := []string{
		"cmd; rm -rf /",
		"cmd && evil",
		"cmd || evil",
		"cmd | grep",
		"cmd `whoami`",
		"cmd $USER",
		"cmd < input",
		"cmd > output",
		"cmd &",
	}
	for _, cmd := range dangerous {
		err := validateCommand(cmd)
		assert.Error(t, err, "expected error for: %s", cmd)
	}
}

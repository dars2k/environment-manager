package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "dot escaping",
			input:    "v1.2.3",
			expected: `v1\.2\.3`,
		},
		{
			name:     "special chars",
			input:    `a+b*c?d^e$f(g)h[i]j{k}l|m`,
			expected: `a\+b\*c\?d\^e\$f\(g\)h\[i\]j\{k\}l\|m`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeRegex(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

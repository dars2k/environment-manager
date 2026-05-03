package environment

// Additional internal tests to cover validateURLStrict IP edge cases.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateURLStrict_Multicast(t *testing.T) {
	// 224.0.0.2 is a multicast address (224.0.0.0/4 range)
	err := validateURLStrict("http://224.0.0.2/path")
	assert.Error(t, err)
}

func TestValidateURLStrict_LinkLocalUnicast(t *testing.T) {
	// 169.254.1.1 is link-local unicast (169.254.0.0/16), not in blocked hostnames list
	err := validateURLStrict("http://169.254.1.1/path")
	assert.Error(t, err)
}

func TestValidateURLStrict_BlockedHostname_Metadata(t *testing.T) {
	// "metadata" is in the blocked hostnames list
	// Will fail at DNS lookup or blocked hostname check
	err := validateURLStrict("http://metadata/path")
	assert.Error(t, err)
}

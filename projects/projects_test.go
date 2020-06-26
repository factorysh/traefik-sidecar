package projects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackendName(t *testing.T) {
	b, err := cutServerName("server-factory-octoback-octoback-1-94d7ce0ad04d-702c4d92a31ff8a6a1fbb1b2c3bd234f")
	assert.NoError(t, err)
	assert.Equal(t, NormalizeName("factory-octoback_octoback_1_94d7ce0ad04d"), b)
}

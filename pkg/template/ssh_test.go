package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSSHKeypair(t *testing.T) {
	privateKey, authorizedKey, err := newSSHKeypair()
	require.NoError(t, err)

	assert.NotEmpty(t, privateKey)
	assert.NotEmpty(t, authorizedKey)
}

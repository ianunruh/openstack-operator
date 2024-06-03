package template

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMustGenerateFernetKey(t *testing.T) {
	require.NotEmpty(t, MustGenerateFernetKey())
}

func TestMustGeneratePassword(t *testing.T) {
	require.Len(t, MustGeneratePassword(), 32)
}

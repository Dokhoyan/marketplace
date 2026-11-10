package password_test

import (
	"strings"
	"testing"

	"github.com/Dokhoyan/marketplace/internal/password"
	"github.com/stretchr/testify/require"
)

func TestPasswordVerification(t *testing.T) {
	original := strings.Repeat("я", 128)
	encoded, err := password.Hash(original)
	require.NoError(t, err)
	require.True(t, password.Verify(original, encoded))
	require.False(t, password.Verify(strings.Repeat("я", 127)+"ю", encoded))
	for _, malformed := range []string{
		"", "plain text", "$argon2id$v=19$m=0,t=2,p=0$x$x",
		"$argon2id$v=19$m=999999999,t=2,p=1$x$x",
		"$argon2id$v=19$m=19456,t=0,p=1$x$x",
		"$argon2id$v=19$m=19456,t=2,p=1$bad-salt$bad-hash",
	} {
		require.False(t, password.Verify(original, malformed))
	}
}

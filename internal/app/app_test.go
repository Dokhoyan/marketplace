package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Dokhoyan/common/pkg/closer"
	"github.com/stretchr/testify/require"
)

func TestAppClosesResources(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run the application lifecycle test")
	}
	t.Setenv("PG_DSN", dsn)
	t.Setenv("HTTP_HOST", "127.0.0.1")
	t.Setenv("HTTP_PORT", "0")
	t.Chdir(t.TempDir())
	configPath = filepath.Join(t.TempDir(), ".env")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a, err := NewApp(ctx)
	require.NoError(t, err)
	require.NoError(t, a.serviceProvider.dbPool.Ping(ctx))

	closed := false
	closer.Add(func() error {
		closed = true
		return nil
	})
	cancel()
	require.NoError(t, a.Run(ctx))
	require.True(t, closed)
	require.Error(t, a.serviceProvider.dbPool.Ping(context.Background()))
}

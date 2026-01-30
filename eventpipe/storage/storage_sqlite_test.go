package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/stretchr/testify/require"
)

func TestSqliteStorage_InitErrors(t *testing.T) {
	t.Run("missing_path", func(t *testing.T) {
		st := &SqliteStorage{}
		opts := map[string]any{
			"timeout": time.Second * 5,
		}

		err := st.Init(context.Background(), opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "need path parameter")
	})

	t.Run("invalid_timeout", func(t *testing.T) {
		tempDir := t.TempDir()
		st := &SqliteStorage{}
		opts := map[string]any{
			"path":    filepath.Join(tempDir, "test.st"),
			"timeout": "invalid",
		}

		err := st.Init(context.Background(), opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse timeout")
	})
}

func TestSqliteStorage_Drop(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_drop.st")

	st := &SqliteStorage{}
	opts := map[string]any{
		"path":    dbPath,
		"timeout": time.Second * 5,
	}

	ctx := context.Background()

	// Initialize and add some data
	err := st.Init(ctx, opts)
	require.NoError(t, err)

	err = st.UpdateCursor("test-source", "test-cursor")
	require.NoError(t, err)

	// Drop the database
	err = st.Drop()
	require.NoError(t, err)

	// Verify files are deleted
	require.NoFileExists(t, dbPath)
	require.NoFileExists(t, dbPath+"-wal")
	require.NoFileExists(t, dbPath+"-shm")
}

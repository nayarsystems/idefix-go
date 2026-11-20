package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // SQLite driver
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

// connect only warns when a PRAGMA fails, so check the driver actually applies them
func TestSqliteStorage_Pragmas(t *testing.T) {
	st := &SqliteStorage{}
	err := st.Init(context.Background(), map[string]any{
		"path":    filepath.Join(t.TempDir(), "test_pragmas.st"),
		"timeout": "5s",
	})
	require.NoError(t, err)
	defer st.Close()

	var journalMode string
	require.NoError(t, st.st.QueryRow("PRAGMA journal_mode").Scan(&journalMode))
	require.Equal(t, "wal", journalMode)

	var synchronous, foreignKeys, busyTimeout int
	require.NoError(t, st.st.QueryRow("PRAGMA synchronous").Scan(&synchronous))
	require.Equal(t, 2, synchronous) // FULL
	require.NoError(t, st.st.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys))
	require.Equal(t, 1, foreignKeys)
	require.NoError(t, st.st.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout))
	require.Equal(t, 5000, busyTimeout)
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

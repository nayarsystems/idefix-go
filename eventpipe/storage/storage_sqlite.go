package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/jaracil/ei"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteStorage struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	dbPath     string
	st         *sql.DB
	timeout    time.Duration
}

func (st *SqliteStorage) Init(ctx context.Context, opts any) error {
	st.ctx, st.cancelFunc = context.WithCancel(ctx)

	st.dbPath = ei.N(opts).M("path").StringZ()
	if st.dbPath == "" {
		return fmt.Errorf("need path parameter to connect to SQLite")
	}

	// Ensure directory exists
	dir := filepath.Dir(st.dbPath)
	if err := ensureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	dbTimeoutStr := ei.N(opts).M("timeout").StringZ()
	if dbTimeoutStr == "" {
		dbTimeoutStr = "60s"
	}
	dbTimeout, err := time.ParseDuration(dbTimeoutStr)
	if err != nil {
		return fmt.Errorf("failed to parse timeout duration %s: %w", dbTimeout, err)
	}

	st.timeout = dbTimeout

	err = st.connect()
	if err != nil {
		return err
	}

	err = st.setupSchema()
	if err != nil {
		return err
	}

	return nil
}

func (st *SqliteStorage) connect() error {
	var err error

	// Simple SQLite connection string
	dsn := fmt.Sprintf("file:%s", st.dbPath)

	st.st, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Set connection pool parameters
	st.st.SetMaxOpenConns(1) // SQLite works best with single connection for writes
	st.st.SetMaxIdleConns(1)
	st.st.SetConnMaxLifetime(time.Hour)

	// Test connection and apply PRAGMA settings
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	// Set query timeout via busy_timeout PRAGMA
	if st.timeout > 0 {
		timeoutMs := int(st.timeout.Milliseconds())
		timeoutPragma := fmt.Sprintf("PRAGMA busy_timeout = %d", timeoutMs)
		if _, err := st.st.ExecContext(ctx, timeoutPragma); err != nil {
			slog.Warn("failed to set busy timeout", "error", err)
		}
	}

	if err := st.st.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	// Apply robustness settings via explicit PRAGMAs
	pragmas := []string{
		// Core robustness settings
		"PRAGMA journal_mode = WAL",  // WAL mode for concurrency and corruption resistance
		"PRAGMA synchronous = FULL",  // Maximum safety against corruption
		"PRAGMA foreign_keys = ON",   // Referential integrity
		"PRAGMA temp_store = MEMORY", // Temporary tables in memory
		"PRAGMA auto_vacuum = FULL",  // Full auto-vacuum (takes effect after VACUUM)

		// Performance and maintenance settings
		"PRAGMA wal_autocheckpoint = 100", // Checkpoint WAL every 100 pages
		"PRAGMA secure_delete = OFF",      // Disable securely delete data

		// Maintenance
		"PRAGMA optimize",        // Optimize query planner
		"PRAGMA integrity_check", // Check database integrity on startup
	}

	for _, pragma := range pragmas {
		if _, err := st.st.ExecContext(ctx, pragma); err != nil {
			slog.Warn("failed to execute pragma", "pragma", pragma, "error", err)
		}
	}

	// Run VACUUM to apply auto_vacuum setting if database already exists
	// This is safe and ensures auto_vacuum takes effect
	if _, err := st.st.ExecContext(ctx, "VACUUM"); err != nil {
		slog.Warn("failed to vacuum database for auto_vacuum", "error", err)
	}

	slog.Info("connected to SQLite database", "path", st.dbPath)
	return nil
}

func (st *SqliteStorage) setupSchema() error {
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	// Create sources table if it does not exist
	sourcesSchema := `
	   CREATE TABLE IF NOT EXISTS sources (
		   id TEXT PRIMARY KEY,
		   cursor TEXT NOT NULL
	   )`

	if _, err := st.st.ExecContext(ctx, sourcesSchema); err != nil {
		return fmt.Errorf("failed to create sources table: %w", err)
	}

	// Create items table if it does not exist (source_id + id is composite primary key)
	itemsSchema := `
	   CREATE TABLE IF NOT EXISTS items (
		   source_id TEXT NOT NULL,
		   id TEXT NOT NULL,
		   data BLOB NOT NULL,
		   context BLOB,
		   locked BOOLEAN NOT NULL DEFAULT 0,
		   order_index INTEGER NOT NULL DEFAULT 0,
		   PRIMARY KEY (source_id, id),
		   FOREIGN KEY (source_id) REFERENCES sources(id) ON DELETE CASCADE
	   )`

	if _, err := st.st.ExecContext(ctx, itemsSchema); err != nil {
		return fmt.Errorf("failed to create items table: %w", err)
	}

	// --- Schema versioning using PRAGMA user_version ---
	const schemaVersion = 1 // Increment this value when you change the schema
	var userVersion int
	err := st.st.QueryRowContext(ctx, "PRAGMA user_version").Scan(&userVersion)
	if err != nil {
		return fmt.Errorf("failed to get user_version: %w", err)
	}

	if userVersion < schemaVersion {
		// Only run ALTER TABLE if the schema version is outdated
		addLockedColumn := `ALTER TABLE items ADD COLUMN locked BOOLEAN NOT NULL DEFAULT 0`
		if _, err := st.st.ExecContext(ctx, addLockedColumn); err != nil {
			// Ignore error if the column already exists
			slog.Debug("locked column may already exist", "error", err)
		}
		// Update the schema version
		if _, err := st.st.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", schemaVersion)); err != nil {
			slog.Warn("failed to update user_version", "error", err)
		}
	}

	// Create index for efficient ordering queries
	orderIndexSchema := `CREATE INDEX IF NOT EXISTS idx_items_order ON items(source_id, order_index)`
	if _, err := st.st.ExecContext(ctx, orderIndexSchema); err != nil {
		slog.Warn("failed to create order index", "error", err)
	}

	return nil
}

func (st *SqliteStorage) Close() error {
	if st.cancelFunc != nil {
		st.cancelFunc()
	}

	if st.st != nil {
		// Run optimize before closing for better performance on next open
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if _, err := st.st.ExecContext(ctx, "PRAGMA optimize"); err != nil {
			slog.Warn("failed to optimize on close", "error", err)
		}

		// Checkpoint WAL to consolidate changes
		if _, err := st.st.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
			slog.Warn("failed to checkpoint WAL on close", "error", err)
		}

		return st.st.Close()
	}

	return nil
}

func (st *SqliteStorage) Drop() error {
	// Use a separate context with timeout for cleanup operations
	ctx, cancel := context.WithTimeout(context.Background(), st.timeout)
	defer cancel()

	// First close the database connection properly
	if st.st != nil {
		// Run optimize and checkpoint before closing
		if _, err := st.st.ExecContext(ctx, "PRAGMA optimize"); err != nil {
			slog.Warn("failed to optimize before drop", "error", err)
		}

		// Checkpoint WAL to consolidate changes
		if _, err := st.st.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
			slog.Warn("failed to checkpoint WAL before drop", "error", err)
		}

		// Close the database connection
		if err := st.st.Close(); err != nil {
			slog.Warn("failed to close database before drop", "error", err)
		}
		st.st = nil
	}

	// Cancel the database context
	if st.cancelFunc != nil {
		st.cancelFunc()
	}

	// Delete the database file and associated files
	if st.dbPath != "" {
		// Remove main database file
		if err := os.Remove(st.dbPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove database file %s: %w", st.dbPath, err)
		}

		// Remove WAL file if exists
		walPath := st.dbPath + "-wal"
		if err := os.Remove(walPath); err != nil && !os.IsNotExist(err) {
			slog.Warn("failed to remove WAL file", "path", walPath, "error", err)
		}

		// Remove SHM file if exists
		shmPath := st.dbPath + "-shm"
		if err := os.Remove(shmPath); err != nil && !os.IsNotExist(err) {
			slog.Warn("failed to remove SHM file", "path", shmPath, "error", err)
		}

		slog.Info("SQLite database dropped successfully", "path", st.dbPath)
	}

	return nil
}

func (st *SqliteStorage) UpdateCursor(id, cursor string) error {
	slog.Debug("updating source cursor", "id", id, "cursor", cursor)
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	query := `UPDATE sources SET cursor = ? WHERE id = ?`

	res, err := st.st.ExecContext(ctx, query, cursor, id)
	if err != nil {
		return fmt.Errorf("failed to update source cursor: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected when updating cursor: %w", err)
	}

	if rowsAffected == 0 {
		// Source does not exist, insert it
		insertQuery := `INSERT INTO sources (id, cursor) VALUES (?, ?)`
		_, err := st.st.ExecContext(ctx, insertQuery, id, cursor)
		if err != nil {
			return fmt.Errorf("failed to insert new source with cursor: %w", err)
		}
	}

	return nil
}

func (st *SqliteStorage) GetCursor(id string) (string, error) {
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	query := "SELECT cursor FROM sources WHERE id = ?"
	var cursor string

	err := st.st.QueryRowContext(ctx, query, id).Scan(&cursor)
	if err != nil {
		return "", fmt.Errorf("failed to get source cursor: %w", err)
	}

	return cursor, nil
}

func (st *SqliteStorage) GetIndex(sourceId string) (uint64, error) {
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	query := "SELECT COALESCE(MAX(order_index), 0) FROM items WHERE source_id = ?"
	var maxIndex uint64

	err := st.st.QueryRowContext(ctx, query, sourceId).Scan(&maxIndex)
	if err != nil {
		return 0, fmt.Errorf("failed to get max order index: %w", err)
	}

	return maxIndex, nil
}

func (st *SqliteStorage) Push(item Item) error {
	itemBytes := item.Bytes()
	itemContext := item.ContextBytes()
	slog.Debug("pushing item", "id", item.Id(), "sourceId", item.SourceId())
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	// Start a transaction to ensure atomicity
	tx, err := st.st.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Ensure source exists (insert with empty cursor if not exists)
	sourceQuery := `INSERT OR IGNORE INTO sources (id, cursor) VALUES (?, '')`
	if _, err := tx.ExecContext(ctx, sourceQuery, item.SourceId()); err != nil {
		return fmt.Errorf("failed to ensure source exists: %w", err)
	}

	// Get the next index for this source
	indexQuery := "SELECT COALESCE(MAX(order_index), 0) + 1 FROM items WHERE source_id = ?"
	var nextIndex uint64
	if err := tx.QueryRowContext(ctx, indexQuery, item.SourceId()).Scan(&nextIndex); err != nil {
		return fmt.Errorf("failed to get next index: %w", err)
	}

	// Insert the item with auto-assigned index
	itemQuery := `INSERT INTO items (source_id, id, data, context, order_index) VALUES (?, ?, ?, ?, ?)`
	if _, err := tx.ExecContext(ctx, itemQuery, item.SourceId(), item.Id(), itemBytes, itemContext, nextIndex); err != nil {
		return fmt.Errorf("failed to push item: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (st *SqliteStorage) Update(item Item) error {
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	// Update the item (preserving existing order_index, will fail if doesn't exist)
	itemQuery := `UPDATE items SET data = ?, context = ? WHERE source_id = ? AND id = ?`
	result, err := st.st.ExecContext(ctx, itemQuery, item.Bytes(), item.ContextBytes(), item.SourceId(), item.Id())
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	// Check if any row was affected (i.e., if the item existed)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item with source_id '%s' and id '%s' not found", item.SourceId(), item.Id())
	}

	return nil
}

func (st *SqliteStorage) Lock(sourceId, itemId string) error {
	slog.Debug("locking item", "sourceId", sourceId, "itemId", itemId)
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	query := "UPDATE items SET locked = 1 WHERE source_id = ? AND id = ?"
	_, err := st.st.ExecContext(ctx, query, sourceId, itemId)
	if err != nil {
		return fmt.Errorf("failed to lock item: %w", err)
	}
	return nil
}

func (st *SqliteStorage) Unlock(sourceId, itemId string) error {
	slog.Debug("unlocking item", "sourceId", sourceId, "itemId", itemId)
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	query := "UPDATE items SET locked = 0 WHERE source_id = ? AND id = ?"
	_, err := st.st.ExecContext(ctx, query, sourceId, itemId)
	if err != nil {
		return fmt.Errorf("failed to unlock item: %w", err)
	}
	return nil
}

func (st *SqliteStorage) Delete(sourceId, itemId string) error {
	slog.Debug("deleting item", "sourceId", sourceId, "itemId", itemId)
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	query := "DELETE FROM items WHERE source_id = ? AND id = ?"
	_, err := st.st.ExecContext(ctx, query, sourceId, itemId)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

func (st *SqliteStorage) GetItems(sourceId string, limit int) ([]Item, error) {
	slog.Debug("getting items", "sourceId", sourceId, "limit", limit)
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	// Default limit if not specified
	if limit <= 0 {
		limit = 100
	}

	query := "SELECT source_id,id,data,context,order_index FROM items WHERE source_id = ? ORDER BY order_index LIMIT ?"
	rows, err := st.st.QueryContext(ctx, query, sourceId, limit)
	if err != nil {
		return []Item{}, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	// Initialize as empty slice to ensure we never return nil
	items := []Item{}
	for rows.Next() {
		var item sqliteItem
		if err := rows.Scan(&item.sourceId, &item.id, &item.data, &item.context, &item.orderIndex); err != nil {
			return []Item{}, fmt.Errorf("failed to scan item row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return []Item{}, fmt.Errorf("error iterating over item rows: %w", err)
	}

	return items, nil
}

func (st *SqliteStorage) GetLocked(sourceId string, limit int) ([]Item, error) {
	slog.Debug("getting locked items", "sourceId", sourceId, "limit", limit)
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	// Default limit if not specified
	if limit <= 0 {
		limit = 100
	}

	query := "SELECT source_id,id,data,context,order_index FROM items WHERE source_id = ? AND locked = 1 ORDER BY order_index LIMIT ?"
	rows, err := st.st.QueryContext(ctx, query, sourceId, limit)
	if err != nil {
		return []Item{}, fmt.Errorf("failed to query locked items: %w", err)
	}
	defer rows.Close()

	// Initialize as empty slice to ensure we never return nil
	items := []Item{}
	for rows.Next() {
		var item sqliteItem
		if err := rows.Scan(&item.sourceId, &item.id, &item.data, &item.context, &item.orderIndex); err != nil {
			return []Item{}, fmt.Errorf("failed to scan locked item row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return []Item{}, fmt.Errorf("error iterating over locked item rows: %w", err)
	}

	return items, nil
}

func (st *SqliteStorage) GetUnlocked(sourceId string, limit int) ([]Item, error) {
	slog.Debug("getting unlocked items", "sourceId", sourceId, "limit", limit)
	ctx, cancel := context.WithTimeout(st.ctx, st.timeout)
	defer cancel()

	// Default limit if not specified
	if limit <= 0 {
		limit = 100
	}

	query := "SELECT source_id,id,data,context,order_index FROM items WHERE source_id = ? AND locked = 0 ORDER BY order_index LIMIT ?"
	rows, err := st.st.QueryContext(ctx, query, sourceId, limit)
	if err != nil {
		return []Item{}, fmt.Errorf("failed to query unlocked items: %w", err)
	}
	defer rows.Close()

	// Initialize as empty slice to ensure we never return nil
	items := []Item{}
	for rows.Next() {
		var item sqliteItem
		if err := rows.Scan(&item.sourceId, &item.id, &item.data, &item.context, &item.orderIndex); err != nil {
			return []Item{}, fmt.Errorf("failed to scan unlocked item row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return []Item{}, fmt.Errorf("error iterating over unlocked item rows: %w", err)
	}

	return items, nil
}

func ensureDir(dir string) error {
	if dir == "" || dir == "." {
		return nil
	}

	return os.MkdirAll(dir, 0755)
}

type sqliteItem struct {
	sourceId   string
	id         string
	data       []byte
	context    []byte
	orderIndex uint64
}

func (si sqliteItem) SourceId() string {
	return si.sourceId
}

func (si sqliteItem) Id() string {
	return si.id
}

func (si sqliteItem) Bytes() []byte {
	return si.data
}

func (si sqliteItem) ContextBytes() []byte {
	return si.context
}

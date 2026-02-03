package storage

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Set default logger
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	m.Run()
}

func initSqliteStorageForTest(t *testing.T) Storage {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.st")
	st := &SqliteStorage{}
	opts := map[string]any{
		"path":    dbPath,
		"timeout": time.Second * 5,
	}

	ctx := context.Background()

	// Test initialization
	err := st.Init(ctx, opts)
	require.NoError(t, err)
	t.Cleanup(func() {
		st.Close()
	})
	return st
}

func TestStorage(t *testing.T) {
	dbs := map[string]Storage{
		"sqlite": initSqliteStorageForTest(t),
		"memory": NewMemoryStorage(),
	}

	for dbName, st := range dbs {
		t.Run(dbName, func(t *testing.T) {
			testBasic(t, st)
			testIndex(t, st)
			testPushUpdate(t, st)
			testLocked(t, st)
		})
	}
}

// Mock item for testing
type mockItem struct {
	sourceId string
	id       string
	data     []byte
	context  []byte
}

func (m mockItem) SourceId() string     { return m.sourceId }
func (m mockItem) Id() string           { return m.id }
func (m mockItem) Bytes() []byte        { return m.data }
func (m mockItem) ContextBytes() []byte { return m.context }

func testBasic(t *testing.T, st Storage) {

	// Test source cursor operations
	t.Run("source_cursor", func(t *testing.T) {
		// Test insert
		err := st.UpdateCursor("test-source", "test-cursor-123")
		require.NoError(t, err)

		// Test get
		cursor, err := st.GetCursor("test-source")
		require.NoError(t, err)
		require.Equal(t, "test-cursor-123", cursor)

		// Test update
		err = st.UpdateCursor("test-source", "test-cursor-456")
		require.NoError(t, err)

		cursor, err = st.GetCursor("test-source")
		require.NoError(t, err)
		require.Equal(t, "test-cursor-456", cursor)

		// Test not found (should return error)
		cursor, err = st.GetCursor("non-existent")
		require.Error(t, err)
		require.Equal(t, "", cursor)
	})

	// Test item operations
	t.Run("items", func(t *testing.T) {
		// Create test items
		item1 := mockItem{
			sourceId: "source1",
			id:       "item1",
			data:     []byte("data1"),
			context:  []byte("context1"),
		}
		item2 := mockItem{
			sourceId: "source1",
			id:       "item2",
			data:     []byte("data2"),
			context:  []byte("context2"),
		}
		item3 := mockItem{
			sourceId: "source2",
			id:       "item3",
			data:     []byte("data3"),
			context:  nil, // Test nil context
		}

		// Test push
		err := st.Push(item1)
		require.NoError(t, err)

		err = st.Push(item2)
		require.NoError(t, err)

		err = st.Push(item3)
		require.NoError(t, err)

		// Test get items by source
		items, err := st.GetItems("source1", 0) // 0 means default limit
		require.NoError(t, err)
		require.Len(t, items, 2)

		// Verify items are returned in push order
		require.Equal(t, "item1", items[0].Id())
		require.Equal(t, "item2", items[1].Id())

		// Verify item data
		for _, item := range items {
			require.Equal(t, "source1", item.SourceId())
			if item.Id() == "item1" {
				require.Equal(t, []byte("data1"), item.Bytes())
				require.Equal(t, []byte("context1"), item.ContextBytes())
			} else if item.Id() == "item2" {
				require.Equal(t, []byte("data2"), item.Bytes())
				require.Equal(t, []byte("context2"), item.ContextBytes())
			}
		}

		// Test update existing item
		item1Updated := mockItem{
			sourceId: "source1",
			id:       "item1",
			data:     []byte("updated_data1"),
			context:  []byte("updated_context1"),
		}
		err = st.Update(item1Updated)
		require.NoError(t, err)

		// Verify update
		items, err = st.GetItems("source1", 0)
		require.NoError(t, err)
		require.Len(t, items, 2)

		for _, item := range items {
			if item.Id() == "item1" {
				require.Equal(t, []byte("updated_data1"), item.Bytes())
				require.Equal(t, []byte("updated_context1"), item.ContextBytes())
			}
		}

		// Verify items are still in push order after update
		require.Equal(t, "item1", items[0].Id())
		require.Equal(t, "item2", items[1].Id())

		// Test delete item
		err = st.Delete("source1", "item1")
		require.NoError(t, err)

		items, err = st.GetItems("source1", 0)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, "item2", items[0].Id())

		// Test update source1 cursor after item operations
		err = st.UpdateCursor("source1", "cursor-after-items")
		require.NoError(t, err)

		cursor, err := st.GetCursor("source1")
		require.NoError(t, err)
		require.Equal(t, "cursor-after-items", cursor)

		// Test get items after cursor update
		items, err = st.GetItems("source1", 0)
		require.NoError(t, err)
		require.Len(t, items, 1) // item2 should still be there

		// Test get items for non-existent source
		items, err = st.GetItems("non-existent", 0)
		require.NoError(t, err)
		require.Len(t, items, 0)
		require.NotNil(t, items, "GetItems should never return nil, always empty slice")

		// Test delete non-existent item (should not error)
		err = st.Delete("source1", "non-existent")
		require.NoError(t, err)
	})

	// Test foreign key constraints with item insertion
	t.Run("foreign_key_auto_create", func(t *testing.T) {
		// Push item without pre-existing source (should auto-create source)
		item := mockItem{
			sourceId: "auto-created-source",
			id:       "item1",
			data:     []byte("data"),
			context:  []byte("context"),
		}

		err := st.Push(item)
		require.NoError(t, err)

		// Verify source was created with empty cursor
		cursor, err := st.GetCursor("auto-created-source")
		require.NoError(t, err)
		require.Equal(t, "", cursor)

		// Verify item was inserted
		items, err := st.GetItems("auto-created-source", 0)
		require.NoError(t, err)
		require.Len(t, items, 1)
	})
}

func testIndex(t *testing.T, st Storage) {
	t.Run("max_order_index", func(t *testing.T) {
		// Test empty source
		maxIndex, err := st.GetIndex("empty-source")
		require.NoError(t, err)
		require.Equal(t, uint64(0), maxIndex)

		// Push items in order
		item1 := mockItem{
			sourceId: "order-test",
			id:       "item1",
			data:     []byte("data1"),
		}
		item2 := mockItem{
			sourceId: "order-test",
			id:       "item2",
			data:     []byte("data2"),
		}
		item3 := mockItem{
			sourceId: "order-test",
			id:       "item3",
			data:     []byte("data3"),
		}

		err = st.Push(item1)
		require.NoError(t, err)
		err = st.Push(item2)
		require.NoError(t, err)
		err = st.Push(item3)
		require.NoError(t, err)

		// Get max order index
		maxIndex, err = st.GetIndex("order-test")
		require.NoError(t, err)
		require.Equal(t, uint64(3), maxIndex)
	})

	t.Run("ordering_and_limit", func(t *testing.T) {
		// Items were pushed in order: item1, item2, item3
		// Should be returned in same order

		// Test default limit
		items, err := st.GetItems("order-test", 0)
		require.NoError(t, err)
		require.Len(t, items, 3)

		// Verify IDs are in correct order based on push order
		require.Equal(t, "item1", items[0].Id()) // pushed first
		require.Equal(t, "item2", items[1].Id()) // pushed second
		require.Equal(t, "item3", items[2].Id()) // pushed third

		// Test with limit
		items, err = st.GetItems("order-test", 2)
		require.NoError(t, err)
		require.Len(t, items, 2)
		require.Equal(t, "item1", items[0].Id())
		require.Equal(t, "item2", items[1].Id())

		// Test with limit 1
		items, err = st.GetItems("order-test", 1)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, "item1", items[0].Id())
	})

	t.Run("multiple_source_ordering", func(t *testing.T) {
		// Verify push order is maintained for different sources
		sourceId := "order-test-2"

		item1 := mockItem{
			sourceId: sourceId,
			id:       "first-pushed",
			data:     []byte("data"),
		}
		item2 := mockItem{
			sourceId: sourceId,
			id:       "second-pushed",
			data:     []byte("data"),
		}
		item3 := mockItem{
			sourceId: sourceId,
			id:       "third-pushed",
			data:     []byte("data"),
		}

		err := st.Push(item1)
		require.NoError(t, err)
		err = st.Push(item2)
		require.NoError(t, err)
		err = st.Push(item3)
		require.NoError(t, err)

		// Should be returned in push order
		items, err := st.GetItems(sourceId, 0)
		require.NoError(t, err)
		require.Len(t, items, 3)

		require.Equal(t, "first-pushed", items[0].Id())
		require.Equal(t, "second-pushed", items[1].Id())
		require.Equal(t, "third-pushed", items[2].Id())
	})
}

func testPushUpdate(t *testing.T, st Storage) {
	// Test push and update operations
	t.Run("push_success", func(t *testing.T) {
		item := mockItem{
			sourceId: "test-source",
			id:       "test-item",
			data:     []byte("test-data"),
			context:  []byte("test-context"),
		}

		err := st.Push(item)
		require.NoError(t, err)

		// Verify item was pushed
		items, err := st.GetItems("test-source", 0)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, "test-item", items[0].Id())
		require.Equal(t, []byte("test-data"), items[0].Bytes())
	})

	t.Run("push_duplicate_succeeds", func(t *testing.T) {
		item := mockItem{
			sourceId: "test-source",
			id:       "test-item-2", // Different ID
			data:     []byte("different-data"),
		}

		err := st.Push(item)
		require.NoError(t, err)

		// Should now have 2 items
		items, err := st.GetItems("test-source", 0)
		require.NoError(t, err)
		require.Len(t, items, 2)
	})

	t.Run("update_success", func(t *testing.T) {
		item := mockItem{
			sourceId: "test-source",
			id:       "test-item",
			data:     []byte("updated-data"),
			context:  []byte("updated-context"),
		}

		err := st.Update(item)
		require.NoError(t, err)

		// Verify item was updated
		items, err := st.GetItems("test-source", 0)
		require.NoError(t, err)
		require.Len(t, items, 2)

		// Find the updated item
		var updatedItem Item
		for _, item := range items {
			if item.Id() == "test-item" {
				updatedItem = item
				break
			}
		}
		require.NotNil(t, updatedItem)
		require.Equal(t, []byte("updated-data"), updatedItem.Bytes())
		require.Equal(t, []byte("updated-context"), updatedItem.ContextBytes())
	})

	t.Run("update_nonexistent_fails", func(t *testing.T) {
		item := mockItem{
			sourceId: "test-source",
			id:       "nonexistent-item",
			data:     []byte("data"),
		}

		err := st.Update(item)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
}

func testLocked(t *testing.T, st Storage) {
	source := "locked-test"
	// Insert 6 items: A, B, C, D, E, F
	ids := []string{"A", "B", "C", "D", "E", "F"}
	for _, id := range ids {
		item := mockItem{sourceId: source, id: id, data: []byte("data-" + id)}
		err := st.Push(item)
		require.NoError(t, err)
	}

	// Get all items to verify insertion
	allItems, err := st.GetItems(source, 0)
	require.NoError(t, err)
	require.Equal(t, ids, extractIds(allItems))

	// Lock C and E
	err = st.Lock(source, "C")
	require.NoError(t, err)
	err = st.Lock(source, "E")
	require.NoError(t, err)

	// Step 3: GetLockedItems returns [C, E] in order
	locked, err := st.GetLocked(source, 0)
	require.NoError(t, err)
	require.Equal(t, []string{"C", "E"}, extractIds(locked))

	// Step 4: GetUnlockedItems returns [A, B, D, F] in order
	unlocked, err := st.GetUnlocked(source, 0)
	require.NoError(t, err)
	require.Equal(t, []string{"A", "B", "D", "F"}, extractIds(unlocked))

	// Step 5: Unlock C, check again
	err = st.Unlock(source, "C")
	require.NoError(t, err)
	locked, err = st.GetLocked(source, 0)
	require.NoError(t, err)
	require.Equal(t, []string{"E"}, extractIds(locked))
	unlocked, err = st.GetUnlocked(source, 0)
	require.NoError(t, err)
	require.Equal(t, []string{"A", "B", "C", "D", "F"}, extractIds(unlocked))

	// Step 6: Lock all, check
	for _, id := range ids {
		err := st.Lock(source, id)
		require.NoError(t, err)
	}
	locked, err = st.GetLocked(source, 0)
	require.NoError(t, err)
	require.Equal(t, ids, extractIds(locked))
	unlocked, err = st.GetUnlocked(source, 0)
	require.NoError(t, err)
	require.NotNil(t, unlocked)
	require.Empty(t, unlocked)

	// Step 7: Unlock all, check
	for _, id := range ids {
		err := st.Unlock(source, id)
		require.NoError(t, err)
	}
	locked, err = st.GetLocked(source, 0)
	require.NoError(t, err)
	require.NotNil(t, locked)
	require.Empty(t, locked)
	unlocked, err = st.GetUnlocked(source, 0)
	require.NoError(t, err)
	require.Equal(t, ids, extractIds(unlocked))
}

// Helper to extract IDs from []Item
func extractIds(items []Item) []string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.Id()
	}
	return ids
}

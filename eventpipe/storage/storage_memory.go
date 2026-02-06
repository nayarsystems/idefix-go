package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type memoryStorage struct {
	mu        sync.RWMutex
	items     map[string]map[string]*memoryItem // sourceId -> itemId -> item
	cursors   map[string]string                 // sourceId -> cursor
	nextIndex map[string]uint64                 // sourceId -> next order index
}

type memoryItem struct {
	sourceId   string
	id         string
	data       []byte
	context    []byte
	orderIndex uint64
	locked     bool
}

func NewMemoryStorage() Storage {
	return &memoryStorage{
		items:     make(map[string]map[string]*memoryItem),
		cursors:   make(map[string]string),
		nextIndex: make(map[string]uint64),
	}
}

func (st *memoryStorage) Init(ctx context.Context, opts any) error {
	// No initialization needed for memory st
	return nil
}

func (st *memoryStorage) Close() error {
	// Nothing to close for memory st
	return nil
}

func (st *memoryStorage) Drop() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	// Clear all data
	st.items = make(map[string]map[string]*memoryItem)
	st.cursors = make(map[string]string)
	st.nextIndex = make(map[string]uint64)

	return nil
}

func (st *memoryStorage) GetCursor(id string) (string, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	cursor, exists := st.cursors[id]
	if !exists {
		return "", fmt.Errorf("cursor for source_id '%s' not found", id)
	}
	return cursor, nil
}

func (st *memoryStorage) UpdateCursor(id string, cursor string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.cursors[id] = cursor
	return nil
}

func (st *memoryStorage) Push(item Item) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	sourceId := item.SourceId()
	itemId := item.Id()

	// Ensure source exists
	if _, exists := st.items[sourceId]; !exists {
		st.items[sourceId] = make(map[string]*memoryItem)
		st.nextIndex[sourceId] = 1
		// Also ensure cursor exists with empty value
		if _, exists := st.cursors[sourceId]; !exists {
			st.cursors[sourceId] = ""
		}
	}

	// Get next index for this source
	orderIndex := st.nextIndex[sourceId]
	st.nextIndex[sourceId] = orderIndex + 1

	// Create memory item
	mItem := &memoryItem{
		sourceId:   sourceId,
		id:         itemId,
		data:       append([]byte(nil), item.Bytes()...), // Copy data
		context:    nil,
		orderIndex: orderIndex,
	}

	// Copy context if present
	if item.ContextBytes() != nil {
		mItem.context = append([]byte(nil), item.ContextBytes()...)
	}

	st.items[sourceId][itemId] = mItem

	return nil
}

func (st *memoryStorage) Update(item Item) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	sourceId := item.SourceId()
	itemId := item.Id()

	sourceItems, exists := st.items[sourceId]
	if !exists {
		return fmt.Errorf("item with source_id '%s' and id '%s' not found", sourceId, itemId)
	}

	existingItem, exists := sourceItems[itemId]
	if !exists {
		return fmt.Errorf("item with source_id '%s' and id '%s' not found", sourceId, itemId)
	}

	// Update data and context, preserving order index
	existingItem.data = append([]byte(nil), item.Bytes()...)
	if item.ContextBytes() != nil {
		existingItem.context = append([]byte(nil), item.ContextBytes()...)
	} else {
		existingItem.context = nil
	}

	return nil
}

func (st *memoryStorage) Delete(sourceId string, itemId string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if sourceItems, exists := st.items[sourceId]; exists {
		delete(sourceItems, itemId)
	}

	return nil
}

func (st *memoryStorage) GetItems(sourceId string, limit int) ([]Item, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	// Default limit if not specified
	if limit <= 0 {
		limit = 100
	}

	sourceItems, exists := st.items[sourceId]
	if !exists {
		return []Item{}, nil
	}

	// Convert map to slice
	allItems := make([]*memoryItem, 0, len(sourceItems))
	for _, item := range sourceItems {
		allItems = append(allItems, item)
	}

	// Sort by order index using sort.Slice
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].orderIndex < allItems[j].orderIndex
	})

	// Apply limit
	if len(allItems) > limit {
		allItems = allItems[:limit]
	}

	// Convert to Item interface and make copies
	items := make([]Item, len(allItems))
	for i, item := range allItems {
		items[i] = &memoryItem{
			sourceId:   item.sourceId,
			id:         item.id,
			data:       append([]byte(nil), item.data...),
			context:    append([]byte(nil), item.context...),
			orderIndex: item.orderIndex,
		}
	}

	return items, nil
}

func (st *memoryStorage) GetLocked(sourceId string, limit int) ([]Item, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	// Default limit if not specified
	if limit <= 0 {
		limit = 100
	}

	sourceItems, exists := st.items[sourceId]
	if !exists {
		return []Item{}, nil
	}

	// Filter locked items
	lockedItems := make([]*memoryItem, 0)
	for _, item := range sourceItems {
		if item.locked {
			lockedItems = append(lockedItems, item)
		}
	}

	// Sort by order index using sort.Slice
	sort.Slice(lockedItems, func(i, j int) bool {
		return lockedItems[i].orderIndex < lockedItems[j].orderIndex
	})

	// Apply limit
	if len(lockedItems) > limit {
		lockedItems = lockedItems[:limit]
	}

	// Convert to Item interface and make copies
	items := make([]Item, len(lockedItems))
	for i, item := range lockedItems {
		items[i] = &memoryItem{
			sourceId:   item.sourceId,
			id:         item.id,
			data:       append([]byte(nil), item.data...),
			context:    append([]byte(nil), item.context...),
			orderIndex: item.orderIndex,
		}
	}

	return items, nil
}

func (st *memoryStorage) GetUnlocked(sourceId string, limit int) ([]Item, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	// Default limit if not specified
	if limit <= 0 {
		limit = 100
	}

	sourceItems, exists := st.items[sourceId]
	if !exists {
		return []Item{}, nil
	}

	// Filter unlocked items
	unlockedItems := make([]*memoryItem, 0)
	for _, item := range sourceItems {
		if !item.locked {
			unlockedItems = append(unlockedItems, item)
		}
	}

	// Sort by order index using sort.Slice
	sort.Slice(unlockedItems, func(i, j int) bool {
		return unlockedItems[i].orderIndex < unlockedItems[j].orderIndex
	})

	// Apply limit
	if len(unlockedItems) > limit {
		unlockedItems = unlockedItems[:limit]
	}

	// Convert to Item interface and make copies
	items := make([]Item, len(unlockedItems))
	for i, item := range unlockedItems {
		items[i] = &memoryItem{
			sourceId:   item.sourceId,
			id:         item.id,
			data:       append([]byte(nil), item.data...),
			context:    append([]byte(nil), item.context...),
			orderIndex: item.orderIndex,
		}
	}

	return items, nil
}

func (st *memoryStorage) Lock(sourceId string, itemId string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	sourceItems, exists := st.items[sourceId]
	if !exists {
		return fmt.Errorf("item with source_id '%s' and id '%s' not found", sourceId, itemId)
	}

	item, exists := sourceItems[itemId]
	if !exists {
		return fmt.Errorf("item with source_id '%s' and id '%s' not found", sourceId, itemId)
	}

	item.locked = true
	return nil
}

func (st *memoryStorage) Unlock(sourceId string, itemId string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	sourceItems, exists := st.items[sourceId]
	if !exists {
		return fmt.Errorf("item with source_id '%s' and id '%s' not found", sourceId, itemId)
	}

	item, exists := sourceItems[itemId]
	if !exists {
		return fmt.Errorf("item with source_id '%s' and id '%s' not found", sourceId, itemId)
	}

	item.locked = false
	return nil
}

// GetIndex returns the maximum order index for a source
func (st *memoryStorage) GetIndex(sourceId string) (uint64, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	// If we have a next index stored, return one less than it
	if nextIdx, exists := st.nextIndex[sourceId]; exists && nextIdx > 0 {
		return nextIdx - 1, nil
	}

	// Otherwise, find the max from existing items
	sourceItems, exists := st.items[sourceId]
	if !exists {
		return 0, nil
	}

	var maxIndex uint64 = 0
	for _, item := range sourceItems {
		if item.orderIndex > maxIndex {
			maxIndex = item.orderIndex
		}
	}

	return maxIndex, nil
}

// Implement Item interface for memoryItem
func (m *memoryItem) SourceId() string {
	return m.sourceId
}

func (m *memoryItem) Id() string {
	return m.id
}

func (m *memoryItem) Bytes() []byte {
	return m.data
}

func (m *memoryItem) ContextBytes() []byte {
	return m.context
}

// Additional helper methods for testing

// GetEventCount returns the number of events for a source
func (st *memoryStorage) GetEventCount(sourceId string) int {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if sourceItems, exists := st.items[sourceId]; exists {
		return len(sourceItems)
	}
	return 0
}

// Clear removes all data from the database
func (st *memoryStorage) Clear() {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.items = make(map[string]map[string]*memoryItem)
	st.cursors = make(map[string]string)
	st.nextIndex = make(map[string]uint64)
}

// GetNextOrderIndex is used for testing to verify order sequence
func (st *memoryStorage) GetNextOrderIndex(sourceId string) uint64 {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if idx, exists := st.nextIndex[sourceId]; exists {
		return idx
	}
	return 1
}

package bstates

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nayarsystems/bstates"
	ifx "github.com/nayarsystems/idefix-go"
)

type SchemasCacheParams struct {
	MaxAge time.Duration
}

// This function must be called once at the start of the program
func InitSchemasCache(params SchemasCacheParams) {
	maxAge = params.MaxAge
}

// GetSchema retrieves a bstates schema by its ID, using a cache to avoid redundant fetches.
func GetSchemaFromId(ic *ifx.Client, schemaId string) (*bstates.StateSchema, error) {
	// Check cache first
	if schema := sc.get(schemaId); schema != nil {
		return schema, nil
	}

	// Not in cache, fetch it
	slog.Info("fetching bstates schema", "schemaId", schemaId)
	res, err := ic.GetSchema(schemaId, time.Second*20)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bstates schema: %w", err)
	}

	// Create schema object
	schema := &bstates.StateSchema{}
	err = schema.UnmarshalJSON([]byte(res.Payload))
	if err != nil {
		return nil, fmt.Errorf("failed to parse bstates schema: %w", err)
	}

	// Cache it for future use
	sc.set(schemaId, schema)
	return schema, nil
}

var maxAge = 2 * 24 * time.Hour

type schemasCacheEntry struct {
	mutex   sync.RWMutex
	lastUse time.Time
	schema  *bstates.StateSchema
}

func (sce *schemasCacheEntry) updateLastUse() {
	sce.mutex.Lock()
	defer sce.mutex.Unlock()
	sce.lastUse = time.Now()
}

func (sce *schemasCacheEntry) getLastUse() time.Time {
	sce.mutex.RLock()
	defer sce.mutex.RUnlock()
	return sce.lastUse
}

type schemasCache struct {
	mutex           sync.RWMutex
	schemasCacheMap map[string]*schemasCacheEntry
}

func (sc *schemasCache) get(schemaId string) *bstates.StateSchema {
	sc.mutex.RLock()

	var schema *bstates.StateSchema
	if entry, ok := sc.schemasCacheMap[schemaId]; ok {
		entry.updateLastUse()
		schema = entry.schema
	}

	sc.mutex.RUnlock()

	sc.removeOldEntries()
	return schema
}

func (sc *schemasCache) set(schemaId string, schema *bstates.StateSchema) {
	sc.mutex.Lock()

	sc.schemasCacheMap[schemaId] = &schemasCacheEntry{
		lastUse: time.Now(),
		schema:  schema,
	}

	sc.mutex.Unlock()

	sc.removeOldEntries()
}

func (sc *schemasCache) removeOldEntries() {
	now := time.Now()

	var toDelete []string

	sc.mutex.RLock()
	for schemaId, entry := range sc.schemasCacheMap {
		if now.Sub(entry.getLastUse()) > maxAge {
			toDelete = append(toDelete, schemaId)
		}
	}
	sc.mutex.RUnlock()

	if len(toDelete) == 0 {
		return
	}

	sc.mutex.Lock()
	for _, schemaId := range toDelete {
		delete(sc.schemasCacheMap, schemaId)
	}
	sc.mutex.Unlock()
}

var sc = &schemasCache{
	schemasCacheMap: make(map[string]*schemasCacheEntry),
}

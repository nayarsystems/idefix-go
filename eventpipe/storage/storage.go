package storage

import (
	"context"
)

type Storage interface {
	Init(ctx context.Context, opts any) error
	Close() error
	Drop() error
	UpdateCursor(id, cursor string) error
	GetCursor(id string) (string, error)
	Push(Item) error
	Update(Item) error
	DeleteItem(sourceId, itemId string) error
	LockItem(sourceId, itemId string) error
	UnlockItem(sourceId, itemId string) error
	GetItems(sourceId string, limit int) ([]Item, error)
	GetUnlockedItems(sourceId string, limit int) ([]Item, error)
	GetLockedItems(sourceId string, limit int) ([]Item, error)
	GetIndex(sourceId string) (uint64, error)
}

type Item interface {
	SourceId() string
	Id() string
	Bytes() []byte
	ContextBytes() []byte
}

func NewStorage(storageType string) Storage {
	switch storageType {
	case "sqlite":
		return &SqliteStorage{}
	case "memory":
		return NewMemoryStorage()
	default:
		return nil
	}
}

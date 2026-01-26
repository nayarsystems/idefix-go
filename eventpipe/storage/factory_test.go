package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStorage_Factory(t *testing.T) {
	tests := []struct {
		name        string
		storageType string
		expected    any
	}{
		{"sqlite", "sqlite", &SqliteStorage{}},
		{"invalid", "invalid", nil},
		{"empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewStorage(tt.storageType)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.IsType(t, tt.expected, result)
			}
		})
	}
}

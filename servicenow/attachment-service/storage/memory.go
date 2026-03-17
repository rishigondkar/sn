package storage

import (
	"context"
	"sync"
)

// MemoryStore is an in-memory Store for tests. Not for production.
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string][]byte // key "bucket/key" -> content
}

// NewMemoryStore returns an in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string][]byte)}
}

func (m *MemoryStore) k(bucket, key string) string { return bucket + "/" + key }

// Put stores content.
func (m *MemoryStore) Put(ctx context.Context, key, bucket string, content []byte) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	dup := make([]byte, len(content))
	copy(dup, content)
	m.data[m.k(bucket, key)] = dup
	return nil
}

// Get returns content.
func (m *MemoryStore) Get(ctx context.Context, key, bucket string) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if b, ok := m.data[m.k(bucket, key)]; ok {
		dup := make([]byte, len(b))
		copy(dup, b)
		return dup, nil
	}
	return nil, ErrNotFound
}

// Delete removes the object.
func (m *MemoryStore) Delete(ctx context.Context, key, bucket string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, m.k(bucket, key))
	return nil
}

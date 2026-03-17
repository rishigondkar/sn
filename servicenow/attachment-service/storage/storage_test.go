package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyGenerator(t *testing.T) {
	k := KeyGenerator("case-1", "att-1", "file.txt")
	assert.Equal(t, "case-1/att-1/file.txt", k)
	// No path traversal in output when inputs are sanitized
	k2 := KeyGenerator("case-1", "att-1", "sub/file.txt")
	assert.Contains(t, k2, "att-1")
}

func TestMemoryStore_PutGetDelete(t *testing.T) {
	ctx := context.Background()
	m := NewMemoryStore()
	content := []byte("hello world")
	err := m.Put(ctx, "key1", "bucket1", content)
	require.NoError(t, err)
	got, err := m.Get(ctx, "key1", "bucket1")
	require.NoError(t, err)
	assert.Equal(t, content, got)
	err = m.Delete(ctx, "key1", "bucket1")
	require.NoError(t, err)
	_, err = m.Get(ctx, "key1", "bucket1")
	assert.ErrorIs(t, err, ErrNotFound)
}

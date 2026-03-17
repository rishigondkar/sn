package storage

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("storage: object not found")

// Store abstracts object storage. Keys are always server-generated; no client-provided paths.
type Store interface {
	// Put writes content to key in bucket. Key must be generated server-side (e.g. caseID/attachmentID/sanitized-filename).
	Put(ctx context.Context, key, bucket string, content []byte) error
	// Get reads content from key in bucket.
	Get(ctx context.Context, key, bucket string) ([]byte, error)
	// Delete removes the object at key in bucket.
	Delete(ctx context.Context, key, bucket string) error
}

// KeyGenerator generates a safe storage key (no path traversal). Example: caseID/attachmentID/filename.
func KeyGenerator(caseID, attachmentID, sanitizedFileName string) string {
	return caseID + "/" + attachmentID + "/" + sanitizedFileName
}

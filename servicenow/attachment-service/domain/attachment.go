package domain

import "time"

// Attachment is the in-service representation of an attachment (metadata only; no binary).
type Attachment struct {
	ID                 string
	CaseID             string
	FileName           string
	FileSizeBytes      int64
	ContentType        string
	StorageProvider    string
	StorageKey         string
	StorageBucket      string
	UploadedByUserID   string
	UploadedAt         time.Time
	IsDeleted         bool
	DeletedAt         *time.Time
	MetadataJSON       string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

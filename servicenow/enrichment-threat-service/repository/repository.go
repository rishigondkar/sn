package repository

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EnrichmentResult is the domain model for enrichment_results table.
type EnrichmentResult struct {
	ID              string
	CaseID          *string
	ObservableID    *string
	EnrichmentType  string
	SourceName      string
	SourceRecordID  *string
	Status          string
	Summary         *string
	ResultData      json.RawMessage
	Score           *float64
	Confidence     *float64
	RequestedAt     *time.Time
	ReceivedAt     time.Time
	ExpiresAt      *time.Time
	LastUpdatedBy   *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ThreatLookupResult is the domain model for threat_lookup_results table.
type ThreatLookupResult struct {
	ID                string
	CaseID            *string
	ObservableID      string
	LookupType        string
	SourceName        string
	SourceRecordID    *string
	Verdict           *string
	RiskScore         *float64
	ConfidenceScore   *float64
	Tags              json.RawMessage
	MatchedIndicators json.RawMessage
	Summary           *string
	ResultData        json.RawMessage
	QueriedAt         *time.Time
	ReceivedAt        time.Time
	ExpiresAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ListFilter holds optional filters for list queries.
type ListFilter struct {
	SourceName   *string
	Type         *string // enrichment_type or lookup_type
	Verdict      *string // threat only
	ActiveOnly   bool
	PageSize     int32
	PageToken    string // offset-based for simplicity: numeric string
}

// ThreatSummary holds aggregated threat lookup summary for an observable.
type ThreatSummary struct {
	ObservableID      string
	TotalCount        int32
	HighestVerdict    *string
	MaxRiskScore      *float64
	SourceNames       []string
	LatestReceivedAt  time.Time
}

// Repository provides data access for enrichment and threat lookup results.
type Repository struct {
	Pool *pgxpool.Pool
}

// NewRepository returns a Repository using the given pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{Pool: pool}
}

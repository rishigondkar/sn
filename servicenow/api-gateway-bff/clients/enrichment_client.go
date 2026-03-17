package clients

import "context"

// EnrichmentQueryService for enrichment read operations.
type EnrichmentQueryService interface {
	ListEnrichmentResultsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*EnrichmentResult, string, error)
}

// ThreatLookupQueryService for threat lookup read operations.
type ThreatLookupQueryService interface {
	ListThreatLookupResultsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*ThreatLookupResult, string, error)
	ListThreatLookupsByObservable(ctx context.Context, observableID string, pageSize int32, pageToken string) ([]*ThreatLookupResult, string, error)
}

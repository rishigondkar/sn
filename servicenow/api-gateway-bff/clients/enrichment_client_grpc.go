package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/servicenow/enrichment-threat-service/proto/enrichment_threat_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	_ EnrichmentQueryService   = (*enrichmentQueryGRPC)(nil)
	_ ThreatLookupQueryService = (*threatLookupQueryGRPC)(nil)
)

type enrichmentQueryGRPC struct {
	client pb.EnrichmentThreatServiceClient
}

type threatLookupQueryGRPC struct {
	client pb.EnrichmentThreatServiceClient
}

// NewEnrichmentClients dials the Enrichment & Threat Service and returns Enrichment and ThreatLookup query clients.
func NewEnrichmentClients(addr string, timeout time.Duration) (EnrichmentQueryService, ThreatLookupQueryService, func(), error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("enrichment service dial: %w", err)
	}
	client := pb.NewEnrichmentThreatServiceClient(conn)
	enrich := &enrichmentQueryGRPC{client: client}
	threat := &threatLookupQueryGRPC{client: client}
	closeFn := func() { _ = conn.Close() }
	return enrich, threat, closeFn, nil
}

func (c *enrichmentQueryGRPC) ListEnrichmentResultsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*EnrichmentResult, string, error) {
	if caseID == "" {
		return nil, "", fmt.Errorf("caseID is required")
	}
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListEnrichmentResultsByCase(ctx, &pb.ListEnrichmentResultsByCaseRequest{
		CaseId:    caseID,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*EnrichmentResult, 0, len(resp.GetItems()))
	for _, r := range resp.GetItems() {
		list = append(list, protoEnrichmentToEnrichmentResult(r))
	}
	return list, resp.GetNextPageToken(), nil
}

func (c *threatLookupQueryGRPC) ListThreatLookupResultsByCase(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*ThreatLookupResult, string, error) {
	if caseID == "" {
		return nil, "", fmt.Errorf("caseID is required")
	}
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListThreatLookupResultsByCase(ctx, &pb.ListThreatLookupResultsByCaseRequest{
		CaseId:    caseID,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*ThreatLookupResult, 0, len(resp.GetItems()))
	for _, r := range resp.GetItems() {
		list = append(list, protoThreatLookupToThreatLookupResult(r))
	}
	return list, resp.GetNextPageToken(), nil
}

func (c *threatLookupQueryGRPC) ListThreatLookupsByObservable(ctx context.Context, observableID string, pageSize int32, pageToken string) ([]*ThreatLookupResult, string, error) {
	if observableID == "" {
		return nil, "", fmt.Errorf("observableID is required")
	}
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListThreatLookupResultsByObservable(ctx, &pb.ListThreatLookupResultsByObservableRequest{
		ObservableId: observableID,
		PageSize:     pageSize,
		PageToken:    pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*ThreatLookupResult, 0, len(resp.GetItems()))
	for _, r := range resp.GetItems() {
		list = append(list, protoThreatLookupToThreatLookupResult(r))
	}
	return list, resp.GetNextPageToken(), nil
}

func protoEnrichmentToEnrichmentResult(r *pb.EnrichmentResult) *EnrichmentResult {
	if r == nil {
		return nil
	}
	out := &EnrichmentResult{
		ID:           r.GetId(),
		ObservableID: r.GetObservableId(),
		CaseID:       r.GetCaseId(),
		Source:       r.GetSourceName(),
		Result:       r.GetSummary(),
	}
	if r.GetReceivedAt() != nil {
		out.CreatedAt = r.GetReceivedAt().AsTime()
	}
	return out
}

func protoThreatLookupToThreatLookupResult(r *pb.ThreatLookupResult) *ThreatLookupResult {
	if r == nil {
		return nil
	}
	out := &ThreatLookupResult{
		ID:           r.GetId(),
		ObservableID: r.GetObservableId(),
		CaseID:       r.GetCaseId(),
		Provider:     r.GetSourceName(),
		Verdict:      r.GetVerdict(),
	}
	if r.GetReceivedAt() != nil {
		out.CreatedAt = r.GetReceivedAt().AsTime()
	}
	return out
}

package handler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	pb "github.com/servicenow/enrichment-threat-service/proto/enrichment_threat_service"
	"github.com/servicenow/enrichment-threat-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) UpsertEnrichmentResult(ctx context.Context, req *pb.UpsertEnrichmentResultRequest) (*pb.UpsertEnrichmentResultResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("UpsertEnrichmentResult failed", "err", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	actor := actorFromContext(ctx)
	in := enrichmentReqToInput(req)
	result, err := h.Service.UpsertEnrichmentResult(ctx, in, actor)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, "validation failed: at least one of case_id or observable_id required; result_data must be valid JSON")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.UpsertEnrichmentResultResponse{Result: enrichmentToProto(result)}, nil
}

func (h *Handler) ListEnrichmentResultsByCase(ctx context.Context, req *pb.ListEnrichmentResultsByCaseRequest) (*pb.ListEnrichmentResultsByCaseResponse, error) {
	if req.GetCaseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "case_id required")
	}
	items, nextToken, err := h.Service.ListEnrichmentResultsByCase(ctx, req.GetCaseId(), listFilterFromEnrichmentCase(req))
	if err != nil {
		slog.Error("ListEnrichmentResultsByCase failed", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	out := &pb.ListEnrichmentResultsByCaseResponse{NextPageToken: nextToken}
	for _, e := range items {
		out.Items = append(out.Items, enrichmentToProto(e))
	}
	return out, nil
}

func (h *Handler) ListEnrichmentResultsByObservable(ctx context.Context, req *pb.ListEnrichmentResultsByObservableRequest) (*pb.ListEnrichmentResultsByObservableResponse, error) {
	if req.GetObservableId() == "" {
		return nil, status.Error(codes.InvalidArgument, "observable_id required")
	}
	items, nextToken, err := h.Service.ListEnrichmentResultsByObservable(ctx, req.GetObservableId(), listFilterFromEnrichmentObservable(req))
	if err != nil {
		slog.Error("ListEnrichmentResultsByObservable failed", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	out := &pb.ListEnrichmentResultsByObservableResponse{NextPageToken: nextToken}
	for _, e := range items {
		out.Items = append(out.Items, enrichmentToProto(e))
	}
	return out, nil
}

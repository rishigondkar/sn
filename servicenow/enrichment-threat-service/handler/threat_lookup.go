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

func (h *Handler) UpsertThreatLookupResult(ctx context.Context, req *pb.UpsertThreatLookupResultRequest) (*pb.UpsertThreatLookupResultResponse, error) {
	start := time.Now()
	var err error
	defer func() {
		if err != nil {
			slog.Error("UpsertThreatLookupResult failed", "err", err, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	actor := actorFromContext(ctx)
	in := threatLookupReqToInput(req)
	result, err := h.Service.UpsertThreatLookupResult(ctx, in, actor)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			return nil, status.Error(codes.InvalidArgument, "validation failed: observable_id required; result_data must be valid JSON")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.UpsertThreatLookupResultResponse{Result: threatLookupToProto(result)}, nil
}

func (h *Handler) ListThreatLookupResultsByCase(ctx context.Context, req *pb.ListThreatLookupResultsByCaseRequest) (*pb.ListThreatLookupResultsByCaseResponse, error) {
	if req.GetCaseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "case_id required")
	}
	items, nextToken, err := h.Service.ListThreatLookupResultsByCase(ctx, req.GetCaseId(), listFilterFromThreatCase(req))
	if err != nil {
		slog.Error("ListThreatLookupResultsByCase failed", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	out := &pb.ListThreatLookupResultsByCaseResponse{NextPageToken: nextToken}
	for _, t := range items {
		out.Items = append(out.Items, threatLookupToProto(t))
	}
	return out, nil
}

func (h *Handler) ListThreatLookupResultsByObservable(ctx context.Context, req *pb.ListThreatLookupResultsByObservableRequest) (*pb.ListThreatLookupResultsByObservableResponse, error) {
	if req.GetObservableId() == "" {
		return nil, status.Error(codes.InvalidArgument, "observable_id required")
	}
	items, nextToken, err := h.Service.ListThreatLookupResultsByObservable(ctx, req.GetObservableId(), listFilterFromThreatObservable(req))
	if err != nil {
		slog.Error("ListThreatLookupResultsByObservable failed", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	out := &pb.ListThreatLookupResultsByObservableResponse{NextPageToken: nextToken}
	for _, t := range items {
		out.Items = append(out.Items, threatLookupToProto(t))
	}
	return out, nil
}

func (h *Handler) GetThreatLookupSummaryByObservable(ctx context.Context, req *pb.GetThreatLookupSummaryByObservableRequest) (*pb.GetThreatLookupSummaryByObservableResponse, error) {
	if req.GetObservableId() == "" {
		return nil, status.Error(codes.InvalidArgument, "observable_id required")
	}
	sum, err := h.Service.GetThreatLookupSummaryByObservable(ctx, req.GetObservableId())
	if err != nil {
		slog.Error("GetThreatLookupSummaryByObservable failed", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
	out := &pb.GetThreatLookupSummaryByObservableResponse{
		ObservableId:     sum.ObservableID,
		TotalCount:       sum.TotalCount,
		SourceNames:      sum.SourceNames,
		LatestReceivedAt: timePtrToProto(&sum.LatestReceivedAt),
	}
	if sum.HighestVerdict != nil {
		out.HighestVerdict = sum.HighestVerdict
	}
	if sum.MaxRiskScore != nil {
		out.MaxRiskScore = sum.MaxRiskScore
	}
	return out, nil
}

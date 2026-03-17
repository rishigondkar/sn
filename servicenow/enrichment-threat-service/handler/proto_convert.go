package handler

import (
	"time"

	"github.com/servicenow/enrichment-threat-service/repository"
	"github.com/servicenow/enrichment-threat-service/service"
	pb "github.com/servicenow/enrichment-threat-service/proto/enrichment_threat_service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func protoToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil || !ts.IsValid() {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func timeToProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func timePtrToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func enrichmentReqToInput(req *pb.UpsertEnrichmentResultRequest) service.UpsertEnrichmentInput {
	in := service.UpsertEnrichmentInput{
		ID:             req.GetId(),
		EnrichmentType: req.GetEnrichmentType(),
		SourceName:     req.GetSourceName(),
		Status:         req.GetStatus(),
		ResultDataJSON: req.GetResultDataJson(),
		CaseID:         req.CaseId,
		ObservableID:   req.ObservableId,
		SourceRecordID: req.SourceRecordId,
		Summary:        req.Summary,
		Score:          req.Score,
		Confidence:     req.Confidence,
		RequestedAt:    protoToTime(req.GetRequestedAt()),
		ExpiresAt:      protoToTime(req.GetExpiresAt()),
		LastUpdatedBy:  req.LastUpdatedBy,
	}
	if req.ReceivedAt != nil {
		in.ReceivedAt = req.ReceivedAt.AsTime()
	}
	return in
}

func enrichmentToProto(e *repository.EnrichmentResult) *pb.EnrichmentResult {
	if e == nil {
		return nil
	}
	out := &pb.EnrichmentResult{
		Id:             e.ID,
		EnrichmentType: e.EnrichmentType,
		SourceName:     e.SourceName,
		Status:         e.Status,
		ResultDataJson: string(e.ResultData),
		ReceivedAt:     timeToProto(e.ReceivedAt),
		CreatedAt:      timeToProto(e.CreatedAt),
		UpdatedAt:      timeToProto(e.UpdatedAt),
	}
	if e.CaseID != nil {
		out.CaseId = *e.CaseID
	}
	if e.ObservableID != nil {
		out.ObservableId = *e.ObservableID
	}
	if e.SourceRecordID != nil {
		out.SourceRecordId = *e.SourceRecordID
	}
	if e.Summary != nil {
		out.Summary = *e.Summary
	}
	if e.Score != nil {
		out.Score = *e.Score
	}
	if e.Confidence != nil {
		out.Confidence = *e.Confidence
	}
	out.RequestedAt = timePtrToProto(e.RequestedAt)
	out.ExpiresAt = timePtrToProto(e.ExpiresAt)
	if e.LastUpdatedBy != nil {
		out.LastUpdatedBy = *e.LastUpdatedBy
	}
	return out
}

func threatLookupReqToInput(req *pb.UpsertThreatLookupResultRequest) service.UpsertThreatLookupInput {
	in := service.UpsertThreatLookupInput{
		ID:             req.GetId(),
		ObservableID:   req.GetObservableId(),
		LookupType:     req.GetLookupType(),
		SourceName:     req.GetSourceName(),
		ResultDataJSON: req.GetResultDataJson(),
		CaseID:         req.CaseId,
		SourceRecordID: req.SourceRecordId,
		Verdict:        req.Verdict,
		RiskScore:      req.RiskScore,
		ConfidenceScore: req.ConfidenceScore,
		Summary:        req.Summary,
		QueriedAt:      protoToTime(req.GetQueriedAt()),
		ExpiresAt:      protoToTime(req.GetExpiresAt()),
	}
	if req.ReceivedAt != nil {
		in.ReceivedAt = req.ReceivedAt.AsTime()
	}
	if req.TagsJson != nil {
		in.TagsJSON = req.TagsJson
	}
	if req.MatchedIndicatorsJson != nil {
		in.MatchedIndicatorsJSON = req.MatchedIndicatorsJson
	}
	return in
}

func threatLookupToProto(t *repository.ThreatLookupResult) *pb.ThreatLookupResult {
	if t == nil {
		return nil
	}
	out := &pb.ThreatLookupResult{
		Id:             t.ID,
		ObservableId:   t.ObservableID,
		LookupType:     t.LookupType,
		SourceName:     t.SourceName,
		ResultDataJson: string(t.ResultData),
		QueriedAt:      timePtrToProto(t.QueriedAt),
		ReceivedAt:     timeToProto(t.ReceivedAt),
		ExpiresAt:      timePtrToProto(t.ExpiresAt),
		CreatedAt:      timeToProto(t.CreatedAt),
		UpdatedAt:      timeToProto(t.UpdatedAt),
	}
	if t.CaseID != nil {
		out.CaseId = *t.CaseID
	}
	if t.SourceRecordID != nil {
		out.SourceRecordId = *t.SourceRecordID
	}
	if t.Verdict != nil {
		out.Verdict = *t.Verdict
	}
	if t.RiskScore != nil {
		out.RiskScore = *t.RiskScore
	}
	if t.ConfidenceScore != nil {
		out.ConfidenceScore = *t.ConfidenceScore
	}
	if t.Tags != nil {
		out.TagsJson = string(t.Tags)
	}
	if t.MatchedIndicators != nil {
		out.MatchedIndicatorsJson = string(t.MatchedIndicators)
	}
	if t.Summary != nil {
		out.Summary = *t.Summary
	}
	return out
}

func listFilterFromEnrichmentCase(req *pb.ListEnrichmentResultsByCaseRequest) repository.ListFilter {
	return repository.ListFilter{
		PageSize:   req.GetPageSize(),
		PageToken:  req.GetPageToken(),
		SourceName: req.SourceName,
		Type:       req.EnrichmentType,
		ActiveOnly: req.GetActiveOnly(),
	}
}

func listFilterFromEnrichmentObservable(req *pb.ListEnrichmentResultsByObservableRequest) repository.ListFilter {
	return repository.ListFilter{
		PageSize:   req.GetPageSize(),
		PageToken:  req.GetPageToken(),
		SourceName: req.SourceName,
		Type:       req.EnrichmentType,
		ActiveOnly: req.GetActiveOnly(),
	}
}

func listFilterFromThreatCase(req *pb.ListThreatLookupResultsByCaseRequest) repository.ListFilter {
	return repository.ListFilter{
		PageSize:   req.GetPageSize(),
		PageToken:  req.GetPageToken(),
		SourceName: req.SourceName,
		Type:       req.LookupType,
		Verdict:    req.Verdict,
		ActiveOnly: req.GetActiveOnly(),
	}
}

func listFilterFromThreatObservable(req *pb.ListThreatLookupResultsByObservableRequest) repository.ListFilter {
	return repository.ListFilter{
		PageSize:   req.GetPageSize(),
		PageToken:  req.GetPageToken(),
		SourceName: req.SourceName,
		Type:       req.LookupType,
		Verdict:    req.Verdict,
		ActiveOnly: req.GetActiveOnly(),
	}
}

package handler

import (
	"context"

	"github.com/org/alert-observable-service/service"
	pb "github.com/org/alert-observable-service/proto/alert_observable_service"
)

func (s *Server) CreateChildObservableRelation(ctx context.Context, req *pb.CreateChildObservableRelationRequest) (*pb.CreateChildObservableRelationResponse, error) {
	actor := actorFromContext(ctx)
	r := &service.CreateChildObservableRelationRequest{
		ParentObservableID:   req.GetParentObservableId(),
		ChildObservableID:    req.GetChildObservableId(),
		RelationshipType:     req.GetRelationshipType(),
		RelationshipDirection: req.GetRelationshipDirection(),
		Confidence:          req.GetConfidence(),
		SourceName:          req.GetSourceName(),
		SourceRecordID:      req.GetSourceRecordId(),
		MetadataJSON:        req.GetMetadataJson(),
	}
	c, err := s.svc.CreateChildObservableRelation(ctx, r, actor)
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.CreateChildObservableRelationResponse{Relation: childObservableToProto(c)}, nil
}

func (s *Server) RecomputeSimilarIncidentsForCase(ctx context.Context, req *pb.RecomputeSimilarIncidentsForCaseRequest) (*pb.RecomputeSimilarIncidentsForCaseResponse, error) {
	actor := actorFromContext(ctx)
	if err := s.svc.RecomputeSimilarIncidentsForCase(ctx, req.GetCaseId(), actor); err != nil {
		return nil, mapErr(err)
	}
	return &pb.RecomputeSimilarIncidentsForCaseResponse{}, nil
}

func (s *Server) ListChildObservables(ctx context.Context, req *pb.ListChildObservablesRequest) (*pb.ListChildObservablesResponse, error) {
	items, nextToken, err := s.svc.ListChildObservables(ctx, req.GetParentObservableId(), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.ChildObservable, len(items))
	for i, c := range items {
		out[i] = childObservableToProto(c)
	}
	return &pb.ListChildObservablesResponse{Items: out, NextPageToken: nextToken}, nil
}

func (s *Server) ListSimilarIncidents(ctx context.Context, req *pb.ListSimilarIncidentsRequest) (*pb.ListSimilarIncidentsResponse, error) {
	items, nextToken, err := s.svc.ListSimilarIncidents(ctx, req.GetCaseId(), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.SimilarIncident, len(items))
	for i, si := range items {
		out[i] = similarIncidentToProto(si)
	}
	return &pb.ListSimilarIncidentsResponse{Items: out, NextPageToken: nextToken}, nil
}

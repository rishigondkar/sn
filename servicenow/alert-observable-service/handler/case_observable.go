package handler

import (
	"context"

	"github.com/org/alert-observable-service/service"
	pb "github.com/org/alert-observable-service/proto/alert_observable_service"
)

func (s *Server) LinkObservableToCase(ctx context.Context, req *pb.LinkObservableToCaseRequest) (*pb.LinkObservableToCaseResponse, error) {
	actor := actorFromContext(ctx)
	r := &service.LinkObservableToCaseRequest{
		CaseID:          req.GetCaseId(),
		ObservableType:  req.GetObservableType(),
		ObservableValue: req.GetObservableValue(),
		NormalizedValue: req.GetNormalizedValue(),
		RoleInCase:      req.GetRoleInCase(),
		TrackingStatus:  req.GetTrackingStatus(),
		IsPrimary:       req.GetIsPrimary(),
		AddedByUserID:   req.GetAddedByUserId(),
	}
	co, err := s.svc.LinkObservableToCase(ctx, r, actor)
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.LinkObservableToCaseResponse{CaseObservable: caseObservableToProto(co)}, nil
}

func (s *Server) UpdateCaseObservable(ctx context.Context, req *pb.UpdateCaseObservableRequest) (*pb.UpdateCaseObservableResponse, error) {
	actor := actorFromContext(ctx)
	var trackingStatus, accuracy, determination, impact *string
	if req.TrackingStatus != nil {
		v := req.GetTrackingStatus()
		trackingStatus = &v
	}
	if req.Accuracy != nil {
		v := req.GetAccuracy()
		accuracy = &v
	}
	if req.Determination != nil {
		v := req.GetDetermination()
		determination = &v
	}
	if req.Impact != nil {
		v := req.GetImpact()
		impact = &v
	}
	co, err := s.svc.UpdateCaseObservable(ctx, req.GetCaseObservableId(), trackingStatus, accuracy, determination, impact, actor)
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.UpdateCaseObservableResponse{CaseObservable: caseObservableToProto(co)}, nil
}

func (s *Server) ListCaseObservables(ctx context.Context, req *pb.ListCaseObservablesRequest) (*pb.ListCaseObservablesResponse, error) {
	var filterType, filterStatus *string
	if req.ObservableType != nil {
		v := req.GetObservableType()
		filterType = &v
	}
	if req.TrackingStatus != nil {
		v := req.GetTrackingStatus()
		filterStatus = &v
	}
	items, nextToken, err := s.svc.ListCaseObservablesWithDetails(ctx, req.GetCaseId(), req.GetPageSize(), req.GetPageToken(), filterType, filterStatus)
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.CaseObservable, len(items))
	for i, d := range items {
		out[i] = caseObservableWithDetailsToProto(d)
	}
	return &pb.ListCaseObservablesResponse{Items: out, NextPageToken: nextToken}, nil
}

func (s *Server) UnlinkObservableFromCase(ctx context.Context, req *pb.UnlinkObservableFromCaseRequest) (*pb.UnlinkObservableFromCaseResponse, error) {
	actor := actorFromContext(ctx)
	if err := s.svc.UnlinkObservableFromCase(ctx, req.GetCaseId(), req.GetObservableId(), actor); err != nil {
		return nil, mapErr(err)
	}
	return &pb.UnlinkObservableFromCaseResponse{}, nil
}

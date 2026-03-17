package handler

import (
	"context"

	"github.com/org/alert-observable-service/service"
	pb "github.com/org/alert-observable-service/proto/alert_observable_service"
)

func (s *Server) CreateOrGetObservable(ctx context.Context, req *pb.CreateOrGetObservableRequest) (*pb.CreateOrGetObservableResponse, error) {
	r := &service.CreateOrGetObservableRequest{
		ObservableType:  req.GetObservableType(),
		ObservableValue: req.GetObservableValue(),
		NormalizedValue: req.GetNormalizedValue(),
		FirstSeenTime:   req.GetFirstSeenTime(),
		LastSeenTime:   req.GetLastSeenTime(),
	}
	o, created, err := s.svc.CreateOrGetObservable(ctx, r)
	if err != nil {
		return nil, mapErr(err)
	}
	if o == nil {
		return nil, mapErr(&service.ErrNotFound{Resource: "observable", ID: ""})
	}
	return &pb.CreateOrGetObservableResponse{Observable: observableToProto(o), Created: created}, nil
}

func (s *Server) GetObservable(ctx context.Context, req *pb.GetObservableRequest) (*pb.GetObservableResponse, error) {
	o, err := s.svc.GetObservable(ctx, req.GetObservableId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.GetObservableResponse{Observable: observableToProto(o)}, nil
}

func (s *Server) UpdateObservable(ctx context.Context, req *pb.UpdateObservableRequest) (*pb.UpdateObservableResponse, error) {
	r := &service.UpdateObservableRequest{ObservableID: req.GetObservableId()}
	if req.ObservableValue != nil {
		v := req.GetObservableValue()
		r.ObservableValue = &v
	}
	if req.ObservableType != nil {
		v := req.GetObservableType()
		r.ObservableType = &v
	}
	if req.Finding != nil {
		v := req.GetFinding()
		r.Finding = &v
	}
	if req.Notes != nil {
		v := req.GetNotes()
		r.Notes = &v
	}
	o, err := s.svc.UpdateObservable(ctx, r)
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.UpdateObservableResponse{Observable: observableToProto(o)}, nil
}

func (s *Server) FindCasesByObservable(ctx context.Context, req *pb.FindCasesByObservableRequest) (*pb.FindCasesByObservableResponse, error) {
	ids, nextToken, err := s.svc.FindCasesByObservable(ctx, req.GetObservableId(), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.FindCasesByObservableResponse{CaseIds: ids, NextPageToken: nextToken}, nil
}

func (s *Server) ListObservables(ctx context.Context, req *pb.ListObservablesRequest) (*pb.ListObservablesResponse, error) {
	search := ""
	if req.Search != nil {
		search = req.GetSearch()
	}
	items, nextToken, err := s.svc.ListObservables(ctx, search, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.Observable, len(items))
	for i, o := range items {
		out[i] = observableToProto(o)
	}
	return &pb.ListObservablesResponse{Items: out, NextPageToken: nextToken}, nil
}

package handler

import (
	"context"

	"github.com/org/alert-observable-service/service"
	pb "github.com/org/alert-observable-service/proto/alert_observable_service"
)

func (s *Server) CreateAlertRule(ctx context.Context, req *pb.CreateAlertRuleRequest) (*pb.CreateAlertRuleResponse, error) {
	if req.GetRuleName() == "" {
		return nil, invalidArg("rule_name", "required")
	}
	id, err := s.svc.CreateAlertRule(ctx, req.GetRuleName(), req.GetRuleType(), req.GetSourceSystem(), req.GetExternalRuleKey(), req.GetDescription(), req.GetIsActive())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.CreateAlertRuleResponse{Id: id.String()}, nil
}

func (s *Server) CreateAlert(ctx context.Context, req *pb.CreateAlertRequest) (*pb.CreateAlertResponse, error) {
	actor := actorFromContext(ctx)
	r := &service.CreateAlertRequest{
		CaseID:            req.GetCaseId(),
		AlertRuleID:       req.GetAlertRuleId(),
		SourceSystem:      req.GetSourceSystem(),
		SourceAlertID:     req.GetSourceAlertId(),
		Title:             req.GetTitle(),
		Description:       req.GetDescription(),
		EventOccurredTime: req.GetEventOccurredTime(),
		EventReceivedTime: req.GetEventReceivedTime(),
		Severity:          req.GetSeverity(),
		RawPayloadJSON:    req.GetRawPayloadJson(),
	}
	id, err := s.svc.CreateAlert(ctx, r, actor)
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.CreateAlertResponse{Id: id.String()}, nil
}

func (s *Server) GetAlert(ctx context.Context, req *pb.GetAlertRequest) (*pb.GetAlertResponse, error) {
	a, err := s.svc.GetAlert(ctx, req.GetAlertId())
	if err != nil {
		return nil, mapErr(err)
	}
	return &pb.GetAlertResponse{Alert: alertToProto(a)}, nil
}

func (s *Server) ListCaseAlerts(ctx context.Context, req *pb.ListCaseAlertsRequest) (*pb.ListCaseAlertsResponse, error) {
	items, nextToken, err := s.svc.ListCaseAlerts(ctx, req.GetCaseId(), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*pb.Alert, len(items))
	for i, a := range items {
		out[i] = alertToProto(a)
	}
	return &pb.ListCaseAlertsResponse{Items: out, NextPageToken: nextToken}, nil
}

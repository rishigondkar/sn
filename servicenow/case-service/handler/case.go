package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/servicenow/case-service/domain"
	"github.com/servicenow/case-service/proto/case_service"
	"github.com/servicenow/case-service/repository"
	"github.com/servicenow/case-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) CreateCase(ctx context.Context, req *case_service.CreateCaseRequest) (*case_service.CreateCaseResponse, error) {
	actor := actorFromContext(ctx)
	state := req.GetState()
	if state == "" {
		state = domain.StateDraft
	}
	input := &domain.Case{
		ShortDescription:     req.GetShortDescription(),
		Description:         req.GetDescription(),
		State:               state,
		Priority:            req.GetPriority(),
		Severity:            req.GetSeverity(),
		OpenedByUserID:      req.GetOpenedByUserId(),
		OpenedTime:          protoTimeToDomain(req.GetOpenedTime()),
		EventOccurredTime:   protoTimePtrToDomain(req.GetEventOccurredTime()),
		EventReceivedTime:   protoTimePtrToDomain(req.GetEventReceivedTime()),
		AffectedUserID:      strPtr(req.GetAffectedUserId()),
		AssignedUserID:      strPtr(req.GetAssignedUserId()),
		AssignmentGroupID:   strPtr(req.GetAssignmentGroupId()),
		AlertRuleID:         strPtr(req.GetAlertRuleId()),
		Accuracy:            strPtr(req.GetAccuracy()),
		Determination:       strPtr(req.GetDetermination()),
		Impact:              strPtr(req.GetImpact()),
		FollowupTime:        protoTimePtrToDomain(req.GetFollowupTime()),
		Category:            strPtr(req.GetCategory()),
		Subcategory:         strPtr(req.GetSubcategory()),
		Source:              strPtr(req.GetSource()),
		SourceTool:          strPtr(req.GetSourceTool()),
		SourceToolFeature:   strPtr(req.GetSourceToolFeature()),
		ConfigurationItem:   strPtr(req.GetConfigurationItem()),
		SOCNotes:            strPtr(req.GetSocNotes()),
		NextSteps:           strPtr(req.GetNextSteps()),
		CSIRTClassification: strPtr(req.GetCsirtClassification()),
		SOCLeadUserID:       strPtr(req.GetSocLeadUserId()),
		NotificationTime:    protoTimePtrToDomain(req.GetNotificationTime()),
		IsAffectedUserVIP:   req.GetIsAffectedUserVip(),
		RequestedByUserID:   strPtr(req.GetRequestedByUserId()),
		EnvironmentLevel:    strPtr(req.GetEnvironmentLevel()),
		EnvironmentType:     strPtr(req.GetEnvironmentType()),
		PDN:                 strPtr(req.GetPdn()),
		ImpactedObject:      strPtr(req.GetImpactedObject()),
		MTTR:                strPtr(req.GetMttr()),
		EngineeringDocument: strPtr(req.GetEngineeringDocument()),
		ResponseDocument:    strPtr(req.GetResponseDocument()),
	}
	id, caseNumber, err := h.Service.CreateCase(ctx, input, actor)
	if err != nil {
		return nil, mapError(err)
	}
	return &case_service.CreateCaseResponse{Id: id, CaseNumber: caseNumber}, nil
}

func (h *Handler) GetCase(ctx context.Context, req *case_service.GetCaseRequest) (*case_service.GetCaseResponse, error) {
	c, err := h.Service.GetCase(ctx, req.GetCaseId())
	if err != nil {
		return nil, mapError(err)
	}
	if c == nil {
		return nil, status.Error(codes.NotFound, "case not found")
	}
	return &case_service.GetCaseResponse{Case: domainCaseToProto(c)}, nil
}

func (h *Handler) UpdateCase(ctx context.Context, req *case_service.UpdateCaseRequest) (*case_service.UpdateCaseResponse, error) {
	actor := actorFromContext(ctx)
	updates := &service.UpdateCaseInput{
		ShortDescription:      req.ShortDescription,
		Description:           req.Description,
		State:                 req.State,
		Priority:              req.Priority,
		Severity:              req.Severity,
		AffectedUserID:        req.AffectedUserId,
		AssignedUserID:        req.AssignedUserId,
		AssignmentGroupID:     req.AssignmentGroupId,
		Accuracy:              req.Accuracy,
		Determination:         req.Determination,
		Impact:                req.Impact,
		VersionNo:             req.GetVersionNo(),
		FollowupTime:          req.FollowupTime,
		Category:              req.Category,
		Subcategory:           req.Subcategory,
		Source:                req.Source,
		SourceTool:            req.SourceTool,
		SourceToolFeature:     req.SourceToolFeature,
		ConfigurationItem:     req.ConfigurationItem,
		SOCNotes:              req.SocNotes,
		NextSteps:             req.NextSteps,
		CSIRTClassification:   req.CsirtClassification,
		SOCLeadUserID:         req.SocLeadUserId,
		NotificationTime:     req.NotificationTime,
		IsAffectedUserVIP:    req.IsAffectedUserVip,
		RequestedByUserID:    req.RequestedByUserId,
		EnvironmentLevel:     req.EnvironmentLevel,
		EnvironmentType:      req.EnvironmentType,
		PDN:                  req.Pdn,
		ImpactedObject:       req.ImpactedObject,
		MTTR:                 req.Mttr,
		ReassignmentCount:    req.ReassignmentCount,
		AssignedToCount:      req.AssignedToCount,
		EngineeringDocument:  req.EngineeringDocument,
		ResponseDocument:     req.ResponseDocument,
		ClosureReason:        req.ClosureReason,
		EventOccurredTime:    req.EventOccurredTime,
	}
	c, err := h.Service.UpdateCase(ctx, req.GetId(), updates, actor)
	if err != nil {
		return nil, mapError(err)
	}
	return &case_service.UpdateCaseResponse{Case: domainCaseToProto(c)}, nil
}

func (h *Handler) AssignCase(ctx context.Context, req *case_service.AssignCaseRequest) (*case_service.AssignCaseResponse, error) {
	actor := actorFromContext(ctx)
	c, err := h.Service.AssignCase(ctx, req.GetCaseId(), req.GetAssignedUserId(), req.GetAssignmentGroupId(), req.GetVersionNo(), actor)
	if err != nil {
		return nil, mapError(err)
	}
	return &case_service.AssignCaseResponse{Case: domainCaseToProto(c)}, nil
}

func (h *Handler) CloseCase(ctx context.Context, req *case_service.CloseCaseRequest) (*case_service.CloseCaseResponse, error) {
	actor := actorFromContext(ctx)
	c, err := h.Service.CloseCase(ctx, req.GetCaseId(), req.GetClosureCode(), req.GetClosureReason(), req.GetVersionNo(), actor)
	if err != nil {
		return nil, mapError(err)
	}
	return &case_service.CloseCaseResponse{Case: domainCaseToProto(c)}, nil
}

func (h *Handler) ListCases(ctx context.Context, req *case_service.ListCasesRequest) (*case_service.ListCasesResponse, error) {
	f := repository.ListCasesFilter{
		PageSize:          req.GetPageSize(),
		PageToken:         req.GetPageToken(),
		State:             req.State,
		Priority:          req.Priority,
		Severity:          req.Severity,
		AssignmentGroupID: req.AssignmentGroupId,
		AssignedUserID:    req.AssignedUserId,
		AffectedUserID:    req.AffectedUserId,
	}
	if t := req.GetOpenedTimeAfter(); t != nil && t.IsValid() {
		s := t.AsTime().Format("2006-01-02T15:04:05Z07:00")
		f.OpenedTimeAfter = &s
	}
	if t := req.GetOpenedTimeBefore(); t != nil && t.IsValid() {
		s := t.AsTime().Format("2006-01-02T15:04:05Z07:00")
		f.OpenedTimeBefore = &s
	}
	items, nextToken, err := h.Service.ListCases(ctx, f)
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]*case_service.Case, len(items))
	for i, c := range items {
		out[i] = domainCaseToProto(c)
	}
	return &case_service.ListCasesResponse{Items: out, NextPageToken: nextToken}, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	slog.Default().With("layer", currentLayer).Error("handler error", "error", err)
	switch {
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, "resource not found")
	case errors.Is(err, repository.ErrVersionConflict):
		return status.Error(codes.FailedPrecondition, "optimistic concurrency conflict")
	case errors.Is(err, domain.ErrInvalidShortDescription),
		errors.Is(err, domain.ErrInvalidState),
		errors.Is(err, domain.ErrInvalidPriority),
		errors.Is(err, domain.ErrInvalidSeverity),
		errors.Is(err, domain.ErrOpenedByUserIDRequired),
		errors.Is(err, domain.ErrOpenedTimeRequired),
		errors.Is(err, domain.ErrClosureRequired),
		errors.Is(err, domain.ErrAlreadyClosed),
		errors.Is(err, domain.ErrTransitionFromClosed),
		errors.Is(err, service.ErrUseCloseCase):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

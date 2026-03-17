package handler

import (
	"time"

	"github.com/servicenow/case-service/domain"
	"github.com/servicenow/case-service/proto/case_service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func domainCaseToProto(c *domain.Case) *case_service.Case {
	if c == nil {
		return nil
	}
	out := &case_service.Case{
		Id:                    c.ID,
		CaseNumber:            c.CaseNumber,
		ShortDescription:      c.ShortDescription,
		Description:           c.Description,
		State:                 c.State,
		Priority:              c.Priority,
		Severity:              c.Severity,
		OpenedByUserId:        c.OpenedByUserID,
		ActiveDurationSeconds: c.ActiveDurationSeconds,
		IsActive:              c.IsActive,
		VersionNo:             c.VersionNo,
		CreatedAt:             timestamppb.New(c.CreatedAt),
		UpdatedAt:             timestamppb.New(c.UpdatedAt),
	}
	if !c.OpenedTime.IsZero() {
		out.OpenedTime = timestamppb.New(c.OpenedTime)
	}
	if c.EventOccurredTime != nil {
		out.EventOccurredTime = timestamppb.New(*c.EventOccurredTime)
	}
	if c.EventReceivedTime != nil {
		out.EventReceivedTime = timestamppb.New(*c.EventReceivedTime)
	}
	if c.AffectedUserID != nil {
		out.AffectedUserId = *c.AffectedUserID
	}
	if c.AssignedUserID != nil {
		out.AssignedUserId = *c.AssignedUserID
	}
	if c.AssignmentGroupID != nil {
		out.AssignmentGroupId = *c.AssignmentGroupID
	}
	if c.AlertRuleID != nil {
		out.AlertRuleId = *c.AlertRuleID
	}
	if c.Accuracy != nil {
		out.Accuracy = *c.Accuracy
	}
	if c.Determination != nil {
		out.Determination = *c.Determination
	}
	if c.Impact != nil {
		out.Impact = *c.Impact
	}
	if c.ClosureCode != nil {
		out.ClosureCode = *c.ClosureCode
	}
	if c.ClosureReason != nil {
		out.ClosureReason = *c.ClosureReason
	}
	if c.ClosedByUserID != nil {
		out.ClosedByUserId = *c.ClosedByUserID
	}
	if c.ClosedTime != nil {
		out.ClosedTime = timestamppb.New(*c.ClosedTime)
	}
	if c.FollowupTime != nil {
		out.FollowupTime = timestamppb.New(*c.FollowupTime)
	}
	if c.Category != nil {
		out.Category = *c.Category
	}
	if c.Subcategory != nil {
		out.Subcategory = *c.Subcategory
	}
	if c.Source != nil {
		out.Source = *c.Source
	}
	if c.SourceTool != nil {
		out.SourceTool = *c.SourceTool
	}
	if c.SourceToolFeature != nil {
		out.SourceToolFeature = *c.SourceToolFeature
	}
	if c.ConfigurationItem != nil {
		out.ConfigurationItem = *c.ConfigurationItem
	}
	if c.SOCNotes != nil {
		out.SocNotes = *c.SOCNotes
	}
	if c.NextSteps != nil {
		out.NextSteps = *c.NextSteps
	}
	if c.CSIRTClassification != nil {
		out.CsirtClassification = *c.CSIRTClassification
	}
	if c.SOCLeadUserID != nil {
		out.SocLeadUserId = *c.SOCLeadUserID
	}
	if c.RequestedByUserID != nil {
		out.RequestedByUserId = *c.RequestedByUserID
	}
	if c.EnvironmentLevel != nil {
		out.EnvironmentLevel = *c.EnvironmentLevel
	}
	if c.EnvironmentType != nil {
		out.EnvironmentType = *c.EnvironmentType
	}
	if c.PDN != nil {
		out.Pdn = *c.PDN
	}
	if c.ImpactedObject != nil {
		out.ImpactedObject = *c.ImpactedObject
	}
	if c.NotificationTime != nil {
		out.NotificationTime = timestamppb.New(*c.NotificationTime)
	}
	if c.MTTR != nil {
		out.Mttr = *c.MTTR
	}
	out.ReassignmentCount = c.ReassignmentCount
	out.AssignedToCount = c.AssignedToCount
	out.IsAffectedUserVip = c.IsAffectedUserVIP
	if c.EngineeringDocument != nil {
		out.EngineeringDocument = *c.EngineeringDocument
	}
	if c.ResponseDocument != nil {
		out.ResponseDocument = *c.ResponseDocument
	}
	return out
}

func domainWorknoteToProto(w *domain.Worknote) *case_service.Worknote {
	if w == nil {
		return nil
	}
	out := &case_service.Worknote{
		Id:              w.ID,
		CaseId:          w.CaseID,
		NoteText:        w.NoteText,
		NoteType:        w.NoteType,
		CreatedByUserId: w.CreatedByUserID,
		CreatedAt:       timestamppb.New(w.CreatedAt),
		IsDeleted:       w.IsDeleted,
	}
	if w.UpdatedAt != nil {
		out.UpdatedAt = timestamppb.New(*w.UpdatedAt)
	}
	return out
}

func protoTimeToDomain(ts *timestamppb.Timestamp) time.Time {
	if ts == nil || !ts.IsValid() {
		return time.Time{}
	}
	return ts.AsTime()
}

func protoTimePtrToDomain(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil || !ts.IsValid() {
		return nil
	}
	t := ts.AsTime()
	return &t
}

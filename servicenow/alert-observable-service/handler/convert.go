package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/org/alert-observable-service/repository"
	pb "github.com/org/alert-observable-service/proto/alert_observable_service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func timeToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func uuidToStr(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func alertToProto(a *repository.Alert) *pb.Alert {
	if a == nil {
		return nil
	}
	rawJSON := ""
	if len(a.RawPayload) > 0 {
		rawJSON = string(a.RawPayload)
	}
	return &pb.Alert{
		Id:                a.ID.String(),
		CaseId:            a.CaseID.String(),
		AlertRuleId:       uuidToStr(a.AlertRuleID),
		SourceSystem:      a.SourceSystem,
		SourceAlertId:     strPtr(a.SourceAlertID),
		Title:             strPtr(a.Title),
		Description:       strPtr(a.Description),
		EventOccurredTime: timeToProto(a.EventOccurredTime),
		EventReceivedTime: timeToProto(a.EventReceivedTime),
		Severity:          strPtr(a.Severity),
		RawPayloadJson:    rawJSON,
		CreatedAt:         timeToProto(&a.CreatedAt),
		UpdatedAt:         timeToProto(&a.UpdatedAt),
	}
}

func observableToProto(o *repository.Observable) *pb.Observable {
	if o == nil {
		return nil
	}
	return &pb.Observable{
		Id:              o.ID.String(),
		ObservableType:  o.ObservableType,
		ObservableValue: o.ObservableValue,
		NormalizedValue: strPtr(o.NormalizedValue),
		FirstSeenTime:   timeToProto(o.FirstSeenTime),
		LastSeenTime:    timeToProto(o.LastSeenTime),
		CreatedAt:       timeToProto(&o.CreatedAt),
		UpdatedAt:       timeToProto(&o.UpdatedAt),
		IncidentCount:   int32(o.IncidentCount),
		Finding:         strPtr(o.Finding),
		Notes:           strPtr(o.Notes),
	}
}

func caseObservableToProto(co *repository.CaseObservable) *pb.CaseObservable {
	if co == nil {
		return nil
	}
	return &pb.CaseObservable{
		Id:             co.ID.String(),
		CaseId:         co.CaseID.String(),
		ObservableId:   co.ObservableID.String(),
		RoleInCase:     strPtr(co.RoleInCase),
		TrackingStatus: strPtr(co.TrackingStatus),
		IsPrimary:      co.IsPrimary,
		Accuracy:       strPtr(co.Accuracy),
		Determination:  strPtr(co.Determination),
		Impact:         strPtr(co.Impact),
		AddedByUserId:  uuidToStr(co.AddedByUserID),
		AddedAt:        timeToProto(&co.AddedAt),
		UpdatedAt:      timeToProto(&co.UpdatedAt),
	}
}

// caseObservableWithDetailsToProto converts CaseObservableWithDetails to proto with observable type/value from JOIN.
func caseObservableWithDetailsToProto(d *repository.CaseObservableWithDetails) *pb.CaseObservable {
	if d == nil {
		return nil
	}
	p := caseObservableToProto(&d.CaseObservable)
	if p == nil {
		return nil
	}
	p.ObservableType = d.ObservableType
	p.ObservableValue = d.ObservableValue
	if p.ObservableValue == "" && d.NormalizedValue != "" {
		p.ObservableValue = d.NormalizedValue
	}
	p.IncidentCount = int32(d.IncidentCount)
	return p
}

func childObservableToProto(c *repository.ChildObservable) *pb.ChildObservable {
	if c == nil {
		return nil
	}
	metaJSON := ""
	if len(c.Metadata) > 0 {
		metaJSON = string(c.Metadata)
	}
	conf := float64(0)
	if c.Confidence != nil {
		conf = *c.Confidence
	}
	return &pb.ChildObservable{
		Id:                   c.ID.String(),
		ParentObservableId:   c.ParentObservableID.String(),
		ChildObservableId:    c.ChildObservableID.String(),
		RelationshipType:     c.RelationshipType,
		RelationshipDirection: strPtr(c.RelationshipDirection),
		Confidence:          conf,
		SourceName:          strPtr(c.SourceName),
		SourceRecordId:      strPtr(c.SourceRecordID),
		MetadataJson:        metaJSON,
		CreatedAt:           timeToProto(&c.CreatedAt),
		UpdatedAt:           timeToProto(&c.UpdatedAt),
	}
}

func similarIncidentToProto(s *repository.SimilarIncident) *pb.SimilarIncident {
	if s == nil {
		return nil
	}
	idsJSON := ""
	if len(s.SharedObservableIDs) > 0 {
		idsJSON = string(s.SharedObservableIDs)
	}
	valsJSON := ""
	if len(s.SharedObservableValues) > 0 {
		valsJSON = string(s.SharedObservableValues)
	}
	score := float64(0)
	if s.SimilarityScore != nil {
		score = *s.SimilarityScore
	}
	return &pb.SimilarIncident{
		Id:                      s.ID.String(),
		CaseId:                  s.CaseID.String(),
		SimilarCaseId:          s.SimilarCaseID.String(),
		MatchReason:             s.MatchReason,
		SharedObservableCount:   int32(s.SharedObservableCount),
		SharedObservableIdsJson: idsJSON,
		SharedObservableValuesJson: valsJSON,
		SimilarityScore:         score,
		LastComputedAt:          timeToProto(&s.LastComputedAt),
		CreatedAt:               timeToProto(&s.CreatedAt),
		UpdatedAt:               timeToProto(&s.UpdatedAt),
	}
}

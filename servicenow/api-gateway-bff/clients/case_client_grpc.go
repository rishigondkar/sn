package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pb "github.com/servicenow/case-service/proto/case_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// seedUserUUID is the POC user from assignment-reference-service seed; case DB expects UUID for opened_by_user_id.
const seedUserUUID = "a0000000-0000-4000-8000-000000000001"

var (
	_ CaseCommandService = (*caseCommandGRPC)(nil)
	_ CaseQueryService   = (*caseQueryGRPC)(nil)
)

// caseCommandGRPC implements CaseCommandService via Case Service gRPC.
type caseCommandGRPC struct {
	client pb.CaseServiceClient
}

// caseQueryGRPC implements CaseQueryService via Case Service gRPC.
type caseQueryGRPC struct {
	client pb.CaseServiceClient
}

// NewCaseClients dials the Case Service and returns command and query clients that use
// the same connection. Outgoing gRPC metadata (x-user-id, x-request-id, x-correlation-id)
// is set from context (see FromContext). Call close when shutting down.
func NewCaseClients(addr string, timeout time.Duration) (CaseCommandService, CaseQueryService, func(), error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("case service dial: %w", err)
	}
	// Use default call timeout via context in each call; connection is shared.
	client := pb.NewCaseServiceClient(conn)
	cmd := &caseCommandGRPC{client: client}
	query := &caseQueryGRPC{client: client}
	closeFn := func() { _ = conn.Close() }
	return cmd, query, closeFn, nil
}

func withMetadata(ctx context.Context) context.Context {
	rc := FromContext(ctx)
	md := metadata.New(nil)
	if rc.ActorUserID != "" {
		md.Set("x-user-id", rc.ActorUserID)
	}
	if rc.RequestID != "" {
		md.Set("x-request-id", rc.RequestID)
	}
	if rc.CorrelationID != "" {
		md.Set("x-correlation-id", rc.CorrelationID)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func protoToCase(c *pb.Case) *Case {
	if c == nil {
		return nil
	}
	out := &Case{
		ID:                   c.GetId(),
		CaseNumber:           c.GetCaseNumber(),
		Title:                c.GetShortDescription(),
		State:                c.GetState(),
		Priority:             c.GetPriority(),
		Severity:             c.GetSeverity(),
		Description:          c.GetDescription(),
		AssignedUserID:       c.GetAssignedUserId(),
		AssignmentGroupID:    c.GetAssignmentGroupId(),
		OpenedByUserID:       c.GetOpenedByUserId(),
		AffectedUserID:       c.GetAffectedUserId(),
		Category:             c.GetCategory(),
		Subcategory:          c.GetSubcategory(),
		Source:               c.GetSource(),
		SourceTool:           c.GetSourceTool(),
		SourceToolFeature:    c.GetSourceToolFeature(),
		ConfigurationItem:    c.GetConfigurationItem(),
		SOCNotes:             c.GetSocNotes(),
		NextSteps:            c.GetNextSteps(),
		CSIRTClassification:  c.GetCsirtClassification(),
		SOCLeadUserID:        c.GetSocLeadUserId(),
		RequestedByUserID:    c.GetRequestedByUserId(),
		EnvironmentLevel:     c.GetEnvironmentLevel(),
		EnvironmentType:     c.GetEnvironmentType(),
		PDN:                  c.GetPdn(),
		ImpactedObject:       c.GetImpactedObject(),
		MTTR:                 c.GetMttr(),
		ReassignmentCount:    c.GetReassignmentCount(),
		AssignedToCount:      c.GetAssignedToCount(),
		IsAffectedUserVIP:    c.GetIsAffectedUserVip(),
		EngineeringDocument:  c.GetEngineeringDocument(),
		ResponseDocument:     c.GetResponseDocument(),
		Accuracy:             c.GetAccuracy(),
		Determination:        c.GetDetermination(),
		ClosureReason:        c.GetClosureReason(),
		ClosedByUserID:       c.GetClosedByUserId(),
	}
	if c.GetOpenedTime() != nil && c.GetOpenedTime().IsValid() {
		t := c.GetOpenedTime().AsTime()
		out.OpenedTime = &t
	}
	if c.GetEventOccurredTime() != nil && c.GetEventOccurredTime().IsValid() {
		t := c.GetEventOccurredTime().AsTime()
		out.EventOccurredTime = &t
	}
	if c.GetNotificationTime() != nil && c.GetNotificationTime().IsValid() {
		t := c.GetNotificationTime().AsTime()
		out.NotificationTime = &t
	}
	if c.GetCreatedAt() != nil {
		out.CreatedAt = c.GetCreatedAt().AsTime()
	}
	if c.GetUpdatedAt() != nil {
		out.UpdatedAt = c.GetUpdatedAt().AsTime()
	}
	if c.GetFollowupTime() != nil && c.GetFollowupTime().IsValid() {
		t := c.GetFollowupTime().AsTime()
		out.FollowupTime = &t
	}
	if c.GetClosedTime() != nil && c.GetClosedTime().IsValid() {
		t := c.GetClosedTime().AsTime()
		out.ClosedTime = &t
	}
	return out
}

func protoToWorknote(w *pb.Worknote) *Worknote {
	if w == nil {
		return nil
	}
	out := &Worknote{
		ID:        w.GetId(),
		CaseID:    w.GetCaseId(),
		Content:   w.GetNoteText(),
		CreatedBy: w.GetCreatedByUserId(),
	}
	if w.GetCreatedAt() != nil {
		out.CreatedAt = w.GetCreatedAt().AsTime()
	}
	return out
}

func (c *caseCommandGRPC) CreateCase(ctx context.Context, req *CreateCaseRequest) (*Case, error) {
	if req == nil {
		return nil, fmt.Errorf("CreateCaseRequest is required")
	}
	ctx = withMetadata(ctx)
	now := timestamppb.Now()
	rc := FromContext(ctx)
	openedBy := rc.ActorUserID
	if openedBy == "" {
		openedBy = seedUserUUID
	} else if _, err := uuid.Parse(openedBy); err != nil {
		// Case DB opened_by_user_id is UUID; non-UUID values (e.g. "frontend-user") map to seed user for POC
		openedBy = seedUserUUID
	}
	pbReq := &pb.CreateCaseRequest{
		ShortDescription:     req.Title,
		Description:          req.Description,
		State:                "Draft",
		Priority:             req.Priority,
		Severity:             req.Severity,
		OpenedByUserId:       openedBy,
		OpenedTime:           now,
		AffectedUserId:       req.AffectedUserID,
		AssignedUserId:       req.AssignedUserID,
		AssignmentGroupId:    req.AssignmentGroupID,
		Category:             req.Category,
		Subcategory:          req.Subcategory,
		Source:               req.Source,
		SourceTool:           req.SourceTool,
		SourceToolFeature:    req.SourceToolFeature,
		ConfigurationItem:    req.ConfigurationItem,
		SocNotes:             req.SOCNotes,
		NextSteps:            req.NextSteps,
		CsirtClassification:  req.CSIRTClassification,
		SocLeadUserId:        req.SOCLeadUserID,
		IsAffectedUserVip:    req.IsAffectedUserVIP,
		RequestedByUserId:   req.RequestedByUserID,
		EnvironmentLevel:    req.EnvironmentLevel,
		EnvironmentType:     req.EnvironmentType,
		Pdn:                  req.PDN,
		ImpactedObject:      req.ImpactedObject,
		Mttr:                 req.MTTR,
		EngineeringDocument: req.EngineeringDocument,
		ResponseDocument:    req.ResponseDocument,
	}
	if req.NotificationTime != nil {
		pbReq.NotificationTime = timestamppb.New(*req.NotificationTime)
	}
	if pbReq.Priority == "" {
		pbReq.Priority = "P3"
	}
	if pbReq.Severity == "" {
		pbReq.Severity = "medium"
	}
	if req.FollowupTime != nil {
		pbReq.FollowupTime = timestamppb.New(*req.FollowupTime)
	}
	resp, err := c.client.CreateCase(ctx, pbReq)
	if err != nil {
		return nil, err
	}
	// Response only has id and case_number; fetch full case for gateway response
	getResp, err := c.client.GetCase(withMetadata(ctx), &pb.GetCaseRequest{CaseId: resp.GetId()})
	if err != nil {
		return &Case{ID: resp.GetId(), CaseNumber: resp.GetCaseNumber(), Title: req.Title, State: "Draft", Priority: pbReq.Priority, CreatedAt: now.AsTime(), UpdatedAt: now.AsTime()}, nil
	}
	return protoToCase(getResp.GetCase()), nil
}

func (c *caseCommandGRPC) UpdateCase(ctx context.Context, caseID string, req *UpdateCaseRequest) (*Case, error) {
	if caseID == "" || req == nil {
		return nil, fmt.Errorf("caseID and UpdateCaseRequest are required")
	}
	ctx = withMetadata(ctx)
	getResp, err := c.client.GetCase(ctx, &pb.GetCaseRequest{CaseId: caseID})
	if err != nil {
		return nil, err
	}
	cur := getResp.GetCase()
	if cur == nil {
		return nil, fmt.Errorf("case not found")
	}
	pbReq := &pb.UpdateCaseRequest{
		Id:         caseID,
		VersionNo:  cur.GetVersionNo(),
	}
	if req.Title != nil {
		s := *req.Title
		pbReq.ShortDescription = &s
	}
	if req.Description != nil {
		pbReq.Description = req.Description
	}
	if req.State != nil {
		pbReq.State = req.State
	}
	if req.Priority != nil {
		pbReq.Priority = req.Priority
	}
	if req.Severity != nil {
		pbReq.Severity = req.Severity
	}
	if req.AffectedUserID != nil {
		pbReq.AffectedUserId = req.AffectedUserID
	}
	if req.AssignedUserID != nil {
		pbReq.AssignedUserId = req.AssignedUserID
	}
	if req.AssignmentGroupID != nil {
		pbReq.AssignmentGroupId = req.AssignmentGroupID
	}
	if req.FollowupTime != nil {
		pbReq.FollowupTime = timestamppb.New(*req.FollowupTime)
	}
	if req.EventOccurredTime != nil {
		pbReq.EventOccurredTime = timestamppb.New(*req.EventOccurredTime)
	}
	if req.Category != nil {
		pbReq.Category = req.Category
	}
	if req.Subcategory != nil {
		pbReq.Subcategory = req.Subcategory
	}
	if req.Source != nil {
		pbReq.Source = req.Source
	}
	if req.SourceTool != nil {
		pbReq.SourceTool = req.SourceTool
	}
	if req.SourceToolFeature != nil {
		pbReq.SourceToolFeature = req.SourceToolFeature
	}
	if req.ConfigurationItem != nil {
		pbReq.ConfigurationItem = req.ConfigurationItem
	}
	if req.SOCNotes != nil {
		pbReq.SocNotes = req.SOCNotes
	}
	if req.NextSteps != nil {
		pbReq.NextSteps = req.NextSteps
	}
	if req.CSIRTClassification != nil {
		pbReq.CsirtClassification = req.CSIRTClassification
	}
	if req.SOCLeadUserID != nil {
		pbReq.SocLeadUserId = req.SOCLeadUserID
	}
	if req.NotificationTime != nil {
		pbReq.NotificationTime = timestamppb.New(*req.NotificationTime)
	}
	if req.IsAffectedUserVIP != nil {
		pbReq.IsAffectedUserVip = req.IsAffectedUserVIP
	}
	if req.RequestedByUserID != nil {
		pbReq.RequestedByUserId = req.RequestedByUserID
	}
	if req.EnvironmentLevel != nil {
		pbReq.EnvironmentLevel = req.EnvironmentLevel
	}
	if req.EnvironmentType != nil {
		pbReq.EnvironmentType = req.EnvironmentType
	}
	if req.PDN != nil {
		pbReq.Pdn = req.PDN
	}
	if req.ImpactedObject != nil {
		pbReq.ImpactedObject = req.ImpactedObject
	}
	if req.MTTR != nil {
		pbReq.Mttr = req.MTTR
	}
	if req.ReassignmentCount != nil {
		pbReq.ReassignmentCount = req.ReassignmentCount
	}
	if req.AssignedToCount != nil {
		pbReq.AssignedToCount = req.AssignedToCount
	}
	if req.EngineeringDocument != nil {
		pbReq.EngineeringDocument = req.EngineeringDocument
	}
	if req.ResponseDocument != nil {
		pbReq.ResponseDocument = req.ResponseDocument
	}
	if req.Accuracy != nil {
		pbReq.Accuracy = req.Accuracy
	}
	if req.Determination != nil {
		pbReq.Determination = req.Determination
	}
	if req.ClosureReason != nil {
		pbReq.ClosureReason = req.ClosureReason
	}
	resp, err := c.client.UpdateCase(ctx, pbReq)
	if err != nil {
		return nil, err
	}
	return protoToCase(resp.GetCase()), nil
}

func (c *caseCommandGRPC) AddWorknote(ctx context.Context, caseID string, req *AddWorknoteRequest) (*Worknote, error) {
	if caseID == "" || req == nil {
		return nil, fmt.Errorf("caseID and AddWorknoteRequest are required")
	}
	ctx = withMetadata(ctx)
	rc := FromContext(ctx)
	createdBy := rc.ActorUserID
	if createdBy == "" {
		createdBy = seedUserUUID
	} else if _, err := uuid.Parse(createdBy); err != nil {
		createdBy = seedUserUUID // case_worknotes.created_by_user_id is UUID
	}
	pbReq := &pb.AddWorknoteRequest{
		CaseId:           caseID,
		NoteText:         req.Content,
		NoteType:         "worknote",
		CreatedByUserId: createdBy,
	}
	resp, err := c.client.AddWorknote(ctx, pbReq)
	if err != nil {
		return nil, err
	}
	return protoToWorknote(resp.GetWorknote()), nil
}

func (c *caseCommandGRPC) AssignCase(ctx context.Context, caseID string, req *AssignCaseRequest) error {
	if caseID == "" || req == nil {
		return fmt.Errorf("caseID and AssignCaseRequest are required")
	}
	ctx = withMetadata(ctx)
	getResp, err := c.client.GetCase(ctx, &pb.GetCaseRequest{CaseId: caseID})
	if err != nil {
		return err
	}
	if getResp.GetCase() == nil {
		return fmt.Errorf("case not found")
	}
	_, err = c.client.AssignCase(ctx, &pb.AssignCaseRequest{
		CaseId:            caseID,
		AssignedUserId:    req.AssignedUserID,
		AssignmentGroupId: req.AssignmentGroupID,
		VersionNo:         getResp.GetCase().GetVersionNo(),
	})
	return err
}

func (c *caseCommandGRPC) CloseCase(ctx context.Context, caseID string, req *CloseCaseRequest) error {
	if caseID == "" || req == nil {
		return fmt.Errorf("caseID and CloseCaseRequest are required")
	}
	ctx = withMetadata(ctx)
	getResp, err := c.client.GetCase(ctx, &pb.GetCaseRequest{CaseId: caseID})
	if err != nil {
		return err
	}
	if getResp.GetCase() == nil {
		return fmt.Errorf("case not found")
	}
	_, err = c.client.CloseCase(ctx, &pb.CloseCaseRequest{
		CaseId:        caseID,
		ClosureCode:   "resolved",
		ClosureReason: req.Resolution,
		VersionNo:     getResp.GetCase().GetVersionNo(),
	})
	return err
}

func (c *caseCommandGRPC) LinkObservable(ctx context.Context, caseID string, req *LinkObservableRequest) error {
	if caseID == "" || req == nil || req.ObservableID == "" {
		return fmt.Errorf("caseID and observableID are required")
	}
	// LinkObservable is implemented by Alert & Observable Service, not Case Service. No-op for case client.
	return nil
}

func (c *caseQueryGRPC) GetCase(ctx context.Context, caseID string) (*Case, error) {
	if caseID == "" {
		return nil, fmt.Errorf("caseID is required")
	}
	ctx = withMetadata(ctx)
	resp, err := c.client.GetCase(ctx, &pb.GetCaseRequest{CaseId: caseID})
	if err != nil {
		return nil, err
	}
	return protoToCase(resp.GetCase()), nil
}

func (c *caseQueryGRPC) ListCases(ctx context.Context, pageSize int32, pageToken string) ([]*Case, string, error) {
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListCases(ctx, &pb.ListCasesRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*Case, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		list = append(list, protoToCase(item))
	}
	return list, resp.GetNextPageToken(), nil
}

func (c *caseQueryGRPC) ListWorknotes(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*Worknote, string, error) {
	if caseID == "" {
		return nil, "", fmt.Errorf("caseID is required")
	}
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListWorknotes(ctx, &pb.ListWorknotesRequest{
		CaseId:    caseID,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*Worknote, 0, len(resp.GetItems()))
	for _, w := range resp.GetItems() {
		list = append(list, protoToWorknote(w))
	}
	return list, resp.GetNextPageToken(), nil
}

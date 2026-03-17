package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pb "github.com/org/alert-observable-service/proto/alert_observable_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	_ ObservableCommandService = (*observableCommandGRPC)(nil)
	_ ObservableQueryService   = (*observableQueryGRPC)(nil)
)

type observableCommandGRPC struct {
	client pb.AlertObservableServiceClient
}

type observableQueryGRPC struct {
	client pb.AlertObservableServiceClient
}

// NewObservableClients dials the Alert & Observable Service and returns command and query clients.
func NewObservableClients(addr string, timeout time.Duration) (ObservableCommandService, ObservableQueryService, func(), error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("observable service dial: %w", err)
	}
	client := pb.NewAlertObservableServiceClient(conn)
	cmd := &observableCommandGRPC{client: client}
	query := &observableQueryGRPC{client: client}
	closeFn := func() { _ = conn.Close() }
	return cmd, query, closeFn, nil
}

func (c *observableCommandGRPC) LinkObservableToCase(ctx context.Context, caseID, observableID string) error {
	if caseID == "" || observableID == "" {
		return fmt.Errorf("caseID and observableID are required")
	}
	ctx = withMetadata(ctx)
	// Backend LinkObservableToCase takes type/value; resolve observable first.
	getResp, err := c.client.GetObservable(ctx, &pb.GetObservableRequest{ObservableId: observableID})
	if err != nil {
		return err
	}
	obs := getResp.GetObservable()
	if obs == nil {
		return fmt.Errorf("observable not found")
	}
	rc := FromContext(ctx)
	addedBy := rc.ActorUserID
	if addedBy != "" {
		if _, err := uuid.Parse(addedBy); err != nil {
			addedBy = seedUserUUID
		}
	} else {
		addedBy = seedUserUUID
	}
	_, err = c.client.LinkObservableToCase(ctx, &pb.LinkObservableToCaseRequest{
		CaseId:          caseID,
		ObservableType:   obs.GetObservableType(),
		ObservableValue:  obs.GetObservableValue(),
		NormalizedValue:  obs.GetNormalizedValue(),
		AddedByUserId:   addedBy,
	})
	return err
}

func (c *observableCommandGRPC) UnlinkObservableFromCase(ctx context.Context, caseID, observableID string) error {
	if caseID == "" || observableID == "" {
		return fmt.Errorf("caseID and observableID are required")
	}
	ctx = withMetadata(ctx)
	_, err := c.client.UnlinkObservableFromCase(ctx, &pb.UnlinkObservableFromCaseRequest{
		CaseId:       caseID,
		ObservableId: observableID,
	})
	return err
}

func (c *observableCommandGRPC) CreateAndLinkObservable(ctx context.Context, caseID, observableType, observableValue string) error {
	if caseID == "" || observableType == "" || observableValue == "" {
		return fmt.Errorf("caseID, observableType and observableValue are required")
	}
	ctx = withMetadata(ctx)
	rc := FromContext(ctx)
	addedBy := rc.ActorUserID
	if addedBy == "" {
		addedBy = seedUserUUID
	} else if _, err := uuid.Parse(addedBy); err != nil {
		addedBy = seedUserUUID
	}
	_, err := c.client.LinkObservableToCase(ctx, &pb.LinkObservableToCaseRequest{
		CaseId:         caseID,
		ObservableType: observableType,
		ObservableValue: observableValue,
		AddedByUserId:  addedBy,
	})
	return err
}

func (c *observableCommandGRPC) CreateOrGetObservable(ctx context.Context, observableType, observableValue string) (*Observable, bool, error) {
	if observableType == "" || observableValue == "" {
		return nil, false, fmt.Errorf("observableType and observableValue are required")
	}
	ctx = withMetadata(ctx)
	resp, err := c.client.CreateOrGetObservable(ctx, &pb.CreateOrGetObservableRequest{
		ObservableType:  observableType,
		ObservableValue: observableValue,
	})
	if err != nil {
		return nil, false, err
	}
	obs := resp.GetObservable()
	if obs == nil {
		return nil, false, fmt.Errorf("observable service returned nil")
	}
	out := &Observable{
		ID:             obs.GetId(),
		Type:           obs.GetObservableType(),
		Value:          obs.GetObservableValue(),
		IncidentCount:  int(obs.GetIncidentCount()),
		Determination:  obs.GetFinding(),
	}
	if obs.GetCreatedAt() != nil {
		out.CreatedAt = obs.GetCreatedAt().AsTime()
	}
	if obs.GetUpdatedAt() != nil {
		out.UpdatedAt = obs.GetUpdatedAt().AsTime()
	}
	return out, resp.GetCreated(), nil
}

func (c *observableCommandGRPC) UpdateObservable(ctx context.Context, observableID string, req *UpdateObservableRequest) (*Observable, error) {
	if observableID == "" {
		return nil, fmt.Errorf("observableID is required")
	}
	ctx = withMetadata(ctx)
	pbReq := &pb.UpdateObservableRequest{ObservableId: observableID}
	if req != nil {
		if req.ObservableValue != nil {
			pbReq.ObservableValue = req.ObservableValue
		}
		if req.ObservableType != nil {
			pbReq.ObservableType = req.ObservableType
		}
		if req.Finding != nil {
			pbReq.Finding = req.Finding
		}
		if req.Notes != nil {
			pbReq.Notes = req.Notes
		}
	}
	resp, err := c.client.UpdateObservable(ctx, pbReq)
	if err != nil {
		return nil, err
	}
	obs := resp.GetObservable()
	if obs == nil {
		return nil, fmt.Errorf("observable service returned nil")
	}
	return protoObservableToObservable(obs), nil
}

func (c *observableQueryGRPC) GetObservable(ctx context.Context, observableID string) (*Observable, error) {
	if observableID == "" {
		return nil, fmt.Errorf("observableID is required")
	}
	ctx = withMetadata(ctx)
	resp, err := c.client.GetObservable(ctx, &pb.GetObservableRequest{ObservableId: observableID})
	if err != nil {
		return nil, err
	}
	obs := resp.GetObservable()
	if obs == nil {
		return nil, fmt.Errorf("observable not found")
	}
	return protoObservableToObservable(obs), nil
}

func (c *observableQueryGRPC) ListObservables(ctx context.Context, search string, pageSize int32, pageToken string) ([]*Observable, string, error) {
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	req := &pb.ListObservablesRequest{PageSize: pageSize, PageToken: pageToken}
	if search != "" {
		req.Search = &search
	}
	resp, err := c.client.ListObservables(ctx, req)
	if err != nil {
		return nil, "", err
	}
	list := make([]*Observable, 0, len(resp.GetItems()))
	for _, o := range resp.GetItems() {
		list = append(list, protoObservableToObservable(o))
	}
	return list, resp.GetNextPageToken(), nil
}

func (c *observableQueryGRPC) ListCaseAlerts(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*Alert, string, error) {
	if caseID == "" {
		return nil, "", fmt.Errorf("caseID is required")
	}
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListCaseAlerts(ctx, &pb.ListCaseAlertsRequest{
		CaseId:    caseID,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*Alert, 0, len(resp.GetItems()))
	for _, a := range resp.GetItems() {
		list = append(list, protoAlertToAlert(a))
	}
	return list, resp.GetNextPageToken(), nil
}

func (c *observableQueryGRPC) ListCaseObservables(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*Observable, string, error) {
	if caseID == "" {
		return nil, "", fmt.Errorf("caseID is required")
	}
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListCaseObservables(ctx, &pb.ListCaseObservablesRequest{
		CaseId:    caseID,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*Observable, 0, len(resp.GetItems()))
	for _, co := range resp.GetItems() {
		list = append(list, protoCaseObservableToObservable(co))
	}
	return list, resp.GetNextPageToken(), nil
}

func (c *observableQueryGRPC) ListChildObservables(ctx context.Context, observableID string, pageSize int32, pageToken string) ([]*ChildObservable, string, error) {
	if observableID == "" {
		return nil, "", fmt.Errorf("observableID is required")
	}
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListChildObservables(ctx, &pb.ListChildObservablesRequest{
		ParentObservableId: observableID,
		PageSize:            pageSize,
		PageToken:           pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*ChildObservable, 0, len(resp.GetItems()))
	for _, ch := range resp.GetItems() {
		list = append(list, protoChildObservableToChildObservable(ch))
	}
	return list, resp.GetNextPageToken(), nil
}

func (c *observableQueryGRPC) ListSimilarIncidents(ctx context.Context, caseID string, pageSize int32, pageToken string) ([]*SimilarIncident, string, error) {
	if caseID == "" {
		return nil, "", fmt.Errorf("caseID is required")
	}
	ctx = withMetadata(ctx)
	if pageSize <= 0 {
		pageSize = 50
	}
	resp, err := c.client.ListSimilarIncidents(ctx, &pb.ListSimilarIncidentsRequest{
		CaseId:    caseID,
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", err
	}
	list := make([]*SimilarIncident, 0, len(resp.GetItems()))
	for _, s := range resp.GetItems() {
		list = append(list, protoSimilarIncidentToSimilarIncident(s))
	}
	return list, resp.GetNextPageToken(), nil
}

func protoAlertToAlert(a *pb.Alert) *Alert {
	if a == nil {
		return nil
	}
	out := &Alert{
		ID:      a.GetId(),
		CaseID:  a.GetCaseId(),
		Summary: a.GetTitle(),
		Severity: a.GetSeverity(),
	}
	if a.GetCreatedAt() != nil {
		out.CreatedAt = a.GetCreatedAt().AsTime()
	}
	return out
}

func protoObservableToObservable(o *pb.Observable) *Observable {
	if o == nil {
		return nil
	}
	out := &Observable{
		ID:            o.GetId(),
		Type:          o.GetObservableType(),
		Value:         o.GetObservableValue(),
		Determination: o.GetFinding(),
		IncidentCount: int(o.GetIncidentCount()),
		Notes:         o.GetNotes(),
	}
	if o.GetCreatedAt() != nil {
		out.CreatedAt = o.GetCreatedAt().AsTime()
	}
	if o.GetUpdatedAt() != nil {
		out.UpdatedAt = o.GetUpdatedAt().AsTime()
	}
	return out
}

func protoCaseObservableToObservable(co *pb.CaseObservable) *Observable {
	if co == nil {
		return nil
	}
	out := &Observable{
		ID:             co.GetObservableId(),
		CaseID:         co.GetCaseId(),
		Type:           co.GetObservableType(),
		Value:          co.GetObservableValue(),
		TrackingStatus: co.GetTrackingStatus(),
		Determination:  co.GetDetermination(),
		IncidentCount:  int(co.GetIncidentCount()),
	}
	if co.GetAddedAt() != nil {
		out.CreatedAt = co.GetAddedAt().AsTime()
	}
	if co.GetUpdatedAt() != nil {
		out.UpdatedAt = co.GetUpdatedAt().AsTime()
	}
	return out
}

func protoChildObservableToChildObservable(ch *pb.ChildObservable) *ChildObservable {
	if ch == nil {
		return nil
	}
	return &ChildObservable{
		ID:       ch.GetChildObservableId(),
		ParentID: ch.GetParentObservableId(),
		Type:     ch.GetRelationshipType(),
		Value:    ch.GetChildObservableId(),
	}
}

func protoSimilarIncidentToSimilarIncident(s *pb.SimilarIncident) *SimilarIncident {
	if s == nil {
		return nil
	}
	out := &SimilarIncident{
		ID:                    s.GetId(),
		CaseID:                s.GetCaseId(),
		SimilarCaseID:         s.GetSimilarCaseId(),
		Summary:               s.GetMatchReason(),
		SharedObservableValues: s.GetSharedObservableValuesJson(),
	}
	if s.GetSimilarityScore() != 0 {
		out.Similarity = fmt.Sprintf("%.2f", s.GetSimilarityScore())
	}
	return out
}

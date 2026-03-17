// Package integration runs gRPC API tests against a live alert-observable-service server.
// Requires server running on localhost:50052 (or set TEST_GRPC_ADDR) and TEST_DB_DSN for server.
// Run: start the server in another terminal, then: go test -v ./integration/ -count=1
package integration

import (
	"context"
	"os"
	"testing"

	pb "github.com/org/alert-observable-service/proto/alert_observable_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func getTestAddr() string {
	if a := os.Getenv("TEST_GRPC_ADDR"); a != "" {
		return a
	}
	return "localhost:50052"
}

func dial(t *testing.T) (*grpc.ClientConn, pb.AlertObservableServiceClient) {
	addr := getTestAddr()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial %s: %v", addr, err)
	}
	return conn, pb.NewAlertObservableServiceClient(conn)
}

func TestCreateAlertRule(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	resp, err := client.CreateAlertRule(ctx, &pb.CreateAlertRuleRequest{
		RuleName: "e2e-rule", RuleType: "test", SourceSystem: "e2e", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateAlertRule: %v", err)
	}
	if resp.GetId() == "" {
		t.Error("expected non-empty rule id")
	}
	t.Logf("CreateAlertRule ok, id=%s", resp.GetId())
}

func TestCreateAlert(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	// Create a rule first for alert_rule_id
	ruleResp, err := client.CreateAlertRule(ctx, &pb.CreateAlertRuleRequest{
		RuleName: "e2e-alert-rule", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateAlertRule: %v", err)
	}
	// Use a random case_id (service may not validate case existence)
	caseID := "018e0000-0000-7000-8000-000000000001"
	resp, err := client.CreateAlert(ctx, &pb.CreateAlertRequest{
		CaseId:       caseID,
		AlertRuleId: ruleResp.GetId(),
		SourceSystem: "e2e",
		Title:        "e2e alert",
		EventOccurredTime: timestamppb.Now(),
		EventReceivedTime: timestamppb.Now(),
	})
	if err != nil {
		t.Fatalf("CreateAlert: %v", err)
	}
	if resp.GetId() == "" {
		t.Error("expected non-empty alert id")
	}
	t.Logf("CreateAlert ok, id=%s", resp.GetId())
}

func TestCreateOrGetObservable(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	resp, err := client.CreateOrGetObservable(ctx, &pb.CreateOrGetObservableRequest{
		ObservableType:  "ipv4",
		ObservableValue: "192.168.1.1",
	})
	if err != nil {
		t.Fatalf("CreateOrGetObservable: %v", err)
	}
	if resp.Observable == nil || resp.Observable.GetId() == "" {
		t.Error("expected non-empty observable in response")
	}
	t.Logf("CreateOrGetObservable ok, id=%s created=%v", resp.Observable.GetId(), resp.Created)
}

func TestGetObservable(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	// Create one first
	cr, err := client.CreateOrGetObservable(ctx, &pb.CreateOrGetObservableRequest{
		ObservableType: "domain", ObservableValue: "e2e-get.example.com",
	})
	if err != nil {
		t.Fatalf("CreateOrGetObservable: %v", err)
	}
	id := cr.Observable.GetId()

	resp, err := client.GetObservable(ctx, &pb.GetObservableRequest{ObservableId: id})
	if err != nil {
		t.Fatalf("GetObservable: %v", err)
	}
	if resp.Observable == nil || resp.Observable.GetId() != id {
		t.Errorf("GetObservable: got %+v", resp)
	}
	t.Logf("GetObservable ok")
}

func TestLinkObservableToCase(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	caseID := "018e0000-0000-7000-8000-000000000002"
	resp, err := client.LinkObservableToCase(ctx, &pb.LinkObservableToCaseRequest{
		CaseId:          caseID,
		ObservableType:  "md5",
		ObservableValue: "e2e123456789012345678901234567890",
		RoleInCase:      "indicator",
	})
	if err != nil {
		t.Fatalf("LinkObservableToCase: %v", err)
	}
	if resp.CaseObservable == nil || resp.CaseObservable.GetId() == "" {
		t.Error("expected case_observable in response")
	}
	t.Logf("LinkObservableToCase ok, co_id=%s", resp.CaseObservable.GetId())
}

func TestListCaseObservables(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	caseID := "018e0000-0000-7000-8000-000000000002"
	resp, err := client.ListCaseObservables(ctx, &pb.ListCaseObservablesRequest{
		CaseId: caseID, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListCaseObservables: %v", err)
	}
	_ = resp.GetItems()
	t.Logf("ListCaseObservables ok, items=%d", len(resp.GetItems()))
}

func TestUpdateCaseObservable(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	caseID := "018e0000-0000-7000-8000-000000000003"
	linkResp, err := client.LinkObservableToCase(ctx, &pb.LinkObservableToCaseRequest{
		CaseId: caseID, ObservableType: "url", ObservableValue: "https://e2e-update.example.com/path",
	})
	if err != nil {
		t.Fatalf("LinkObservableToCase: %v", err)
	}
	coID := linkResp.CaseObservable.GetId()

	status := "under_review"
	resp, err := client.UpdateCaseObservable(ctx, &pb.UpdateCaseObservableRequest{
		CaseObservableId: coID,
		TrackingStatus:   &status,
	})
	if err != nil {
		t.Fatalf("UpdateCaseObservable: %v", err)
	}
	if resp.CaseObservable == nil {
		t.Error("expected case_observable in response")
	}
	t.Logf("UpdateCaseObservable ok")
}

func TestGetAlert(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	ruleResp, _ := client.CreateAlertRule(ctx, &pb.CreateAlertRuleRequest{RuleName: "e2e-get-alert", IsActive: true})
	alertResp, err := client.CreateAlert(ctx, &pb.CreateAlertRequest{
		CaseId: "018e0000-0000-7000-8000-000000000010", AlertRuleId: ruleResp.GetId(),
		SourceSystem: "e2e", Title: "get-alert",
		EventOccurredTime: timestamppb.Now(), EventReceivedTime: timestamppb.Now(),
	})
	if err != nil {
		t.Fatalf("CreateAlert: %v", err)
	}

	resp, err := client.GetAlert(ctx, &pb.GetAlertRequest{AlertId: alertResp.GetId()})
	if err != nil {
		t.Fatalf("GetAlert: %v", err)
	}
	if resp.Alert == nil {
		t.Error("expected alert in response")
	}
	t.Logf("GetAlert ok")
}

func TestListCaseAlerts(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	resp, err := client.ListCaseAlerts(ctx, &pb.ListCaseAlertsRequest{
		CaseId: "018e0000-0000-7000-8000-000000000010", PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListCaseAlerts: %v", err)
	}
	t.Logf("ListCaseAlerts ok, items=%d", len(resp.GetItems()))
}

func TestFindCasesByObservable(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	cr, err := client.CreateOrGetObservable(ctx, &pb.CreateOrGetObservableRequest{
		ObservableType: "email", ObservableValue: "e2e-find@example.com",
	})
	if err != nil {
		t.Fatalf("CreateOrGetObservable: %v", err)
	}

	resp, err := client.FindCasesByObservable(ctx, &pb.FindCasesByObservableRequest{
		ObservableId: cr.Observable.GetId(), PageSize: 10,
	})
	if err != nil {
		t.Fatalf("FindCasesByObservable: %v", err)
	}
	t.Logf("FindCasesByObservable ok, case_ids=%d", len(resp.GetCaseIds()))
}

func TestCreateChildObservableRelation(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	p, err := client.CreateOrGetObservable(ctx, &pb.CreateOrGetObservableRequest{
		ObservableType: "domain", ObservableValue: "parent-e2e.example.com",
	})
	if err != nil || p == nil || p.Observable == nil {
		t.Fatalf("CreateOrGetObservable parent: %v", err)
	}
	c, err := client.CreateOrGetObservable(ctx, &pb.CreateOrGetObservableRequest{
		ObservableType: "ipv4", ObservableValue: "10.0.0.1",
	})
	if err != nil || c == nil || c.Observable == nil {
		t.Fatalf("CreateOrGetObservable child: %v", err)
	}
	resp, err := client.CreateChildObservableRelation(ctx, &pb.CreateChildObservableRelationRequest{
		ParentObservableId: p.Observable.GetId(),
		ChildObservableId:  c.Observable.GetId(),
		RelationshipType:   "resolves_to",
		Confidence:        0.9,
	})
	if err != nil {
		t.Fatalf("CreateChildObservableRelation: %v", err)
	}
	if resp.Relation == nil {
		t.Error("expected relation in response")
	}
	t.Logf("CreateChildObservableRelation ok")
}

func TestListChildObservables(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	cr, err := client.CreateOrGetObservable(ctx, &pb.CreateOrGetObservableRequest{
		ObservableType: "domain", ObservableValue: "parent-list-e2e.example.com",
	})
	if err != nil || cr == nil || cr.Observable == nil {
		t.Fatalf("CreateOrGetObservable: %v", err)
	}
	resp, err := client.ListChildObservables(ctx, &pb.ListChildObservablesRequest{
		ParentObservableId: cr.Observable.GetId(), PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListChildObservables: %v", err)
	}
	t.Logf("ListChildObservables ok, items=%d", len(resp.GetItems()))
}

func TestRecomputeSimilarIncidentsForCase(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	caseID := "018e0000-0000-7000-8000-000000000020"
	_, err := client.RecomputeSimilarIncidentsForCase(ctx, &pb.RecomputeSimilarIncidentsForCaseRequest{CaseId: caseID})
	if err != nil {
		t.Fatalf("RecomputeSimilarIncidentsForCase: %v", err)
	}
	t.Logf("RecomputeSimilarIncidentsForCase ok")
}

func TestListSimilarIncidents(t *testing.T) {
	conn, client := dial(t)
	defer conn.Close()
	ctx := context.Background()

	resp, err := client.ListSimilarIncidents(ctx, &pb.ListSimilarIncidentsRequest{
		CaseId: "018e0000-0000-7000-8000-000000000020", PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListSimilarIncidents: %v", err)
	}
	t.Logf("ListSimilarIncidents ok, items=%d", len(resp.GetItems()))
}

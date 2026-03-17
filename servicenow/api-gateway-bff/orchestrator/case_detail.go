package orchestrator

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/servicenow/api-gateway/clients"
)

var errBadRequest = errors.New("caseId is required")

// CaseDetail is the aggregated view model for GET /cases/{caseId}/detail.
type CaseDetail struct {
	Case          *clients.Case           `json:"case"`
	Worknotes     []*clients.Worknote     `json:"worknotes"`
	Alerts        []*clients.Alert        `json:"alerts"`
	Observables   []*clients.Observable   `json:"observables"`
	ChildObservables []*clients.ChildObservable `json:"child_observables,omitempty"`
	SimilarIncidents []*clients.SimilarIncident `json:"similar_incidents,omitempty"`
	EnrichmentResults []*clients.EnrichmentResult `json:"enrichment_results,omitempty"`
	ThreatLookups  []*clients.ThreatLookupResult `json:"threat_lookups,omitempty"`
	Attachments   []*clients.Attachment   `json:"attachments"`
	AuditEvents   []*clients.AuditEvent   `json:"audit_events"`
	// Expanded reference data (optional)
	AssignedUser   *clients.User `json:"assigned_user,omitempty"`
	AssignmentGroup *clients.Group `json:"assignment_group,omitempty"`
	OpenedByUser   *clients.User `json:"opened_by_user,omitempty"`
	// Section-level degradation: if a section failed, it's logged and optionally flagged
	DegradedSections []string `json:"degraded_sections,omitempty"`
}

const defaultDetailPageSize = 50

// GetCaseDetail fetches case core first, then fans out to supporting services in parallel.
// Tolerates optional section failures and records degraded sections.
func (o *Orchestrator) GetCaseDetail(ctx context.Context, caseID string, logFn func(msg string, args ...any)) (*CaseDetail, error) {
	if caseID == "" {
		return nil, errBadRequest
	}

	// 1. Fetch case core first (required)
	c, err := o.CaseQuery.GetCase(ctx, caseID)
	if err != nil {
		return nil, err
	}
	out := &CaseDetail{Case: c}

	pageSize := int32(defaultDetailPageSize)
	var degraded []string

	var wg sync.WaitGroup
	type result struct {
		name string
		err  error
	}
	results := make(chan result, 9)

	// 2. Fan out in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		worknotes, _, err := o.CaseQuery.ListWorknotes(ctx, caseID, pageSize, "")
		if err != nil {
			results <- result{"worknotes", err}
			return
		}
		out.Worknotes = worknotes
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		alerts, _, err := o.ObsQuery.ListCaseAlerts(ctx, caseID, pageSize, "")
		if err != nil {
			results <- result{"alerts", err}
			return
		}
		out.Alerts = alerts
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		obs, _, err := o.ObsQuery.ListCaseObservables(ctx, caseID, pageSize, "")
		if err != nil {
			results <- result{"observables", err}
			return
		}
		out.Observables = obs
	}()

	// child_observables: we need observable IDs from ListCaseObservables first; do in phase 2

	wg.Add(1)
	go func() {
		defer wg.Done()
		sim, _, err := o.ObsQuery.ListSimilarIncidents(ctx, caseID, pageSize, "")
		if err != nil {
			results <- result{"similar_incidents", err}
			return
		}
		// Enrich with similar case number, title, created
		for i := range sim {
			if sim[i].SimilarCaseID == "" {
				continue
			}
			c, err := o.CaseQuery.GetCase(ctx, sim[i].SimilarCaseID)
			if err != nil {
				continue
			}
			sim[i].SimilarCaseNumber = c.CaseNumber
			sim[i].SimilarCaseTitle = c.Title
			if c.OpenedTime != nil {
				sim[i].SimilarCaseCreatedAt = c.OpenedTime.Format(time.RFC3339)
			}
		}
		out.SimilarIncidents = sim
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		enr, _, err := o.Enrichment.ListEnrichmentResultsByCase(ctx, caseID, pageSize, "")
		if err != nil {
			results <- result{"enrichment_results", err}
			return
		}
		out.EnrichmentResults = enr
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		thr, _, err := o.ThreatLookup.ListThreatLookupResultsByCase(ctx, caseID, pageSize, "")
		if err != nil {
			results <- result{"threat_lookups", err}
			return
		}
		out.ThreatLookups = thr
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		att, _, err := o.Attachment.ListAttachmentsByCase(ctx, caseID, pageSize, "")
		if err != nil {
			results <- result{"attachments", err}
			return
		}
		out.Attachments = att
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		aud, _, err := o.Audit.ListAuditEventsByCase(ctx, caseID, pageSize, "")
		if err != nil {
			results <- result{"audit_events", err}
			return
		}
		out.AuditEvents = aud
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r.err != nil {
			degraded = append(degraded, r.name)
			if logFn != nil {
				logFn("case detail section failed", "section", r.name, "error", r.err)
			}
		}
	}

	// Phase 2: child observables graph summary (needs observable IDs from phase 1)
	if len(out.Observables) > 0 {
		child, _, err := o.ObsQuery.ListChildObservables(ctx, out.Observables[0].ID, pageSize, "")
		if err != nil {
			degraded = append(degraded, "child_observables")
			if logFn != nil {
				logFn("case detail section failed", "section", "child_observables", "error", err)
			}
		} else {
			out.ChildObservables = child
		}
	}

	// Optional: expand reference data
	if c.AssignedUserID != "" {
		if u, err := o.Reference.GetUser(ctx, c.AssignedUserID); err == nil {
			out.AssignedUser = u
		} else {
			degraded = append(degraded, "assigned_user")
		}
	}
	if c.AssignmentGroupID != "" {
		if g, err := o.Reference.GetGroup(ctx, c.AssignmentGroupID); err == nil {
			out.AssignmentGroup = g
		} else {
			degraded = append(degraded, "assignment_group")
		}
	}
	if c.OpenedByUserID != "" {
		if u, err := o.Reference.GetUser(ctx, c.OpenedByUserID); err == nil {
			out.OpenedByUser = u
		} else {
			degraded = append(degraded, "opened_by_user")
		}
	}

	out.DegradedSections = degraded
	return out, nil
}

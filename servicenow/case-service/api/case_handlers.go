package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/servicenow/case-service/domain"
	"github.com/servicenow/case-service/repository"
	"github.com/servicenow/case-service/service"
)

const (
	headerUserID        = "X-User-Id"
	headerRequestID     = "X-Request-Id"
	headerCorrelationID = "X-Correlation-Id"
)

func actorFromRequest(r *http.Request) *service.ActorMeta {
	return &service.ActorMeta{
		UserID:        r.Header.Get(headerUserID),
		RequestID:     r.Header.Get(headerRequestID),
		CorrelationID:  r.Header.Get(headerCorrelationID),
	}
}

type createCaseRequest struct {
	ShortDescription   string     `json:"short_description"`
	Description        string     `json:"description"`
	State              string     `json:"state"`
	Priority           string     `json:"priority"`
	Severity           string     `json:"severity"`
	OpenedByUserID     string     `json:"opened_by_user_id"`
	OpenedTime         time.Time  `json:"opened_time"`
	EventOccurredTime  *time.Time `json:"event_occurred_time,omitempty"`
	EventReceivedTime  *time.Time `json:"event_received_time,omitempty"`
	AffectedUserID     string     `json:"affected_user_id"`
	AssignedUserID     string     `json:"assigned_user_id"`
	AssignmentGroupID   string     `json:"assignment_group_id"`
	AlertRuleID        string     `json:"alert_rule_id"`
	Accuracy           string     `json:"accuracy"`
	Determination      string     `json:"determination"`
	Impact             string     `json:"impact"`
}

func (a *API) CreateCase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req createCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	input := &domain.Case{
		ShortDescription:  req.ShortDescription,
		Description:       req.Description,
		State:             req.State,
		Priority:          req.Priority,
		Severity:          req.Severity,
		OpenedByUserID:    req.OpenedByUserID,
		OpenedTime:        req.OpenedTime,
		EventOccurredTime: req.EventOccurredTime,
		EventReceivedTime: req.EventReceivedTime,
		AffectedUserID:    strPtr(req.AffectedUserID),
		AssignedUserID:    strPtr(req.AssignedUserID),
		AssignmentGroupID: strPtr(req.AssignmentGroupID),
		AlertRuleID:       strPtr(req.AlertRuleID),
		Accuracy:          strPtr(req.Accuracy),
		Determination:     strPtr(req.Determination),
		Impact:            strPtr(req.Impact),
	}
	id, caseNumber, err := a.Service.CreateCase(r.Context(), input, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id, "case_number": caseNumber})
}

func (a *API) GetCase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing case id")
		return
	}
	c, err := a.Service.GetCase(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if c == nil {
		writeError(w, http.StatusNotFound, "case not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(caseToJSON(c))
}

func (a *API) ListCases(w http.ResponseWriter, r *http.Request) {
	f := repository.ListCasesFilter{
		PageSize:  50,
		PageToken: r.URL.Query().Get("page_token"),
	}
	if s := r.URL.Query().Get("state"); s != "" {
		f.State = &s
	}
	if s := r.URL.Query().Get("priority"); s != "" {
		f.Priority = &s
	}
	if s := r.URL.Query().Get("severity"); s != "" {
		f.Severity = &s
	}
	if s := r.URL.Query().Get("assignment_group_id"); s != "" {
		f.AssignmentGroupID = &s
	}
	if s := r.URL.Query().Get("assigned_user_id"); s != "" {
		f.AssignedUserID = &s
	}
	if s := r.URL.Query().Get("affected_user_id"); s != "" {
		f.AffectedUserID = &s
	}
	items, nextToken, err := a.Service.ListCases(r.Context(), f)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]interface{}, len(items))
	for i, c := range items {
		out[i] = caseToJSON(c)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"items": out, "next_page_token": nextToken})
}

func (a *API) PatchCase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing case id")
		return
	}
	var body struct {
		ShortDescription   *string `json:"short_description"`
		Description        *string `json:"description"`
		State              *string `json:"state"`
		Priority           *string `json:"priority"`
		Severity           *string `json:"severity"`
		AffectedUserID     *string `json:"affected_user_id"`
		AssignedUserID     *string `json:"assigned_user_id"`
		AssignmentGroupID  *string `json:"assignment_group_id"`
		Accuracy           *string `json:"accuracy"`
		Determination      *string `json:"determination"`
		Impact             *string `json:"impact"`
		VersionNo          int32   `json:"version_no"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	updates := &service.UpdateCaseInput{
		ShortDescription:   body.ShortDescription,
		Description:        body.Description,
		State:              body.State,
		Priority:           body.Priority,
		Severity:           body.Severity,
		AffectedUserID:    body.AffectedUserID,
		AssignedUserID:    body.AssignedUserID,
		AssignmentGroupID:  body.AssignmentGroupID,
		Accuracy:           body.Accuracy,
		Determination:      body.Determination,
		Impact:             body.Impact,
		VersionNo:          body.VersionNo,
	}
	c, err := a.Service.UpdateCase(r.Context(), id, updates, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(caseToJSON(c))
}

func (a *API) AssignCase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing case id")
		return
	}
	var body struct {
		AssignedUserID    string `json:"assigned_user_id"`
		AssignmentGroupID string `json:"assignment_group_id"`
		VersionNo         int32  `json:"version_no"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	c, err := a.Service.AssignCase(r.Context(), id, body.AssignedUserID, body.AssignmentGroupID, body.VersionNo, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(caseToJSON(c))
}

func (a *API) CloseCase(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing case id")
		return
	}
	var body struct {
		ClosureCode   string `json:"closure_code"`
		ClosureReason string `json:"closure_reason"`
		VersionNo     int32  `json:"version_no"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	c, err := a.Service.CloseCase(r.Context(), id, body.ClosureCode, body.ClosureReason, body.VersionNo, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(caseToJSON(c))
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func caseToJSON(c *domain.Case) map[string]interface{} {
	if c == nil {
		return nil
	}
	m := map[string]interface{}{
		"id": c.ID, "case_number": c.CaseNumber, "short_description": c.ShortDescription,
		"description": c.Description, "state": c.State, "priority": c.Priority, "severity": c.Severity,
		"opened_by_user_id": c.OpenedByUserID, "opened_time": c.OpenedTime,
		"active_duration_seconds": c.ActiveDurationSeconds, "is_active": c.IsActive,
		"version_no": c.VersionNo, "created_at": c.CreatedAt, "updated_at": c.UpdatedAt,
	}
	if c.AffectedUserID != nil {
		m["affected_user_id"] = *c.AffectedUserID
	}
	if c.AssignedUserID != nil {
		m["assigned_user_id"] = *c.AssignedUserID
	}
	if c.AssignmentGroupID != nil {
		m["assignment_group_id"] = *c.AssignmentGroupID
	}
	if c.ClosureCode != nil {
		m["closure_code"] = *c.ClosureCode
	}
	if c.ClosureReason != nil {
		m["closure_reason"] = *c.ClosureReason
	}
	if c.ClosedByUserID != nil {
		m["closed_by_user_id"] = *c.ClosedByUserID
	}
	if c.ClosedTime != nil {
		m["closed_time"] = *c.ClosedTime
	}
	return m
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]string{"code": "ERROR", "message": message}})
}

func writeServiceError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, service.ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, repository.ErrVersionConflict):
		code = http.StatusConflict
	case errors.Is(err, domain.ErrInvalidShortDescription),
		errors.Is(err, domain.ErrInvalidState),
		errors.Is(err, domain.ErrInvalidPriority),
		errors.Is(err, domain.ErrInvalidSeverity),
		errors.Is(err, domain.ErrClosureRequired),
		errors.Is(err, domain.ErrAlreadyClosed),
		errors.Is(err, domain.ErrTransitionFromClosed):
		code = http.StatusBadRequest
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": map[string]string{"code": "ERROR", "message": err.Error()}})
}

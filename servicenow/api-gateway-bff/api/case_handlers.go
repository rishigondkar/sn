package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/servicenow/api-gateway/clients"
)

func (h *Handler) CreateCase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r)
		return
	}
	var req struct {
		Title                string `json:"title"`
		Description          string `json:"description"`
		Priority             string `json:"priority"`
		Severity             string `json:"severity"`
		AffectedUserID       string `json:"affected_user_id"`
		AssignedUserID       string `json:"assigned_user_id"`
		AssignmentGroupID    string `json:"assignment_group_id"`
		FollowupTime         string `json:"followup_time"`
		NotificationTime     string `json:"notification_time"`
		Category             string `json:"category"`
		Subcategory          string `json:"subcategory"`
		Source               string `json:"source"`
		SourceTool           string `json:"source_tool"`
		SourceToolFeature    string `json:"source_tool_feature"`
		ConfigurationItem    string `json:"configuration_item"`
		SOCNotes             string `json:"soc_notes"`
		NextSteps            string `json:"next_steps"`
		CSIRTClassification  string `json:"csirt_classification"`
		SOCLeadUserID        string `json:"soc_lead_user_id"`
		RequestedByUserID    string `json:"requested_by_user_id"`
		EnvironmentLevel     string `json:"environment_level"`
		EnvironmentType      string `json:"environment_type"`
		PDN                  string `json:"pdn"`
		ImpactedObject       string `json:"impacted_object"`
		MTTR                 string `json:"mttr"`
		IsAffectedUserVIP    bool   `json:"is_affected_user_vip"`
		EngineeringDocument  string `json:"engineering_document"`
		ResponseDocument     string `json:"response_document"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, r, "body", "invalid JSON")
		return
	}
	if req.Title == "" {
		writeValidationError(w, r, "title", "required")
		return
	}
	var followup, notificationTime *time.Time
	if req.FollowupTime != "" {
		if t, err := time.Parse(time.RFC3339, req.FollowupTime); err == nil {
			followup = &t
		}
	}
	if req.NotificationTime != "" {
		if t, err := time.Parse(time.RFC3339, req.NotificationTime); err == nil {
			notificationTime = &t
		}
	}
	cmd := &clients.CreateCaseRequest{
		Title:                req.Title,
		Description:          req.Description,
		Priority:             req.Priority,
		Severity:             req.Severity,
		AffectedUserID:       req.AffectedUserID,
		AssignedUserID:       req.AssignedUserID,
		AssignmentGroupID:    req.AssignmentGroupID,
		FollowupTime:         followup,
		NotificationTime:     notificationTime,
		Category:             req.Category,
		Subcategory:          req.Subcategory,
		Source:               req.Source,
		SourceTool:           req.SourceTool,
		SourceToolFeature:    req.SourceToolFeature,
		ConfigurationItem:    req.ConfigurationItem,
		SOCNotes:             req.SOCNotes,
		NextSteps:            req.NextSteps,
		CSIRTClassification: req.CSIRTClassification,
		SOCLeadUserID:        req.SOCLeadUserID,
		IsAffectedUserVIP:    req.IsAffectedUserVIP,
		RequestedByUserID:    req.RequestedByUserID,
		EnvironmentLevel:     req.EnvironmentLevel,
		EnvironmentType:      req.EnvironmentType,
		PDN:                  req.PDN,
		ImpactedObject:       req.ImpactedObject,
		MTTR:                 req.MTTR,
		EngineeringDocument: req.EngineeringDocument,
		ResponseDocument:     req.ResponseDocument,
	}
	if cmd.Priority == "" {
		cmd.Priority = "P3"
	}
	c, err := h.Orch.CaseCmd.CreateCase(r.Context(), cmd)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

func (h *Handler) ListCases(w http.ResponseWriter, r *http.Request) {
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.CaseQuery.ListCases(r.Context(), pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	if list == nil {
		list = []*clients.Case{}
	}
	resp := struct {
		Cases         []*clients.Case `json:"cases"`
		NextPageToken string          `json:"next_page_token,omitempty"`
	}{Cases: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetCase(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	c, err := h.Orch.CaseQuery.GetCase(r.Context(), caseID)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(c)
}

func (h *Handler) UpdateCase(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	var body struct {
		Title                *string `json:"title"`
		Description          *string `json:"description"`
		State                *string `json:"state"`
		Priority             *string `json:"priority"`
		Severity             *string `json:"severity"`
		AffectedUserID       *string `json:"affected_user_id"`
		AssignedUserID       *string `json:"assigned_user_id"`
		AssignmentGroupID    *string `json:"assignment_group_id"`
		FollowupTime         *string `json:"followup_time"`
		NotificationTime     *string `json:"notification_time"`
		EventOccurredTime    *string `json:"event_occurred_time"`
		Category             *string `json:"category"`
		Subcategory          *string `json:"subcategory"`
		Source               *string `json:"source"`
		SourceTool           *string `json:"source_tool"`
		SourceToolFeature    *string `json:"source_tool_feature"`
		ConfigurationItem    *string `json:"configuration_item"`
		SOCNotes             *string `json:"soc_notes"`
		NextSteps            *string `json:"next_steps"`
		CSIRTClassification  *string `json:"csirt_classification"`
		SOCLeadUserID        *string `json:"soc_lead_user_id"`
		RequestedByUserID    *string `json:"requested_by_user_id"`
		EnvironmentLevel     *string `json:"environment_level"`
		EnvironmentType      *string `json:"environment_type"`
		PDN                  *string `json:"pdn"`
		ImpactedObject       *string `json:"impacted_object"`
		MTTR                 *string `json:"mttr"`
		ReassignmentCount    *int32  `json:"reassignment_count"`
		AssignedToCount      *int32  `json:"assigned_to_count"`
		IsAffectedUserVIP    *bool   `json:"is_affected_user_vip"`
		EngineeringDocument  *string `json:"engineering_document"`
		ResponseDocument     *string `json:"response_document"`
		Accuracy             *string `json:"accuracy"`
		Determination        *string `json:"determination"`
		ClosureReason        *string `json:"closure_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeValidationError(w, r, "body", "invalid JSON")
		return
	}
	var followup, notificationTime, eventOccurredTime *time.Time
	if body.FollowupTime != nil && *body.FollowupTime != "" {
		if t, err := time.Parse(time.RFC3339, *body.FollowupTime); err == nil {
			followup = &t
		}
	}
	if body.NotificationTime != nil && *body.NotificationTime != "" {
		if t, err := time.Parse(time.RFC3339, *body.NotificationTime); err == nil {
			notificationTime = &t
		}
	}
	if body.EventOccurredTime != nil && *body.EventOccurredTime != "" {
		if t, err := time.Parse(time.RFC3339, *body.EventOccurredTime); err == nil {
			eventOccurredTime = &t
		}
	}
	cmd := &clients.UpdateCaseRequest{
		Title:                body.Title,
		Description:          body.Description,
		State:                body.State,
		Priority:             body.Priority,
		Severity:             body.Severity,
		AffectedUserID:       body.AffectedUserID,
		AssignedUserID:       body.AssignedUserID,
		AssignmentGroupID:    body.AssignmentGroupID,
		FollowupTime:         followup,
		NotificationTime:     notificationTime,
		EventOccurredTime:    eventOccurredTime,
		Category:             body.Category,
		Subcategory:          body.Subcategory,
		Source:               body.Source,
		SourceTool:           body.SourceTool,
		SourceToolFeature:    body.SourceToolFeature,
		ConfigurationItem:    body.ConfigurationItem,
		SOCNotes:             body.SOCNotes,
		NextSteps:            body.NextSteps,
		CSIRTClassification: body.CSIRTClassification,
		SOCLeadUserID:        body.SOCLeadUserID,
		RequestedByUserID:    body.RequestedByUserID,
		EnvironmentLevel:     body.EnvironmentLevel,
		EnvironmentType:      body.EnvironmentType,
		PDN:                  body.PDN,
		ImpactedObject:       body.ImpactedObject,
		MTTR:                 body.MTTR,
		ReassignmentCount:   body.ReassignmentCount,
		AssignedToCount:     body.AssignedToCount,
		IsAffectedUserVIP:   body.IsAffectedUserVIP,
		EngineeringDocument: body.EngineeringDocument,
		ResponseDocument:    body.ResponseDocument,
		Accuracy:            body.Accuracy,
		Determination:       body.Determination,
		ClosureReason:       body.ClosureReason,
	}
	c, err := h.Orch.CaseCmd.UpdateCase(r.Context(), caseID, cmd)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(c)
}

func (h *Handler) AddWorknote(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, r, "body", "invalid JSON")
		return
	}
	if req.Content == "" {
		writeValidationError(w, r, "content", "required")
		return
	}
	note, err := h.Orch.CaseCmd.AddWorknote(r.Context(), caseID, &clients.AddWorknoteRequest{Content: req.Content})
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(note)
}

func (h *Handler) ListWorknotes(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.CaseQuery.ListWorknotes(r.Context(), caseID, pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	resp := struct {
		Worknotes     []*clients.Worknote `json:"worknotes"`
		NextPageToken string              `json:"next_page_token,omitempty"`
	}{Worknotes: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) AssignCase(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	var req struct {
		AssignedUserID    string `json:"assigned_user_id"`
		AssignmentGroupID string `json:"assignment_group_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeValidationError(w, r, "body", "invalid JSON")
		return
	}
	err := h.Orch.CaseCmd.AssignCase(r.Context(), caseID, &clients.AssignCaseRequest{
		AssignedUserID:    req.AssignedUserID,
		AssignmentGroupID: req.AssignmentGroupID,
	})
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CloseCase(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	var req struct {
		Resolution string `json:"resolution"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	err := h.Orch.CaseCmd.CloseCase(r.Context(), caseID, &clients.CloseCaseRequest{Resolution: req.Resolution})
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetCaseDetail(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		writeValidationError(w, r, "caseId", "required")
		return
	}
	detail, err := h.Orch.GetCaseDetail(r.Context(), caseID, nil)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(detail)
}

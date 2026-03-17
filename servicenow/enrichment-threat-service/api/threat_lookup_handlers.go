package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/servicenow/enrichment-threat-service/repository"
	"github.com/servicenow/enrichment-threat-service/service"
)

type threatLookupBody struct {
	ID                   string   `json:"id"`
	CaseID               *string  `json:"case_id,omitempty"`
	ObservableID         string   `json:"observable_id"`
	LookupType           string   `json:"lookup_type"`
	SourceName           string   `json:"source_name"`
	SourceRecordID       *string  `json:"source_record_id,omitempty"`
	Verdict              *string `json:"verdict,omitempty"`
	RiskScore            *float64 `json:"risk_score,omitempty"`
	ConfidenceScore      *float64 `json:"confidence_score,omitempty"`
	Tags                 any     `json:"tags,omitempty"`
	MatchedIndicators    any     `json:"matched_indicators,omitempty"`
	Summary              *string  `json:"summary,omitempty"`
	ResultData           any     `json:"result_data"`
	QueriedAt            *string  `json:"queried_at,omitempty"`
	ReceivedAt           string   `json:"received_at"`
	ExpiresAt            *string  `json:"expires_at,omitempty"`
}

func (a *API) PostThreatLookup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	requestID, correlationID := getIDs(r)
	var body threatLookupBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteValidationError(w, "invalid JSON", nil, requestID, correlationID)
		return
	}
	resultDataJSON, err := json.Marshal(body.ResultData)
	if err != nil {
		WriteValidationError(w, "result_data must be valid JSON", nil, requestID, correlationID)
		return
	}
	receivedAt, err := parseTime(body.ReceivedAt)
	if err != nil || receivedAt == nil {
		WriteValidationError(w, "received_at required (RFC3339)", nil, requestID, correlationID)
		return
	}
	var tagsJSON, matchedJSON *string
	if body.Tags != nil {
		b, _ := json.Marshal(body.Tags)
		s := string(b)
		tagsJSON = &s
	}
	if body.MatchedIndicators != nil {
		b, _ := json.Marshal(body.MatchedIndicators)
		s := string(b)
		matchedJSON = &s
	}
	var queriedAt, expiresAt *time.Time
	if body.QueriedAt != nil {
		queriedAt, _ = parseTime(*body.QueriedAt)
	}
	if body.ExpiresAt != nil {
		expiresAt, _ = parseTime(*body.ExpiresAt)
	}

	in := service.UpsertThreatLookupInput{
		ID:                   body.ID,
		CaseID:               body.CaseID,
		ObservableID:         body.ObservableID,
		LookupType:           body.LookupType,
		SourceName:           body.SourceName,
		SourceRecordID:       body.SourceRecordID,
		Verdict:              body.Verdict,
		RiskScore:            body.RiskScore,
		ConfidenceScore:      body.ConfidenceScore,
		TagsJSON:             tagsJSON,
		MatchedIndicatorsJSON: matchedJSON,
		Summary:              body.Summary,
		ResultDataJSON:       string(resultDataJSON),
		QueriedAt:            queriedAt,
		ReceivedAt:           *receivedAt,
		ExpiresAt:            expiresAt,
	}
	actor := service.ActorInfo{UserID: getUserID(r), RequestID: requestID, CorrelationID: correlationID}
	result, err := a.Service.UpsertThreatLookupResult(r.Context(), in, actor)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			WriteValidationError(w, "validation failed: observable_id required", nil, requestID, correlationID)
			return
		}
		WriteInternal(w, requestID, correlationID)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(threatLookupToJSON(result))
}

func (a *API) PutThreatLookup(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		reqID, corrID := getIDs(r)
		WriteValidationError(w, "id required", nil, reqID, corrID)
		return
	}
	requestID, correlationID := getIDs(r)
	var body threatLookupBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteValidationError(w, "invalid JSON", nil, requestID, correlationID)
		return
	}
	resultDataJSON, err := json.Marshal(body.ResultData)
	if err != nil {
		WriteValidationError(w, "result_data must be valid JSON", nil, requestID, correlationID)
		return
	}
	receivedAt, err := parseTime(body.ReceivedAt)
	if err != nil || receivedAt == nil {
		WriteValidationError(w, "received_at required (RFC3339)", nil, requestID, correlationID)
		return
	}
	var tagsJSON, matchedJSON *string
	if body.Tags != nil {
		b, _ := json.Marshal(body.Tags)
		s := string(b)
		tagsJSON = &s
	}
	if body.MatchedIndicators != nil {
		b, _ := json.Marshal(body.MatchedIndicators)
		s := string(b)
		matchedJSON = &s
	}
	var queriedAt, expiresAt *time.Time
	if body.QueriedAt != nil {
		queriedAt, _ = parseTime(*body.QueriedAt)
	}
	if body.ExpiresAt != nil {
		expiresAt, _ = parseTime(*body.ExpiresAt)
	}

	in := service.UpsertThreatLookupInput{
		ID:                   id,
		CaseID:               body.CaseID,
		ObservableID:         body.ObservableID,
		LookupType:           body.LookupType,
		SourceName:           body.SourceName,
		SourceRecordID:       body.SourceRecordID,
		Verdict:              body.Verdict,
		RiskScore:            body.RiskScore,
		ConfidenceScore:      body.ConfidenceScore,
		TagsJSON:             tagsJSON,
		MatchedIndicatorsJSON: matchedJSON,
		Summary:              body.Summary,
		ResultDataJSON:       string(resultDataJSON),
		QueriedAt:            queriedAt,
		ReceivedAt:           *receivedAt,
		ExpiresAt:            expiresAt,
	}
	actor := service.ActorInfo{UserID: getUserID(r), RequestID: requestID, CorrelationID: correlationID}
	result, err := a.Service.UpsertThreatLookupResult(r.Context(), in, actor)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			WriteValidationError(w, "validation failed", nil, requestID, correlationID)
			return
		}
		WriteInternal(w, requestID, correlationID)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(threatLookupToJSON(result))
}

func (a *API) GetThreatLookupsByCase(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		reqID, corrID := getIDs(r)
		WriteValidationError(w, "caseId required", nil, reqID, corrID)
		return
	}
	f := listFilterFromQueryThreat(r)
	items, nextToken, err := a.Service.ListThreatLookupResultsByCase(r.Context(), caseID, f)
	if err != nil {
		reqID, corrID := getIDs(r)
		WriteInternal(w, reqID, corrID)
		return
	}
	out := map[string]interface{}{
		"items":            threatLookupsToJSON(items),
		"next_page_token": nextToken,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}

func (a *API) GetThreatLookupsByObservable(w http.ResponseWriter, r *http.Request) {
	observableID := r.PathValue("observableId")
	if observableID == "" {
		reqID, corrID := getIDs(r)
		WriteValidationError(w, "observableId required", nil, reqID, corrID)
		return
	}
	f := listFilterFromQueryThreat(r)
	items, nextToken, err := a.Service.ListThreatLookupResultsByObservable(r.Context(), observableID, f)
	if err != nil {
		reqID, corrID := getIDs(r)
		WriteInternal(w, reqID, corrID)
		return
	}
	out := map[string]interface{}{
		"items":            threatLookupsToJSON(items),
		"next_page_token": nextToken,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}

func listFilterFromQueryThreat(r *http.Request) repository.ListFilter {
	f := repository.ListFilter{}
	if v := r.URL.Query().Get("page_size"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			f.PageSize = int32(n)
			if f.PageSize > 100 {
				f.PageSize = 100
			}
		}
	}
	f.PageToken = r.URL.Query().Get("page_token")
	if v := r.URL.Query().Get("source_name"); v != "" {
		f.SourceName = &v
	}
	if v := r.URL.Query().Get("lookup_type"); v != "" {
		f.Type = &v
	}
	if v := r.URL.Query().Get("verdict"); v != "" {
		f.Verdict = &v
	}
	f.ActiveOnly = r.URL.Query().Get("active_only") == "true" || r.URL.Query().Get("active_only") == "1"
	return f
}

func threatLookupToJSON(t *repository.ThreatLookupResult) map[string]interface{} {
	if t == nil {
		return nil
	}
	out := map[string]interface{}{
		"id":             t.ID,
		"observable_id":  t.ObservableID,
		"lookup_type":    t.LookupType,
		"source_name":   t.SourceName,
		"result_data":   json.RawMessage(t.ResultData),
		"received_at":   t.ReceivedAt.Format(time.RFC3339),
		"created_at":     t.CreatedAt.Format(time.RFC3339),
		"updated_at":     t.UpdatedAt.Format(time.RFC3339),
	}
	if t.CaseID != nil {
		out["case_id"] = *t.CaseID
	}
	if t.SourceRecordID != nil {
		out["source_record_id"] = *t.SourceRecordID
	}
	if t.Verdict != nil {
		out["verdict"] = *t.Verdict
	}
	if t.RiskScore != nil {
		out["risk_score"] = *t.RiskScore
	}
	if t.ConfidenceScore != nil {
		out["confidence_score"] = *t.ConfidenceScore
	}
	if t.Tags != nil {
		out["tags"] = json.RawMessage(t.Tags)
	}
	if t.MatchedIndicators != nil {
		out["matched_indicators"] = json.RawMessage(t.MatchedIndicators)
	}
	if t.Summary != nil {
		out["summary"] = *t.Summary
	}
	if t.QueriedAt != nil {
		out["queried_at"] = t.QueriedAt.Format(time.RFC3339)
	}
	if t.ExpiresAt != nil {
		out["expires_at"] = t.ExpiresAt.Format(time.RFC3339)
	}
	return out
}

func threatLookupsToJSON(items []*repository.ThreatLookupResult) []map[string]interface{} {
	if items == nil {
		return nil
	}
	out := make([]map[string]interface{}, len(items))
	for i, t := range items {
		out[i] = threatLookupToJSON(t)
	}
	return out
}

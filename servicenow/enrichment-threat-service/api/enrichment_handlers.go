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

type enrichmentResultBody struct {
	ID              string   `json:"id"`
	CaseID          *string  `json:"case_id,omitempty"`
	ObservableID    *string  `json:"observable_id,omitempty"`
	EnrichmentType  string   `json:"enrichment_type"`
	SourceName      string   `json:"source_name"`
	SourceRecordID  *string  `json:"source_record_id,omitempty"`
	Status          string   `json:"status"`
	Summary         *string  `json:"summary,omitempty"`
	ResultData      any      `json:"result_data"` // encoded to JSON for storage
	Score           *float64 `json:"score,omitempty"`
	Confidence      *float64 `json:"confidence,omitempty"`
	RequestedAt     *string  `json:"requested_at,omitempty"` // RFC3339
	ReceivedAt     string   `json:"received_at"`            // RFC3339
	ExpiresAt       *string  `json:"expires_at,omitempty"`
	LastUpdatedBy   *string  `json:"last_updated_by,omitempty"`
}

func parseTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (a *API) PostEnrichmentResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	requestID, correlationID := getIDs(r)
	var body enrichmentResultBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteValidationError(w, "invalid JSON", []DetailEntry{{Field: "body", Issue: "invalid"}}, requestID, correlationID)
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
	var requestedAt, expiresAt *time.Time
	if body.RequestedAt != nil {
		requestedAt, _ = parseTime(*body.RequestedAt)
	}
	if body.ExpiresAt != nil {
		expiresAt, _ = parseTime(*body.ExpiresAt)
	}

	in := service.UpsertEnrichmentInput{
		ID:             body.ID,
		CaseID:         body.CaseID,
		ObservableID:   body.ObservableID,
		EnrichmentType: body.EnrichmentType,
		SourceName:     body.SourceName,
		SourceRecordID: body.SourceRecordID,
		Status:         body.Status,
		Summary:        body.Summary,
		ResultDataJSON: string(resultDataJSON),
		Score:          body.Score,
		Confidence:     body.Confidence,
		RequestedAt:    requestedAt,
		ReceivedAt:     *receivedAt,
		ExpiresAt:      expiresAt,
		LastUpdatedBy:  body.LastUpdatedBy,
	}
	actor := service.ActorInfo{UserID: getUserID(r), RequestID: requestID, CorrelationID: correlationID}
	result, err := a.Service.UpsertEnrichmentResult(r.Context(), in, actor)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			WriteValidationError(w, "validation failed", nil, requestID, correlationID)
			return
		}
		WriteInternal(w, requestID, correlationID)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(enrichmentToJSON(result))
}

func (a *API) PutEnrichmentResult(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		reqID, corrID := getIDs(r)
		WriteValidationError(w, "id required", nil, reqID, corrID)
		return
	}
	requestID, correlationID := getIDs(r)
	var body enrichmentResultBody
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
	var requestedAt, expiresAt *time.Time
	if body.RequestedAt != nil {
		requestedAt, _ = parseTime(*body.RequestedAt)
	}
	if body.ExpiresAt != nil {
		expiresAt, _ = parseTime(*body.ExpiresAt)
	}

	in := service.UpsertEnrichmentInput{
		ID:             id,
		CaseID:         body.CaseID,
		ObservableID:   body.ObservableID,
		EnrichmentType: body.EnrichmentType,
		SourceName:     body.SourceName,
		SourceRecordID: body.SourceRecordID,
		Status:         body.Status,
		Summary:        body.Summary,
		ResultDataJSON: string(resultDataJSON),
		Score:          body.Score,
		Confidence:     body.Confidence,
		RequestedAt:    requestedAt,
		ReceivedAt:     *receivedAt,
		ExpiresAt:      expiresAt,
		LastUpdatedBy:  body.LastUpdatedBy,
	}
	actor := service.ActorInfo{UserID: getUserID(r), RequestID: requestID, CorrelationID: correlationID}
	result, err := a.Service.UpsertEnrichmentResult(r.Context(), in, actor)
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
	_ = json.NewEncoder(w).Encode(enrichmentToJSON(result))
}

func (a *API) GetEnrichmentResultsByCase(w http.ResponseWriter, r *http.Request) {
	caseID := r.PathValue("caseId")
	if caseID == "" {
		reqID, corrID := getIDs(r)
		WriteValidationError(w, "caseId required", nil, reqID, corrID)
		return
	}
	f := listFilterFromQueryEnrichment(r)
	items, nextToken, err := a.Service.ListEnrichmentResultsByCase(r.Context(), caseID, f)
	if err != nil {
		reqID, corrID := getIDs(r)
		WriteInternal(w, reqID, corrID)
		return
	}
	out := map[string]interface{}{
		"items":          enrichmentsToJSON(items),
		"next_page_token": nextToken,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}

func (a *API) GetEnrichmentResultsByObservable(w http.ResponseWriter, r *http.Request) {
	observableID := r.PathValue("observableId")
	if observableID == "" {
		reqID, corrID := getIDs(r)
		WriteValidationError(w, "observableId required", nil, reqID, corrID)
		return
	}
	f := listFilterFromQueryEnrichment(r)
	items, nextToken, err := a.Service.ListEnrichmentResultsByObservable(r.Context(), observableID, f)
	if err != nil {
		reqID, corrID := getIDs(r)
		WriteInternal(w, reqID, corrID)
		return
	}
	out := map[string]interface{}{
		"items":           enrichmentsToJSON(items),
		"next_page_token": nextToken,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}

func getIDs(r *http.Request) (requestID, correlationID string) {
	requestID, _ = r.Context().Value(ctxKeyRequestID).(string)
	correlationID, _ = r.Context().Value(ctxKeyCorrelationID).(string)
	return requestID, correlationID
}

func getUserID(r *http.Request) string {
	u, _ := r.Context().Value(ctxKeyUserID).(string)
	return u
}

func listFilterFromQueryEnrichment(r *http.Request) repository.ListFilter {
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
	if v := r.URL.Query().Get("enrichment_type"); v != "" {
		f.Type = &v
	}
	f.ActiveOnly = r.URL.Query().Get("active_only") == "true" || r.URL.Query().Get("active_only") == "1"
	return f
}

func enrichmentToJSON(e *repository.EnrichmentResult) map[string]interface{} {
	if e == nil {
		return nil
	}
	out := map[string]interface{}{
		"id":              e.ID,
		"enrichment_type": e.EnrichmentType,
		"source_name":     e.SourceName,
		"status":          e.Status,
		"result_data":     json.RawMessage(e.ResultData),
		"received_at":     e.ReceivedAt.Format(time.RFC3339),
		"created_at":      e.CreatedAt.Format(time.RFC3339),
		"updated_at":      e.UpdatedAt.Format(time.RFC3339),
	}
	if e.CaseID != nil {
		out["case_id"] = *e.CaseID
	}
	if e.ObservableID != nil {
		out["observable_id"] = *e.ObservableID
	}
	if e.SourceRecordID != nil {
		out["source_record_id"] = *e.SourceRecordID
	}
	if e.Summary != nil {
		out["summary"] = *e.Summary
	}
	if e.Score != nil {
		out["score"] = *e.Score
	}
	if e.Confidence != nil {
		out["confidence"] = *e.Confidence
	}
	if e.RequestedAt != nil {
		out["requested_at"] = e.RequestedAt.Format(time.RFC3339)
	}
	if e.ExpiresAt != nil {
		out["expires_at"] = e.ExpiresAt.Format(time.RFC3339)
	}
	if e.LastUpdatedBy != nil {
		out["last_updated_by"] = *e.LastUpdatedBy
	}
	return out
}

func enrichmentsToJSON(items []*repository.EnrichmentResult) []map[string]interface{} {
	if items == nil {
		return nil
	}
	out := make([]map[string]interface{}, len(items))
	for i, e := range items {
		out[i] = enrichmentToJSON(e)
	}
	return out
}

package api

import (
	"encoding/json"
	"net/http"

	"github.com/servicenow/case-service/domain"
)

func (a *API) AddWorknote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing case id")
		return
	}
	var body struct {
		NoteText        string `json:"note_text"`
		NoteType        string `json:"note_type"`
		CreatedByUserID string `json:"created_by_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.CreatedByUserID == "" {
		writeError(w, http.StatusBadRequest, "created_by_user_id is required")
		return
	}
	noteType := body.NoteType
	if noteType == "" {
		noteType = domain.DefaultNoteType
	}
	w2, err := a.Service.AddWorknote(r.Context(), id, body.NoteText, noteType, body.CreatedByUserID, actorFromRequest(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(worknoteToJSON(w2))
}

func (a *API) ListWorknotes(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing case id")
		return
	}
	pageSize := int32(50)
	pageToken := r.URL.Query().Get("page_token")
	items, nextToken, err := a.Service.ListWorknotes(r.Context(), id, pageSize, pageToken)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]interface{}, len(items))
	for i, w := range items {
		out[i] = worknoteToJSON(w)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"items": out, "next_page_token": nextToken})
}

func worknoteToJSON(w *domain.Worknote) map[string]interface{} {
	if w == nil {
		return nil
	}
	m := map[string]interface{}{
		"id": w.ID, "case_id": w.CaseID, "note_text": w.NoteText, "note_type": w.NoteType,
		"created_by_user_id": w.CreatedByUserID, "created_at": w.CreatedAt, "is_deleted": w.IsDeleted,
	}
	if w.UpdatedAt != nil {
		m["updated_at"] = *w.UpdatedAt
	}
	return m
}

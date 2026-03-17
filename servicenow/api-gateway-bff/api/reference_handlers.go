package api

import (
	"encoding/json"
	"net/http"

	"github.com/servicenow/api-gateway/clients"
)

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.Reference.ListUsers(r.Context(), pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	resp := struct {
		Users         []*clients.User `json:"users"`
		NextPageToken string          `json:"next_page_token,omitempty"`
	}{Users: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	pageSize, pageToken := parsePagination(r)
	list, nextToken, err := h.Orch.Reference.ListGroups(r.Context(), pageSize, pageToken)
	if err != nil {
		writeDownstreamError(w, r, err)
		return
	}
	if list == nil {
		list = []*clients.Group{}
	}
	resp := struct {
		Groups        []*clients.Group `json:"groups"`
		NextPageToken string          `json:"next_page_token,omitempty"`
	}{Groups: list, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

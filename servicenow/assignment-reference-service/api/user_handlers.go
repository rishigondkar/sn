package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/servicenow/assignment-reference-service/repository"
	"github.com/servicenow/assignment-reference-service/service"
)

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-Id")
	correlationID := r.Header.Get("X-Correlation-Id")

	pageSize := int32(50)
	if v := r.URL.Query().Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			pageSize = int32(n)
		}
	}
	pageToken := r.URL.Query().Get("page_token")
	activeOnly := true
	if v := r.URL.Query().Get("active_only"); v == "false" || v == "0" {
		activeOnly = false
	}
	var filterDisplayName, filterUsername, filterEmail *string
	if v := r.URL.Query().Get("filter_display_name"); v != "" {
		filterDisplayName = &v
	}
	if v := r.URL.Query().Get("filter_username"); v != "" {
		filterUsername = &v
	}
	if v := r.URL.Query().Get("filter_email"); v != "" {
		filterEmail = &v
	}

	items, nextToken, err := s.Service.ListUsers(r.Context(),
		pageSize, pageToken, &activeOnly, filterDisplayName, filterUsername, filterEmail)
	if err != nil {
		code := "INTERNAL_ERROR"
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrUserNotFound) {
			code, statusCode = "NOT_FOUND", http.StatusNotFound
		} else if errors.Is(err, service.ErrInvalidEmail) || errors.Is(err, service.ErrUsernameTaken) || errors.Is(err, service.ErrEmailTaken) {
			code, statusCode = "VALIDATION_ERROR", http.StatusBadRequest
		}
		writeRESTError(w, code, err.Error(), requestID, correlationID, statusCode)
		return
	}

	out := make([]userResp, len(items))
	for i, u := range items {
		out[i] = userToResp(u)
	}
	resp := listUsersResp{Items: out, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

type userResp struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type listUsersResp struct {
	Items         []userResp `json:"items"`
	NextPageToken string     `json:"next_page_token,omitempty"`
}

func userToResp(u *repository.User) userResp {
	if u == nil {
		return userResp{}
	}
	return userResp{
		ID: u.ID, Username: u.Username, Email: u.Email, DisplayName: u.DisplayName,
		IsActive: u.IsActive,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"errors"

	"github.com/servicenow/assignment-reference-service/repository"
	"github.com/servicenow/assignment-reference-service/service"
)

func (s *Server) listGroups(w http.ResponseWriter, r *http.Request) {
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
	var filterGroupName *string
	if v := r.URL.Query().Get("filter_group_name"); v != "" {
		filterGroupName = &v
	}

	items, nextToken, err := s.Service.ListGroups(r.Context(),
		pageSize, pageToken, &activeOnly, filterGroupName)
	if err != nil {
		code := "INTERNAL_ERROR"
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrGroupNotFound) {
			code, statusCode = "NOT_FOUND", http.StatusNotFound
		} else if errors.Is(err, service.ErrGroupNameTaken) {
			code, statusCode = "VALIDATION_ERROR", http.StatusBadRequest
		}
		writeRESTError(w, code, err.Error(), requestID, correlationID, statusCode)
		return
	}

	out := make([]groupResp, len(items))
	for i, g := range items {
		out[i] = groupToResp(g)
	}
	resp := listGroupsResp{Items: out, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) listGroupMembers(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-Id")
	correlationID := r.Header.Get("X-Correlation-Id")
	groupID := r.PathValue("groupId")
	if groupID == "" {
		writeRESTError(w, "VALIDATION_ERROR", "missing groupId", requestID, correlationID, http.StatusBadRequest)
		return
	}

	pageSize := int32(50)
	if v := r.URL.Query().Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			pageSize = int32(n)
		}
	}
	pageToken := r.URL.Query().Get("page_token")

	items, nextToken, err := s.Service.ListGroupMembers(r.Context(), groupID, pageSize, pageToken)
	if err != nil {
		code := "INTERNAL_ERROR"
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrGroupNotFound) {
			code, statusCode = "NOT_FOUND", http.StatusNotFound
		}
		writeRESTError(w, code, err.Error(), requestID, correlationID, statusCode)
		return
	}

	out := make([]groupMemberResp, len(items))
	for i, m := range items {
		out[i] = groupMemberToResp(m)
	}
	resp := listGroupMembersResp{Items: out, NextPageToken: nextToken}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

type groupResp struct {
	ID          string `json:"id"`
	GroupName   string `json:"group_name"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type listGroupsResp struct {
	Items         []groupResp `json:"items"`
	NextPageToken string      `json:"next_page_token,omitempty"`
}

func groupToResp(g *repository.Group) groupResp {
	if g == nil {
		return groupResp{}
	}
	return groupResp{
		ID: g.ID, GroupName: g.GroupName, Description: g.Description,
		IsActive: g.IsActive,
		CreatedAt: g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: g.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type groupMemberResp struct {
	ID         string `json:"id"`
	GroupID    string `json:"group_id"`
	UserID     string `json:"user_id"`
	MemberRole string `json:"member_role,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type listGroupMembersResp struct {
	Items         []groupMemberResp `json:"items"`
	NextPageToken string           `json:"next_page_token,omitempty"`
}

func groupMemberToResp(m *repository.GroupMember) groupMemberResp {
	if m == nil {
		return groupMemberResp{}
	}
	return groupMemberResp{
		ID: m.ID, GroupID: m.GroupID, UserID: m.UserID, MemberRole: m.MemberRole,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

package handler

import (
	"context"

	"github.com/servicenow/case-service/domain"
	"github.com/servicenow/case-service/proto/case_service"
)

func (h *Handler) AddWorknote(ctx context.Context, req *case_service.AddWorknoteRequest) (*case_service.AddWorknoteResponse, error) {
	actor := actorFromContext(ctx)
	noteType := req.GetNoteType()
	if noteType == "" {
		noteType = domain.DefaultNoteType
	}
	w, err := h.Service.AddWorknote(ctx, req.GetCaseId(), req.GetNoteText(), noteType, req.GetCreatedByUserId(), actor)
	if err != nil {
		return nil, mapError(err)
	}
	return &case_service.AddWorknoteResponse{Worknote: domainWorknoteToProto(w)}, nil
}

func (h *Handler) ListWorknotes(ctx context.Context, req *case_service.ListWorknotesRequest) (*case_service.ListWorknotesResponse, error) {
	items, nextToken, err := h.Service.ListWorknotes(ctx, req.GetCaseId(), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]*case_service.Worknote, len(items))
	for i, w := range items {
		out[i] = domainWorknoteToProto(w)
	}
	return &case_service.ListWorknotesResponse{Items: out, NextPageToken: nextToken}, nil
}

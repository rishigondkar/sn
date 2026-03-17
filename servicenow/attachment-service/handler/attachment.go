package handler

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/soc-platform/attachment-service/proto/attachment_service"
	"github.com/soc-platform/attachment-service/service"
	"github.com/soc-platform/attachment-service/domain"
)

const (
	metaUserID        = "x-user-id"
	metaRequestID     = "x-request-id"
	metaCorrelationID = "x-correlation-id"
)

func getMetadata(ctx context.Context) (userID, requestID, correlationID, actorName string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", "", ""
	}
	if v := md.Get(metaUserID); len(v) > 0 {
		userID = v[0]
	}
	if v := md.Get(metaRequestID); len(v) > 0 {
		requestID = v[0]
	}
	if v := md.Get(metaCorrelationID); len(v) > 0 {
		correlationID = v[0]
	}
	// actor name not in standard metadata; could add x-actor-name
	return userID, requestID, correlationID, actorName
}

// CreateAttachment implements AttachmentService.CreateAttachment.
func (h *Handler) CreateAttachment(ctx context.Context, req *pb.CreateAttachmentRequest) (*pb.CreateAttachmentResponse, error) {
	start := time.Now()
	method := "CreateAttachment"
	logs := []string{"starting CreateAttachment handler"}
	var err error
	defer func() {
		if err != nil {
			slog.Error("CreateAttachment handler failed", "layer", currentLayer, "method", method, "error", err, "logs", logs, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	_, requestID, correlationID, actorName := getMetadata(ctx)
	a, err := h.Service.CreateAttachment(ctx,
		req.GetCaseId(),
		req.GetFileName(),
		req.GetContentType(),
		req.GetUploadedByUserId(),
		req.GetContent(),
		requestID, correlationID, actorName,
	)
	if err != nil {
		err = mapServiceError(err)
		return nil, err
	}
	return &pb.CreateAttachmentResponse{
		Id:         a.ID,
		Attachment: domainToProto(a),
	}, nil
}

// ListAttachmentsByCase implements AttachmentService.ListAttachmentsByCase.
func (h *Handler) ListAttachmentsByCase(ctx context.Context, req *pb.ListAttachmentsByCaseRequest) (*pb.ListAttachmentsByCaseResponse, error) {
	start := time.Now()
	method := "ListAttachmentsByCase"
	logs := []string{"starting ListAttachmentsByCase handler"}
	var err error
	defer func() {
		if err != nil {
			slog.Error("ListAttachmentsByCase handler failed", "layer", currentLayer, "method", method, "error", err, "logs", logs, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	items, nextToken, err := h.Service.ListAttachmentsByCase(ctx, req.GetCaseId(), req.GetPageSize(), req.GetPageToken(), req.GetIncludeDeleted())
	if err != nil {
		err = mapServiceError(err)
		return nil, err
	}
	out := make([]*pb.Attachment, len(items))
	for i, a := range items {
		out[i] = domainToProto(a)
	}
	return &pb.ListAttachmentsByCaseResponse{Items: out, NextPageToken: nextToken}, nil
}

// DeleteAttachment implements AttachmentService.DeleteAttachment.
func (h *Handler) DeleteAttachment(ctx context.Context, req *pb.DeleteAttachmentRequest) (*pb.DeleteAttachmentResponse, error) {
	start := time.Now()
	method := "DeleteAttachment"
	logs := []string{"starting DeleteAttachment handler"}
	var err error
	defer func() {
		if err != nil {
			slog.Error("DeleteAttachment handler failed", "layer", currentLayer, "method", method, "error", err, "logs", logs, "duration_ms", time.Since(start).Milliseconds())
		}
	}()

	userID, requestID, correlationID, actorName := getMetadata(ctx)
	err = h.Service.DeleteAttachment(ctx, req.GetAttachmentId(), requestID, correlationID, userID, actorName)
	if err != nil {
		err = mapServiceError(err)
		return nil, err
	}
	return &pb.DeleteAttachmentResponse{}, nil
}

func domainToProto(a *domain.Attachment) *pb.Attachment {
	out := &pb.Attachment{
		Id:               a.ID,
		CaseId:           a.CaseID,
		FileName:         a.FileName,
		FileSizeBytes:    a.FileSizeBytes,
		ContentType:      a.ContentType,
		StorageProvider:  a.StorageProvider,
		UploadedByUserId: a.UploadedByUserID,
		IsDeleted:        a.IsDeleted,
		MetadataJson:     a.MetadataJSON,
	}
	out.UploadedAt = timestamppb.New(a.UploadedAt)
	out.CreatedAt = timestamppb.New(a.CreatedAt)
	out.UpdatedAt = timestamppb.New(a.UpdatedAt)
	if a.DeletedAt != nil {
		out.DeletedAt = timestamppb.New(*a.DeletedAt)
	}
	return out
}

func mapServiceError(err error) error {
	switch e := err.(type) {
	case service.ErrValidation:
		return status.Error(codes.InvalidArgument, e.Error())
	}
	if err == service.ErrNotFound {
		return status.Error(codes.NotFound, "attachment not found")
	}
	return status.Error(codes.Internal, "internal error")
}

package handler

import (
	pb "github.com/soc-platform/attachment-service/proto/attachment_service"
	"github.com/soc-platform/attachment-service/service"
)

const currentLayer = "handler"

// Handler implements AttachmentService gRPC server.
type Handler struct {
	pb.UnimplementedAttachmentServiceServer
	Service *service.Service
}

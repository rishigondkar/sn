package handler

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/servicenow/enrichment-threat-service/config"
	pb "github.com/servicenow/enrichment-threat-service/proto/enrichment_threat_service"
	"github.com/servicenow/enrichment-threat-service/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_UpsertEnrichmentResult_InvalidArgument(t *testing.T) {
	cfg := &config.Config{MaxPayloadBytes: 1024}
	svc := service.NewService(nil, nil, cfg)
	h := NewHandler(svc)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	s := grpc.NewServer()
	pb.RegisterEnrichmentThreatServiceServer(s, h)
	go func() { _ = s.Serve(lis) }()
	defer s.GracefulStop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	client := pb.NewEnrichmentThreatServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// No case_id and no observable_id -> InvalidArgument
	_, err = client.UpsertEnrichmentResult(ctx, &pb.UpsertEnrichmentResultRequest{
		EnrichmentType: "geo",
		SourceName:     "test",
		Status:         "ok",
		ResultDataJson: "{}",
		ReceivedAt:     timestamppb.New(time.Now().UTC()),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "InvalidArgument")
}

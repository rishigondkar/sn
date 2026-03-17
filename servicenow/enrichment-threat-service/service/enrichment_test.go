package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/servicenow/enrichment-threat-service/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertEnrichmentResult_Validation(t *testing.T) {
	cfg := &config.Config{MaxPayloadBytes: 1024}
	svc := NewService(nil, nil, cfg)
	ctx := context.Background()
	actor := ActorInfo{}

	tests := []struct {
		name    string
		in      UpsertEnrichmentInput
		wantErr bool
	}{
		{
			name: "missing both case_id and observable_id",
			in: UpsertEnrichmentInput{
				EnrichmentType: "geo",
				SourceName:     "test",
				Status:         "ok",
				ResultDataJSON: "{}",
				ReceivedAt:     time.Now().UTC(),
			},
			wantErr: true,
		},
		{
			name: "empty result_data",
			in: UpsertEnrichmentInput{
				ObservableID:   strPtr("obs-1"),
				EnrichmentType: "geo",
				SourceName:     "test",
				Status:         "ok",
				ResultDataJSON: "",
				ReceivedAt:     time.Now().UTC(),
			},
			wantErr: true,
		},
		{
			name: "invalid JSON result_data",
			in: UpsertEnrichmentInput{
				ObservableID:   strPtr("obs-1"),
				EnrichmentType: "geo",
				SourceName:     "test",
				Status:         "ok",
				ResultDataJSON: "{invalid",
				ReceivedAt:     time.Now().UTC(),
			},
			wantErr: true,
		},
		{
			name: "empty enrichment_type",
			in: UpsertEnrichmentInput{
				ObservableID:   strPtr("obs-1"),
				SourceName:     "test",
				Status:         "ok",
				ResultDataJSON: "{}",
				ReceivedAt:     time.Now().UTC(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.UpsertEnrichmentResult(ctx, tt.in, actor)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrValidation))
		})
	}
}

func strPtr(s string) *string { return &s }

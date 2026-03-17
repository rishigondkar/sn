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

func TestUpsertThreatLookupResult_Validation(t *testing.T) {
	cfg := &config.Config{MaxPayloadBytes: 1024}
	svc := NewService(nil, nil, cfg)
	ctx := context.Background()
	actor := ActorInfo{}

	tests := []struct {
		name    string
		in      UpsertThreatLookupInput
		wantErr bool
	}{
		{
			name: "missing observable_id",
			in: UpsertThreatLookupInput{
				LookupType:     "vt",
				SourceName:     "test",
				ResultDataJSON: "{}",
				ReceivedAt:     time.Now().UTC(),
			},
			wantErr: true,
		},
		{
			name: "empty result_data",
			in: UpsertThreatLookupInput{
				ObservableID:   "obs-1",
				LookupType:     "vt",
				SourceName:     "test",
				ResultDataJSON: "",
				ReceivedAt:     time.Now().UTC(),
			},
			wantErr: true,
		},
		{
			name: "invalid JSON result_data",
			in: UpsertThreatLookupInput{
				ObservableID:   "obs-1",
				LookupType:     "vt",
				SourceName:     "test",
				ResultDataJSON: "not json",
				ReceivedAt:     time.Now().UTC(),
			},
			wantErr: true,
		},
		{
			name: "empty lookup_type",
			in: UpsertThreatLookupInput{
				ObservableID:   "obs-1",
				SourceName:     "test",
				ResultDataJSON: "{}",
				ReceivedAt:     time.Now().UTC(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.UpsertThreatLookupResult(ctx, tt.in, actor)
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrValidation))
		})
	}
}

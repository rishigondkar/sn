package consumer

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateEnvelope_RequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*AuditEnvelope)
		wantErr string
	}{
		{"missing event_id", func(e *AuditEnvelope) { e.EventID = "" }, "event_id"},
		{"missing event_type", func(e *AuditEnvelope) { e.EventType = "" }, "event_type"},
		{"missing source_service", func(e *AuditEnvelope) { e.SourceService = "" }, "source_service"},
		{"missing entity_type", func(e *AuditEnvelope) { e.EntityType = "" }, "entity_type"},
		{"missing entity_id", func(e *AuditEnvelope) { e.EntityID = "" }, "entity_id"},
		{"missing action", func(e *AuditEnvelope) { e.Action = "" }, "action"},
		{"missing occurred_at", func(e *AuditEnvelope) { e.OccurredAt = "" }, "occurred_at"},
		{"invalid occurred_at", func(e *AuditEnvelope) { e.OccurredAt = "not-a-date" }, "occurred_at"},
		{"valid RFC3339", func(e *AuditEnvelope) { e.OccurredAt = "2026-03-14T12:00:00Z" }, ""},
		{"valid RFC3339Nano", func(e *AuditEnvelope) { e.OccurredAt = "2026-03-14T12:00:00.123456789Z" }, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AuditEnvelope{
				EventID:       "ev-1",
				EventType:     "case.updated",
				SourceService: "case-service",
				EntityType:    "case",
				EntityID:      "uuid-1",
				Action:        "update",
				OccurredAt:    "2026-03-14T12:00:00Z",
			}
			if tt.setup != nil {
				tt.setup(e)
			}
			err := ValidateEnvelope(e)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestProcessOne_InvalidJSON(t *testing.T) {
	// Use a mock repo that we never call (processOne will fail before repo)
	// We need a Repository - use nil would panic on insert. So we need a minimal test that only checks JSON/malformed.
	// Test processOne with invalid JSON returns error.
	cons := &Consumer{Repo: nil, Source: nil}
	err := cons.processOne(context.Background(), []byte("not json"))
	assert.Error(t, err)
}

func TestProcessOne_MissingRequiredField(t *testing.T) {
	// Body missing required field -> validation error after unmarshal
	body := []byte(`{"event_id":"e1","event_type":"x","source_service":"s","entity_type":"e","entity_id":"id"}`)
	// missing "action" and "occurred_at"
	cons := &Consumer{Repo: nil, Source: nil}
	err := cons.processOne(context.Background(), body)
	assert.Error(t, err)
}

func TestChannelSource_SendReceive(t *testing.T) {
	src := NewChannelSource(1)
	defer src.Close()

	body := []byte(`{"event_id":"ev-1","event_type":"case.created","source_service":"case-service","entity_type":"case","entity_id":"uuid-1","action":"create","occurred_at":"2026-03-14T12:00:00Z"}`)
	src.Send(body)

	ctx := context.Background()
	msg, err := src.Receive(ctx)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, body, msg.Body)
	msg.Ack()
}

func TestEnvelope_JSONRoundTrip(t *testing.T) {
	// Platform contract envelope round-trip
	raw := `{
		"event_id": "ev-1",
		"event_type": "case.updated",
		"source_service": "case-service",
		"entity_type": "case",
		"entity_id": "550e8400-e29b-41d4-a716-446655440000",
		"action": "update",
		"actor_user_id": "user-1",
		"request_id": "req-1",
		"correlation_id": "corr-1",
		"occurred_at": "2026-03-14T15:20:00Z"
	}`
	var e AuditEnvelope
	err := json.Unmarshal([]byte(raw), &e)
	require.NoError(t, err)
	err = ValidateEnvelope(&e)
	require.NoError(t, err)
	assert.Equal(t, "ev-1", e.EventID)
	assert.Equal(t, "case.updated", e.EventType)
}

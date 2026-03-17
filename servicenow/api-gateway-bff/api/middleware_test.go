package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	var captured string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.Context().Value(CtxKeyRequestID).(string)
	})
	handler := RequestID(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if captured == "" {
		t.Error("expected request_id in context")
	}
	if rec.Header().Get("X-Request-Id") != captured {
		t.Error("X-Request-Id header should match context")
	}
}

func TestRequestID_Forwarded(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(CtxKeyRequestID).(string)
		if id != "forwarded-id" {
			t.Errorf("expected forwarded request_id, got %s", id)
		}
	})
	handler := RequestID(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-Id", "forwarded-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Header().Get("X-Request-Id") != "forwarded-id" {
		t.Error("X-Request-Id should be forwarded")
	}
}

func TestAuth(t *testing.T) {
	var userID, userName string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ = r.Context().Value(CtxKeyActorID).(string)
		userName, _ = r.Context().Value(CtxKeyActorName).(string)
	})
	handler := Auth(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Id", "user-123")
	req.Header.Set("X-User-Name", "Jane")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if userID != "user-123" {
		t.Errorf("expected actor_user_id user-123, got %s", userID)
	}
	if userName != "Jane" {
		t.Errorf("expected actor_name Jane, got %s", userName)
	}
}

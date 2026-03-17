package api

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/google/uuid"
	"github.com/servicenow/api-gateway/logging"
)

// Context keys for middleware (must match logging package and clients.FromContext)
const (
	CtxKeyRequestID    = "request_id"
	CtxKeyCorrelationID = "correlation_id"
	CtxKeyActorID      = "actor_user_id"
	CtxKeyActorName    = "actor_name"
)

// RequestID generates or forwards X-Request-Id and sets in context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.Must(uuid.NewV7()).String()
		}
		ctx := context.WithValue(r.Context(), CtxKeyRequestID, id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CorrelationID generates or forwards X-Correlation-Id.
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-Id")
		if id == "" {
			id = uuid.Must(uuid.NewV7()).String()
		}
		ctx := context.WithValue(r.Context(), CtxKeyCorrelationID, id)
		w.Header().Set("X-Correlation-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Auth extracts trusted upstream identity headers and sets in context.
// POC: X-User-Id, X-User-Name (or X-Actor-Name). Do not trust from untrusted clients without upstream auth.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if uid := r.Header.Get("X-User-Id"); uid != "" {
			ctx = context.WithValue(ctx, CtxKeyActorID, uid)
		}
		if name := r.Header.Get("X-User-Name"); name != "" {
			ctx = context.WithValue(ctx, CtxKeyActorName, name)
		}
		if name := r.Header.Get("X-Actor-Name"); name != "" && ctx.Value(CtxKeyActorName) == nil {
			ctx = context.WithValue(ctx, CtxKeyActorName, name)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logging logs request and response with request_id and correlation_id.
func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logging.FromContext(r.Context(), log)
			log.Info("request", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

// Recovery recovers panics and returns 500 with REST error envelope.
func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic recovered", "error", err, "stack", string(debug.Stack()))
					requestID, _ := r.Context().Value(CtxKeyRequestID).(string)
					correlationID, _ := r.Context().Value(CtxKeyCorrelationID).(string)
					WriteError(w, http.StatusInternalServerError, CodeInternal, "Internal server error", requestID, correlationID, nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORS sets CORS headers. Allowed origins and methods should come from config.
func CORS(origins, methods []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, o := range origins {
				if o == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					break
				}
				w.Header().Add("Access-Control-Allow-Origin", o)
			}
			w.Header().Set("Access-Control-Allow-Methods", join(methods))
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-Id, X-Correlation-Id, X-User-Id, X-User-Name, Idempotency-Key")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func join(s []string) string {
	if len(s) == 0 {
		return ""
	}
	out := s[0]
	for i := 1; i < len(s); i++ {
		out += ", " + s[i]
	}
	return out
}

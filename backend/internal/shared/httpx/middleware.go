package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
)

// MaxBodyBytes caps JSON request bodies (1 MiB).
const MaxBodyBytes = 1 << 20

type ctxKey int

const requestIDKey ctxKey = iota

// RequestID propagates an X-Request-ID header (generating one when absent)
// through the request context and the response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

// RequestIDFromContext returns the request id set by the RequestID middleware.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// BodyLimit caps the request body size.
func BodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		next.ServeHTTP(w, r)
	})
}

// Recovery converts panics into a 500 INTERNAL error envelope and logs the
// stack trace.
func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						slog.String("error", toErrorString(rec)),
						slog.String("stack", string(debug.Stack())),
						slog.String("request_id", RequestIDFromContext(r.Context())),
					)
					WriteError(w, http.StatusInternalServerError, "INTERNAL", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logging emits one structured log line per request with method, route
// pattern, status, duration, and request id.
func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := newStatusRecorder(w)
			next.ServeHTTP(rec, r)
			log.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", RoutePattern(r)),
				slog.Int("status", rec.status),
				slog.Duration("duration", time.Since(start)),
				slog.String("request_id", RequestIDFromContext(r.Context())),
			)
		})
	}
}

// Metrics records http_requests_total and http_request_duration_seconds
// using the Chi route pattern (never the raw URL) to bound cardinality.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := newStatusRecorder(w)
		next.ServeHTTP(rec, r)
		pattern := RoutePattern(r)
		httpRequestsTotal.WithLabelValues(r.Method, pattern, statusCodeLabel(rec.status)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, pattern).Observe(time.Since(start).Seconds())
	})
}

// RoutePattern returns the Chi route pattern for the request (e.g.
// "/api/v1/shops/{id}"), or "unmatched" when no route matched.
func RoutePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
		return rctx.RoutePattern()
	}
	return "unmatched"
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(buf)
}

func statusCodeLabel(code int) string {
	switch {
	case code >= 500:
		return "5xx"
	case code >= 400:
		return "4xx"
	case code >= 300:
		return "3xx"
	default:
		return "2xx"
	}
}

func toErrorString(v any) string {
	if err, ok := v.(error); ok {
		return err.Error()
	}
	return "panic value is not an error"
}
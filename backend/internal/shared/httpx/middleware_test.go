package httpx

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
}

func TestMetricsUseRoutePatternNotRawURL(t *testing.T) {
	before := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/items/{id}", "2xx"))

	r := chi.NewRouter()
	r.Use(Metrics)
	r.Get("/items/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/items/42")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()

	after := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/items/{id}", "2xx"))
	if after-before != 1 {
		t.Fatalf("expected +1 for route pattern /items/{id}, got %v", after-before)
	}
}

func TestRoutePatternUnmatched(t *testing.T) {
	r := chi.NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)
	if got := RoutePattern(req); got != "unmatched" {
		t.Fatalf("RoutePattern = %q, want unmatched", got)
	}
	_ = r
}

func TestRequestIDGeneratedAndPropagated(t *testing.T) {
	r := chi.NewRouter()
	r.Use(RequestID)
	var seen string
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		seen = RequestIDFromContext(req.Context())
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()

	header := resp.Header.Get("X-Request-ID")
	if header == "" {
		t.Fatal("expected X-Request-ID response header")
	}
	if seen != header {
		t.Fatalf("context id %q != header %q", seen, header)
	}
}

func TestRequestIDPassthrough(t *testing.T) {
	r := chi.NewRouter()
	r.Use(RequestID)
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/", nil)
	req.Header.Set("X-Request-ID", "abc123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()
	if got := resp.Header.Get("X-Request-ID"); got != "abc123" {
		t.Fatalf("X-Request-ID = %q, want abc123", got)
	}
}

func TestBodyLimitAndReadJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	h := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var p payload
		_ = ReadJSON(w, req, &p)
	}))

	rec := httptest.NewRecorder()
	bigBody := `{"name":"` + strings.Repeat("a", MaxBodyBytes) + `"}`
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(bigBody)))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body: status %d, want 413", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{invalid")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid JSON: status %d, want 400", rec.Code)
	}
}

func TestReadJSONValidAndUnknownField(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	handler := func(w http.ResponseWriter, req *http.Request) {
		var p payload
		if err := ReadJSON(w, req, &p); err != nil {
			return
		}
		WriteData(w, http.StatusOK, p)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"name":"x"}`)))
	handler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid JSON: status %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"name":"x","extra":1}`)))
	handler(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown field: status %d, want 400", rec.Code)
	}
}

func TestRecoveryWritesErrorEnvelope(t *testing.T) {
	r := chi.NewRouter()
	r.Use(Recovery(testLogger()))
	r.Get("/boom", func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/boom")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", resp.StatusCode)
	}
	var body ErrorBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error.Code != "INTERNAL" {
		t.Fatalf("code = %q, want INTERNAL", body.Error.Code)
	}
}

func TestErrorEnvelopeShape(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusNotFound, "SHOP_NOT_FOUND", "shop not found")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("content-type %q", ct)
	}
	var body map[string]map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"]["code"] != "SHOP_NOT_FOUND" {
		t.Fatalf("body = %+v", body)
	}
}
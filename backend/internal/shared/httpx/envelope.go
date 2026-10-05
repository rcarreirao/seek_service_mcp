// Package httpx provides shared HTTP plumbing: the JSON error envelope,
// request/response helpers, and the Chi middleware chain.
package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrorDetail is the payload of the standard error envelope.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorBody is the wire shape of every error response.
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

// DataBody wraps successful payloads as {"data": ...}.
type DataBody struct {
	Data any `json:"data"`
}

// WriteError writes the standard error envelope with the given status.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorBody{Error: ErrorDetail{Code: code, Message: message}})
}

// WriteData writes a successful response wrapped in {"data": ...}.
func WriteData(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, DataBody{Data: data})
}

// ReadJSON decodes the request body into dst. Malformed JSON yields a 400
// INVALID_JSON envelope; bodies over the configured limit yield 413
// BODY_TOO_LARGE. Unknown fields are rejected.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		switch {
		case errors.As(err, &maxErr):
			WriteError(w, http.StatusRequestEntityTooLarge, "BODY_TOO_LARGE", "request body exceeds the maximum allowed size")
		default:
			WriteError(w, http.StatusBadRequest, "INVALID_JSON", "request body is not valid JSON")
		}
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
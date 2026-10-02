// Package v1 provides HTTP handler utilities for the Appointly API v1.
// All handlers use the same response envelope format for consistency.
package v1

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/appointly/appointly/backend/internal/middleware"
	"github.com/appointly/appointly/backend/internal/pkg/apperror"
	"github.com/appointly/appointly/backend/internal/pkg/pagination"
)

// --- Response Envelope -------------------------------------------------------

// Meta holds request metadata included in every response.
type Meta struct {
	RequestID  string          `json:"request_id,omitempty"`
	Timestamp  string          `json:"timestamp"`
	Pagination *pagination.OffsetMeta `json:"pagination,omitempty"`
}

// Response is the standard success response envelope.
type Response[T any] struct {
	Data T    `json:"data"`
	Meta Meta `json:"meta"`
}

// ListResponse is the standard paginated list response envelope.
type ListResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

// ErrorResponse is the standard error response envelope.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
	Meta  Meta      `json:"meta"`
}

// ErrorBody holds the error details.
type ErrorBody struct {
	Code    apperror.Code        `json:"code"`
	Message string               `json:"message"`
	Details []apperror.FieldError `json:"details,omitempty"`
}

// --- Response Writers --------------------------------------------------------

// RespondJSON writes a success JSON response with the standard envelope.
func RespondJSON[T any](w http.ResponseWriter, r *http.Request, status int, data T) {
	resp := Response[T]{
		Data: data,
		Meta: buildMeta(r, nil),
	}
	writeJSON(w, status, resp)
}

// RespondList writes a paginated list JSON response.
func RespondList[T any](w http.ResponseWriter, r *http.Request, status int, data []T, paginationMeta *pagination.OffsetMeta) {
	if data == nil {
		data = []T{}
	}
	resp := ListResponse[T]{
		Data: data,
		Meta: buildMeta(r, paginationMeta),
	}
	writeJSON(w, status, resp)
}

// RespondError writes an error JSON response with the standard envelope.
// In production, stack traces and internal details MUST NOT be included.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	appErr := apperror.AsAppError(err)
	if appErr == nil {
		// Unknown error — return generic 500
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error: ErrorBody{
				Code:    apperror.CodeInternal,
				Message: "an internal error occurred",
			},
			Meta: buildMeta(r, nil),
		})
		return
	}

	writeJSON(w, appErr.HTTPCode, ErrorResponse{
		Error: ErrorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
		Meta: buildMeta(r, nil),
	})
}

// RespondCreated writes a 201 Created response.
func RespondCreated[T any](w http.ResponseWriter, r *http.Request, data T) {
	RespondJSON(w, r, http.StatusCreated, data)
}

// RespondNoContent writes a 204 No Content response.
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// --- Request Decoding --------------------------------------------------------

// DecodeJSON decodes the request body into the given value.
// Returns a validation error if the body is malformed.
func DecodeJSON(r *http.Request, v interface{}) *apperror.AppError {
	if r.Body == nil {
		return apperror.BadRequest("request body is required")
	}
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		return apperror.BadRequest("invalid request body: " + err.Error())
	}
	return nil
}

// --- Internal Helpers --------------------------------------------------------

func buildMeta(r *http.Request, page *pagination.OffsetMeta) Meta {
	return Meta{
		RequestID:  middleware.RequestIDFromContext(r.Context()),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Pagination: page,
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Can't do much here — headers already sent
		// Logger should catch this upstream
		_ = err
	}
}

// Package apperror defines structured application error types used
// throughout the application. All errors originating from the domain
// and use case layers should use these types so that handlers can
// translate them to appropriate HTTP responses without coupling domain
// logic to transport concerns.
package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is a machine-readable error code string.
type Code string

const (
	// 4xx — Client errors
	CodeBadRequest          Code = "BAD_REQUEST"
	CodeValidationFailed    Code = "VALIDATION_FAILED"
	CodeUnauthorized        Code = "UNAUTHORIZED"
	CodeForbidden           Code = "FORBIDDEN"
	CodeNotFound            Code = "NOT_FOUND"
	CodeConflict            Code = "CONFLICT"
	CodeGone                Code = "GONE"
	CodeUnprocessable       Code = "UNPROCESSABLE_ENTITY"
	CodeTooManyRequests     Code = "TOO_MANY_REQUESTS"

	// 5xx — Server errors
	CodeInternal            Code = "INTERNAL_SERVER_ERROR"
	CodeServiceUnavailable  Code = "SERVICE_UNAVAILABLE"
	CodeTimeout             Code = "TIMEOUT"

	// Domain-specific
	CodeSlotUnavailable     Code = "SLOT_UNAVAILABLE"
	CodeBookingConflict     Code = "BOOKING_CONFLICT"
	CodePaymentFailed       Code = "PAYMENT_FAILED"
	CodeSubscriptionExpired Code = "SUBSCRIPTION_EXPIRED"
	CodePlanLimitExceeded   Code = "PLAN_LIMIT_EXCEEDED"
	CodeInvalidCredentials  Code = "INVALID_CREDENTIALS"
	CodeTokenExpired        Code = "TOKEN_EXPIRED"
	CodeTokenInvalid        Code = "TOKEN_INVALID"
	CodeEmailNotVerified    Code = "EMAIL_NOT_VERIFIED"
	CodeAlreadyExists       Code = "ALREADY_EXISTS"
	CodeInvalidStatus       Code = "INVALID_STATUS_TRANSITION"
)

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// AppError is the structured error type returned by domain and use case layers.
// Handlers translate AppError into HTTP responses.
type AppError struct {
	Code     Code         `json:"code"`
	Message  string       `json:"message"`
	Details  []FieldError `json:"details,omitempty"`
	HTTPCode int          `json:"-"`
	Err      error        `json:"-"` // wrapped underlying error (never exposed to client)
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap allows errors.Is and errors.As to traverse the error chain.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if the target error has the same Code.
func (e *AppError) Is(target error) bool {
	var t *AppError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// WithField adds a field-level validation error to the AppError.
func (e *AppError) WithField(field, message string) *AppError {
	e.Details = append(e.Details, FieldError{Field: field, Message: message})
	return e
}

// Wrap wraps an underlying error into the AppError (for logging/tracing).
func (e *AppError) Wrap(err error) *AppError {
	e.Err = err
	return e
}

// --- Constructors ------------------------------------------------------------

// New creates a new AppError with the given code, HTTP status, and message.
func New(code Code, httpStatus int, message string) *AppError {
	return &AppError{
		Code:     code,
		HTTPCode: httpStatus,
		Message:  message,
	}
}

func BadRequest(message string) *AppError {
	return New(CodeBadRequest, http.StatusBadRequest, message)
}

func ValidationFailed(message string) *AppError {
	return New(CodeValidationFailed, http.StatusBadRequest, message)
}

func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, http.StatusUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(CodeForbidden, http.StatusForbidden, message)
}

func NotFound(resource string) *AppError {
	return New(CodeNotFound, http.StatusNotFound, fmt.Sprintf("%s not found", resource))
}

func Conflict(message string) *AppError {
	return New(CodeConflict, http.StatusConflict, message)
}

func AlreadyExists(resource string) *AppError {
	return New(CodeAlreadyExists, http.StatusConflict, fmt.Sprintf("%s already exists", resource))
}

func Unprocessable(message string) *AppError {
	return New(CodeUnprocessable, http.StatusUnprocessableEntity, message)
}

func TooManyRequests(message string) *AppError {
	return New(CodeTooManyRequests, http.StatusTooManyRequests, message)
}

func Internal(message string) *AppError {
	return New(CodeInternal, http.StatusInternalServerError, message)
}

func ServiceUnavailable(message string) *AppError {
	return New(CodeServiceUnavailable, http.StatusServiceUnavailable, message)
}

func SlotUnavailable() *AppError {
	return New(CodeSlotUnavailable, http.StatusConflict, "the requested time slot is no longer available")
}

func BookingConflict() *AppError {
	return New(CodeBookingConflict, http.StatusConflict, "booking conflicts with an existing appointment")
}

func InvalidCredentials() *AppError {
	return New(CodeInvalidCredentials, http.StatusUnauthorized, "invalid email or password")
}

func TokenExpired() *AppError {
	return New(CodeTokenExpired, http.StatusUnauthorized, "token has expired")
}

func TokenInvalid() *AppError {
	return New(CodeTokenInvalid, http.StatusUnauthorized, "token is invalid")
}

func InvalidStatusTransition(from, to string) *AppError {
	return New(CodeInvalidStatus, http.StatusUnprocessableEntity,
		fmt.Sprintf("cannot transition appointment status from %s to %s", from, to))
}

func PlanLimitExceeded(resource string) *AppError {
	return New(CodePlanLimitExceeded, http.StatusForbidden,
		fmt.Sprintf("plan limit exceeded: %s", resource))
}

// --- Helpers -----------------------------------------------------------------

// AsAppError attempts to extract an *AppError from err.
// Returns nil if err is not an *AppError.
func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// HTTPStatus returns the HTTP status code for the error.
// If err is not an *AppError, returns 500.
func HTTPStatus(err error) int {
	if appErr := AsAppError(err); appErr != nil {
		return appErr.HTTPCode
	}
	return http.StatusInternalServerError
}

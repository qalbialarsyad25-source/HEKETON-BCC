package errors

import (
	"fmt"
	"net/http"
)

// Issue is one invalid field in a request, sent to the client in "issues".
type Issue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// AppError is an error the client is allowed to see: status, a stable code and a message.
// Err holds the cause; it is logged, never sent to the client.
type AppError struct {
	StatusCode int
	Code       string // stable, SCREAMING_SNAKE, e.g. POST_NOT_FOUND
	Message    string
	Issues     []Issue
	Err        error
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil *AppError>"
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	return ok && t != nil && e != nil && t.Code == e.Code
}

func (e *AppError) Wrap(cause error) *AppError {
	wrapped := *e
	wrapped.Issues = append([]Issue(nil), e.Issues...)
	wrapped.Err = cause
	return &wrapped
}

func New(statusCode int, code string, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

func BadRequest(code string, message string) *AppError {
	return New(http.StatusBadRequest, code, message)
}

func Unauthorized(code string, message string) *AppError {
	return New(http.StatusUnauthorized, code, message)
}

func Forbidden(code string, message string) *AppError {
	return New(http.StatusForbidden, code, message)
}

func NotFound(code string, message string) *AppError {
	return New(http.StatusNotFound, code, message)
}

func Conflict(code string, message string) *AppError {
	return New(http.StatusConflict, code, message)
}

func TooManyRequests(code string, message string) *AppError {
	return New(http.StatusTooManyRequests, code, message)
}

// BadGateway is a 502: a service this API depends on (Google, a payment gateway) failed.
func BadGateway(code string, message string) *AppError {
	return New(http.StatusBadGateway, code, message)
}

// PayloadTooLarge is a 413: the request body or an uploaded file is over its limit.
func PayloadTooLarge(code string, message string) *AppError {
	return New(http.StatusRequestEntityTooLarge, code, message)
}

// UnsupportedMediaType is a 415: the body or an uploaded file has a type the endpoint refuses.
func UnsupportedMediaType(code string, message string) *AppError {
	return New(http.StatusUnsupportedMediaType, code, message)
}

// Validation is a 400 VALIDATION_FAILED listing every invalid field.
func Validation(issues []Issue) *AppError {
	err := New(http.StatusBadRequest, "VALIDATION_FAILED", "request validation failed")
	err.Issues = issues
	return err
}

// Internal hides cause behind a generic 500. The cause is only logged.
func Internal(cause error) *AppError {
	err := New(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	err.Err = cause
	return err
}

// Shared errors. Use Wrap to attach a cause; never change these values.
var (
	ErrUnauthorized     = Unauthorized("UNAUTHORIZED", "request unauthorized")
	ErrInvalidToken     = Unauthorized("INVALID_TOKEN", "invalid token")
	ErrTokenExpired     = Unauthorized("TOKEN_EXPIRED", "token expired")
	ErrForbidden        = Forbidden("FORBIDDEN", "request forbidden")
	ErrRouteNotFound    = NotFound("ROUTE_NOT_FOUND", "route not found")
	ErrMethodNotAllowed = New(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	ErrMalformedBody    = BadRequest("MALFORMED_BODY", "malformed request body")
	ErrInvalidRequest   = BadRequest("INVALID_REQUEST", "invalid request data")
	ErrPayloadTooLarge  = PayloadTooLarge("PAYLOAD_TOO_LARGE", "request body too large")
)

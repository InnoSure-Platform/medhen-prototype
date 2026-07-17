package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode is a domain-specific error code.
type ErrorCode string

// AppError represents an RFC 7807 Problem Details error.
type AppError struct {
	Type     string      `json:"type"`
	Title    string      `json:"title"`
	Status   int         `json:"status"`
	Detail   string      `json:"detail"`
	Instance string      `json:"instance,omitempty"`
	Code     ErrorCode   `json:"code,omitempty"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s: %s (code: %s)", e.Status, e.Title, e.Detail, e.Code)
}

// New creates a new AppError.
func New(status int, title, detail string, code ErrorCode) *AppError {
	return &AppError{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
		Code:   code,
	}
}

// Common errors
func NewInternalServerError(detail string) *AppError {
	return New(http.StatusInternalServerError, "Internal Server Error", detail, "INTERNAL_ERROR")
}

func NewBadRequestError(detail string) *AppError {
	return New(http.StatusBadRequest, "Bad Request", detail, "BAD_REQUEST")
}

func NewNotFoundError(detail string) *AppError {
	return New(http.StatusNotFound, "Not Found", detail, "NOT_FOUND")
}

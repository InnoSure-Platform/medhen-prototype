// Package errors defines typed platform error codes for edge mapping.
package errors

import "fmt"

type Code string

const (
	CodeNotFound      Code = "NOT_FOUND"
	CodeConflict      Code = "CONFLICT"
	CodeValidation    Code = "VALIDATION"
	CodeUnauthorized  Code = "UNAUTHORIZED"
	CodeForbidden     Code = "FORBIDDEN"
	CodeIdempotency   Code = "IDEMPOTENCY_REPLAY"
	CodeUnavailable   Code = "UNAVAILABLE"
	CodeInternal      Code = "INTERNAL"
	CodeUWDeclined    Code = "UW_DECLINED"
	CodePaymentFailed Code = "PAYMENT_FAILED"
)

type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.Err }

func E(code Code, msg string) *Error { return &Error{Code: code, Message: msg} }
func Wrap(code Code, msg string, err error) *Error {
	return &Error{Code: code, Message: msg, Err: err}
}

func HTTPStatus(code Code) int {
	switch code {
	case CodeNotFound:
		return 404
	case CodeConflict, CodeIdempotency:
		return 409
	case CodeValidation, CodeUWDeclined:
		return 400
	case CodeUnauthorized:
		return 401
	case CodeForbidden:
		return 403
	case CodeUnavailable, CodePaymentFailed:
		return 502
	default:
		return 500
	}
}

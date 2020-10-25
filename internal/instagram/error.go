package instagram

import (
	"errors"
	"fmt"
)

// Error is the response returned when a call is unsuccessful.
type Error struct {
	// Message is a human-readable description of the error, for example,
	// "Tried accessing nonexisting field (media_type) on node type (User)".
	Message string `json:"message"`
	// Type is Instagram defined error type, e.g., IGApiException.
	Type string `json:"type"`
	// Code is a machine-readable error code as defined in
	// https://developers.facebook.com/docs/graph-api/using-graph-api/error-handling.
	Code int `json:"code"`
	// TraceID is Instagram's internal support identifier to help them find related logs when debugging.
	TraceID string `json:"fbtrace_id"`

	// HTTPStatusCode is an HTTP status code returned by a server.
	HTTPStatusCode int
	// Body is the raw response returned by a server.
	Body string
	// Inner is a wrapped error, e.g., JSON serialization error.
	Inner error
}

func (e Error) Error() string {
	if e.Inner != nil {
		return e.Inner.Error()
	}
	return fmt.Sprintf("%s %d: %s", e.Type, e.Code, e.Message)
}

func (e Error) Unwrap() error {
	return e.Inner
}

// ErrorCode returns a machine-readable error code, if available.
// See available codes at https://developers.facebook.com/docs/graph-api/using-graph-api/error-handling.
func ErrorCode(err error) int {
	var e Error
	if errors.As(err, &e) {
		return e.Code
	}
	return 0
}

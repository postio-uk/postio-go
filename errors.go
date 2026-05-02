package postio

import (
	"errors"
	"fmt"
)

// Error is every typed Postio API error. Use errors.As to discriminate.
//
//	var e *postio.Error
//	if errors.As(err, &e) && e.Status == 401 { ... }
//
// Or use the sentinel values:
//
//	if errors.Is(err, postio.ErrInvalidKey) { ... }
type Error struct {
	// Status is the HTTP status code (0 for network/timeout errors).
	Status int
	// Code is the API's error code (e.g. "invalid_api_key"). Empty for
	// transport-level errors.
	Code string
	// Message is the human-readable error string from the envelope.
	Message string
	// Details is the optional details field from the envelope.
	Details string
	// RequestID is the API's request_id — quote this in support tickets.
	RequestID string
	// Envelope is the raw decoded error body, for any field this struct
	// doesn't surface explicitly.
	Envelope *ErrorEnvelope
	// RetryAfter is set on rate-limit errors when the API sends a
	// Retry-After header.
	RetryAfter float64
	// Cause is the underlying transport error, if any.
	Cause error
}

func (e *Error) Error() string {
	if e.Code != "" && e.Message != "" {
		return fmt.Sprintf("postio: %s (status=%d, code=%s, request_id=%s)", e.Message, e.Status, e.Code, e.RequestID)
	}
	if e.Cause != nil {
		return fmt.Sprintf("postio: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("postio: %s (status=%d)", e.Message, e.Status)
}

func (e *Error) Unwrap() error { return e.Cause }

// Is implements errors.Is dispatch against the sentinel errors below.
func (e *Error) Is(target error) bool {
	switch target {
	case ErrInvalidKey:
		return e.Status == 401
	case ErrOutOfCredit:
		return e.Status == 402
	case ErrForbidden:
		return e.Status == 403
	case ErrNotFound:
		return e.Status == 404
	case ErrValidation:
		return e.Status == 400 || e.Status == 422
	case ErrRateLimit:
		return e.Status == 429
	case ErrServer:
		return e.Status >= 500
	case ErrTimeout:
		return e.Code == "request_timeout"
	case ErrConnection:
		return e.Code == "network_error"
	}
	return false
}

// Sentinel errors. Use errors.Is(err, ...) to match.
var (
	ErrInvalidKey   = errors.New("postio: invalid api key (401)")
	ErrOutOfCredit  = errors.New("postio: out of credit (402)")
	ErrForbidden    = errors.New("postio: forbidden (403)")
	ErrNotFound     = errors.New("postio: not found (404)")
	ErrValidation   = errors.New("postio: validation error (400/422)")
	ErrRateLimit    = errors.New("postio: rate limit (429)")
	ErrServer       = errors.New("postio: server error (5xx)")
	ErrTimeout      = errors.New("postio: request timeout")
	ErrConnection   = errors.New("postio: connection error")
)

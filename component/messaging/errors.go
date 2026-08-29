package messaging

import (
	"errors"
	"fmt"

	types "go.wasmcloud.dev/component/imports/wasmcloud_messaging_0_3_0_types"
)

// Sentinel errors for the named cases of wasmcloud:messaging/types.error.
// Test for them with errors.Is; backend-specific failures that do not map to
// a named case are returned as ordinary errors carrying the backend message.
var (
	// ErrSubjectInvalid means the subject is malformed, empty, or otherwise
	// rejected by the broker.
	ErrSubjectInvalid = errors.New("messaging: subject invalid")
	// ErrAccessDenied means the component may not publish to this subject.
	ErrAccessDenied = errors.New("messaging: access denied")
	// ErrTimeout means a [Request] got no reply within its timeout.
	ErrTimeout = errors.New("messaging: timeout")
	// ErrBrokerUnavailable means the broker is currently unreachable. The
	// operation may succeed if retried later.
	ErrBrokerUnavailable = errors.New("messaging: broker unavailable")
	// ErrMessageTooLarge means the body exceeded the broker's maximum
	// payload size.
	ErrMessageTooLarge = errors.New("messaging: message too large")
	// ErrQuotaExceeded means a quota or rate limit was exceeded.
	ErrQuotaExceeded = errors.New("messaging: quota exceeded")
)

// Dispositions a handler returns to tell the host whether redelivering the
// message could help. Return one from the callback registered with
// messaging/handler.HandleFunc, on its own or wrapped with %w; any other
// error is reported as a handler-specific failure, which hosts treat the
// same as ErrReject for redelivery.
var (
	// ErrReject says the message cannot be processed and redelivering the
	// same message will not help — a malformed payload, an unrecognized
	// subject shape.
	ErrReject = errors.New("messaging: reject")
	// ErrRetry says processing failed transiently and redelivering later may
	// succeed. Backends with redelivery semantics (JetStream, say) map it to
	// a negative acknowledgement.
	ErrRetry = errors.New("messaging: retry")
)

func convertError(e types.Error) error {
	switch e.Tag() {
	case types.ErrorSubjectInvalid:
		return ErrSubjectInvalid
	case types.ErrorAccessDenied:
		return ErrAccessDenied
	case types.ErrorTimeout:
		return ErrTimeout
	case types.ErrorBrokerUnavailable:
		return ErrBrokerUnavailable
	case types.ErrorMessageTooLarge:
		return ErrMessageTooLarge
	case types.ErrorQuotaExceeded:
		return ErrQuotaExceeded
	case types.ErrorOther:
		return fmt.Errorf("messaging: %s", e.Other())
	default:
		// The variant is non-exhaustive; future hosts may add cases.
		return fmt.Errorf("messaging: unknown error (tag %d)", e.Tag())
	}
}

// ToDisposition maps an error returned by a handler callback onto the
// per-delivery disposition the host expects. It is exported for the
// messaging/handler subpackage; applications should not need it.
func ToDisposition(err error) types.HandleMessageError {
	switch {
	case errors.Is(err, ErrRetry):
		return types.MakeHandleMessageErrorRetry()
	case errors.Is(err, ErrReject):
		return types.MakeHandleMessageErrorReject()
	default:
		return types.MakeHandleMessageErrorOther(err.Error())
	}
}

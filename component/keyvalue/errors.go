package keyvalue

import (
	"errors"
	"fmt"

	types "go.wasmcloud.dev/component/imports/wasmcloud_keyvalue_0_1_0_types"
)

// Sentinel errors for the named cases of wasmcloud:keyvalue/types.error.
// Test for them with errors.Is; backend-specific failures that do not map to
// a named case are returned as ordinary errors carrying the backend message.
var (
	ErrNoSuchStore        = errors.New("keyvalue: no such store")
	ErrAccessDenied       = errors.New("keyvalue: access denied")
	ErrInvalidArgument    = errors.New("keyvalue: invalid argument")
	ErrPreconditionFailed = errors.New("keyvalue: precondition failed")
	ErrTimeout            = errors.New("keyvalue: timeout")
	ErrStoreUnavailable   = errors.New("keyvalue: store unavailable")
	ErrQuotaExceeded      = errors.New("keyvalue: quota exceeded")
)

func convertError(e types.Error) error {
	switch e.Tag() {
	case types.ErrorNoSuchStore:
		return ErrNoSuchStore
	case types.ErrorAccessDenied:
		return ErrAccessDenied
	case types.ErrorInvalidArgument:
		return ErrInvalidArgument
	case types.ErrorPreconditionFailed:
		return ErrPreconditionFailed
	case types.ErrorTimeout:
		return ErrTimeout
	case types.ErrorStoreUnavailable:
		return ErrStoreUnavailable
	case types.ErrorQuotaExceeded:
		return ErrQuotaExceeded
	case types.ErrorOther:
		return fmt.Errorf("keyvalue: %s", e.Other())
	default:
		// The variant is non-exhaustive; future hosts may add cases.
		return fmt.Errorf("keyvalue: unknown error (tag %d)", e.Tag())
	}
}

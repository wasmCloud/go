package blobstore

import (
	"errors"
	"fmt"

	types "go.wasmcloud.dev/component/imports/wasmcloud_blobstore_0_1_0_types"
)

// Sentinel errors for the named cases of wasmcloud:blobstore/types.error.
// Test for them with errors.Is; backend-specific failures that do not map to
// a named case are returned as ordinary errors carrying the backend message.
var (
	ErrNoSuchContainer        = errors.New("blobstore: no such container")
	ErrContainerAlreadyExists = errors.New("blobstore: container already exists")
	ErrNoSuchObject           = errors.New("blobstore: no such object")
	ErrAccessDenied           = errors.New("blobstore: access denied")
	ErrTimeout                = errors.New("blobstore: timeout")
	ErrStoreUnavailable       = errors.New("blobstore: store unavailable")
	ErrQuotaExceeded          = errors.New("blobstore: quota exceeded")
)

func convertError(e types.Error) error {
	switch e.Tag() {
	case types.ErrorNoSuchContainer:
		return ErrNoSuchContainer
	case types.ErrorContainerAlreadyExists:
		return ErrContainerAlreadyExists
	case types.ErrorNoSuchObject:
		return ErrNoSuchObject
	case types.ErrorAccessDenied:
		return ErrAccessDenied
	case types.ErrorTimeout:
		return ErrTimeout
	case types.ErrorStoreUnavailable:
		return ErrStoreUnavailable
	case types.ErrorQuotaExceeded:
		return ErrQuotaExceeded
	case types.ErrorOther:
		return fmt.Errorf("blobstore: %s", e.Other())
	default:
		// The variant is non-exhaustive; future hosts may add cases.
		return fmt.Errorf("blobstore: unknown error (tag %d)", e.Tag())
	}
}

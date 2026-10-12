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
	case types.Error_NoSuchContainer:
		return ErrNoSuchContainer
	case types.Error_ContainerAlreadyExists:
		return ErrContainerAlreadyExists
	case types.Error_NoSuchObject:
		return ErrNoSuchObject
	case types.Error_AccessDenied:
		return ErrAccessDenied
	case types.Error_Timeout:
		return ErrTimeout
	case types.Error_StoreUnavailable:
		return ErrStoreUnavailable
	case types.Error_QuotaExceeded:
		return ErrQuotaExceeded
	case types.Error_Other:
		return fmt.Errorf("blobstore: %s", e.Other())
	default:
		// The variant is non-exhaustive; future hosts may add cases.
		return fmt.Errorf("blobstore: unknown error (tag %d)", e.Tag())
	}
}

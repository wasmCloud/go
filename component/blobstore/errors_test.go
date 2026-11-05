package blobstore

import (
	"errors"
	"testing"

	types "go.wasmcloud.dev/component/imports/wasmcloud_blobstore_0_1_0_types"
)

func TestConvertError(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   types.Error
		want error
	}{
		{"no such container", types.MakeErrorNoSuchContainer(), ErrNoSuchContainer},
		{"container already exists", types.MakeErrorContainerAlreadyExists(), ErrContainerAlreadyExists},
		{"no such object", types.MakeErrorNoSuchObject(), ErrNoSuchObject},
		{"access denied", types.MakeErrorAccessDenied(), ErrAccessDenied},
		{"timeout", types.MakeErrorTimeout(), ErrTimeout},
		{"store unavailable", types.MakeErrorStoreUnavailable(), ErrStoreUnavailable},
		{"quota exceeded", types.MakeErrorQuotaExceeded(), ErrQuotaExceeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertError(tc.in); !errors.Is(got, tc.want) {
				t.Errorf("convertError() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestConvertErrorOther(t *testing.T) {
	got := convertError(types.MakeErrorOther("disk full"))
	if want := "blobstore: disk full"; got.Error() != want {
		t.Errorf("convertError() = %q, want %q", got, want)
	}
}

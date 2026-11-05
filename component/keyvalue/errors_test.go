package keyvalue

import (
	"errors"
	"testing"

	types "go.wasmcloud.dev/component/imports/wasmcloud_keyvalue_0_2_0_types"
)

func TestConvertError(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   types.Error
		want error
	}{
		{"no such store", types.MakeErrorNoSuchStore(), ErrNoSuchStore},
		{"access denied", types.MakeErrorAccessDenied(), ErrAccessDenied},
		{"invalid argument", types.MakeErrorInvalidArgument(), ErrInvalidArgument},
		{"precondition failed", types.MakeErrorPreconditionFailed(), ErrPreconditionFailed},
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
	got := convertError(types.MakeErrorOther("backend exploded"))
	if want := "keyvalue: backend exploded"; got.Error() != want {
		t.Errorf("convertError() = %q, want %q", got, want)
	}
}

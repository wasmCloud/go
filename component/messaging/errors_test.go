package messaging

import (
	"errors"
	"fmt"
	"testing"

	types "go.wasmcloud.dev/component/imports/wasmcloud_messaging_0_3_0_types"
)

func TestConvertError(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   types.Error
		want error
	}{
		{"subject invalid", types.MakeErrorSubjectInvalid(), ErrSubjectInvalid},
		{"access denied", types.MakeErrorAccessDenied(), ErrAccessDenied},
		{"timeout", types.MakeErrorTimeout(), ErrTimeout},
		{"broker unavailable", types.MakeErrorBrokerUnavailable(), ErrBrokerUnavailable},
		{"message too large", types.MakeErrorMessageTooLarge(), ErrMessageTooLarge},
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
	got := convertError(types.MakeErrorOther("broker said no"))
	if want := "messaging: broker said no"; got.Error() != want {
		t.Errorf("convertError() = %q, want %q", got, want)
	}
}

func TestToDisposition(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   error
		tag  uint8
	}{
		{"retry", ErrRetry, types.HandleMessageError_Retry},
		{"wrapped retry", fmt.Errorf("busy: %w", ErrRetry), types.HandleMessageError_Retry},
		{"reject", ErrReject, types.HandleMessageError_Reject},
		{"wrapped reject", fmt.Errorf("bad payload: %w", ErrReject), types.HandleMessageError_Reject},
		{"other", errors.New("boom"), types.HandleMessageError_Other},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ToDisposition(tc.in); got.Tag() != tc.tag {
				t.Errorf("ToDisposition() tag = %d, want %d", got.Tag(), tc.tag)
			}
		})
	}
}

func TestToDispositionOtherCarriesMessage(t *testing.T) {
	got := ToDisposition(errors.New("boom"))
	if got.Other() != "boom" {
		t.Errorf("ToDisposition().Other() = %q, want %q", got.Other(), "boom")
	}
}

package nats

import (
	"errors"
	"testing"

	types "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_types"
)

func TestConvertErrorSentinels(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   types.NatsError
		want error
	}{
		{"timeout", types.MakeNatsErrorTimeout("request timed out: deadline has elapsed"), ErrTimeout},
		{"no responders", types.MakeNatsErrorNoResponders(), ErrNoResponders},
		{"key not found", types.MakeNatsErrorKeyNotFound(), ErrKeyNotFound},
		{"no messages", types.MakeNatsErrorNoMessages(), ErrNoMessages},
		{"disconnected", types.MakeNatsErrorDisconnected(), ErrDisconnected},
		{"already settled", types.MakeNatsErrorAlreadySettled(), ErrAlreadySettled},
		{"ack owned by host", types.MakeNatsErrorAckOwnedByHost(), ErrAckOwnedByHost},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertError(tc.in); !errors.Is(got, tc.want) {
				t.Errorf("convertError() = %v, want %v", got, tc.want)
			}
		})
	}
}

// The timeout detail names the call that timed out, so it must survive the
// wrapping that makes ErrTimeout matchable.
func TestConvertErrorTimeoutKeepsDetail(t *testing.T) {
	got := convertError(types.MakeNatsErrorTimeout("kv history on flags did not complete within 10000ms"))
	if want := "nats: timeout: kv history on flags did not complete within 10000ms"; got.Error() != want {
		t.Errorf("convertError() = %q, want %q", got, want)
	}
	if errors.Is(got, ErrNoResponders) {
		t.Error("a timeout must not match ErrNoResponders")
	}
}

func TestConvertErrorTyped(t *testing.T) {
	t.Run("denied message", func(t *testing.T) {
		var d *DeniedError
		err := convertError(types.MakeNatsErrorDenied(types.Denial{
			Reason: types.DenialReason_NotGranted,
			Target: types.MakeDeniedResourceMessage(42),
			Name:   "ORDERS",
		}))
		if !errors.As(err, &d) {
			t.Fatalf("convertError() = %T, want *DeniedError", err)
		}
		if d.Reason != DenialNotGranted || d.Target != DeniedMessage || d.Name != "ORDERS" || d.Sequence != 42 {
			t.Errorf("DeniedError = %+v", *d)
		}
	})
	t.Run("max payload exceeded", func(t *testing.T) {
		var e *MaxPayloadExceededError
		if err := convertError(types.MakeNatsErrorMaxPayloadExceeded(1 << 20)); !errors.As(err, &e) || e.Limit != 1<<20 {
			t.Errorf("convertError() = %v", err)
		}
	})
	t.Run("revision mismatch", func(t *testing.T) {
		var e *RevisionMismatchError
		if err := convertError(types.MakeNatsErrorRevisionMismatch(7)); !errors.As(err, &e) || e.Current != 7 {
			t.Errorf("convertError() = %v", err)
		}
	})
	t.Run("invalid header", func(t *testing.T) {
		var e *InvalidHeaderError
		if err := convertError(types.MakeNatsErrorInvalidHeader("bad:name")); !errors.As(err, &e) || e.Detail != "bad:name" {
			t.Errorf("convertError() = %v", err)
		}
	})
	t.Run("limit exceeded", func(t *testing.T) {
		var e *LimitExceededError
		if err := convertError(types.MakeNatsErrorLimitExceeded("batch over max-request-batch")); !errors.As(err, &e) || e.Detail != "batch over max-request-batch" {
			t.Errorf("convertError() = %v", err)
		}
	})
	t.Run("not found", func(t *testing.T) {
		var e *NotFoundError
		if err := convertError(types.MakeNatsErrorNotFound("stream ORDERS")); !errors.As(err, &e) || e.Resource != "stream ORDERS" {
			t.Errorf("convertError() = %v", err)
		}
	})
	t.Run("unsupported by server", func(t *testing.T) {
		var e *UnsupportedByServerError
		if err := convertError(types.MakeNatsErrorUnsupportedByServer("needs 2.11")); !errors.As(err, &e) || e.Detail != "needs 2.11" {
			t.Errorf("convertError() = %v", err)
		}
	})
}

func TestConvertErrorDetailOnly(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   types.NatsError
		want string
	}{
		{"connection", types.MakeNatsErrorConnection("broken pipe"), "nats: connection: broken pipe"},
		{"jetstream", types.MakeNatsErrorJetstream("stream sealed"), "nats: jetstream: stream sealed"},
		{"unexpected", types.MakeNatsErrorUnexpected("boom"), "nats: boom"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertError(tc.in); got.Error() != tc.want {
				t.Errorf("convertError() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDeniedErrorMessage(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   DeniedError
		want string
	}{
		{
			"subject",
			DeniedError{Reason: DenialNotGranted, Target: DeniedSubject, Name: "orders.>"},
			`nats: subject "orders.>" denied: not granted`,
		},
		{
			"bucket",
			DeniedError{Reason: DenialReserved, Target: DeniedBucket, Name: "config"},
			`nats: bucket "config" denied: reserved`,
		},
		{
			"message",
			DeniedError{Reason: DenialWildcardNotAllowed, Target: DeniedMessage, Name: "ORDERS", Sequence: 42},
			`nats: message 42 in stream "ORDERS" denied: wildcard not allowed`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.in.Error(); got != tc.want {
				t.Errorf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUnknownEnumStrings(t *testing.T) {
	if got := DenialReason(99).String(); got != "unknown" {
		t.Errorf("DenialReason(99).String() = %q", got)
	}
	if got := DeniedTarget(99).String(); got != "unknown" {
		t.Errorf("DeniedTarget(99).String() = %q", got)
	}
}

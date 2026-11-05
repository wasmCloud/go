package nats

import "testing"

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

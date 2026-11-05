package postgres

import (
	"errors"
	"math"
	"testing"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	types "go.wasmcloud.dev/component/imports/wasmcloud_postgres_0_2_0_types"
)

func TestConvertErrorSentinels(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   types.Error
		want error
	}{
		{"unknown prepared statement", types.MakeErrorUnknownPreparedStatement(), ErrUnknownPreparedStatement},
		{"access denied", types.MakeErrorAccessDenied(), ErrAccessDenied},
		{"timeout", types.MakeErrorTimeout(), ErrTimeout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertError(tc.in); !errors.Is(got, tc.want) {
				t.Errorf("convertError() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestConvertErrorMessages(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   types.Error
		want string
	}{
		{"connection failed", types.MakeErrorConnectionFailed("refused"), "postgres: connection failed: refused"},
		{"invalid params", types.MakeErrorInvalidParams("want 2"), "postgres: invalid params: want 2"},
		{"invalid query", types.MakeErrorInvalidQuery("syntax"), "postgres: invalid query: syntax"},
		{"value conversion failed", types.MakeErrorValueConversionFailed("int8"), "postgres: value conversion failed: int8"},
		{"other", types.MakeErrorOther("odd"), "postgres: odd"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := convertError(tc.in); got.Error() != tc.want {
				t.Errorf("convertError() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestConvertErrorQueryFailed(t *testing.T) {
	in := types.MakeErrorQueryFailed(types.DbError{
		Code:     "23505",
		Severity: "ERROR",
		Message:  "duplicate key value violates unique constraint",
		Detail:   witTypes.Some("Key (id)=(1) already exists."),
		Extras:   []witTypes.Tuple2[string, string]{{F0: "constraint", F1: "users_pkey"}},
	})

	var db *DBError
	if !errors.As(convertError(in), &db) {
		t.Fatalf("convertError() is not a *DBError")
	}
	if db.Code != "23505" || db.Severity != "ERROR" || db.Detail != "Key (id)=(1) already exists." {
		t.Errorf("DBError = %+v", db)
	}
	if len(db.Extras) != 1 || db.Extras[0] != [2]string{"constraint", "users_pkey"} {
		t.Errorf("DBError.Extras = %v", db.Extras)
	}
	if want := "postgres: query failed: ERROR 23505: duplicate key value violates unique constraint"; db.Error() != want {
		t.Errorf("DBError.Error() = %q, want %q", db.Error(), want)
	}
}

func TestConvertErrorQueryFailedNoDetail(t *testing.T) {
	in := types.MakeErrorQueryFailed(types.DbError{
		Code:     "42P01",
		Severity: "ERROR",
		Message:  "relation does not exist",
		Detail:   witTypes.None[string](),
	})

	var db *DBError
	if !errors.As(convertError(in), &db) {
		t.Fatalf("convertError() is not a *DBError")
	}
	if db.Detail != "" || db.Extras != nil {
		t.Errorf("DBError = %+v, want no detail and no extras", db)
	}
}

func TestEncodeFloat64(t *testing.T) {
	// Values checked against Rust's num::Float::integer_decode.
	for _, tc := range []struct {
		in       float64
		mantissa uint64
		exponent int16
		sign     int8
	}{
		{1, 4503599627370496, -52, 1},
		{-2.5, 5629499534213120, -51, -1},
		{0, 0, -1075, 1},
	} {
		got := EncodeFloat64(tc.in)
		if got.F0 != tc.mantissa || got.F1 != tc.exponent || got.F2 != tc.sign {
			t.Errorf("EncodeFloat64(%v) = (%d, %d, %d), want (%d, %d, %d)",
				tc.in, got.F0, got.F1, got.F2, tc.mantissa, tc.exponent, tc.sign)
		}
	}
}

func TestFloat64RoundTrip(t *testing.T) {
	for _, f := range []float64{
		0, 1, -1, 0.1, -2.5, math.Pi, 1e300, -1e-300,
		math.MaxFloat64, math.SmallestNonzeroFloat64,
	} {
		if got := DecodeFloat64(EncodeFloat64(f)); got != f {
			t.Errorf("DecodeFloat64(EncodeFloat64(%v)) = %v", f, got)
		}
	}
}

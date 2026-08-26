package postgres

import (
	"math"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	types "go.wasmcloud.dev/component/imports/wasmcloud_postgres_0_2_0_types"
)

// Value is a Postgres data value, usable as a query parameter or returned in
// a result row. It aliases the generated `pg-value` variant; the helpers
// below construct the common cases, and the full set of constructors
// (MakePgValue*) lives in the generated
// go.wasmcloud.dev/component/imports/wasmcloud_postgres_0_2_0_types package.
type Value = types.PgValue

// Null returns a SQL NULL value.
func Null() Value { return types.MakePgValueNull() }

// Bool returns a boolean value.
func Bool(v bool) Value { return types.MakePgValueBool(v) }

// Int16 returns a smallint (int2) value.
func Int16(v int16) Value { return types.MakePgValueInt2(v) }

// Int32 returns an integer (int4) value.
func Int32(v int32) Value { return types.MakePgValueInt4(v) }

// Int64 returns a bigint (int8) value.
func Int64(v int64) Value { return types.MakePgValueInt8(v) }

// Float64 returns a double-precision (float8) value. The wire encoding is
// the sign/mantissa/exponent integer decomposition required by the WIT
// `hashable-f64` type.
func Float64(v float64) Value { return types.MakePgValueDouble(EncodeFloat64(v)) }

// Text returns a text value.
func Text(v string) Value { return types.MakePgValueText(v) }

// Bytea returns a bytea (raw bytes) value.
func Bytea(v []byte) Value { return types.MakePgValueBytea(v) }

// JSON returns a json value from an already-encoded JSON string.
func JSON(v string) Value { return types.MakePgValueJson(v) }

// JSONB returns a jsonb value from an already-encoded JSON string.
func JSONB(v string) Value { return types.MakePgValueJsonb(v) }

// UUID returns a uuid value from its canonical string form.
func UUID(v string) Value { return types.MakePgValueUuid(v) }

// Numeric returns an arbitrary-precision numeric value from its decimal
// string form.
func Numeric(v string) Value { return types.MakePgValueNumeric(v) }

// EncodeFloat64 decomposes a float64 into the (mantissa, exponent, sign)
// triple used by the WIT `hashable-f64` type. The decomposition matches
// Rust's num::Float::integer_decode: value = sign * mantissa * 2^exponent.
func EncodeFloat64(f float64) types.HashableF64 {
	bits := math.Float64bits(f)
	sign := int8(1)
	if bits>>63 != 0 {
		sign = -1
	}
	exponent := int16((bits >> 52) & 0x7ff)
	mantissa := bits & ((uint64(1) << 52) - 1)
	if exponent == 0 {
		mantissa <<= 1
	} else {
		mantissa |= uint64(1) << 52
	}
	exponent -= 1075 // 1023 (bias) + 52 (mantissa bits)
	return witTypes.Tuple3[uint64, int16, int8]{F0: mantissa, F1: exponent, F2: sign}
}

// DecodeFloat64 reassembles a float64 from the WIT `hashable-f64`
// sign/mantissa/exponent triple produced by the database or [EncodeFloat64].
func DecodeFloat64(v types.HashableF64) float64 {
	return float64(v.F2) * float64(v.F0) * math.Pow(2, float64(v.F1))
}

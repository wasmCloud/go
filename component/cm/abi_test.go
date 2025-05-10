package cm

import (
	"math"
	"testing"
)

func TestIntConversions(t *testing.T) {
	for i := int8(math.MinInt8); i < math.MaxInt8; i++ {
		testIntRoundTrip[uint32](t, i)
		testIntRoundTrip[uint64](t, i)
	}

	for i := uint8(0); i < math.MaxUint8; i++ {
		testIntRoundTrip[uint32](t, i)
		testIntRoundTrip[uint64](t, i)
	}

	for i := int16(math.MinInt16); i < math.MaxInt16; i++ {
		testIntRoundTrip[uint32](t, i)
		testIntRoundTrip[uint64](t, i)
	}

	for i := uint16(0); i < math.MaxUint16; i++ {
		testIntRoundTrip[uint32](t, i)
		testIntRoundTrip[uint64](t, i)
	}

	// int32/uint32 into uint32
	testIntRoundTrip[uint32](t, int32(0))
	testIntRoundTrip[uint32](t, int32(math.MinInt8))
	testIntRoundTrip[uint32](t, int32(math.MinInt16))
	testIntRoundTrip[uint32](t, int32(math.MinInt32))
	testIntRoundTrip[uint32](t, int32(math.MaxInt8))
	testIntRoundTrip[uint32](t, int32(math.MaxInt16))
	testIntRoundTrip[uint32](t, int32(math.MaxInt32))
	testIntRoundTrip[uint32](t, uint32(0))
	testIntRoundTrip[uint32](t, uint32(math.MaxUint8))
	testIntRoundTrip[uint32](t, uint32(math.MaxUint16))
	testIntRoundTrip[uint32](t, uint32(math.MaxUint32))

	// int32/uint32 into uint64
	testIntRoundTrip[uint64](t, int32(0))
	testIntRoundTrip[uint64](t, int32(math.MinInt8))
	testIntRoundTrip[uint64](t, int32(math.MinInt16))
	testIntRoundTrip[uint64](t, int32(math.MinInt32))
	testIntRoundTrip[uint64](t, int32(math.MaxInt8))
	testIntRoundTrip[uint64](t, int32(math.MaxInt16))
	testIntRoundTrip[uint64](t, int32(math.MaxInt32))
	testIntRoundTrip[uint64](t, uint32(0))
	testIntRoundTrip[uint64](t, uint32(math.MaxUint8))
	testIntRoundTrip[uint64](t, uint32(math.MaxUint16))
	testIntRoundTrip[uint64](t, uint32(math.MaxUint32))

	// int64/uint64 into uint64
	testIntRoundTrip[uint64](t, int64(0))
	testIntRoundTrip[uint64](t, int64(math.MinInt8))
	testIntRoundTrip[uint64](t, int64(math.MinInt16))
	testIntRoundTrip[uint64](t, int64(math.MinInt32))
	testIntRoundTrip[uint64](t, int64(math.MaxInt8))
	testIntRoundTrip[uint64](t, int64(math.MaxInt16))
	testIntRoundTrip[uint64](t, int64(math.MaxInt32))
	testIntRoundTrip[uint64](t, uint64(0))
	testIntRoundTrip[uint64](t, uint64(math.MaxUint8))
	testIntRoundTrip[uint64](t, uint64(math.MaxUint16))
	testIntRoundTrip[uint64](t, uint64(math.MaxUint32))
}

func testIntRoundTrip[Core CoreIntegers, From Integers](t *testing.T, want From) {
	core := Core(want) // Convert to a core integer type
	got := From(core)  // Convert back to original type
	if got != want {
		t.Errorf("testLowerLift[%T, %T](t, %v): got %v, expected %v", want, core, want, got, want)
	}
}

type Integers interface {
	int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | uintptr
}

type CoreIntegers interface {
	uint32 | uint64
}

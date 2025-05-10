package cm

import (
	"math"
	"testing"
)

func TestTuple(t *testing.T) {
	var HL HostLayout
	_ = Tuple[string, bool]{HL, "hello", false}
	_ = Tuple3[string, bool, uint8]{HL, "hello", false, 1}
	_ = Tuple4[string, bool, uint8, uint16]{HL, "hello", false, 1, 32000}
	_ = Tuple5[string, bool, uint8, uint16, uint32]{HL, "hello", false, 1, 32000, 1_000_000}
	_ = Tuple6[string, bool, uint8, uint16, uint32, uint64]{HL, "hello", false, 1, 32000, 1_000_000, 5_000_000_000}
	_ = Tuple7[string, bool, uint8, uint16, uint32, uint64, float32]{HL, "hello", false, math.MaxUint8, math.MaxUint16, math.MaxUint32, math.MaxUint64, math.MaxFloat32}
	_ = Tuple8[string, bool, uint8, uint16, uint32, uint64, float32, float64]{HL, "hello", false, math.MaxUint8, math.MaxUint16, math.MaxUint32, math.MaxUint64, math.MaxFloat32, math.MaxFloat64}
}

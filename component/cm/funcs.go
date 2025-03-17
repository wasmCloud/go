package cm

import (
	"unsafe"

	bcm "go.bytecodealliance.org/cm"
)

func BoolToU32[B ~bool](v B) uint32 { return bcm.BoolToU32(v) }

func Case[T any, V AnyVariant[Tag, Shape, Align], Tag Discriminant, Shape, Align any](v *V, tag Tag) *T {
	return bcm.Case[T](v, tag)
}

func Err[R AnyResult[Shape, OK, Err], Shape, OK, Err any](err Err) R {
	return bcm.Err[R](err)
}

func F32ToU32(v float32) uint32 {
	return bcm.F32ToU32(v)
}

func F32ToU64(v float32) uint64 {
	return bcm.F32ToU64(v)
}

func F64ToU64(v float64) uint64 {
	return bcm.F64ToU64(v)
}

func LiftList[L AnyList[T], T any, Data unsafe.Pointer | uintptr | *T, Len AnyInteger](data Data, datalen Len) L {
	return bcm.LiftList[L](data, datalen)
}

func LiftString[T ~string, Data unsafe.Pointer | uintptr | *uint8, Len AnyInteger](data Data, datalen Len) T {
	return bcm.LiftString[T](data, datalen)
}

func LowerList[L AnyList[T], T any](list L) (*T, uint32) {
	return bcm.LowerList(list)
}

func LowerString[S ~string](s S) (*byte, uint32) {
	return bcm.LowerString(s)
}

func New[V AnyVariant[Tag, Shape, Align], Tag Discriminant, Shape, Align any, T any](tag Tag, data T) V {
	return bcm.New[V](tag, data)
}

func OK[R AnyResult[Shape, OK, Err], Shape, OK, Err any](ok OK) R {
	return bcm.OK[R](ok)
}

func PointerToU32[T any](v *T) uint32 {
	return bcm.PointerToU32(v)
}

func PointerToU64[T any](v *T) uint64 {
	return bcm.PointerToU64(v)
}

func Reinterpret[T, From any](from From) (to T) {
	return bcm.Reinterpret[T](from)
}

func U32ToBool(v uint32) bool {
	return bcm.U32ToBool(v)
}

func U32ToF32(v uint32) float32 {
	return bcm.U32ToF32(v)
}

func U32ToPointer[T any](v uint32) *T {
	return bcm.U32ToPointer[T](v)
}

func U64ToF32(v uint64) float32 {
	return bcm.U64ToF32(v)
}

func U64ToF64(v uint64) float64 {
	return bcm.U64ToF64(v)
}

func U64ToPointer[T any](v uint64) *T {
	return bcm.U64ToPointer[T](v)
}

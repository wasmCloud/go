package cm_idiomatic

import (
	"go.bytecodealliance.org/cm"
)

// ToPtr converts a Component Model Option to a Go pointer.
// Returns nil if the Option is None, otherwise returns a pointer to the value.
func ToPtr[T any](opt cm.Option[T]) *T {
	return opt.Some()
}

// FromPtr converts a Go pointer to a Component Model Option.
// Returns None if the pointer is nil, otherwise returns Some with the dereferenced value.
func FromPtr[T any](ptr *T) cm.Option[T] {
	if ptr == nil {
		return cm.None[T]()
	}
	return cm.Some(*ptr)
}

// ToSlice converts a Component Model List to a Go slice.
func ToSlice[T any](list cm.List[T]) []T {
	return list.Slice()
}

// FromSlice converts a Go slice to a Component Model List.
func FromSlice[T any](slice []T) cm.List[T] {
	return cm.ToList(slice)
}

// ToMap converts a Component Model List of tuples to a Go map.
func ToMap[K comparable, V any](list cm.List[cm.Tuple[K, V]]) map[K]V {
	slice := ToSlice(list)
	m := make(map[K]V, len(slice))

	for i := 0; i < len(slice); i++ {
		tuple := slice[i]
		m[tuple.F0] = tuple.F1
	}

	return m
}

// FromMap converts a Go map to a Component Model List of tuples.
func FromMap[K comparable, V any](m map[K]V) cm.List[cm.Tuple[K, V]] {
	tuples := make([]cm.Tuple[K, V], 0, len(m))
	for k, v := range m {
		tuples = append(tuples, cm.Tuple[K, V]{F0: k, F1: v})
	}
	return FromSlice(tuples)
}

// FromResult converts Go's (value, error) pattern to a Component Model Result.
// Returns OK(value) if error is nil, otherwise returns Err(error).
func FromResult[R cm.AnyResult[Shape, T, E], Shape, T, E any](value T, err E) R {
	var zero E
	// Check if err is the zero value (e.g., nil for error interface)
	if any(err) == any(zero) {
		return cm.OK[R](value)
	}
	return cm.Err[R](err)
}

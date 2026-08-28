# cm_idiomatic

Idiomatic Go conversions for WebAssembly Component Model types.

## Overview

This library provides simple functions to convert between Go types and Component Model types:

- **Option** ↔ **Pointer**: Convert between `cm.Option[T]` and `*T`
- **List** ↔ **Slice**: Convert between `cm.List[T]` and `[]T`
- **Map** ↔ **Tuples**: Convert between Go maps and Component Model tuple lists
- **Result** ↔ **Error**: Convert between `cm.Result[T, T, E]` and Go's `(T, error)` pattern

### Option

```go
// Go pointer to Component Model Option
value := 42
opt := cm_idiomatic.FromPtr(&value) // cm.Option[int]

// Component Model Option to Go pointer
ptr := cm_idiomatic.ToPtr(opt) // *int
```

### List

```go
// Go slice to Component Model List
slice := []string{"hello", "world"}
list := cm_idiomatic.FromSlice(slice) // cm.List[string]

// Component Model List to Go slice
result := cm_idiomatic.ToSlice(list) // []string
```

### Map

```go
// Go map to Component Model tuple list
m := map[string]int{"foo": 1, "bar": 2}
tuples := cm_idiomatic.FromMap(m) // cm.List[cm.Tuple[string, int]]

// Component Model tuple list to Go map
result := cm_idiomatic.ToMap(tuples) // map[string]int
```

### Result

```go
// Go (value, error) to Component Model Result
result := cm_idiomatic.FromError[string, error]("success", nil)
// Returns cm.Result[string, string, error]

// Component Model Result to Go (value, error)
value, err := cm_idiomatic.ToError(result) // (string, error)
```

## Installation

```bash
go get go.wasmcloud.dev/x/cm_idiomatic
```

## Usage

```go
import "go.wasmcloud.dev/x/cm_idiomatic"
```

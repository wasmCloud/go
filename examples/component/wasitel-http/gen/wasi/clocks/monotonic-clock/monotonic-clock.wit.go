// Code generated by wit-bindgen-go. DO NOT EDIT.

// Package monotonicclock represents the imported interface "wasi:clocks/monotonic-clock@0.2.0".
package monotonicclock

import (
	"github.com/wasmCloud/go/examples/component/wasitel-http/gen/wasi/io/poll"
	"go.bytecodealliance.org/cm"
)

// Pollable represents the imported type alias "wasi:clocks/monotonic-clock@0.2.0#pollable".
//
// See [poll.Pollable] for more information.
type Pollable = poll.Pollable

// Instant represents the u64 "wasi:clocks/monotonic-clock@0.2.0#instant".
//
//	type instant = u64
type Instant uint64

// Duration represents the u64 "wasi:clocks/monotonic-clock@0.2.0#duration".
//
//	type duration = u64
type Duration uint64

// Now represents the imported function "now".
//
//	now: func() -> instant
//
//go:nosplit
func Now() (result Instant) {
	result0 := wasmimport_Now()
	result = (Instant)((uint64)(result0))
	return
}

// Resolution represents the imported function "resolution".
//
//	resolution: func() -> duration
//
//go:nosplit
func Resolution() (result Duration) {
	result0 := wasmimport_Resolution()
	result = (Duration)((uint64)(result0))
	return
}

// SubscribeInstant represents the imported function "subscribe-instant".
//
//	subscribe-instant: func(when: instant) -> pollable
//
//go:nosplit
func SubscribeInstant(when Instant) (result Pollable) {
	when0 := (uint64)(when)
	result0 := wasmimport_SubscribeInstant((uint64)(when0))
	result = cm.Reinterpret[Pollable]((uint32)(result0))
	return
}

// SubscribeDuration represents the imported function "subscribe-duration".
//
//	subscribe-duration: func(when: duration) -> pollable
//
//go:nosplit
func SubscribeDuration(when Duration) (result Pollable) {
	when0 := (uint64)(when)
	result0 := wasmimport_SubscribeDuration((uint64)(when0))
	result = cm.Reinterpret[Pollable]((uint32)(result0))
	return
}

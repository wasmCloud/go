// Package sleep provides a sleep that works inside async-lifted exports,
// where [time.Sleep] does not.
//
// A Go handler serving an async WIT export — every wasmcloud:nats handler,
// for one — traps if it parks a goroutine on a Go runtime timer: the export
// returns without producing a result, and the host reports
//
//	wasm trap: async-lifted export failed to produce a result
//
// Only the timer path is affected; a handler that spins or burns CPU for the
// same duration completes. That removes debouncing, backoff, rate-limiting
// and polling from such handlers — unless the wait goes through the host
// instead of the Go runtime, which is what this package does: [Sleep] awaits
// `wasi:clocks/monotonic-clock@0.3.0.wait-for`, an ordinary async host
// import, the same mechanism every wasmcloud:nats call already uses.
//
// # Requirements
//
// The component's world must import the P3 monotonic clock:
//
//	import wasi:clocks/monotonic-clock@0.3.0;
//
// The `wasmcloud:nats-guest@0.1.0` worlds import it already, so a world built
// as
//
//	world app {
//	  include wasmcloud:nats-guest/subscriber@0.1.0;
//	}
//
// needs nothing further. And the call must run under componentize-go's async
// runtime — any world with an async export or import qualifies, which every
// wasmcloud:nats world is.
//
// Calling [Sleep] outside an async world, or in a world without the import,
// fails at componentize time with an unresolved-import error naming
// `wasi:clocks/monotonic-clock@0.3.0` — loudly, not at run time.
package sleep

import (
	"time"

	witAsync "go.bytecodealliance.org/pkg/wit/async"
)

// The generated shape for an async-lowered import, mirrored by hand so this
// package needs no bindings regeneration: componentize-go emits exactly this
// for `wait-for` when the world imports the interface.
//
//go:wasmimport wasi:clocks/monotonic-clock@0.3.0 [async-lower]wait-for
func wasmimport_wait_for(howLongNanos int64) int32

// Sleep pauses the calling goroutine for at least d, without parking the Go
// runtime: the wait happens in the host's clock, and other work on the
// instance keeps running meanwhile.
//
// A zero or negative d returns immediately.
func Sleep(d time.Duration) {
	if d <= 0 {
		return
	}
	witAsync.SubtaskWait(uint32(wasmimport_wait_for(int64(d))))
}

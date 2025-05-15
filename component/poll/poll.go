package io

import (
	"runtime"
	"time"

	monotonicclock "go.wasmcloud.dev/component/gen/wasi/clocks/monotonic-clock"
	"go.wasmcloud.dev/component/gen/wasi/http/types"
)

// PollWithBackoff is a utility function that polls a given Pollable object
// until it is ready, using an exponential backoff strategy starting at 1ms
// and capping at 5 seconds. It uses a wasi:clocks/monotonic-clock pollable
// to backoff and yield the thread to the Go runtime scheduler.
func PollWithBackoff(pollable types.Pollable) {
	backoffDuration := 1 * time.Millisecond
	for !pollable.Ready() {
		runtime.Gosched()
		backoff := monotonicclock.SubscribeDuration(monotonicclock.Duration(backoffDuration))
		backoff.Block()
		backoffDuration *= 2
		if backoffDuration > 5*time.Second {
			backoffDuration = 5 * time.Second // Cap the backoff duration
		}
	}
}

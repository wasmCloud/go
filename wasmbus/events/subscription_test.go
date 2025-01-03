package events

import (
	"context"
	"testing"
	"time"

	"go.wasmcloud.dev/wasmbus"
	"go.wasmcloud.dev/wasmbus/wasmbustest"
)

func TestEventSubscription(t *testing.T) {
	nc, teardown := wasmbustest.WithWash(t)
	defer teardown(t)

	sub, err := Subscribe(
		wasmbus.NewNatsBus(nc),
		"default",
		wasmbus.PatternAll,
		wasmbus.NoBackLog,
		DiscardErrorsHandler(func(ctx context.Context, ev Event) {
			switch bv := ev.BusEvent.(type) {
			default:
				t.Logf("Unknown event type %s %+T", ev.CloudEvent.Type(), bv)
			}
		}))
	if err != nil {
		t.Fatal(err)
	}

	<-time.After(1 * time.Minute)
	t.Log("unsubscribing")
	sub.Drain()
	<-time.After(30 * time.Second)

	t.Log("exiting")
}

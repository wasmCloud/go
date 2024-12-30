package events

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"go.wasmcloud.dev/wasmbus"
)

func TestEventSubscription(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatal(err)
	}

	unsub, err := Subscribe(
		wasmbus.NewNatsBus(nc),
		"default",
		wasmbus.PatternAll,
		wasmbus.NoBackLog,
		func(ctx context.Context, msg *wasmbus.Message, ev Event, err error) {
			if err != nil {
				t.Log("ERR", err)
				return
			}
			switch bv := ev.BusEvent.(type) {
			default:
				t.Logf("Unknown event type %s %+T", ev.CloudEvent.Type(), bv)
			}

		})

	if err != nil {
		t.Fatal(err)
	}

	<-time.After(1 * time.Minute)
	t.Log("unsubscribing")
	unsub()
	<-time.After(30 * time.Second)

	t.Log("exiting")
}

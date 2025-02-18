package events

import (
	"context"
	"testing"
	"time"

	"go.wasmcloud.dev/x/wasmbus"
	"go.wasmcloud.dev/x/wasmbus/wasmbustest"
)

func TestRouter(t *testing.T) {
	defer wasmbustest.MustStartNats(t)()
	nc, err := wasmbus.NatsConnect(wasmbus.NatsDefaultURL)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()

	bus := wasmbus.NewNatsBus(nc)

	t.Run("valid route", func(t *testing.T) {
		router := NewEventRouter()

		evChan := make(chan *HostHeartbeat, 1)
		sub, err := Subscribe(
			bus,
			"default",
			wasmbus.PatternAll,
			wasmbus.NoBackLog,
			router)
		if err != nil {
			t.Fatal(err)
		}
		router.AddRoute("heartbeat", Route(func(ctx context.Context, ev *HostHeartbeat) {
			evChan <- ev
		}))

		// Publish an event
		hbEv, err := EncodeEvent("com.wasmcloud.lattice.host_heartbeat", "test", "test", HostHeartbeat{
			HostId: "my-host-name",
		})
		if err != nil {
			t.Fatal(err)
		}
		evMsg, err := wasmbus.Encode("wasmbus.evt.default.host_heartbeat", &hbEv.CloudEvent)
		if err != nil {
			t.Fatal(err)
		}

		if err := bus.Publish(evMsg); err != nil {
			t.Fatal(err)
		}

		var ev *HostHeartbeat
		select {
		case ev = <-evChan:
			break
		case <-time.After(1 * time.Second):
			t.Fatal("expected event, got none")
		}

		if want, got := "my-host-name", ev.HostId; want != got {
			t.Fatalf("want %q, got %q", want, got)
		}

		if err := sub.Drain(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid route", func(t *testing.T) {
		router := NewEventRouter()

		evChan := make(chan *HostStopped, 1)
		sub, err := Subscribe(
			bus,
			"default",
			wasmbus.PatternAll,
			wasmbus.NoBackLog,
			router)
		if err != nil {
			t.Fatal(err)
		}
		router.AddRoute("host-stop", Route(func(ctx context.Context, ev *HostStopped) {
			evChan <- ev
		}))

		// Publish an event
		hbEv, err := EncodeEvent("com.wasmcloud.lattice.host_heartbeat", "test", "test", HostHeartbeat{
			HostId: "my-host-name",
		})
		if err != nil {
			t.Fatal(err)
		}
		evMsg, err := wasmbus.Encode("wasmbus.evt.default.host_heartbeat", &hbEv.CloudEvent)
		if err != nil {
			t.Fatal(err)
		}

		if err := bus.Publish(evMsg); err != nil {
			t.Fatal(err)
		}

		if err := sub.Drain(); err != nil {
			t.Fatal(err)
		}

		select {
		case <-evChan:
			t.Fatal("expected no event, got one")
		default:
		}
	})
}

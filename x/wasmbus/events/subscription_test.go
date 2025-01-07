package events

import (
	"context"
	"testing"
	"time"

	"go.wasmcloud.dev/x/wasmbus"
	"go.wasmcloud.dev/x/wasmbus/wasmbustest"
)

type errHandler struct {
	handleEventFunc func(context.Context, Event)
	handleErrorFunc func(context.Context, *wasmbus.Message, error)
}

func (e errHandler) HandleEvent(ctx context.Context, ev Event) {
	if e.handleEventFunc != nil {
		e.handleEventFunc(ctx, ev)
	}
}

func (e errHandler) HandleError(ctx context.Context, msg *wasmbus.Message, err error) {
	if e.handleErrorFunc != nil {
		e.handleErrorFunc(ctx, msg, err)
	}
}

func TestErrorHandlerSubscription(t *testing.T) {
	defer wasmbustest.MustStartNats(t)()
	nc, err := wasmbus.NatsConnect(wasmbus.NatsDefaultURL)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()

	bus := wasmbus.NewNatsBus(nc)

	evChan := make(chan Event, 1)
	errChan := make(chan error, 1)
	sub, err := Subscribe(
		bus,
		"default",
		wasmbus.PatternAll,
		wasmbus.NoBackLog,
		&errHandler{
			handleEventFunc: func(ctx context.Context, ev Event) {
				evChan <- ev
			},
			handleErrorFunc: func(ctx context.Context, msg *wasmbus.Message, err error) {
				errChan <- err
			},
		})
	if err != nil {
		t.Fatal(err)
	}

	// Publish an event ( missing event id )
	hbEv, err := EncodeEvent("com.wasmcloud.lattice.host_heartbeat", "test", "", HostHeartbeat{})
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

	select {
	case <-evChan:
		t.Error("expected error, got event")
	case <-errChan:
		break
	case <-time.After(1 * time.Second):
		t.Fatal("expected event, got none")
	}

	if err := sub.Drain(); err != nil {
		t.Fatal(err)
	}
}

func TestEventSubscription(t *testing.T) {
	defer wasmbustest.MustStartNats(t)()
	nc, err := wasmbus.NatsConnect(wasmbus.NatsDefaultURL)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Close()

	bus := wasmbus.NewNatsBus(nc)

	evChan := make(chan Event, 1)
	sub, err := Subscribe(
		bus,
		"default",
		wasmbus.PatternAll,
		wasmbus.NoBackLog,
		DiscardErrorsHandler(func(ctx context.Context, ev Event) {
			evChan <- ev
		}))
	if err != nil {
		t.Fatal(err)
	}

	// Publish an event
	hbEv, err := EncodeEvent("com.wasmcloud.lattice.host_heartbeat", "test", "test", HostHeartbeat{})
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

	var ev Event
	select {
	case ev = <-evChan:
		break
	case <-time.After(1 * time.Second):
		t.Fatal("expected event, got none")
	}

	if ev.CloudEvent.Type() != "com.wasmcloud.lattice.host_heartbeat" {
		t.Fatalf("expected event type 'com.wasmcloud.lattice.host_heartbeat', got '%s'", ev.CloudEvent.Type())
	}

	if err := sub.Drain(); err != nil {
		t.Fatal(err)
	}
}

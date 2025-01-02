package events

import (
	"context"
	"strings"

	"go.wasmcloud.dev/wasmbus"
)

// Event is a parsed event from the bus
type EventHandler interface {
	HandleEvent(context.Context, Event)
	HandleError(context.Context, *wasmbus.Message, error)
}

// DiscardErrorsHandler is a simple handler that discards errors
type DiscardErrorsHandler func(context.Context, Event)

func (h DiscardErrorsHandler) HandleError(context.Context, *wasmbus.Message, error) {}
func (h DiscardErrorsHandler) HandleEvent(ctx context.Context, ev Event)            { h(ctx, ev) }

// Subscribe creates a subscription to events on the specified lattice and pattern.
// The pattern is a glob pattern that is matched against the event type. Use `wasmbus.PatternAll` for all events.
// The backlog parameter is the maximum number of messages that can be buffered in memory.
// See `DiscardErrorsHandler` for a simple handler implementation that ignores errors.
func Subscribe(b wasmbus.Bus, lattice string, pattern string, backlog int, handler EventHandler) (wasmbus.Subscription, error) {
	subject := strings.Join([]string{wasmbus.PrefixEvents, lattice, pattern}, ".")
	sub, err := b.Subscribe(subject, backlog)
	if err != nil {
		return nil, err
	}

	go sub.Handle(func(msg *wasmbus.Message) {
		ctx := context.Background()
		ev, err := ParseEvent(msg.Data)
		if err != nil {
			handler.HandleError(ctx, msg, err)
			return
		}
		handler.HandleEvent(ctx, ev)
	})

	return sub, nil
}

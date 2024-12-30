package events

import (
	"context"
	"strings"

	"go.wasmcloud.dev/wasmbus"
)

type EventCallback func(context.Context, *wasmbus.Message, Event, error)

func Subscribe(b wasmbus.Bus, lattice string, pattern string, backlog int, callback EventCallback) (func(), error) {
	subject := strings.Join([]string{wasmbus.PrefixEvents, lattice, pattern}, ".")
	sub, err := b.Subscribe(subject, backlog)
	if err != nil {
		return nil, err
	}

	go sub.Handle(func(msg *wasmbus.Message) {
		ev, err := ParseEvent(msg.Data)
		callback(context.Background(), msg, ev, err)
	})

	return func() { _ = sub.Drain() }, nil
}

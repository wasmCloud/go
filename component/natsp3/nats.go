package natsp3

import (
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	core "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_2_0_core"
	types "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_2_0_types"
)

// Header is a single NATS message header.
type Header struct {
	Name  string
	Value string
}

// Message is a NATS message, for both core and JetStream.
type Message struct {
	Subject string
	Body    []byte
	// ReplyTo is the subject a reply should be published to, or "" if the
	// message expects no reply.
	ReplyTo string
	Headers []Header
}

// Header returns the first value for name, and whether it was present.
func (m Message) Header(name string) (string, bool) {
	for _, h := range m.Headers {
		if h.Name == name {
			return h.Value, true
		}
	}
	return "", false
}

// Publish sends msg to its subject and returns once it is written to the
// connection — not once a subscriber has seen it. Core NATS is
// fire-and-forget; use [JetStreamPublish] for durable delivery.
func Publish(msg Message) error {
	if res := core.Publish(toWitMessage(msg)); res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// Request publishes msg and waits up to timeout for a single reply.
//
// Replies arrive on a per-workload inbox prefix, so two workloads on one
// host cannot observe each other's responses.
func Request(msg Message, timeout time.Duration) (Message, error) {
	res := core.Request(toWitMessage(msg), uint32(timeout.Milliseconds()))
	if res.IsErr() {
		return Message{}, convertError(res.Err())
	}
	return FromWitMessage(res.Ok()), nil
}

func toWitHeaders(headers []Header) witTypes.Option[[]types.HeaderEntry] {
	if len(headers) == 0 {
		return witTypes.None[[]types.HeaderEntry]()
	}
	out := make([]types.HeaderEntry, 0, len(headers))
	for _, h := range headers {
		out = append(out, types.HeaderEntry{Name: h.Name, Value: h.Value})
	}
	return witTypes.Some(out)
}

func toWitMessage(msg Message) types.NatsMessage {
	replyTo := witTypes.None[string]()
	if msg.ReplyTo != "" {
		replyTo = witTypes.Some(msg.ReplyTo)
	}
	return types.NatsMessage{
		Subject: msg.Subject,
		Body:    msg.Body,
		ReplyTo: replyTo,
		Headers: toWitHeaders(msg.Headers),
	}
}

// FromWitMessage converts a generated binding message into a [Message]. It is
// exported for the nats/*handler subpackages; applications should not need it.
func FromWitMessage(msg types.NatsMessage) Message {
	var headers []Header
	if msg.Headers.IsSome() {
		entries := msg.Headers.Some()
		headers = make([]Header, 0, len(entries))
		for _, e := range entries {
			headers = append(headers, Header{Name: e.Name, Value: e.Value})
		}
	}
	return Message{
		Subject: msg.Subject,
		Body:    msg.Body,
		ReplyTo: msg.ReplyTo.SomeOr(""),
		Headers: headers,
	}
}

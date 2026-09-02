package nats

import (
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	core "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_core"
	types "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_types"
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

// Core is the core-NATS surface this package calls: exactly the two functions
// `wasmcloud:nats/core@0.1.0` defines, in the generated bindings' own types.
//
// The default implementation is the SDK's committed bindings, which name the
// plain (unlabeled) instance. A component that imports the interface under an
// `(implements ..)` label — because its manifest's hostInterfaces entry sets
// `name:` — generates its own bindings for that label and satisfies this
// interface with a few lines of forwarding, rather than reimplementing the
// package. See the "Labeled bindings" section of the package doc.
type Core interface {
	Publish(msg types.NatsMessage) witTypes.Result[witTypes.Unit, types.NatsError]
	Request(msg types.NatsMessage, timeoutMs uint32) witTypes.Result[types.NatsMessage, types.NatsError]
}

// Conn is one core-NATS binding: the connection, credentials, and grants the
// host attached to a single instance of `wasmcloud:nats/core@0.1.0`. A
// component with more than one binding holds one [Conn] per label.
type Conn struct{ core Core }

// NewConn wraps a generated core bindings package as a [Conn].
func NewConn(c Core) *Conn { return &Conn{core: c} }

// plainCore is [Core] over the committed bindings for the plain, unlabeled
// instance — the only instance name a committed binding can carry, because
// //go:wasmimport takes it as a compile-time literal.
type plainCore struct{}

func (plainCore) Publish(msg types.NatsMessage) witTypes.Result[witTypes.Unit, types.NatsError] {
	return core.Publish(msg)
}

func (plainCore) Request(msg types.NatsMessage, timeoutMs uint32) witTypes.Result[types.NatsMessage, types.NatsError] {
	return core.Request(msg, timeoutMs)
}

// Default is the connection bound to the plain (unlabeled) instance, the one
// an unnamed hostInterfaces entry routes to. The package-level [Publish] and
// [Request] are its methods.
var Default = NewConn(plainCore{})

// Publish sends msg to its subject and returns once it is written to the
// connection — not once a subscriber has seen it. Core NATS is
// fire-and-forget; use [JetStreamPublish] for durable delivery.
func Publish(msg Message) error { return Default.Publish(msg) }

// Request publishes msg and waits up to timeout for a single reply.
//
// Replies arrive on a per-workload inbox prefix, so two workloads on one
// host cannot observe each other's responses.
func Request(msg Message, timeout time.Duration) (Message, error) {
	return Default.Request(msg, timeout)
}

// Publish sends msg to its subject over this connection. See [Publish].
func (c *Conn) Publish(msg Message) error {
	if res := c.core.Publish(toWitMessage(msg)); res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// Request publishes msg over this connection and waits up to timeout for a
// single reply. See [Request].
func (c *Conn) Request(msg Message, timeout time.Duration) (Message, error) {
	res := c.core.Request(toWitMessage(msg), uint32(timeout.Milliseconds()))
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

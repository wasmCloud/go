package nats

import (
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	js "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_jetstream"
)

// MsgIDHeader deduplicates a JetStream publish within the stream's duplicate
// window. The window is stream configuration, not a guarantee this interface
// makes: without one configured, setting this header changes nothing.
const MsgIDHeader = "Nats-Msg-Id"

// PublishAck is the server's acknowledgement of a JetStream publish.
type PublishAck struct {
	Stream   string
	Sequence uint64
	// Duplicate reports that the server recognised the message as a
	// duplicate within the stream's duplicate window and did not store it
	// again. Only ever true when the publish carried [MsgIDHeader].
	Duplicate bool
}

// StoredMessage is a read-only snapshot of a message held in a stream.
type StoredMessage struct {
	Subject  string
	Sequence uint64
	Data     []byte
	Headers  []Header
}

// JetStreamPublish publishes msg durably and waits for the server ack.
//
// Unlike core [Publish], this returns only once the message is stored.
func JetStreamPublish(msg Message) (PublishAck, error) {
	res := js.Publish(toWitMessage(msg))
	if res.IsErr() {
		return PublishAck{}, convertError(res.Err())
	}
	ack := res.Ok()
	return PublishAck{
		Stream:    ack.StreamName,
		Sequence:  ack.Sequence,
		Duplicate: ack.Duplicate,
	}, nil
}

// GetBySequence fetches one message from a stream by its sequence number.
//
// Requires NATS server 2.9 or newer; older servers return an
// [UnsupportedByServerError].
func GetBySequence(stream string, sequence uint64) (StoredMessage, error) {
	res := js.GetBySequence(stream, sequence)
	if res.IsErr() {
		return StoredMessage{}, convertError(res.Err())
	}
	return fromWitStored(res.Ok()), nil
}

// Scan replays up to maxCount messages from stream, starting at
// startSequence. It creates no durable consumer, so it does not affect any
// other reader's position — this is the replay path that
// wasmcloud:messaging cannot express.
//
// The host caps both the count and the time one call may take, so a large
// range may come back short; call again from the last sequence seen.
func Scan(stream string, startSequence uint64, maxCount uint32) ([]StoredMessage, error) {
	res := js.Scan(stream, startSequence, maxCount)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	raw := res.Ok()
	out := make([]StoredMessage, 0, len(raw))
	for _, m := range raw {
		out = append(out, fromWitStored(m))
	}
	return out, nil
}

// PullConsumer is a batched reader over an existing durable consumer.
type PullConsumer struct{ inner *js.PullConsumer }

// OpenPullConsumer attaches to a consumer that already exists on stream.
//
// It never creates or reconfigures one: consumer lifecycle is deliberately
// outside this interface, so provision it out-of-band (the NATS CLI, or
// deployment tooling). A missing consumer returns a [NotFoundError].
func OpenPullConsumer(stream, consumer string) (*PullConsumer, error) {
	res := js.OpenPullConsumer(stream, consumer)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	return &PullConsumer{inner: res.Ok()}, nil
}

// Fetch reads up to batch messages, waiting at most timeout for the first.
// It returns [ErrNoMessages] when the timeout elapses with none available,
// which is an idle result rather than a failure.
//
// Every returned handle must be settled — Ack, Nak, or Term — or the
// consumer stalls until ack-wait expires.
func (c *PullConsumer) Fetch(batch uint32, timeout time.Duration) ([]*MessageHandle, error) {
	res := c.inner.Fetch(batch, uint32(timeout.Milliseconds()))
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	raw := res.Ok()
	out := make([]*MessageHandle, 0, len(raw))
	for _, h := range raw {
		out = append(out, &MessageHandle{inner: h})
	}
	return out, nil
}

// Close releases the consumer handle. The consumer itself is untouched.
func (c *PullConsumer) Close() { c.inner.Drop() }

// MessageHandle is one JetStream delivery plus the ability to settle it.
//
// Delivery is at-least-once: the same message arrives again after any
// failure downstream of processing it, so the work must be idempotent.
// Settling is one-shot — a second Ack, Nak, or Term reports an error.
type MessageHandle struct{ inner *js.MessageHandle }

// NewMessageHandle wraps a generated handle. It is exported for the
// nats/jetstreamhandler subpackage; applications should not need it.
func NewMessageHandle(inner *js.MessageHandle) *MessageHandle {
	return &MessageHandle{inner: inner}
}

// Message returns the delivered message.
func (h *MessageHandle) Message() Message { return FromWitMessage(h.inner.Message()) }

// Sequence returns the message's position in the stream.
func (h *MessageHandle) Sequence() uint64 { return h.inner.Sequence() }

// DeliveryCount returns how many times this message has been delivered; 1 on
// the first attempt. A value above 1 means an earlier attempt did not ack.
func (h *MessageHandle) DeliveryCount() uint32 { return h.inner.DeliveryCount() }

// Ack acknowledges successful processing, so the message is not redelivered.
func (h *MessageHandle) Ack() error {
	if res := h.inner.Ack(); res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// Nak asks for redelivery after delay. A zero delay redelivers as soon as the
// server can, which spins if the failure is permanent — prefer a delay, or
// Term for something that can never succeed.
func (h *MessageHandle) Nak(delay time.Duration) error {
	d := witTypes.None[uint32]()
	if delay > 0 {
		d = witTypes.Some(uint32(delay.Milliseconds()))
	}
	if res := h.inner.Nak(d); res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// InProgress resets the ack-wait timer while work continues. Unlike the
// settling methods it may be called repeatedly.
func (h *MessageHandle) InProgress() error {
	if res := h.inner.InProgress(); res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// Term rejects the message permanently, so it is never redelivered. This is
// the right answer for a malformed payload that no retry can fix.
//
// Under a binding configured with `ack-mode: auto` the host owns the
// acknowledgement and this reports an error; use `ack-mode: manual` to settle
// from the guest.
func (h *MessageHandle) Term() error {
	if res := h.inner.Term(); res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

func fromWitStored(m js.StoredMessage) StoredMessage {
	var headers []Header
	if m.Headers.IsSome() {
		entries := m.Headers.Some()
		headers = make([]Header, 0, len(entries))
		for _, e := range entries {
			headers = append(headers, Header{Name: e.Name, Value: e.Value})
		}
	}
	return StoredMessage{
		Subject:  m.Subject,
		Sequence: m.Sequence,
		Data:     m.Data,
		Headers:  headers,
	}
}

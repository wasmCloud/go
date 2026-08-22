package natsp3

import (
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	js "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_2_0_jetstream"
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

// StreamInfo is a read-only snapshot of a stream.
type StreamInfo struct {
	Name string
	// Subjects the stream is configured to capture.
	Subjects      []string
	Messages      uint64
	Bytes         uint64
	FirstSequence uint64
	LastSequence  uint64
	ConsumerCount uint64
}

// SubjectCount is one subject a stream currently holds messages on.
type SubjectCount struct {
	Subject string
	Count   uint64
}

// ConsumerInfo is a read-only snapshot of a consumer, including the limits it
// was provisioned with. Those limits are set out-of-band — this interface
// reads them, it does not administer them.
type ConsumerInfo struct {
	Name   string
	Stream string
	// FilterSubject is empty when the consumer captures the stream's whole
	// subject space.
	FilterSubject string
	// MaxAckPending, MaxWaiting, and MaxDeliver are zero when unset.
	MaxAckPending uint64
	MaxWaiting    uint64
	MaxDeliver    uint64
	AckWait       time.Duration
	// NumAckPending is delivered but not yet acknowledged.
	NumAckPending uint64
	// NumPending is waiting to be delivered.
	NumPending uint64
	// NumRedelivered is awaiting redelivery after a nak or ack-wait expiry.
	NumRedelivered uint64
}

// GetStreamInfo returns a stream's configuration and current state.
func GetStreamInfo(stream string) (StreamInfo, error) {
	res := js.GetStreamInfo(stream)
	if res.IsErr() {
		return StreamInfo{}, convertError(res.Err())
	}
	i := res.Ok()
	return StreamInfo{
		Name:          i.Name,
		Subjects:      i.Subjects,
		Messages:      i.Messages,
		Bytes:         i.Bytes,
		FirstSequence: i.FirstSequence,
		LastSequence:  i.LastSequence,
		ConsumerCount: i.ConsumerCount,
	}, nil
}

// ListStreamSubjects returns the subjects a stream currently holds messages
// on, with per-subject counts. Pass ">" for the whole stream.
func ListStreamSubjects(stream, subjectFilter string) ([]SubjectCount, error) {
	res := js.ListStreamSubjects(stream, subjectFilter)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	raw := res.Ok()
	out := make([]SubjectCount, 0, len(raw))
	for _, c := range raw {
		out = append(out, SubjectCount{Subject: c.Subject, Count: c.Count})
	}
	return out, nil
}

// GetConsumerInfo reads a consumer's configuration and state without
// attaching to it.
func GetConsumerInfo(stream, consumer string) (ConsumerInfo, error) {
	res := js.GetConsumerInfo(stream, consumer)
	if res.IsErr() {
		return ConsumerInfo{}, convertError(res.Err())
	}
	return fromWitConsumerInfo(res.Ok()), nil
}

func fromWitConsumerInfo(i js.ConsumerInfo) ConsumerInfo {
	return ConsumerInfo{
		Name:           i.Name,
		Stream:         i.StreamName,
		FilterSubject:  i.FilterSubject,
		MaxAckPending:  i.MaxAckPending,
		MaxWaiting:     i.MaxWaiting,
		MaxDeliver:     i.MaxDeliver,
		AckWait:        time.Duration(i.AckWaitMs) * time.Millisecond,
		NumAckPending:  i.NumAckPending,
		NumPending:     i.NumPending,
		NumRedelivered: i.NumRedelivered,
	}
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

// FetchWithLimits is [PullConsumer.Fetch] bounded by bytes as well as count,
// for a consumer provisioned with a max-bytes limit. A maxBytes of 0 means no
// byte bound.
func (c *PullConsumer) FetchWithLimits(batch uint32, maxBytes uint64, timeout time.Duration) ([]*MessageHandle, error) {
	res := c.inner.FetchWithLimits(batch, maxBytes, uint32(timeout.Milliseconds()))
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

// Info returns the consumer's current configuration and state, including the
// limits it was provisioned with.
func (c *PullConsumer) Info() (ConsumerInfo, error) {
	res := c.inner.Info()
	if res.IsErr() {
		return ConsumerInfo{}, convertError(res.Err())
	}
	return fromWitConsumerInfo(res.Ok()), nil
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

// AckSync acknowledges and waits for the server to confirm it (double-ack).
//
// One round trip slower than Ack, and the only way to know the delivery will
// not be repeated after a server or client failure.
func (h *MessageHandle) AckSync() error {
	if res := h.inner.AckSync(); res.IsErr() {
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

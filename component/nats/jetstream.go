package nats

import (
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	js "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_jetstream"
	types "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_types"
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
	// FilterSubject is the singular filter, empty when unset. A consumer may
	// instead carry FilterSubjects; it captures the stream's whole subject
	// space only when both are empty.
	FilterSubject string
	// FilterSubjects is the multi-subject filter servers 2.10 and newer
	// support. Empty when the consumer uses the singular FilterSubject, or no
	// filter at all.
	FilterSubjects []string
	// MaxAckPending, MaxWaiting, and MaxDeliver are zero when unset.
	MaxAckPending uint64
	MaxWaiting    uint64
	MaxDeliver    uint64
	// MaxRequestBatch is the largest batch a single pull may ask for, zero
	// when unset. A Fetch above it is refused with a [LimitExceededError]
	// rather than trimmed, so size against this instead of discovering it.
	MaxRequestBatch uint64
	// MaxRequestMaxBytes is the largest byte bound a single pull may ask for,
	// zero when unset. It counts subject, reply subject, and payload — around
	// 63 bytes of overhead per small message — so leave margin.
	MaxRequestMaxBytes uint64
	AckWait            time.Duration
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
		Name:               i.Name,
		Stream:             i.StreamName,
		FilterSubject:      i.FilterSubject,
		FilterSubjects:     i.FilterSubjects,
		MaxAckPending:      i.MaxAckPending,
		MaxWaiting:         i.MaxWaiting,
		MaxDeliver:         i.MaxDeliver,
		MaxRequestBatch:    i.MaxRequestBatch,
		MaxRequestMaxBytes: i.MaxRequestMaxBytes,
		AckWait:            time.Duration(i.AckWaitMs) * time.Millisecond,
		NumAckPending:      i.NumAckPending,
		NumPending:         i.NumPending,
		NumRedelivered:     i.NumRedelivered,
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

// FetchStop says why a fetch stopped delivering.
//
// It is the difference between "that was everything" and "there is more, a
// bound cut this batch short" — without it a short batch and a drained
// consumer look identical, and a reader that stops on a short batch leaves
// messages behind.
type FetchStop uint8

const (
	// FetchStopBatchFilled means the batch was filled: batch messages came back.
	FetchStopBatchFilled FetchStop = FetchStop(js.FetchStopBatchFilled)
	// FetchStopDrained means the consumer had no more to give before the
	// timeout elapsed. What came back is everything that was there.
	FetchStopDrained FetchStop = FetchStop(js.FetchStopDrained)
	// FetchStopByteLimit means a byte bound ended the batch early — maxBytes
	// on the call, or the consumer's MaxRequestMaxBytes. More messages are
	// waiting, and the next fetch picks up where this one stopped.
	FetchStopByteLimit FetchStop = FetchStop(js.FetchStopByteLimit)
)

func (s FetchStop) String() string {
	switch s {
	case FetchStopBatchFilled:
		return "batch-filled"
	case FetchStopDrained:
		return "drained"
	case FetchStopByteLimit:
		return "byte-limit"
	default:
		return "unknown"
	}
}

// FetchedBatch is what one fetch returned, and why it ended.
type FetchedBatch struct {
	Messages []*MessageHandle
	Stop     FetchStop
}

// Close releases every handle in the batch.
//
// Settling a message does not release it: the host holds the delivered
// payload until the handle is dropped, and a worker that loops Fetch while
// only acking grows the host's memory by everything it has ever been
// delivered. `defer batch.Close()` after a successful Fetch is the safe
// shape; Close is idempotent and safe to call after settling each message
// individually.
func (b FetchedBatch) Close() {
	for _, h := range b.Messages {
		h.Close()
	}
}

// Fetch reads up to batch messages, waiting at most timeout for the first.
// It returns [ErrNoMessages] when the timeout elapses with none available,
// which is an idle result rather than a failure, and a [LimitExceededError]
// when batch is over the consumer's MaxRequestBatch — the server refuses such
// a request outright rather than trimming it, so retrying unchanged fails the
// same way.
//
// batch must be at least 1.
//
// Every returned handle must be settled — Ack, Nak, or Term — or the
// consumer stalls until ack-wait expires, and must be released with Close, or
// the host holds its payload for the life of the component. `defer
// batch.Close()` covers the whole batch.
func (c *PullConsumer) Fetch(batch uint32, timeout time.Duration) (FetchedBatch, error) {
	return fetched(c.inner.Fetch(batch, uint32(timeout.Milliseconds())))
}

// FetchWithLimits is [PullConsumer.Fetch] bounded by bytes as well as count,
// for a consumer provisioned with a max-bytes limit. A maxBytes of 0 means no
// byte bound.
//
// Check the returned [FetchStop]: a byte bound that cuts the batch short
// leaves messages waiting, and only Stop distinguishes that from a drained
// consumer.
func (c *PullConsumer) FetchWithLimits(batch uint32, maxBytes uint64, timeout time.Duration) (FetchedBatch, error) {
	return fetched(c.inner.FetchWithLimits(batch, maxBytes, uint32(timeout.Milliseconds())))
}

func fetched(res witTypes.Result[js.FetchedBatch, types.NatsError]) (FetchedBatch, error) {
	if res.IsErr() {
		return FetchedBatch{}, convertError(res.Err())
	}
	raw := res.Ok()
	out := FetchedBatch{
		Messages: make([]*MessageHandle, 0, len(raw.Messages)),
		Stop:     FetchStop(raw.Stop),
	}
	for _, h := range raw.Messages {
		out.Messages = append(out.Messages, &MessageHandle{inner: h})
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
type MessageHandle struct {
	inner *js.MessageHandle
	// Set once the underlying host resource has been dropped, so Close is
	// idempotent and a use-after-close is caught here rather than trapping
	// on a dangling resource index.
	closed bool
}

// NewMessageHandle wraps a generated handle. It is exported for the
// nats/jetstreamhandler subpackage; applications should not need it.
func NewMessageHandle(inner *js.MessageHandle) *MessageHandle {
	return &MessageHandle{inner: inner}
}

// Close releases the handle's host-side resource.
//
// This is not the same as settling. Ack, Nak, and Term tell the *server* what
// to do with the message; Close tells the *host* it may free the delivered
// payload. Until it is called the payload stays resident in the host, so a
// pull worker that loops Fetch and only acks walks the host's memory up by
// every byte it has ever been delivered, and eventually the host refuses
// further fetches (or is OOM-killed). Close after settling:
//
//	batch, err := consumer.Fetch(10, time.Second)
//	if err != nil {
//	  return err
//	}
//	defer batch.Close()
//	for _, h := range batch.Messages {
//	  if err := process(h.Message()); err != nil {
//	    _ = h.Nak(time.Second)
//	    continue
//	  }
//	  _ = h.Ack()
//	}
//
// Calling Close twice is a no-op. Any other method on a closed handle
// reports [ErrHandleClosed] rather than reaching the host.
//
// A handle delivered to a jetstreamhandler callback is owned by the host and
// does not need closing; Close on one is harmless but unnecessary.
func (h *MessageHandle) Close() {
	if h.closed {
		return
	}
	h.closed = true
	h.inner.Drop()
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
	if h.closed {
		return ErrHandleClosed
	}
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
	if h.closed {
		return ErrHandleClosed
	}
	if res := h.inner.AckSync(); res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// Nak asks for redelivery after delay. A zero delay redelivers as soon as the
// server can, which spins if the failure is permanent — prefer a delay, or
// Term for something that can never succeed.
func (h *MessageHandle) Nak(delay time.Duration) error {
	if h.closed {
		return ErrHandleClosed
	}
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
	if h.closed {
		return ErrHandleClosed
	}
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
	if h.closed {
		return ErrHandleClosed
	}
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

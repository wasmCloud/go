package messaging

import (
	"io"
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	consumer "go.wasmcloud.dev/component/imports/wasmcloud_messaging_0_3_0_consumer"
	types "go.wasmcloud.dev/component/imports/wasmcloud_messaging_0_3_0_types"
)

// BrokerMessage is a message sent to or received from the broker.
type BrokerMessage struct {
	Subject string
	// Body is the payload, which flows as a stream rather than as one
	// buffered value. Publishing reads it to exhaustion; on a message
	// received from the broker it is a stream from the host that also
	// implements io.Closer, and the handler must drain or close it.
	//
	// Use bytes.NewReader for a payload already in memory.
	Body io.Reader
	// ReplyTo is the subject a reply should be published to, or "" if the
	// message does not expect a reply.
	ReplyTo string
}

// Publish sends msg to its subject without awaiting a response.
//
// It returns once the host has consumed the whole body and the broker has
// accepted the message — not once any subscriber has received it.
func Publish(msg BrokerMessage) error {
	body, done := lowerBody(msg.Body)
	defer done()
	res := consumer.Publish(types.BrokerMessage{
		Subject: msg.Subject,
		Body:    body,
		ReplyTo: optionalString(msg.ReplyTo),
	})
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// Request publishes body to subject and waits for a reply.
//
// The timeout bounds the wait for the reply, measured from the point the
// request body has been fully sent — the caller controls how fast its own
// body is produced, so that time is not charged against the broker. A zero
// timeout uses the backend's own default. No reply in time returns
// [ErrTimeout].
//
// The reply's [BrokerMessage.Body] is a live stream: read it, then close it.
func Request(subject string, body io.Reader, timeout time.Duration) (BrokerMessage, error) {
	lowered, done := lowerBody(body)
	defer done()
	timeoutMs := witTypes.None[uint32]()
	if timeout > 0 {
		timeoutMs = witTypes.Some(uint32(timeout.Milliseconds()))
	}
	res := consumer.Request(subject, lowered, timeoutMs)
	if res.IsErr() {
		return BrokerMessage{}, convertError(res.Err())
	}
	return FromWit(res.Ok()), nil
}

// lowerBody turns an io.Reader into the stream<u8> the host reads the body
// from. The bytes are pumped by a goroutine running concurrently with the
// host call, because the call does not return until the host has read the
// stream to its end — writing the body first would deadlock. The returned
// function waits for that goroutine, so the caller must defer it.
func lowerBody(body io.Reader) (*witTypes.StreamReader[uint8], func()) {
	tx, rx := consumer.MakeStreamU8()
	if body == nil {
		tx.Drop()
		return rx, func() {}
	}
	pumped := make(chan struct{})
	go func() {
		defer close(pumped)
		defer tx.Drop()
		buf := make([]uint8, 16*1024)
		for !tx.ReaderDropped() {
			n, err := body.Read(buf)
			if n > 0 {
				tx.WriteAll(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()
	return rx, func() { <-pumped }
}

func optionalString(s string) witTypes.Option[string] {
	if s == "" {
		return witTypes.None[string]()
	}
	return witTypes.Some(s)
}

// FromWit converts a generated binding message into a [BrokerMessage]. It is
// exported for use by the messaging/handler subpackage; applications should
// not need it.
func FromWit(msg types.BrokerMessage) BrokerMessage {
	return BrokerMessage{
		Subject: msg.Subject,
		Body:    &streamReadCloser{stream: msg.Body},
		ReplyTo: msg.ReplyTo.SomeOr(""),
	}
}

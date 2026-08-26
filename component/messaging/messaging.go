package messaging

import (
	"errors"
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	consumer "go.wasmcloud.dev/component/imports/wasmcloud_messaging_0_2_0_consumer"
	types "go.wasmcloud.dev/component/imports/wasmcloud_messaging_0_2_0_types"
)

// BrokerMessage is a message sent to or received from the broker.
type BrokerMessage struct {
	Subject string
	Body    []byte
	// ReplyTo is the subject a reply should be published to, or "" if the
	// message does not expect a reply.
	ReplyTo string
}

// Publish sends msg to its subject without awaiting a response.
func Publish(msg BrokerMessage) error {
	res := consumer.Publish(toWit(msg))
	if res.IsErr() {
		return errors.New(res.Err())
	}
	return nil
}

// Request publishes body to subject and waits up to timeout for a reply.
func Request(subject string, body []byte, timeout time.Duration) (BrokerMessage, error) {
	res := consumer.Request(subject, body, uint32(timeout.Milliseconds()))
	if res.IsErr() {
		return BrokerMessage{}, errors.New(res.Err())
	}
	return FromWit(res.Ok()), nil
}

func toWit(msg BrokerMessage) types.BrokerMessage {
	replyTo := witTypes.None[string]()
	if msg.ReplyTo != "" {
		replyTo = witTypes.Some(msg.ReplyTo)
	}
	return types.BrokerMessage{
		Subject: msg.Subject,
		Body:    msg.Body,
		ReplyTo: replyTo,
	}
}

// FromWit converts a generated binding message into a [BrokerMessage]. It is
// exported for use by the messaging/handler subpackage; applications should
// not need it.
func FromWit(msg types.BrokerMessage) BrokerMessage {
	return BrokerMessage{
		Subject: msg.Subject,
		Body:    msg.Body,
		ReplyTo: msg.ReplyTo.SomeOr(""),
	}
}

// A NATS request-reply service, in Go.
//
// The host delivers every message on the workload's core subscriptions to
// the exported handler; the component answers by publishing to the subject
// the requester named in reply-to. It holds no subscription and no
// connection of its own, so it is per-request and scales down to nothing
// between calls.
//
// Core NATS has no acknowledgement and no redelivery. Returning an error
// from the handler is logged by the host and nothing else happens — the
// requester is left waiting out its timeout. A failure the requester should
// know about has to travel back in the reply, which is what the
// Nats-Service-Error headers below are for.
package main

import (
	"errors"
	"fmt"
	"strings"

	"go.bytecodealliance.org/pkg/wasilog"
	"go.wasmcloud.dev/component/nats"
	"go.wasmcloud.dev/component/nats/corehandler"
)

// The error headers NATS micro services use. A requester reads these before
// treating the body as a result: an error reply arrives the same way a
// successful one does, so the headers are the only thing distinguishing it.
const (
	errorHeader     = "Nats-Service-Error"
	errorCodeHeader = "Nats-Service-Error-Code"
)

// Longest reply this service will send.
const maxReplyBytes = 8 << 10

func init() {
	corehandler.HandleFunc(handle)
}

// serviceError is a failure meant for the requester rather than the log.
type serviceError struct {
	code    int
	message string
}

func (e *serviceError) Error() string { return fmt.Sprintf("%d %s", e.code, e.message) }

func handle(msg nats.Message) error {
	log := wasilog.ContextLogger("handle")

	// A publish with no reply-to is not a request. Answering anyway would
	// mean publishing to "", which the host rejects.
	if msg.ReplyTo == "" {
		log.Warn("ignoring message with no reply subject", "subject", msg.Subject)
		return nil
	}

	result, err := dispatch(msg)
	if err != nil {
		svcErr := &serviceError{code: 500, message: err.Error()}
		errors.As(err, &svcErr)

		log.Warn("request failed", "subject", msg.Subject, "code", svcErr.code, "error", svcErr.message)
		return reply(msg.ReplyTo, nil, []nats.Header{
			{Name: errorHeader, Value: svcErr.message},
			{Name: errorCodeHeader, Value: fmt.Sprint(svcErr.code)},
		})
	}

	log.Info("served request", "subject", msg.Subject, "bytes", len(result))
	return reply(msg.ReplyTo, result, nil)
}

// dispatch routes on the last token of the subject, so one `service.*`
// subscription serves every endpoint.
func dispatch(msg nats.Message) ([]byte, error) {
	endpoint := msg.Subject
	if i := strings.LastIndex(endpoint, "."); i >= 0 {
		endpoint = endpoint[i+1:]
	}

	switch endpoint {
	case "greet":
		name := strings.TrimSpace(string(msg.Body))
		if name == "" {
			name = "world"
		}
		return fmt.Appendf(nil, "hello, %s", name), nil

	case "upper":
		if len(msg.Body) == 0 {
			return nil, &serviceError{code: 400, message: "empty body"}
		}
		return []byte(strings.ToUpper(string(msg.Body))), nil

	case "echo":
		// Echo the request headers back as well, so a caller can see what
		// survived the round trip.
		var b strings.Builder
		for _, h := range msg.Headers {
			fmt.Fprintf(&b, "%s: %s\n", h.Name, h.Value)
		}
		b.Write(msg.Body)
		return []byte(b.String()), nil

	default:
		return nil, &serviceError{code: 404, message: "no endpoint named " + endpoint}
	}
}

// reply publishes body to the requester's reply subject.
//
// Every publish is checked against the workload's grant, and the reply
// subject is an inbox the requester generated rather than anything this
// service knows in advance — so `subject-allow` has to cover the inbox
// prefix (`_INBOX.>`) or every answer comes back as a
// [nats.SubjectDeniedError] that the requester only ever sees as a timeout.
func reply(subject string, body []byte, headers []nats.Header) error {
	if len(body) > maxReplyBytes {
		body = body[:maxReplyBytes]
	}

	err := nats.Publish(nats.Message{Subject: subject, Body: body, Headers: headers})

	// The connected server's maximum payload is smaller than this service
	// assumed. The error carries the real limit, so trim to it and try once
	// more rather than leaving the requester with nothing.
	var tooBig *nats.MaxPayloadExceededError
	if errors.As(err, &tooBig) && uint64(len(body)) > tooBig.Limit {
		return nats.Publish(nats.Message{
			Subject: subject,
			Body:    body[:tooBig.Limit],
			Headers: headers,
		})
	}
	return err
}

// main is required for a Go component but never runs: the host invokes the
// exported handler instead.
func main() {}

package nats

import (
	"errors"
	"fmt"

	types "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_types"
)

// Sentinel errors for the payload-free cases of wasmcloud:nats/types.nats-error.
// Test for them with errors.Is.
var (
	// ErrNoResponders means nothing is subscribed to the subject. It is
	// distinct from a timeout: retrying immediately fails the same way
	// until a responder appears, so back off or fail fast rather than spin.
	ErrNoResponders = errors.New("nats: no responders on subject")
	// ErrKeyNotFound is returned by KV reads for an absent key.
	ErrKeyNotFound = errors.New("nats: key not found")
	// ErrNoMessages means a pull-consumer fetch returned empty within its
	// timeout. It is an ordinary idle result, not a failure.
	ErrNoMessages = errors.New("nats: no messages")
	// ErrDisconnected means the connection is down, as opposed to a
	// transport error on a live connection.
	ErrDisconnected = errors.New("nats: disconnected")
)

// SubjectDeniedError reports a subject outside the grant the workload was
// bound with. The check happens host-side, so the message never reaches the
// server. Widen `subject-allow` on the interface binding to permit it.
type SubjectDeniedError struct{ Subject string }

func (e *SubjectDeniedError) Error() string {
	return fmt.Sprintf("nats: subject %q is outside this workload's grant", e.Subject)
}

// MaxPayloadExceededError reports a payload larger than the connected
// server's advertised maximum, carrying that limit in bytes.
type MaxPayloadExceededError struct{ Limit uint64 }

func (e *MaxPayloadExceededError) Error() string {
	return fmt.Sprintf("nats: payload exceeds the server maximum of %d bytes", e.Limit)
}

// RevisionMismatchError reports a failed compare-and-swap on a KV key. It
// carries the revision the key actually holds, so a retry can reapply
// without re-reading.
type RevisionMismatchError struct{ Current uint64 }

func (e *RevisionMismatchError) Error() string {
	return fmt.Sprintf("nats: revision mismatch, current revision is %d", e.Current)
}

// NotFoundError reports a missing stream, consumer, or bucket. The string
// names the resource.
type NotFoundError struct{ Resource string }

func (e *NotFoundError) Error() string { return "nats: not found: " + e.Resource }

// UnsupportedByServerError reports an operation the connected NATS server is
// too old to perform, carrying the minimum version required.
type UnsupportedByServerError struct{ Detail string }

func (e *UnsupportedByServerError) Error() string { return "nats: unsupported by server: " + e.Detail }

// convertError maps a generated nats-error onto a Go error.
func convertError(e types.NatsError) error {
	switch e.Tag() {
	case types.NatsErrorConnection:
		return fmt.Errorf("nats: connection: %s", e.Connection())
	case types.NatsErrorTimeout:
		return fmt.Errorf("nats: timeout: %s", e.Timeout())
	case types.NatsErrorNoResponders:
		return ErrNoResponders
	case types.NatsErrorSubjectDenied:
		return &SubjectDeniedError{Subject: e.SubjectDenied()}
	case types.NatsErrorMaxPayloadExceeded:
		return &MaxPayloadExceededError{Limit: e.MaxPayloadExceeded()}
	case types.NatsErrorJetstream:
		return fmt.Errorf("nats: jetstream: %s", e.Jetstream())
	case types.NatsErrorKeyNotFound:
		return ErrKeyNotFound
	case types.NatsErrorRevisionMismatch:
		return &RevisionMismatchError{Current: e.RevisionMismatch()}
	case types.NatsErrorNoMessages:
		return ErrNoMessages
	case types.NatsErrorNotFound:
		return &NotFoundError{Resource: e.NotFound()}
	case types.NatsErrorUnsupportedByServer:
		return &UnsupportedByServerError{Detail: e.UnsupportedByServer()}
	case types.NatsErrorDisconnected:
		return ErrDisconnected
	case types.NatsErrorUnexpected:
		return fmt.Errorf("nats: %s", e.Unexpected())
	default:
		// The variant is non-exhaustive; future hosts may add cases.
		return fmt.Errorf("nats: unknown error (tag %d)", e.Tag())
	}
}

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
	// ErrKeyNotFound is returned by KV reads for an absent, deleted, or
	// purged key.
	ErrKeyNotFound = errors.New("nats: key not found")
	// ErrNoMessages means a pull-consumer fetch returned empty within its
	// timeout. It is an ordinary idle result, not a failure.
	ErrNoMessages = errors.New("nats: no messages")
	// ErrDisconnected means the connection is down, as opposed to a
	// transport error on a live connection.
	ErrDisconnected = errors.New("nats: disconnected")
)

// DenialReason is why the host refused a name.
type DenialReason uint8

const (
	// DenialReserved means the name is in a space the host keeps for
	// itself — the JetStream API, the KV/object-store key spaces, $SYS, or
	// the host's own lattice. No grant widens it.
	DenialReserved DenialReason = iota
	// DenialNotGranted means no grant on this binding covers the name.
	DenialNotGranted
	// DenialWildcardNotAllowed means a publish or request subject
	// contained * or >, which would let it satisfy a narrower grant.
	DenialWildcardNotAllowed
)

func (r DenialReason) String() string {
	switch r {
	case DenialReserved:
		return "reserved"
	case DenialNotGranted:
		return "not granted"
	case DenialWildcardNotAllowed:
		return "wildcard not allowed"
	default:
		return "unknown"
	}
}

// DeniedTarget is what kind of name was refused, and so which grant to
// widen: `subject-allow`, `stream-allow`, or `bucket-allow`.
type DeniedTarget uint8

const (
	DeniedSubject DeniedTarget = iota
	DeniedStream
	DeniedBucket
)

func (t DeniedTarget) String() string {
	switch t {
	case DeniedSubject:
		return "subject"
	case DeniedStream:
		return "stream"
	case DeniedBucket:
		return "bucket"
	default:
		return "unknown"
	}
}

// DeniedError reports a name outside the grant the workload was bound with,
// or in a space the host reserves. The check happens host-side, so nothing
// reaches the server.
//
// Which grant to widen follows from Target: `subject-allow` for a subject or
// subscription, `stream-allow` for a stream, `bucket-allow` for a bucket. A
// Reason of [DenialReserved] cannot be widened by any grant.
type DeniedError struct {
	Reason DenialReason
	Target DeniedTarget
	Name   string
}

func (e *DeniedError) Error() string {
	return fmt.Sprintf("nats: %s %q denied: %s", e.Target, e.Name, e.Reason)
}

// InvalidHeaderError reports a header name or value that cannot go on the
// NATS wire: names must be printable ASCII without ':', and values may not
// contain CR or LF.
type InvalidHeaderError struct{ Detail string }

func (e *InvalidHeaderError) Error() string { return "nats: invalid header: " + e.Detail }

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

// LimitExceededError means the server refused a pull request outright because
// it asks for more than the consumer was provisioned to allow — a batch over
// MaxRequestBatch, a byte bound over MaxRequestMaxBytes, or too many waiting
// pulls.
//
// Nothing was delivered, and retrying unchanged fails the same way. It is
// distinct from [ErrNoMessages], which means the request ran against an idle
// consumer. [PullConsumer.Info] reports the limits to size the next request
// against.
type LimitExceededError struct{ Detail string }

func (e *LimitExceededError) Error() string { return "nats: limit exceeded: " + e.Detail }

// convertError maps a generated nats-error onto a Go error.
func convertError(e types.NatsError) error {
	switch e.Tag() {
	case types.NatsErrorConnection:
		return fmt.Errorf("nats: connection: %s", e.Connection())
	case types.NatsErrorTimeout:
		return fmt.Errorf("nats: timeout: %s", e.Timeout())
	case types.NatsErrorNoResponders:
		return ErrNoResponders
	case types.NatsErrorDenied:
		d := e.Denied()
		return &DeniedError{
			Reason: DenialReason(d.Reason),
			Target: DeniedTarget(d.Target),
			Name:   d.Name,
		}
	case types.NatsErrorMaxPayloadExceeded:
		return &MaxPayloadExceededError{Limit: e.MaxPayloadExceeded()}
	case types.NatsErrorInvalidHeader:
		return &InvalidHeaderError{Detail: e.InvalidHeader()}
	case types.NatsErrorJetstream:
		return fmt.Errorf("nats: jetstream: %s", e.Jetstream())
	case types.NatsErrorKeyNotFound:
		return ErrKeyNotFound
	case types.NatsErrorRevisionMismatch:
		return &RevisionMismatchError{Current: e.RevisionMismatch()}
	case types.NatsErrorNoMessages:
		return ErrNoMessages
	case types.NatsErrorLimitExceeded:
		return &LimitExceededError{Detail: e.LimitExceeded()}
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

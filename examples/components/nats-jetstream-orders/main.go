// Order processing over JetStream, in Go.
//
// The host pushes each delivery into the registered handler; the component
// accumulates a per-order total in JetStream KV and publishes a
// processed-order notification. It holds no consumer and no stream, so it is
// per-request and scales down to nothing between bursts.
package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.wasmcloud.dev/component/nats"
	"go.wasmcloud.dev/component/nats/jetstreamhandler"
)

const (
	totalsBucket     = "order-totals"
	processedSubject = "orders.processed"
	// Attempts to reapply a total before giving up. A conflict means a
	// concurrent handler wrote first, which is expected under load.
	maxCASAttempts = 5
)

func init() {
	jetstreamhandler.HandleFunc(handleOrder)
}

// total is a running sum plus the stream sequence that last contributed to
// it. Storing the sequence is what makes the update idempotent: delivery is
// at-least-once, so a bare `total += amount` double-counts a redelivery.
type total struct {
	running      uint64
	lastSequence uint64
}

func parseTotal(raw []byte) (total, bool) {
	running, seq, found := strings.Cut(strings.TrimSpace(string(raw)), "@")
	if !found {
		return total{}, false
	}
	r, err := strconv.ParseUint(running, 10, 64)
	if err != nil {
		return total{}, false
	}
	s, err := strconv.ParseUint(seq, 10, 64)
	if err != nil {
		return total{}, false
	}
	return total{running: r, lastSequence: s}, true
}

func (t total) encode() []byte {
	return fmt.Appendf(nil, "%d@%d", t.running, t.lastSequence)
}

// parseOrder reads `order-id:amount` from a message body.
func parseOrder(body []byte) (string, uint64, bool) {
	id, amount, found := strings.Cut(strings.TrimSpace(string(body)), ":")
	if !found || id == "" {
		return "", 0, false
	}
	n, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		return "", 0, false
	}
	return id, n, true
}

func handleOrder(h *nats.MessageHandle) error {
	msg := h.Message()
	sequence := h.Sequence()

	orderID, amount, ok := parseOrder(msg.Body)
	if !ok {
		// A malformed body never parses, so returning an error would nak it
		// and retry forever. Under the binding's `ack-mode: auto` the host
		// owns the acknowledgement and Term is not ours to call, so report
		// success and let the message be dropped.
		fmt.Printf("dropping malformed order at sequence %d\n", sequence)
		return nil
	}

	if delivery := h.DeliveryCount(); delivery > 1 {
		fmt.Printf("order %s redelivered (attempt %d)\n", orderID, delivery)
	}

	bucket, err := nats.OpenBucket(totalsBucket)
	if err != nil {
		return fmt.Errorf("open %s: %w", totalsBucket, err)
	}
	defer bucket.Close()

	running, err := accumulate(bucket, orderID, amount, sequence)
	if err != nil {
		return fmt.Errorf("accumulate %s: %w", orderID, err)
	}

	// Nats-Msg-Id deduplicates within the stream's duplicate window, so a
	// redelivered order does not publish a second notification.
	ack, err := nats.JetStreamPublish(nats.Message{
		Subject: processedSubject,
		Body:    fmt.Appendf(nil, "%s:%d", orderID, running),
		Headers: []nats.Header{{
			Name:  nats.MsgIDHeader,
			Value: fmt.Sprintf("processed-%s-%d", orderID, sequence),
		}},
	})
	if err != nil {
		return fmt.Errorf("publish notification: %w", err)
	}

	fmt.Printf("order %s +%d -> %d (stream seq %d, duplicate: %t)\n",
		orderID, amount, running, ack.Sequence, ack.Duplicate)

	// Returning nil acks under `ack-mode: auto`.
	return nil
}

// accumulate adds amount to the order's total exactly once per sequence.
func accumulate(bucket *nats.Bucket, key string, amount, sequence uint64) (uint64, error) {
	for attempt := 0; attempt < maxCASAttempts; attempt++ {
		entry, found, err := bucket.Get(key)
		if err != nil {
			return 0, err
		}

		var current total
		if found {
			if parsed, ok := parseTotal(entry.Value); ok {
				current = parsed
			}
			// This sequence is already counted, so this is a redelivery.
			// Report the total unchanged rather than adding again.
			if sequence <= current.lastSequence {
				return current.running, nil
			}
		}

		next := total{running: current.running + amount, lastSequence: sequence}

		if found {
			_, err = bucket.Update(key, next.encode(), entry.Revision)
		} else {
			_, err = bucket.Create(key, next.encode())
		}
		if err == nil {
			return next.running, nil
		}

		// Someone wrote between the read and the write. Re-read and reapply
		// rather than clobbering their value.
		var mismatch *nats.RevisionMismatchError
		if !errors.As(err, &mismatch) {
			return 0, err
		}
	}
	return 0, fmt.Errorf("gave up after %d contended attempts", maxCASAttempts)
}

// main is required for a Go component but never runs: the host invokes the
// exported handler instead.
func main() {}

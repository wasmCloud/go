// A JetStream KV watcher that maintains a derived view, in Go.
//
// The bucket holds one key per feature flag (`flag.<name>` = "on" | "off").
// The host pushes every change on that prefix into the exported handler,
// which recomputes the set of enabled flags into a single `active` key and
// announces the change over core NATS.
//
// Two things make a watcher different from a queue consumer:
//
//   - There is no acknowledgement and no replay. Returning an error is
//     logged by the host and the event is gone. So the handler never applies
//     a delta — it rebuilds the derived value from the bucket, which is
//     correct whether it ran once, twice, or missed the event before it.
//   - Writing to the bucket you are watching feeds your own writes back to
//     you. Nothing here is clever enough to break that loop after the fact,
//     so the derived key is deliberately kept outside the watched prefix.
package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.bytecodealliance.org/pkg/wasilog"
	"go.wasmcloud.dev/component/nats"
	"go.wasmcloud.dev/component/nats/kvhandler"
)

const (
	flagsBucket = "feature-flags"
	// Watched keys. The host is configured to deliver only this prefix.
	flagPrefix = "flag."
	// The derived key, deliberately outside flagPrefix so writing it does
	// not re-trigger this handler.
	activeKey = "active"
	// Where change announcements go.
	changeSubject = "flags.changed"
	// Attempts to write the derived value before giving up. A conflict
	// means another event was processed concurrently, which is expected.
	maxCASAttempts = 5
)

func init() {
	kvhandler.HandleFunc(handleChange)
}

func handleChange(bucket string, entry nats.Entry) error {
	log := wasilog.ContextLogger("handleChange")

	// Belt and braces: the host's `kv-watches` filter should never deliver
	// the derived key, but a widened filter must not turn into a write loop.
	if !strings.HasPrefix(entry.Key, flagPrefix) {
		log.Debug("ignoring key outside the watched prefix", "key", entry.Key)
		return nil
	}

	log.Info("flag changed",
		"bucket", bucket,
		"key", entry.Key,
		"operation", entry.Operation.String(),
		"revision", entry.Revision,
	)

	b, err := nats.OpenBucket(flagsBucket)
	if err != nil {
		return fmt.Errorf("open %s: %w", flagsBucket, err)
	}
	defer b.Close()

	// The event says something changed; the bucket says what is true now.
	// Reading current state rather than applying entry.Value is what makes
	// a missed or duplicated event harmless.
	active, err := rebuildActive(b)
	if err != nil {
		return fmt.Errorf("rebuild active set: %w", err)
	}

	if err := storeActive(b, active); err != nil {
		return fmt.Errorf("store active set: %w", err)
	}

	// Core NATS is fire-and-forget, so this returns once the announcement
	// is written to the connection — not once anyone has read it. A missed
	// announcement is survivable here: `active` is the durable answer, and
	// this is only a nudge to go read it.
	if err := nats.Publish(nats.Message{
		Subject: changeSubject,
		Body:    fmt.Appendf(nil, "%s %s", entry.Key, entry.Operation),
		Headers: []nats.Header{{Name: "Kv-Revision", Value: fmt.Sprint(entry.Revision)}},
	}); err != nil {
		// A denied subject is a deployment problem, not a transient one:
		// say so plainly rather than burying it in a generic message.
		var denied *nats.SubjectDeniedError
		if errors.As(err, &denied) {
			return fmt.Errorf("announce: %w (add it to subject-allow)", err)
		}
		return fmt.Errorf("announce: %w", err)
	}

	log.Info("active flags updated", "count", len(active), "flags", strings.Join(active, ","))
	return nil
}

// rebuildActive lists the flag keys and returns the enabled ones, sorted so
// the stored value is stable and two concurrent rebuilds of the same state
// produce the same bytes.
func rebuildActive(b *nats.Bucket) ([]string, error) {
	keys, err := b.Keys()
	if err != nil {
		return nil, err
	}

	var active []string
	for _, key := range keys {
		if !strings.HasPrefix(key, flagPrefix) {
			continue
		}
		entry, found, err := b.Get(key)
		if err != nil {
			return nil, fmt.Errorf("get %s: %w", key, err)
		}
		// Keys() can name a key that a concurrent delete has already
		// removed, so an absent key here is ordinary, not an error.
		if !found || entry.Operation != nats.OperationPut {
			continue
		}
		if enabled(entry.Value) {
			active = append(active, strings.TrimPrefix(key, flagPrefix))
		}
	}
	sort.Strings(active)
	return active, nil
}

func enabled(value []byte) bool {
	switch strings.ToLower(strings.TrimSpace(string(value))) {
	case "on", "true", "1", "enabled":
		return true
	default:
		return false
	}
}

// storeActive writes the derived value with compare-and-swap, so a rebuild
// that raced with a newer one loses instead of overwriting it.
func storeActive(b *nats.Bucket, active []string) error {
	value := []byte(strings.Join(active, ","))

	for attempt := range maxCASAttempts {
		entry, found, err := b.Get(activeKey)
		if err != nil {
			return err
		}
		// Nothing to do, and skipping the write also stops an unchanged
		// rebuild from burning a revision on every unrelated flag event.
		if found && string(entry.Value) == string(value) {
			return nil
		}

		if found {
			_, err = b.Update(activeKey, value, entry.Revision)
		} else {
			_, err = b.Create(activeKey, value)
		}
		if err == nil {
			return nil
		}

		// Someone wrote between the read and the write. Their rebuild saw
		// state at least as new as ours, so re-read and reapply.
		var mismatch *nats.RevisionMismatchError
		if !errors.As(err, &mismatch) {
			return err
		}
		wasilog.ContextLogger("storeActive").Debug("lost a compare-and-swap, retrying",
			"attempt", attempt+1, "currentRevision", mismatch.Current)
	}
	return fmt.Errorf("gave up after %d contended attempts", maxCASAttempts)
}

// main is required for a Go component but never runs: the host invokes the
// exported handler instead.
func main() {}

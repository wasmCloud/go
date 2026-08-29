package nats

import (
	"time"

	kv "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_kv"
)

// Operation is what produced a KV entry.
type Operation uint8

const (
	OperationPut Operation = iota
	OperationDelete
	OperationPurge
)

func (o Operation) String() string {
	switch o {
	case OperationPut:
		return "put"
	case OperationDelete:
		return "delete"
	case OperationPurge:
		return "purge"
	default:
		return "unknown"
	}
}

// Entry is a single KV entry as seen by get, history, and watch.
type Entry struct {
	Key   string
	Value []byte
	// Revision is the entry's version, and the value to pass to
	// [Bucket.Update] for a compare-and-swap.
	Revision  uint64
	CreatedAt time.Time
	Operation Operation
}

// BucketStatus is a snapshot of a bucket's state, read fresh from the server.
type BucketStatus struct {
	Bucket string
	// Values is the stored message count — live values plus retained
	// history, matching what `nats kv status` reports.
	Values  uint64
	History uint8
	TTL     time.Duration
	Bytes   uint64
}

// Bucket is an open JetStream KV bucket.
type Bucket struct{ inner *kv.Bucket }

// OpenBucket opens an existing bucket.
//
// It never creates one: bucket lifecycle is deliberately outside this
// interface, so provision it out-of-band. A missing bucket returns a
// [NotFoundError]; one outside the workload's `bucket-allow` grant returns a
// [DeniedError].
func OpenBucket(bucket string) (*Bucket, error) {
	res := kv.Open(bucket)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	return &Bucket{inner: res.Ok()}, nil
}

// Close releases the bucket handle. The bucket itself is untouched.
func (b *Bucket) Close() { b.inner.Drop() }

// Get returns the latest entry for key. An absent, deleted, or purged key
// returns [ErrKeyNotFound]; test for it with errors.Is.
func (b *Bucket) Get(key string) (Entry, error) {
	res := b.inner.Get(key)
	if res.IsErr() {
		return Entry{}, convertError(res.Err())
	}
	return fromWitEntry(res.Ok()), nil
}

// Put writes value unconditionally, last-write-wins, and returns the new
// revision. A concurrent writer's value is overwritten — use [Bucket.Update]
// to make the write conditional instead.
func (b *Bucket) Put(key string, value []byte) (uint64, error) {
	res := b.inner.Put(key, value)
	if res.IsErr() {
		return 0, convertError(res.Err())
	}
	return res.Ok(), nil
}

// Create writes value only if key is absent, and returns the new revision.
func (b *Bucket) Create(key string, value []byte) (uint64, error) {
	res := b.inner.Create(key, value)
	if res.IsErr() {
		return 0, convertError(res.Err())
	}
	return res.Ok(), nil
}

// Update writes value only if key still holds expectedRevision, and returns
// the new revision. On conflict it returns a [RevisionMismatchError] carrying
// the current revision, so a retry can reapply without re-reading:
//
//	for {
//	    entry, err := b.Get(key)
//	    // ...
//	    _, err = b.Update(key, next, entry.Revision)
//	    var mismatch *nats.RevisionMismatchError
//	    if errors.As(err, &mismatch) {
//	        continue // someone else wrote; re-read and reapply
//	    }
//	    return err
//	}
func (b *Bucket) Update(key string, value []byte, expectedRevision uint64) (uint64, error) {
	res := b.inner.Update(key, value, expectedRevision)
	if res.IsErr() {
		return 0, convertError(res.Err())
	}
	return res.Ok(), nil
}

// Delete writes a tombstone for key, preserving its history.
func (b *Bucket) Delete(key string) error {
	if res := b.inner.Delete(key); res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// Purge removes key and its history.
func (b *Bucket) Purge(key string) error {
	if res := b.inner.Purge(key); res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// KeyPage is a listing of a bucket's keys, and whether it is the whole of it.
type KeyPage struct {
	Keys []string
	// Truncated reports that more keys match the filter than this page
	// carries. The listing is capped host-side, so a large bucket cannot be
	// enumerated in full through one call — treat a truncated page as a
	// signal to narrow the filter, not as a complete set.
	Truncated bool
}

// Keys lists the bucket's keys matching filter, which is a NATS subject
// pattern over the key space — pass ">" for all of them.
//
// The listing is capped host-side, and it is key cardinality rather than
// value size that the cap bounds: check [KeyPage.Truncated] before treating a
// page as the whole bucket, and pass a narrower filter to walk one that holds
// more than a page.
func (b *Bucket) Keys(filter string) (KeyPage, error) {
	res := b.inner.Keys(filter)
	if res.IsErr() {
		return KeyPage{}, convertError(res.Err())
	}
	page := res.Ok()
	return KeyPage{Keys: page.Keys, Truncated: page.Truncated}, nil
}

// History returns every retained revision of key, oldest first, including
// delete and purge tombstones. How much is retained is bucket configuration.
// A key with no history at all returns [ErrKeyNotFound].
//
// The call is bounded: the host probes for the key before opening the history
// consumer, and the drain runs under the binding's `request-timeout-ms` (ten
// seconds if it names none), returning a [TimeoutError] rather than blocking.
// It cannot pin the instance's admission permit indefinitely, which an earlier
// revision could.
func (b *Bucket) History(key string) ([]Entry, error) {
	res := b.inner.History(key)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	raw := res.Ok()
	out := make([]Entry, 0, len(raw))
	for _, e := range raw {
		out = append(out, fromWitEntry(e))
	}
	return out, nil
}

// Status returns a snapshot of the bucket's state.
func (b *Bucket) Status() (BucketStatus, error) {
	res := b.inner.Status()
	if res.IsErr() {
		return BucketStatus{}, convertError(res.Err())
	}
	s := res.Ok()
	return BucketStatus{
		Bucket:  s.Bucket,
		Values:  s.Values,
		History: s.History,
		TTL:     time.Duration(s.TtlSeconds) * time.Second,
		Bytes:   s.Bytes,
	}, nil
}

// FromWitEntry converts a generated KV entry into an [Entry]. It is exported
// for the nats/kvhandler subpackage; applications should not need it.
func FromWitEntry(e kv.Entry) Entry { return fromWitEntry(e) }

func fromWitEntry(e kv.Entry) Entry {
	return Entry{
		Key:       e.Key,
		Value:     e.Value,
		Revision:  e.Revision,
		CreatedAt: time.Unix(0, int64(e.CreatedAtUnixNanos)).UTC(),
		Operation: Operation(e.Operation),
	}
}

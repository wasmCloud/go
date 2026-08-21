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

// BucketStatus is a snapshot of a bucket's state.
type BucketStatus struct {
	Bucket  string
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
// [SubjectDeniedError].
func OpenBucket(bucket string) (*Bucket, error) {
	res := kv.Open(bucket)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	return &Bucket{inner: res.Ok()}, nil
}

// Close releases the bucket handle. The bucket itself is untouched.
func (b *Bucket) Close() { b.inner.Drop() }

// Get returns the latest entry for key. A missing key returns ok=false with
// a nil error, so an absent key is not an error condition.
func (b *Bucket) Get(key string) (entry Entry, ok bool, err error) {
	res := b.inner.Get(key)
	if res.IsErr() {
		return Entry{}, false, convertError(res.Err())
	}
	opt := res.Ok()
	if opt.IsNone() {
		return Entry{}, false, nil
	}
	return fromWitEntry(opt.Some()), true, nil
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
//	    entry, ok, err := b.Get(key)
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

// Keys lists every key in the bucket.
func (b *Bucket) Keys() ([]string, error) {
	res := b.inner.Keys()
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	return res.Ok(), nil
}

// History returns every retained revision of key, oldest first, including
// delete and purge tombstones. How much is retained is bucket configuration.
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

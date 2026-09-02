package keyvalue

import (
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	atomics "go.wasmcloud.dev/component/imports/wasmcloud_keyvalue_0_2_0_atomics"
	batch "go.wasmcloud.dev/component/imports/wasmcloud_keyvalue_0_2_0_batch"
	cas "go.wasmcloud.dev/component/imports/wasmcloud_keyvalue_0_2_0_cas"
	store "go.wasmcloud.dev/component/imports/wasmcloud_keyvalue_0_2_0_store"
	types "go.wasmcloud.dev/component/imports/wasmcloud_keyvalue_0_2_0_types"
)

// Bucket is a collection of key-value pairs provided by the host.
type Bucket struct {
	inner *types.Bucket
}

// Store is the one label-routed call of `wasmcloud:keyvalue`: `store.open`.
// Everything else — atomics, cas, batch, and the bucket's own methods —
// operates on a [types.Bucket] that already carries the backend it was opened
// through, so only this needs to know which binding it is talking to.
//
// The default implementation is the SDK's committed bindings, which name the
// plain (unlabeled) instance. A component that imports `store` under an
// `(implements ..)` label satisfies this interface with its own generated
// bindings for that label; see the package doc.
type Store interface {
	Open(identifier string) witTypes.Result[*types.Bucket, types.Error]
}

// plainStore is [Store] over the committed bindings for the plain, unlabeled
// instance — the only instance name a committed binding can carry, because
// //go:wasmimport takes it as a compile-time literal.
type plainStore struct{}

func (plainStore) Open(identifier string) witTypes.Result[*types.Bucket, types.Error] {
	return store.Open(identifier)
}

// Open returns the bucket with the given host-provided identifier, from the
// plain (unlabeled) instance an unnamed hostInterfaces entry routes to.
func Open(identifier string) (*Bucket, error) {
	return OpenFrom(plainStore{}, identifier)
}

// OpenFrom returns the bucket with the given host-provided identifier, opened
// through s rather than the plain instance. The returned [Bucket] is an
// ordinary bucket: every method on it routes to the backend it was opened
// through.
func OpenFrom(s Store, identifier string) (*Bucket, error) {
	res := s.Open(identifier)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	return &Bucket{inner: res.Ok()}, nil
}

// Drop releases the host-side bucket handle. The handle is also released by
// the garbage collector if the Bucket becomes unreachable.
func (b *Bucket) Drop() {
	b.inner.Drop()
}

// Get returns the value stored at key. ok is false (with a nil error) when
// the key does not exist.
func (b *Bucket) Get(key string) (value []byte, ok bool, err error) {
	res := b.inner.Get(key)
	if res.IsErr() {
		return nil, false, convertError(res.Err())
	}
	opt := res.Ok()
	if opt.IsNone() {
		return nil, false, nil
	}
	return opt.Some(), true, nil
}

// SetOptions modify the behavior of [Bucket.Set].
type SetOptions struct {
	// TTL, when nonzero, expires the entry that long after the write.
	// Sub-millisecond durations are rounded down; hosts whose backend cannot
	// honor expiry raise an error rather than silently ignoring it.
	TTL time.Duration
	// IfNotExists makes the write succeed only if the key does not already
	// exist, failing with [ErrPreconditionFailed] otherwise.
	IfNotExists bool
}

// Set stores value at key, overwriting any existing entry. opts may be nil
// for the default overwrite-on-exists behavior with no expiry.
func (b *Bucket) Set(key string, value []byte, opts *SetOptions) error {
	witOpts := witTypes.None[types.SetOptions]()
	if opts != nil {
		ttl := witTypes.None[uint64]()
		if opts.TTL > 0 {
			ttl = witTypes.Some(uint64(opts.TTL.Milliseconds()))
		}
		witOpts = witTypes.Some(types.SetOptions{TtlMs: ttl, IfNotExists: opts.IfNotExists})
	}
	res := b.inner.Set(key, value, witOpts)
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// Delete removes the entry at key. Deleting a missing key is not an error.
func (b *Bucket) Delete(key string) error {
	res := b.inner.Delete(key)
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// Exists reports whether an entry exists at key.
func (b *Bucket) Exists(key string) (bool, error) {
	res := b.inner.Exists(key)
	if res.IsErr() {
		return false, convertError(res.Err())
	}
	return res.Ok(), nil
}

// ListKeys returns a page of keys, optionally restricted to those beginning
// with prefix ("" for all keys). cursor continues a previous listing ("" for
// the first page); the returned next cursor is "" when there are no more
// keys. Keys are returned in no particular order.
func (b *Bucket) ListKeys(prefix, cursor string) (keys []string, next string, err error) {
	p := witTypes.None[string]()
	if prefix != "" {
		p = witTypes.Some(prefix)
	}
	c := witTypes.None[string]()
	if cursor != "" {
		c = witTypes.Some(cursor)
	}
	res := b.inner.ListKeys(p, c)
	if res.IsErr() {
		return nil, "", convertError(res.Err())
	}
	kr := res.Ok()
	return kr.Keys, kr.Cursor.SomeOr(""), nil
}

// Increment atomically adds delta (which may be negative) to the numeric
// value at key, creating the entry with value delta if it does not exist,
// and returns the new value.
//
// Requires the app's world to import wasmcloud:keyvalue/atomics@0.2.0.
func (b *Bucket) Increment(key string, delta int64) (int64, error) {
	res := atomics.Increment(b.inner, key, delta)
	if res.IsErr() {
		return 0, convertError(res.Err())
	}
	return res.Ok(), nil
}

// Entry is a key's value together with its backend-defined version token.
// Only equality of versions is meaningful.
type Entry struct {
	Value   []byte
	Version string
}

// Current reads a key's current value and version for use as the
// precondition of a later [Bucket.Swap]. It returns nil (with a nil error)
// when the key does not exist.
//
// Requires the app's world to import wasmcloud:keyvalue/cas@0.2.0.
func (b *Bucket) Current(key string) (*Entry, error) {
	res := cas.Current(b.inner, key)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	opt := res.Ok()
	if opt.IsNone() {
		return nil, nil
	}
	e := opt.Some()
	return &Entry{Value: e.Value, Version: e.Version}, nil
}

// SwapOptions are the preconditions for a [Bucket.Swap]. At least one must
// be set; a Swap with no preconditions fails with [ErrInvalidArgument] (use
// [Bucket.Set] for an unconditional write).
type SwapOptions struct {
	// RequireVersion, when non-nil, requires the entry's current version to
	// equal it (the ABA-safe check).
	RequireVersion *string
	// RequireValue, when non-nil, requires the entry's current value to
	// equal it.
	RequireValue []byte
}

// SwapResult is the outcome of a [Bucket.Swap].
type SwapResult struct {
	// Swapped is true when the new value was written.
	Swapped bool
	// Stale is the current entry when a precondition failed (nil if the key
	// is absent), letting the caller recompute and retry without a separate
	// read. Only meaningful when Swapped is false.
	Stale *Entry
}

// Swap conditionally and atomically sets key to value if every precondition
// in opts holds. A lost race is reported via SwapResult.Swapped == false,
// not an error.
//
// Requires the app's world to import wasmcloud:keyvalue/cas@0.2.0.
func (b *Bucket) Swap(key string, value []byte, opts SwapOptions) (SwapResult, error) {
	witOpts := cas.CasOptions{
		RequireVersion: witTypes.None[string](),
		RequireValue:   witTypes.None[[]uint8](),
	}
	if opts.RequireVersion != nil {
		witOpts.RequireVersion = witTypes.Some(*opts.RequireVersion)
	}
	if opts.RequireValue != nil {
		witOpts.RequireValue = witTypes.Some[[]uint8](opts.RequireValue)
	}
	res := cas.Swap(b.inner, key, value, witOpts)
	if res.IsErr() {
		return SwapResult{}, convertError(res.Err())
	}
	outcome := res.Ok()
	if outcome.Tag() == cas.CasResultSwapped {
		return SwapResult{Swapped: true}, nil
	}
	var stale *Entry
	if opt := outcome.Stale(); opt.IsSome() {
		e := opt.Some()
		stale = &Entry{Value: e.Value, Version: e.Version}
	}
	return SwapResult{Stale: stale}, nil
}

// GetMany returns the values for the given keys in one round trip. Keys that
// do not exist are absent from the returned map.
//
// Requires the app's world to import wasmcloud:keyvalue/batch@0.2.0.
func (b *Bucket) GetMany(keys []string) (map[string][]byte, error) {
	res := batch.GetMany(b.inner, keys)
	if res.IsErr() {
		return nil, convertError(res.Err())
	}
	out := make(map[string][]byte, len(keys))
	for _, opt := range res.Ok() {
		if opt.IsSome() {
			pair := opt.Some()
			out[pair.F0] = pair.F1
		}
	}
	return out, nil
}

// SetMany stores the given entries in one round trip, overwriting existing
// values. The operation is not atomic: on error, some entries may have been
// written.
//
// Requires the app's world to import wasmcloud:keyvalue/batch@0.2.0.
func (b *Bucket) SetMany(entries map[string][]byte) error {
	kvs := make([]witTypes.Tuple2[string, []uint8], 0, len(entries))
	for k, v := range entries {
		kvs = append(kvs, witTypes.Tuple2[string, []uint8]{F0: k, F1: v})
	}
	res := batch.SetMany(b.inner, kvs)
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

// DeleteMany removes the given keys in one round trip, skipping keys that do
// not exist. The operation is not atomic: on error, some entries may have
// been deleted.
//
// Requires the app's world to import wasmcloud:keyvalue/batch@0.2.0.
func (b *Bucket) DeleteMany(keys []string) error {
	res := batch.DeleteMany(b.inner, keys)
	if res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

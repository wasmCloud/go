package nats

import (
	"fmt"
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	kvadmin "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_4_kv_admin"
)

// StorageType is where a bucket's backing stream keeps its data.
type StorageType uint8

const (
	// StorageFile persists to disk. It is the only storage a binding allows
	// unless the host adds `memory` to `bucket-storage-allow`.
	StorageFile StorageType = iota
	// StorageMemory keeps the bucket in server RAM.
	StorageMemory
)

func (s StorageType) String() string {
	switch s {
	case StorageFile:
		return "file"
	case StorageMemory:
		return "memory"
	default:
		return "unknown"
	}
}

// BucketConfig is a bucket's configuration, as requested from [CreateBucket]
// or as read back by [GetBucketConfig].
//
// Zero values ask for the defaults: History 0 keeps one revision, MaxAge 0
// keeps values forever, MaxBytes 0 takes the host's `bucket-max-bytes`, and
// MaxValueSize 0 is unlimited. Replicas, placement, mirrors and sources are
// not exposed — the host's `bucket-replicas` decides the replica count.
type BucketConfig struct {
	// Bucket is the name: letters, digits, '-' and '_'.
	Bucket string
	// History is how many revisions to keep per key, 1 to 64.
	History uint8
	// MaxAge is how long a value lives. Millisecond precision.
	MaxAge time.Duration
	// MaxBytes caps the bucket's total size.
	MaxBytes uint64
	// MaxValueSize caps a single value, at most 2^31 - 1 bytes.
	MaxValueSize uint32
	Storage      StorageType
	// LimitMarkerTTL, when non-zero, enables per-key TTL and sets how long the
	// server keeps the marker it writes when a key expires or ages out. At
	// least one second, rounded up to whole seconds. Needs NATS 2.11+.
	LimitMarkerTTL time.Duration
}

// ConfigConflictError reports that [CreateBucket] found the bucket already
// exists with a configuration that differs from the request, or that is
// outside this binding's limits (a size over a lowered `bucket-max-bytes`).
// The bucket was left unchanged. Existing carries the configuration it has,
// to compare against or adopt with a second CreateBucket call.
type ConfigConflictError struct{ Existing BucketConfig }

func (e *ConfigConflictError) Error() string {
	return fmt.Sprintf("nats: bucket %q exists with a conflicting configuration", e.Existing.Bucket)
}

// CreateBucket ensures a JetStream KV bucket exists with a compatible
// configuration and opens it. It is idempotent, so it is safe to call on
// every start or every request:
//
//   - No bucket: it is created as asked, with the host's choices applied
//     (MaxBytes 0 takes `bucket-max-bytes`; replicas are `bucket-replicas`).
//   - A bucket matching every field set in cfg: it is opened unchanged.
//     MaxBytes 0 matches any size within the host's limit, and replicas are
//     never compared.
//   - Any other bucket: it is left unchanged, and the call fails with a
//     [*ConfigConflictError] carrying the configuration it has.
//
// The binding needs the bucket in both `bucket-allow` and
// `bucket-create-allow`; otherwise it fails with a [*DeniedError] even when
// the bucket already exists. Storage outside `bucket-storage-allow` is
// refused too, and a MaxBytes over `bucket-max-bytes` on a new bucket is a
// [*LimitExceededError]. Needs NATS 2.10+.
func CreateBucket(cfg BucketConfig) (*Bucket, error) {
	res := kvadmin.CreateBucket(toWitBucketConfig(cfg))
	if res.IsErr() {
		e := res.Err()
		switch e.Tag() {
		case kvadmin.CreateBucketError_ConfigConflict:
			return nil, &ConfigConflictError{Existing: fromWitBucketConfig(e.ConfigConflict())}
		case kvadmin.CreateBucketError_Nats:
			return nil, convertError(e.Nats())
		default:
			return nil, fmt.Errorf("nats: unknown create-bucket error (tag %d)", e.Tag())
		}
	}
	return &Bucket{inner: res.Ok()}, nil
}

// GetBucketConfig reads a bucket's configuration. It needs only
// `bucket-allow`. A missing bucket returns a [*NotFoundError].
func GetBucketConfig(bucket string) (BucketConfig, error) {
	res := kvadmin.GetBucketConfig(bucket)
	if res.IsErr() {
		return BucketConfig{}, convertError(res.Err())
	}
	return fromWitBucketConfig(res.Ok()), nil
}

// DeleteBucket deletes a bucket and every key in it. It cannot be undone, and
// open [Bucket] handles fail on their next call. The binding needs the bucket
// in `bucket-allow` and `bucket-delete-allow`. A missing bucket returns a
// [*NotFoundError].
func DeleteBucket(bucket string) error {
	if res := kvadmin.DeleteBucket(bucket); res.IsErr() {
		return convertError(res.Err())
	}
	return nil
}

func toWitBucketConfig(c BucketConfig) kvadmin.BucketConfig {
	marker := witTypes.None[uint64]()
	if c.LimitMarkerTTL > 0 {
		marker = witTypes.Some(uint64(c.LimitMarkerTTL.Milliseconds()))
	}
	return kvadmin.BucketConfig{
		Bucket:           c.Bucket,
		History:          c.History,
		MaxAgeMs:         uint64(c.MaxAge.Milliseconds()),
		MaxBytes:         c.MaxBytes,
		MaxValueSize:     c.MaxValueSize,
		Storage:          kvadmin.StorageType(c.Storage),
		LimitMarkerTtlMs: marker,
	}
}

func fromWitBucketConfig(c kvadmin.BucketConfig) BucketConfig {
	return BucketConfig{
		Bucket:         c.Bucket,
		History:        c.History,
		MaxAge:         time.Duration(c.MaxAgeMs) * time.Millisecond,
		MaxBytes:       c.MaxBytes,
		MaxValueSize:   c.MaxValueSize,
		Storage:        StorageType(c.Storage),
		LimitMarkerTTL: time.Duration(c.LimitMarkerTtlMs.SomeOr(0)) * time.Millisecond,
	}
}

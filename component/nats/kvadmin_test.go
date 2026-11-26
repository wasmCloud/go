package nats

import (
	"testing"
	"time"
)

func TestBucketConfigRoundTrip(t *testing.T) {
	in := BucketConfig{
		Bucket:         "carts",
		History:        5,
		MaxAge:         90 * time.Second,
		MaxBytes:       64 << 20,
		MaxValueSize:   1 << 16,
		Storage:        StorageMemory,
		LimitMarkerTTL: 2 * time.Second,
	}
	if got := fromWitBucketConfig(toWitBucketConfig(in)); got != in {
		t.Fatalf("round trip = %+v, want %+v", got, in)
	}
}

// A zero LimitMarkerTTL must reach the host as `none`, which leaves per-key
// TTL off. Some(0) would ask for a marker TTL below the 1s minimum.
func TestBucketConfigZeroMarkerTTLIsNone(t *testing.T) {
	w := toWitBucketConfig(BucketConfig{Bucket: "b"})
	if w.LimitMarkerTtlMs.IsSome() {
		t.Fatalf("LimitMarkerTtlMs = Some(%d), want None", w.LimitMarkerTtlMs.Some())
	}
	if w.Storage != 0 {
		t.Fatalf("default storage tag = %d, want file (0)", w.Storage)
	}
}

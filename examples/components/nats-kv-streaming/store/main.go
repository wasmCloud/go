// The store component: objects of any size on a NATS KV bucket.
//
// A KV value is one NATS message. It is capped by the server's `max_payload`
// (1 MiB by default), and reading one hands the whole value to the guest in a
// single allocation. So an object is never one value. It is a manifest plus
// fixed-size chunks, which is the layout NATS's own object store uses:
//
//	meta.<name>                   the manifest: size, chunk count, SHA-256, …
//	chunk.<name>.<upload>.<n>     chunk n of one version of the object
//
// Every put writes its chunks under a fresh upload id and writes the manifest
// last, so the manifest swap is the commit: a reader that started before it
// keeps reading the old chunks, one that started after it reads the new ones,
// and nobody sees a mix. The replaced version's chunks are kept until the
// next commit of the same name, which gives a reader that was mid-way through
// an object until then to finish.
//
// Delete is a commit too: it swaps in a deleted marker with the same
// compare-and-swap, so it only ever removes the version it read.
//
// Bytes come in and go out as `stream<u8>`, a chunk at a time, so this
// component's memory use is one chunk however large the object is.
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.bytecodealliance.org/wasi/wasilog"
	"go.wasmcloud.dev/component/nats"

	exports "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/storeexports/export_wasmcloud_archive_store"
	store "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/storeexports/wasmcloud_archive_store"
	"github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/streamio"
	witTypes "go.bytecodealliance.org/pkg/wit/types"

	// Links the //go:wasmexport glue for wasmcloud:archive/store.
	_ "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/storeexports/wit_exports"

	// Puts the SDK's WIT, and the headless world this component is built
	// with, where componentize-go looks for them.
	_ "go.wasmcloud.dev/component"
)

const (
	// The bucket this component is granted (`bucket-allow`). It never creates
	// it: provisioning storage is the operator's job, not the workload's.
	bucketName = "archive"

	// 512 KiB leaves room under the default 1 MiB `max_payload` for the
	// message headers. Raising it means fewer, larger KV round trips and a
	// larger allocation per chunk read.
	chunkSize = 512 << 10

	metaPrefix  = "meta."
	chunkPrefix = "chunk."

	// How many times a put retries its manifest swap when another put of the
	// same name commits first.
	maxCommitAttempts = 5
)

// Object names become KV key tokens, so they are kept to characters that are
// valid in a key and contain no `.` token separator.
var validName = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// manifest is what meta.<name> holds.
type manifest struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        uint64 `json:"size"`
	ChunkSize   uint32 `json:"chunkSize"`
	Chunks      uint32 `json:"chunks"`
	SHA256      string `json:"sha256"`
	CreatedAtMs uint64 `json:"createdAtMs"`
	// Upload namespaces this version's chunk keys.
	Upload string `json:"upload"`
	// The version this one replaced, whose chunks are kept for one more
	// generation so a read of it that was in flight can finish.
	Previous       string `json:"previous,omitempty"`
	PreviousChunks uint32 `json:"previousChunks,omitempty"`
	// Deleted marks a manifest left by delete. It names no chunks, and the
	// object reads as not found until the next put of the name.
	Deleted bool `json:"deleted,omitempty"`
}

func (m *manifest) info() store.ObjectInfo {
	return store.ObjectInfo{
		Name:        m.Name,
		ContentType: m.ContentType,
		Size:        m.Size,
		ChunkSize:   m.ChunkSize,
		Chunks:      m.Chunks,
		Sha256:      m.SHA256,
		CreatedAtMs: m.CreatedAtMs,
	}
}

func metaKey(name string) string { return metaPrefix + name }

func chunkKey(name, upload string, n uint32) string {
	return fmt.Sprintf("%s%s.%s.%06d", chunkPrefix, name, upload, n)
}

func init() {
	exports.Exports.Put = put
	exports.Exports.Get = get
	exports.Exports.Stat = stat
	exports.Exports.ListObjects = listObjects
	exports.Exports.Delete = deleteObject
}

// put streams data into chunks, then commits the manifest.
func put(name, contentType string, data *witTypes.StreamReader[uint8], done *exports.Done) witTypes.Result[store.ObjectInfo, store.Error] {
	log := wasilog.ContextLogger("put")

	// Release both handles on every path. Until they are dropped the writer
	// on the other side keeps waiting for us to read.
	body := streamio.NewReader(data, done)
	defer body.Close()

	if !validName.MatchString(name) {
		return fail[store.ObjectInfo](store.MakeErrorInvalidName(name))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	bucket, err := nats.OpenBucket(bucketName)
	if err != nil {
		return fail[store.ObjectInfo](storageError(err))
	}
	defer bucket.Close()

	started := time.Now()
	m := manifest{
		Name:        name,
		ContentType: contentType,
		ChunkSize:   chunkSize,
		Upload:      newUploadID(),
	}

	// One buffer, reused for every chunk. io.ReadFull fills it straight from
	// the stream, and Put sends it from there: the guest never holds more
	// than this one chunk of the object.
	buf := make([]byte, chunkSize)
	sum := sha256.New()
	for {
		n, readErr := io.ReadFull(body, buf)
		if n > 0 {
			if _, err := bucket.Put(chunkKey(name, m.Upload, m.Chunks), buf[:n]); err != nil {
				purgeChunks(bucket, name, m.Upload, m.Chunks)
				return fail[store.ObjectInfo](storageError(fmt.Errorf("writing chunk %d: %w", m.Chunks, err)))
			}
			sum.Write(buf[:n])
			m.Chunks++
			m.Size += uint64(n)
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			// The stream ended and its writer reported success.
			break
		}
		if readErr != nil {
			// The writer reported a failure, or the stream broke. Whatever
			// arrived is a fragment: throw it away and leave the previous
			// version, if there is one, exactly as it was.
			purgeChunks(bucket, name, m.Upload, m.Chunks)
			log.Warn("put aborted", "name", name, "after_bytes", m.Size, "err", readErr)
			return fail[store.ObjectInfo](store.MakeErrorAborted(readErr.Error()))
		}
	}
	m.SHA256 = hex.EncodeToString(sum.Sum(nil))
	m.CreatedAtMs = uint64(time.Now().UnixMilli())

	replaced, err := commit(bucket, &m)
	if err != nil {
		purgeChunks(bucket, name, m.Upload, m.Chunks)
		return fail[store.ObjectInfo](storageError(err))
	}

	// The version before the one just replaced was kept for its readers
	// until this commit. A reader still on it now fails part way, with an
	// error on its future, rather than reading a mix of versions.
	if replaced != nil && replaced.Previous != "" {
		purgeChunks(bucket, name, replaced.Previous, replaced.PreviousChunks)
	}

	log.Info("stored object",
		"name", name, "bytes", m.Size, "chunks", m.Chunks, "sha256", m.SHA256,
		"elapsed", time.Since(started).Round(time.Millisecond).String())
	return witTypes.Ok[store.ObjectInfo, store.Error](m.info())
}

// commit swaps the manifest in with compare-and-swap, so two puts of the same
// name cannot both believe they replaced the same version. It returns the
// manifest it replaced, if there was one.
func commit(bucket *nats.Bucket, m *manifest) (*manifest, error) {
	var lastErr error
	for range maxCommitAttempts {
		current, revision, err := readManifest(bucket, m.Name)
		if err != nil && !errors.Is(err, nats.ErrKeyNotFound) {
			return nil, err
		}
		m.Previous, m.PreviousChunks = "", 0
		if current != nil {
			m.Previous, m.PreviousChunks = current.Upload, current.Chunks
		}
		value, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		if current == nil {
			_, err = bucket.Create(metaKey(m.Name), value)
		} else {
			_, err = bucket.Update(metaKey(m.Name), value, revision)
		}
		if err == nil {
			return current, nil
		}
		if !lostRace(bucket, m.Name, current, err) {
			return nil, err
		}
		// Another put committed between our read and our write. Re-read and
		// replace whatever it wrote instead.
		lastErr = err
	}
	return nil, fmt.Errorf("manifest for %q kept changing; gave up after %d attempts: %w", m.Name, maxCommitAttempts, lastErr)
}

// lostRace reports whether a failed manifest write failed because another
// writer committed first, rather than because storage failed.
//
// A failed Update says so itself, with a RevisionMismatchError. A failed
// Create does not say why, so the manifest is read back: if one exists now,
// someone else created it.
func lostRace(bucket *nats.Bucket, name string, current *manifest, err error) bool {
	if current != nil {
		var mismatch *nats.RevisionMismatchError
		return errors.As(err, &mismatch)
	}
	_, _, readErr := readManifest(bucket, name)
	return readErr == nil
}

// get returns the manifest and a stream that a goroutine fills one chunk at a
// time. The export returns as soon as the stream exists; the goroutine keeps
// running after it, which is what lets the caller start forwarding bytes
// before this component has read the second chunk.
func get(name string) witTypes.Result[witTypes.Tuple3[store.ObjectInfo, *witTypes.StreamReader[uint8], *exports.Done], store.Error] {
	type result = witTypes.Tuple3[store.ObjectInfo, *witTypes.StreamReader[uint8], *exports.Done]

	bucket, m, err := openObject(name)
	if err != nil {
		return fail[result](err.(storeError).wit)
	}

	data, dataReader := store.MakeStreamU8()
	done, doneReader := store.MakeFutureResultUnitString()
	out := streamio.NewWriter(data, done)

	go func() {
		defer bucket.Close()
		err := copyChunks(out, bucket, m)
		if err != nil && !errors.Is(err, streamio.ErrReaderGone) {
			wasilog.ContextLogger("get").Warn("read failed", "name", name, "err", err)
		}
		streamio.Close(out, err)
	}()

	return witTypes.Ok[result, store.Error](result{F0: m.info(), F1: dataReader, F2: doneReader})
}

// copyChunks writes every chunk of m to out, in order, and checks the size
// and SHA-256 against the manifest.
//
// The last chunk is checked before it is written, so an object that fails
// the check always reaches the reader short. A reader that knows the size,
// as an HTTP client does from Content-Length, can then tell it is incomplete
// without waiting for the error on the future.
func copyChunks(out io.Writer, bucket *nats.Bucket, m *manifest) error {
	sum := sha256.New()
	var read uint64
	for n := range m.Chunks {
		entry, err := bucket.Get(chunkKey(m.Name, m.Upload, n))
		if err != nil {
			return fmt.Errorf("chunk %d of %d: %w", n, m.Chunks, err)
		}
		sum.Write(entry.Value)
		read += uint64(len(entry.Value))

		if read > m.Size {
			return fmt.Errorf("chunk %d of %d runs past the manifest size of %d bytes", n, m.Chunks, m.Size)
		}
		if n == m.Chunks-1 {
			if err := verify(m, read, sum.Sum(nil)); err != nil {
				return err
			}
		}

		// Write blocks until the reader takes the chunk, so a slow reader
		// paces the KV reads instead of chunks piling up here.
		if _, err := out.Write(entry.Value); err != nil {
			return err
		}
	}
	if m.Chunks == 0 {
		return verify(m, 0, sum.Sum(nil))
	}
	return nil
}

func verify(m *manifest, size uint64, sum []byte) error {
	if size != m.Size {
		return fmt.Errorf("read %d bytes, manifest says %d", size, m.Size)
	}
	if got := hex.EncodeToString(sum); got != m.SHA256 {
		return fmt.Errorf("sha256 is %s, manifest says %s", got, m.SHA256)
	}
	return nil
}

func stat(name string) witTypes.Result[store.ObjectInfo, store.Error] {
	bucket, m, err := openObject(name)
	if err != nil {
		return fail[store.ObjectInfo](err.(storeError).wit)
	}
	bucket.Close()
	return witTypes.Ok[store.ObjectInfo, store.Error](m.info())
}

func listObjects() witTypes.Result[[]store.ObjectInfo, store.Error] {
	bucket, err := nats.OpenBucket(bucketName)
	if err != nil {
		return fail[[]store.ObjectInfo](storageError(err))
	}
	defer bucket.Close()

	// The filter is applied host-side, so the host's listing cap counts
	// objects rather than every chunk key in the bucket.
	page, err := bucket.Keys(metaPrefix + "*")
	if err != nil {
		return fail[[]store.ObjectInfo](storageError(err))
	}
	if page.Truncated {
		wasilog.ContextLogger("list").Warn("listing truncated by the host's key cap", "returned", len(page.Keys))
	}

	infos := make([]store.ObjectInfo, 0, len(page.Keys))
	for _, key := range page.Keys {
		m, _, err := readManifest(bucket, strings.TrimPrefix(key, metaPrefix))
		if errors.Is(err, nats.ErrKeyNotFound) {
			continue // deleted since the listing
		}
		if err != nil {
			return fail[[]store.ObjectInfo](storageError(err))
		}
		if m.Deleted {
			continue
		}
		infos = append(infos, m.info())
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return witTypes.Ok[[]store.ObjectInfo, store.Error](infos)
}

// deleteObject replaces the manifest with a deleted marker, then purges the
// chunks of the version it replaced and of the one before that.
//
// Purging the manifest instead would race a put: a put that commits between
// the read and the purge would lose its manifest, and its chunks would be
// left with nothing naming them. The marker is written with compare-and-swap
// against the manifest that was read, so delete only removes what it read,
// and once the marker is in, no manifest names either version's chunks.
//
// A download of the deleted version that is still in progress fails part way.
func deleteObject(name string) witTypes.Result[witTypes.Unit, store.Error] {
	if !validName.MatchString(name) {
		return fail[witTypes.Unit](store.MakeErrorInvalidName(name))
	}
	bucket, err := nats.OpenBucket(bucketName)
	if err != nil {
		return fail[witTypes.Unit](storageError(err))
	}
	defer bucket.Close()

	marker, err := json.Marshal(manifest{Name: name, Deleted: true})
	if err != nil {
		return fail[witTypes.Unit](storageError(err))
	}

	for range maxCommitAttempts {
		m, revision, err := readManifest(bucket, name)
		if errors.Is(err, nats.ErrKeyNotFound) || (err == nil && m.Deleted) {
			return fail[witTypes.Unit](store.MakeErrorNotFound())
		}
		if err != nil {
			return fail[witTypes.Unit](storageError(err))
		}

		_, err = bucket.Update(metaKey(name), marker, revision)
		var mismatch *nats.RevisionMismatchError
		if errors.As(err, &mismatch) {
			continue // a put committed first; delete what it wrote instead
		}
		if err != nil {
			return fail[witTypes.Unit](storageError(err))
		}

		purgeChunks(bucket, name, m.Upload, m.Chunks)
		if m.Previous != "" {
			purgeChunks(bucket, name, m.Previous, m.PreviousChunks)
		}
		return witTypes.Ok[witTypes.Unit, store.Error](witTypes.Unit{})
	}
	return fail[witTypes.Unit](storageError(fmt.Errorf("manifest for %q kept changing; gave up after %d attempts", name, maxCommitAttempts)))
}

// --- helpers -----------------------------------------------------------------

// storeError carries a WIT error through a Go error return.
type storeError struct{ wit store.Error }

func (e storeError) Error() string { return fmt.Sprintf("store error (tag %d)", e.wit.Tag()) }

func openObject(name string) (*nats.Bucket, *manifest, error) {
	if !validName.MatchString(name) {
		return nil, nil, storeError{store.MakeErrorInvalidName(name)}
	}
	bucket, err := nats.OpenBucket(bucketName)
	if err != nil {
		return nil, nil, storeError{storageError(err)}
	}
	m, _, err := readManifest(bucket, name)
	if err == nil && m.Deleted {
		err = nats.ErrKeyNotFound
	}
	if err != nil {
		bucket.Close()
		if errors.Is(err, nats.ErrKeyNotFound) {
			return nil, nil, storeError{store.MakeErrorNotFound()}
		}
		return nil, nil, storeError{storageError(err)}
	}
	return bucket, m, nil
}

func readManifest(bucket *nats.Bucket, name string) (*manifest, uint64, error) {
	entry, err := bucket.Get(metaKey(name))
	if err != nil {
		return nil, 0, err
	}
	var m manifest
	if err := json.Unmarshal(entry.Value, &m); err != nil {
		return nil, 0, fmt.Errorf("manifest for %q: %w", name, err)
	}
	return &m, entry.Revision, nil
}

// purgeChunks removes chunks 0..count-1 of one upload. It is best-effort: a
// chunk that cannot be purged is unreachable once no manifest names its
// upload, so the failure is logged rather than returned.
func purgeChunks(bucket *nats.Bucket, name, upload string, count uint32) {
	for n := range count {
		if err := bucket.Purge(chunkKey(name, upload, n)); err != nil {
			wasilog.ContextLogger("purge").Warn("could not purge chunk",
				"name", name, "upload", upload, "chunk", n, "err", err)
		}
	}
}

func newUploadID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func storageError(err error) store.Error { return store.MakeErrorStorage(err.Error()) }

func fail[T any](e store.Error) witTypes.Result[T, store.Error] {
	return witTypes.Err[T](e)
}

func main() {}

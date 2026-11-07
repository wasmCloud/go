// The api component: HTTP in front of the store and query components.
//
// Every handler moves bytes as streams. An upload is piped from the request
// body into the store; a download is the store's stream copied into the
// response; a query is the store's stream handed to the query component and
// its output copied into the response. None of them reads a body into memory,
// so a 10 MB object costs this component the same as a 10 KB one.
//
//	GET    /objects                 list objects
//	PUT    /objects/{name}          store the request body
//	GET    /objects/{name}          download
//	GET    /objects/{name}/query    filter readings: ?site= &sensor= &min= &limit= &format=csv|ndjson
//	DELETE /objects/{name}          delete
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	"go.bytecodealliance.org/wasi/wasihttp"
	"go.bytecodealliance.org/wasi/wasilog"

	"github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/client/wasmcloud_archive_query"
	store "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/client/wasmcloud_archive_store"
	"github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/streamio"

	// Anchor go.wasmcloud.dev/component in go.mod: it carries the wasmCloud
	// worlds' WIT and componentize-go.toml used at build time.
	_ "go.wasmcloud.dev/component"
)

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /objects", listObjects)
	mux.HandleFunc("PUT /objects/{name}", putObject)
	mux.HandleFunc("GET /objects/{name}", getObject)
	mux.HandleFunc("GET /objects/{name}/query", queryObject)
	mux.HandleFunc("DELETE /objects/{name}", deleteObject)
	wasihttp.Handle(mux)
}

// putObject pipes the request body into the store.
func putObject(w http.ResponseWriter, r *http.Request) {
	log := wasilog.ContextLogger("put")
	started := time.Now()

	data, dataReader := store.MakeStreamU8()
	done, doneReader := store.MakeFutureResultUnitString()
	body := streamio.NewWriter(data, done)

	// The copy runs alongside the Put call below: Put reads the stream while
	// this goroutine fills it. If the client goes away mid-upload, io.Copy
	// returns its error and the store discards what it has so far, rather
	// than committing a truncated object.
	go func() {
		_, err := io.Copy(body, r.Body)
		streamio.Close(body, err)
	}()

	res := store.Put(r.PathValue("name"), r.Header.Get("Content-Type"), dataReader, doneReader)
	if res.IsErr() {
		writeStoreError(w, res.Err())
		return
	}
	info := res.Ok()
	log.Info("uploaded", "name", info.Name, "bytes", info.Size, "elapsed", time.Since(started).Round(time.Millisecond).String())
	writeJSON(w, http.StatusCreated, toJSON(info))
}

// getObject copies the store's stream into the response.
func getObject(w http.ResponseWriter, r *http.Request) {
	res := store.Get(r.PathValue("name"))
	if res.IsErr() {
		writeStoreError(w, res.Err())
		return
	}
	info, data, done := res.Ok().F0, res.Ok().F1, res.Ok().F2
	body := streamio.NewReader(data, done)
	defer body.Close()

	h := w.Header()
	h.Set("Content-Type", info.ContentType)
	// The size is known up front, so the client can tell a truncated
	// download from a complete one even without the checksum.
	h.Set("Content-Length", strconv.FormatUint(info.Size, 10))
	h.Set("ETag", strconv.Quote(info.Sha256))
	h.Set("X-Archive-Sha256", info.Sha256)

	if n, err := io.Copy(w, body); err != nil {
		// Too late for a status code: the headers are gone. Ending the body
		// short of Content-Length is how HTTP/1.1 says "this is broken", and
		// the store has already logged why.
		wasilog.ContextLogger("get").Warn("download incomplete", "name", info.Name, "sent", n, "size", info.Size, "err", err)
	}
}

// queryObject hands the store's stream to the query component and copies the
// query's output into the response. The object's bytes go from the store to
// the query component directly; this component only ever sees the matches.
func queryObject(w http.ResponseWriter, r *http.Request) {
	filter, err := parseFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := store.Get(r.PathValue("name"))
	if res.IsErr() {
		writeStoreError(w, res.Err())
		return
	}
	data, done := res.Ok().F1, res.Ok().F2

	// Ownership of both handles moves to the query component here.
	out, summary := wasmcloud_archive_query.Run(data, done, filter)
	body := streamio.NewReader(out, summary)
	defer body.Close()

	h := w.Header()
	if filter.Format == wasmcloud_archive_query.FormatCsv {
		h.Set("Content-Type", "text/csv")
	} else {
		h.Set("Content-Type", "application/x-ndjson")
	}
	// How many matched is only known at the end, so it goes in trailers.
	// Declare each one as its own value: wasihttp matches them one by one
	// and, unlike net/http, does not split a comma-separated list.
	for _, name := range []string{"X-Query-Scanned", "X-Query-Matched", "X-Query-Malformed", "X-Query-Error"} {
		h.Add("Trailer", name)
	}

	_, err = io.Copy(w, body)
	s := body.Value()
	h.Set("X-Query-Scanned", strconv.FormatUint(s.Scanned, 10))
	h.Set("X-Query-Matched", strconv.FormatUint(s.Matched, 10))
	h.Set("X-Query-Malformed", strconv.FormatUint(s.Malformed, 10))
	if err != nil {
		// The status line was sent with the first match, so the failure goes
		// in a trailer: a client that checks it knows the results stopped
		// short.
		h.Set("X-Query-Error", err.Error())
		wasilog.ContextLogger("query").Warn("query incomplete", "name", r.PathValue("name"), "err", err)
		return
	}
	wasilog.ContextLogger("query").Info("query complete", "name", r.PathValue("name"),
		"scanned", s.Scanned, "matched", s.Matched, "malformed", s.Malformed)
}

func listObjects(w http.ResponseWriter, _ *http.Request) {
	res := store.ListObjects()
	if res.IsErr() {
		writeStoreError(w, res.Err())
		return
	}
	out := make([]objectJSON, 0, len(res.Ok()))
	for _, info := range res.Ok() {
		out = append(out, toJSON(info))
	}
	writeJSON(w, http.StatusOK, out)
}

func deleteObject(w http.ResponseWriter, r *http.Request) {
	if res := store.Delete(r.PathValue("name")); res.IsErr() {
		writeStoreError(w, res.Err())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers -----------------------------------------------------------------

func parseFilter(r *http.Request) (wasmcloud_archive_query.Filter, error) {
	q := r.URL.Query()
	f := wasmcloud_archive_query.Filter{
		Site:     optional(q.Get("site")),
		Sensor:   optional(q.Get("sensor")),
		MinValue: witTypes.None[float64](),
		Limit:    witTypes.None[uint32](),
		Format:   wasmcloud_archive_query.FormatNdjson,
	}
	if v := q.Get("min"); v != "" {
		min, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return f, fmt.Errorf("min: %w", err)
		}
		f.MinValue = witTypes.Some(min)
	}
	if v := q.Get("limit"); v != "" {
		limit, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return f, fmt.Errorf("limit: %w", err)
		}
		f.Limit = witTypes.Some(uint32(limit))
	}
	switch q.Get("format") {
	case "", "ndjson":
	case "csv":
		f.Format = wasmcloud_archive_query.FormatCsv
	default:
		return f, errors.New("format: want csv or ndjson")
	}
	return f, nil
}

func optional(s string) witTypes.Option[string] {
	if s == "" {
		return witTypes.None[string]()
	}
	return witTypes.Some(s)
}

type objectJSON struct {
	Name        string    `json:"name"`
	ContentType string    `json:"contentType"`
	Size        uint64    `json:"size"`
	ChunkSize   uint32    `json:"chunkSize"`
	Chunks      uint32    `json:"chunks"`
	SHA256      string    `json:"sha256"`
	CreatedAt   time.Time `json:"createdAt"`
}

func toJSON(info store.ObjectInfo) objectJSON {
	return objectJSON{
		Name:        info.Name,
		ContentType: info.ContentType,
		Size:        info.Size,
		ChunkSize:   info.ChunkSize,
		Chunks:      info.Chunks,
		SHA256:      info.Sha256,
		CreatedAt:   time.UnixMilli(int64(info.CreatedAtMs)).UTC(),
	}
}

func writeStoreError(w http.ResponseWriter, e store.Error) {
	switch e.Tag() {
	case store.ErrorNotFound:
		http.Error(w, "not found", http.StatusNotFound)
	case store.ErrorInvalidName:
		http.Error(w, "invalid name "+strconv.Quote(e.InvalidName())+": use 1-64 of [A-Za-z0-9_-]", http.StatusBadRequest)
	case store.ErrorAborted:
		http.Error(w, "upload aborted: "+e.Aborted(), http.StatusBadRequest)
	case store.ErrorStorage:
		http.Error(w, "storage: "+e.Storage(), http.StatusBadGateway)
	default:
		http.Error(w, "unknown store error", http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func main() {}

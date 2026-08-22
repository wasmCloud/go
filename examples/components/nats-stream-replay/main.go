// Reading a JetStream stream from a component, in Go.
//
// Three ways to get messages out of a stream, each with different
// consequences for everyone else reading it:
//
//	GET  /streams/{stream}/messages/{seq}          one message, by sequence
//	GET  /streams/{stream}/replay?from=&count=     a range, consuming nothing
//	POST /streams/{stream}/consumers/{name}/drain  a batch from a durable consumer
//
// The replay path is the one `wasmcloud:messaging` cannot express. It reads
// directly at a sequence and creates no consumer, so it does not move any
// other reader's position — that is what makes it safe to point at a live
// stream. The drain path does the opposite on purpose: it takes messages
// from a durable consumer and settles them, so what it reads, nobody else
// gets.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.bytecodealliance.org/pkg/wasihttp"
	"go.bytecodealliance.org/pkg/wasilog"
	"go.wasmcloud.dev/component/nats"

	// Anchor go.wasmcloud.dev/component in go.mod: it carries the wasmCloud
	// worlds' WIT and componentize-go.toml used at build time.
	_ "go.wasmcloud.dev/component"
)

const (
	defaultCount = 20
	maxCount     = 500
	// How long a fetch waits for the first message. A pull consumer with
	// nothing pending is the common case, so this stays short: the caller
	// is an HTTP request, not a background loop.
	fetchTimeout = time.Second
	// Redelivery backoff for a message that failed transiently, multiplied
	// by the delivery count so a persistently failing message backs off
	// instead of spinning.
	nakBackoff = 5 * time.Second
)

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", index)
	mux.HandleFunc("GET /streams/{stream}/messages/{seq}", getMessage)
	mux.HandleFunc("GET /streams/{stream}/replay", replay)
	mux.HandleFunc("POST /streams/{stream}/consumers/{consumer}/drain", drain)
	wasihttp.Handle(mux)
}

func index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "text/plain")
	io.WriteString(w, "GET  /streams/{stream}/messages/{seq}\n"+
		"GET  /streams/{stream}/replay?from=1&count=20\n"+
		"POST /streams/{stream}/consumers/{consumer}/drain?batch=10\n")
}

// getMessage reads one message by its stream sequence.
func getMessage(w http.ResponseWriter, r *http.Request) {
	stream := r.PathValue("stream")
	seq, err := strconv.ParseUint(r.PathValue("seq"), 10, 64)
	if err != nil {
		http.Error(w, "sequence must be a positive integer", http.StatusBadRequest)
		return
	}

	msg, err := nats.GetBySequence(stream, seq)
	if err != nil {
		writeNatsError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toJSON(msg))
}

// replay returns a range of messages without consuming anything.
func replay(w http.ResponseWriter, r *http.Request) {
	log := wasilog.ContextLogger("replay")
	stream := r.PathValue("stream")

	from, err := uintParam(r, "from", 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	count, err := uintParam(r, "count", defaultCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if count == 0 || count > maxCount {
		http.Error(w, fmt.Sprintf("count must be between 1 and %d", maxCount), http.StatusBadRequest)
		return
	}

	messages, err := nats.Scan(stream, from, uint32(count))
	if err != nil {
		writeNatsError(w, err)
		return
	}

	// The host caps both the number of messages and the time one scan may
	// take, so a short read is normal rather than the end of the stream.
	// `next` is where to resume; there is no "done" flag to report, because
	// an empty scan is the only thing that means caught up — and even that
	// only until the next publish.
	out := struct {
		Stream   string          `json:"stream"`
		From     uint64          `json:"from"`
		Count    int             `json:"count"`
		Next     uint64          `json:"next"`
		Messages []storedMessage `json:"messages"`
	}{
		Stream:   stream,
		From:     from,
		Count:    len(messages),
		Next:     from,
		Messages: make([]storedMessage, 0, len(messages)),
	}
	for _, m := range messages {
		out.Messages = append(out.Messages, toJSON(m))
		out.Next = m.Sequence + 1
	}

	log.Info("replayed", "stream", stream, "from", from, "returned", len(messages))
	writeJSON(w, http.StatusOK, out)
}

// drain fetches a batch from an existing durable consumer and settles every
// message it takes.
func drain(w http.ResponseWriter, r *http.Request) {
	log := wasilog.ContextLogger("drain")
	stream, consumer := r.PathValue("stream"), r.PathValue("consumer")

	batch, err := uintParam(r, "batch", 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if batch == 0 || batch > maxCount {
		http.Error(w, fmt.Sprintf("batch must be between 1 and %d", maxCount), http.StatusBadRequest)
		return
	}

	// The consumer must already exist: its lifecycle is deliberately outside
	// this interface, so a workload cannot provision durable state it was
	// not granted. Create it with `nats consumer add`.
	c, err := nats.OpenPullConsumer(stream, consumer)
	if err != nil {
		writeNatsError(w, err)
		return
	}
	defer c.Close()

	handles, err := c.Fetch(uint32(batch), fetchTimeout)
	if errors.Is(err, nats.ErrNoMessages) {
		// An idle consumer, not a failure.
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		writeNatsError(w, err)
		return
	}

	// Every handle has to be settled — ack, nak, or term — or the consumer
	// stalls until ack-wait expires and redelivers the whole batch.
	results := make([]result, 0, len(handles))
	for _, h := range handles {
		results = append(results, settle(h))
	}

	log.Info("drained", "stream", stream, "consumer", consumer, "messages", len(results))
	writeJSON(w, http.StatusOK, struct {
		Stream   string   `json:"stream"`
		Consumer string   `json:"consumer"`
		Results  []result `json:"results"`
	}{stream, consumer, results})
}

type result struct {
	Sequence      uint64 `json:"sequence"`
	Subject       string `json:"subject"`
	DeliveryCount uint32 `json:"deliveryCount"`
	Outcome       string `json:"outcome"`
	Detail        string `json:"detail,omitempty"`
}

// settle processes one delivery and reports which of the three outcomes it
// chose. Which one matters more than it looks: an ack that should have been
// a nak drops work silently, and a nak that should have been a term retries
// forever.
func settle(h *nats.MessageHandle) result {
	msg := h.Message()
	res := result{
		Sequence:      h.Sequence(),
		Subject:       msg.Subject,
		DeliveryCount: h.DeliveryCount(),
	}

	err := process(h)
	switch {
	case err == nil:
		res.Outcome, res.Detail = "ack", settleErr(h.Ack())

	case errors.Is(err, errPoison):
		// No number of retries fixes a malformed payload, so take it off
		// the consumer for good rather than redelivering it forever.
		res.Outcome, res.Detail = "term", settleErr(h.Term())

	default:
		// Transient. Back off proportionally to how often this has already
		// come back, so a message that keeps failing stops hammering.
		delay := nakBackoff * time.Duration(h.DeliveryCount())
		res.Outcome = "nak"
		res.Detail = err.Error()
		if nakErr := settleErr(h.Nak(delay)); nakErr != "" {
			res.Detail = nakErr
		}
	}
	return res
}

// errPoison marks a message no retry can fix.
var errPoison = errors.New("malformed payload")

// process stands in for real work. It reads `key=value` bodies: anything
// else is poison, and a value of "fail" is a transient failure.
func process(h *nats.MessageHandle) error {
	body := string(h.Message().Body)

	key, value, found := strings.Cut(body, "=")
	if !found || key == "" {
		return fmt.Errorf("%w: %q", errPoison, truncate(body, 64))
	}

	if value == "slow" {
		// Work that outlives the consumer's ack-wait must say so, or the
		// message is redelivered while this attempt is still running.
		// Unlike the settling calls, this one may be made repeatedly.
		if err := h.InProgress(); err != nil {
			return fmt.Errorf("extending ack-wait: %w", err)
		}
	}
	if value == "fail" {
		return errors.New("downstream unavailable")
	}
	return nil
}

// settleErr reports a settling failure, which is worth surfacing rather
// than ignoring: under a binding configured with `ack-mode: auto` the host
// owns the acknowledgement and these calls are refused.
func settleErr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

type storedMessage struct {
	Subject  string            `json:"subject"`
	Sequence uint64            `json:"sequence"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     string            `json:"body,omitempty"`
	// Set instead of Body when the payload is not valid UTF-8.
	BodyBase64 []byte `json:"bodyBase64,omitempty"`
}

func toJSON(m nats.StoredMessage) storedMessage {
	out := storedMessage{Subject: m.Subject, Sequence: m.Sequence}
	if len(m.Headers) > 0 {
		out.Headers = make(map[string]string, len(m.Headers))
		for _, h := range m.Headers {
			out.Headers[h.Name] = h.Value
		}
	}
	if utf8.Valid(m.Data) {
		out.Body = string(m.Data)
	} else {
		out.BodyBase64 = m.Data
	}
	return out
}

// writeNatsError maps a JetStream read failure onto the HTTP status that
// means the same thing, using the SDK's typed errors rather than matching
// on message text.
func writeNatsError(w http.ResponseWriter, err error) {
	var notFound *nats.NotFoundError
	var denied *nats.SubjectDeniedError
	var unsupported *nats.UnsupportedByServerError

	switch {
	case errors.As(err, &notFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, nats.ErrKeyNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.As(err, &denied):
		// Outside `stream-allow`; the read never reached the server.
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.As(err, &unsupported):
		// Reading by sequence needs NATS server 2.9 or newer.
		http.Error(w, err.Error(), http.StatusNotImplemented)
	case errors.Is(err, nats.ErrDisconnected):
		http.Error(w, err.Error(), http.StatusBadGateway)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		wasilog.ContextLogger("writeJSON").Error("encoding response", "error", err)
	}
}

func uintParam(r *http.Request, name string, fallback uint64) (uint64, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a non-negative integer", name)
	}
	return v, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// main is required for a Go component but never runs: the host invokes the
// exported HTTP handler instead.
func main() {}

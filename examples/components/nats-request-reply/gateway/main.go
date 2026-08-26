// An HTTP-to-NATS request-reply gateway, in Go.
//
// It turns an HTTP request into a NATS request and the NATS reply into an
// HTTP response, which is the other half of the pattern the sibling
// `service` component implements. Everything interesting is in the error
// mapping: a NATS request fails in ways HTTP has names for, and
// `wasmcloud:nats` reports each of them as a distinct Go error rather than
// a string to match on.
package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.bytecodealliance.org/pkg/wasihttp"
	"go.bytecodealliance.org/pkg/wasilog"
	"go.wasmcloud.dev/component/nats"

	// Anchor go.wasmcloud.dev/component in go.mod: it carries the wasmCloud
	// worlds' WIT and componentize-go.toml used at build time.
	_ "go.wasmcloud.dev/component"
)

const (
	defaultTimeout = 2 * time.Second
	maxTimeout     = 30 * time.Second
	// Largest request body forwarded to NATS. The server enforces its own
	// maximum too, but rejecting here keeps an oversized upload from being
	// read into memory first.
	maxRequestBytes = 1 << 20
)

// Set by the sibling service on a failed reply.
const (
	errorHeader     = "Nats-Service-Error"
	errorCodeHeader = "Nats-Service-Error-Code"
)

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", index)
	// The trailing "..." captures the remaining path segments, so
	// POST /ask/service/greet requests the subject `service.greet`.
	mux.HandleFunc("POST /ask/{subject...}", ask)
	wasihttp.Handle(mux)
}

func index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "text/plain")
	io.WriteString(w, "POST /ask/<subject-path>[?timeout=2s] - request/reply over NATS\n"+
		"  e.g. POST /ask/service/greet -> subject service.greet\n")
}

func ask(w http.ResponseWriter, r *http.Request) {
	log := wasilog.ContextLogger("ask")

	subject := strings.ReplaceAll(r.PathValue("subject"), "/", ".")
	if subject == "" {
		http.Error(w, "no subject in path", http.StatusBadRequest)
		return
	}

	timeout, err := parseTimeout(r.URL.Query().Get("timeout"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBytes+1))
	if err != nil {
		http.Error(w, "reading request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(body) > maxRequestBytes {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	// Request blocks until a reply arrives or the timeout elapses. Nothing
	// is retried: a retried request is a second request as far as the
	// responder is concerned, and only the caller knows whether that is safe.
	start := time.Now()
	replyMsg, err := nats.Request(nats.Message{
		Subject: subject,
		Body:    body,
		Headers: forwardedHeaders(r),
	}, timeout)
	if err != nil {
		status, detail := statusFor(subject, err)
		log.Warn("request failed", "subject", subject, "status", status, "error", err)
		http.Error(w, detail, status)
		return
	}

	log.Info("request served", "subject", subject, "took", time.Since(start).String())
	writeReply(w, replyMsg)
}

// statusFor maps a NATS failure onto the HTTP status that means the same
// thing. Each case is a typed error from the SDK, so none of this is string
// matching.
func statusFor(subject string, err error) (int, string) {
	var denied *nats.DeniedError
	var tooBig *nats.MaxPayloadExceededError

	switch {
	case errors.Is(err, nats.ErrNoResponders):
		// Nothing is subscribed. Retrying immediately fails identically
		// until a responder appears, so this is a 503 and not a 504.
		return http.StatusServiceUnavailable, "no responders on " + subject

	case errors.As(err, &denied):
		// The name is outside this workload's grant, so the message never
		// left the host. The host declares the grants, so widening one is
		// the operator's call and never something to work around here.
		return http.StatusForbidden, fmt.Sprintf("%s %s is not permitted", denied.Target, denied.Name)

	case errors.As(err, &tooBig):
		return http.StatusRequestEntityTooLarge,
			fmt.Sprintf("body exceeds the server maximum of %d bytes", tooBig.Limit)

	case errors.Is(err, nats.ErrDisconnected):
		return http.StatusBadGateway, "not connected to NATS"

	default:
		// A responder existed but did not answer in time, or the
		// connection failed mid-request. Both leave the request in an
		// unknown state — it may well have been processed.
		return http.StatusGatewayTimeout, "no reply within the timeout"
	}
}

// writeReply turns a service reply into an HTTP response, honouring the
// NATS micro error headers the sibling service sets.
func writeReply(w http.ResponseWriter, msg nats.Message) {
	if detail, ok := msg.Header(errorHeader); ok {
		status := http.StatusBadGateway
		if raw, ok := msg.Header(errorCodeHeader); ok {
			if code, err := strconv.Atoi(raw); err == nil && code >= 400 && code <= 599 {
				status = code
			}
		}
		http.Error(w, detail, status)
		return
	}

	for _, h := range msg.Headers {
		w.Header().Add("X-Nats-"+h.Name, h.Value)
	}
	w.Header().Set("content-type", "application/octet-stream")
	w.Write(msg.Body)
}

// forwardedHeaders carries a couple of HTTP headers through as NATS headers.
// Only an explicit allow list travels: forwarding everything would leak
// cookies and authorization into the message bus.
func forwardedHeaders(r *http.Request) []nats.Header {
	var out []nats.Header
	for _, name := range []string{"Content-Type", "X-Request-Id"} {
		if v := r.Header.Get(name); v != "" {
			out = append(out, nats.Header{Name: name, Value: v})
		}
	}
	return out
}

func parseTimeout(raw string) (time.Duration, error) {
	if raw == "" {
		return defaultTimeout, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout %q: %w", raw, err)
	}
	if d <= 0 || d > maxTimeout {
		return 0, fmt.Errorf("timeout must be between 0 and %s", maxTimeout)
	}
	return d, nil
}

// main is required for a Go component but never runs: the host invokes the
// exported HTTP handler instead.
func main() {}

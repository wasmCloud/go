// An HTTP handler that calls the callee over wasi:http and reports the result.
// On a host running with same-host local routing, the outgoing request to a
// hostname the co-located callee offers via `localRoute` is served in-memory.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"go.bytecodealliance.org/pkg/wasihttp"

	// Anchor go.wasmcloud.dev/component in go.mod: it carries the wasmCloud
	// worlds' WIT and componentize-go.toml used at build time.
	_ "go.wasmcloud.dev/component"
)

// defaultTarget resolves nowhere in real DNS, so a successful call proves the
// host short-circuited it.
const defaultTarget = "http://callee.internal/hello"

func init() {
	wasihttp.HandleFunc(callHandler)
}

// callHandler resolves the target from `?url=`, then CALLEE_URL, then
// defaultTarget.
func callHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("url")
	if target == "" {
		target = os.Getenv("CALLEE_URL")
	}
	if target == "" {
		target = defaultTarget
	}

	resp, err := wasihttp.DefaultClient.Get(target)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "caller -> %s\nrequest failed: %v\n", target, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte(fmt.Sprintf("<unreadable body: %v>\n", err))
	}
	fmt.Fprintf(w, "caller -> %s\nupstream status: %d %s\nupstream body: %s",
		target, resp.StatusCode, http.StatusText(resp.StatusCode), body)
}

func main() {}

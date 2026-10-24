// An HTTP handler that answers with a greeting echoing the path and Host it
// received. Reached through same-host local routing, the echoed host is the
// authority the caller dialed, a name no DNS server knows.
package main

import (
	"fmt"
	"net/http"

	"go.bytecodealliance.org/pkg/wasihttp"

	// Anchor go.wasmcloud.dev/component in go.mod: it carries the wasmCloud
	// worlds' WIT and componentize-go.toml used at build time.
	_ "go.wasmcloud.dev/component"
)

func init() {
	wasihttp.HandleFunc(helloHandler)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello from the Go callee! (path: %s, host: %s)\n", r.URL.Path, r.Host)
}

func main() {}

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
	// Register the handleRequest function as the handler for all incoming requests.
	wasihttp.HandleFunc(handleRequest)
}

//nolint:revive
func handleRequest(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Hello from Go!\n")
}

// Since we don't run this program like a CLI, the `main` function is empty. Instead,
// we call the `handleRequest` function when an HTTP request is received.
func main() {}

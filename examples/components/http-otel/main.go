package main

import (
	"io"
	"net/http"

	"go.bytecodealliance.org/pkg/wasihttp"
	"go.opentelemetry.io/otel"

	// Anchor go.wasmcloud.dev/component in go.mod: it carries the wasmCloud
	// worlds' WIT and componentize-go.toml used at build time.
	_ "go.wasmcloud.dev/component"
)

const (
	name        = "github.com/wasmCloud/go/examples/components/http-otel"
	serviceName = "http-otel"
)

var tracer = otel.Tracer(name)

func init() {
	if err := setupOTelSDK(); err != nil {
		panic(err)
	}

	router := http.NewServeMux()
	router.HandleFunc("/", echoHandler)
	wasihttp.Handle(router)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	// Parent this request's span into the host's trace for the incoming
	// HTTP request.
	ctx := parentFromHost(r.Context())
	_, span := tracer.Start(ctx, serviceName)
	defer span.End()

	w.WriteHeader(http.StatusOK)
	_, err := io.Copy(w, r.Body)
	if err != nil {
		http.Error(w, "failed to copy input to response", http.StatusInternalServerError)
		return
	}
}

func main() {}

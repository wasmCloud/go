package main

import (
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.wasmcloud.dev/component/net/wasihttp"
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

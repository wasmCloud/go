//go:generate go tool wit-bindgen-go generate --world wasmcloud:wasitel-echo/wasitel-echo --out gen ./wit

package main

import (
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.wasmcloud.dev/component/net/wasihttp"
)

const (
	name        = "github.com/wasmCloud/go/examples/component/wasitel-echo"
	serviceName = "wasitel-http"
)

var tracer = otel.Tracer(name)

func init() {
	router := http.NewServeMux()
	router.HandleFunc("/", echoHandler)
	wasihttp.Handle(router)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	if err := setupOTelSDK(); err != nil {
		http.Error(w, "failed setting up otel", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
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

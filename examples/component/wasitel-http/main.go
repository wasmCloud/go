//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --world example --out gen ./wit

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
	ctx := r.Context()
	setupOTelSDK()
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

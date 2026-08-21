package main

import (
	"io"
	"net/http"

	"go.bytecodealliance.org/pkg/wasihttp"

	// Anchor go.wasmcloud.dev/component in go.mod: it carries the wasmCloud
	// worlds' WIT and componentize-go.toml used at build time.
	_ "go.wasmcloud.dev/component"
)

var (
	wasiTransport = &wasihttp.Transport{}
	httpClient    = &http.Client{Transport: wasiTransport}
)

func init() {
	wasihttp.HandleFunc(proxyHandler)
}

func proxyHandler(w http.ResponseWriter, _ *http.Request) {
	url := "https://dog.ceo/api/breeds/image/random"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusBadGateway)
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, "failed to make outbound request", http.StatusBadGateway)
		return
	}

	w.Header().Set("X-Custom-Header", "proxied")
	w.WriteHeader(resp.StatusCode)

	_, _ = io.Copy(w, resp.Body)
}
func main() {}

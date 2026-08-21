// An async WASI P3 component demonstrating streaming bodies and native
// concurrency: request bodies echo back without buffering, and plain
// goroutines fan out concurrent outbound requests.
package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"go.bytecodealliance.org/pkg/wasihttp"

	// Anchor go.wasmcloud.dev/component in go.mod: it carries the wasmCloud
	// worlds' WIT and componentize-go.toml used at build time.
	_ "go.wasmcloud.dev/component"
)

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /echo", echoHandler)
	mux.HandleFunc("GET /fanout", fanoutHandler)
	wasihttp.Handle(mux)
}

// echoHandler streams the request body straight back to the response —
// bytes flow through as they arrive, with no buffering, regardless of body
// size.
func echoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", r.Header.Get("content-type"))
	if _, err := io.Copy(w, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// fanoutHandler issues concurrent outbound requests with ordinary
// goroutines; the P3 runtime multiplexes them on the component's single
// thread while each blocked goroutine parks.
func fanoutHandler(w http.ResponseWriter, r *http.Request) {
	urls := []string{
		"https://dog.ceo/api/breeds/image/random",
		"https://dog.ceo/api/breeds/list/all",
	}

	type result struct {
		url    string
		status string
		bytes  int64
		err    error
	}

	results := make([]result, len(urls))
	var wg sync.WaitGroup
	for i, url := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := result{url: url}
			defer func() { results[i] = res }()

			resp, err := wasihttp.DefaultClient.Get(url)
			if err != nil {
				res.err = err
				return
			}
			defer resp.Body.Close()
			res.status = resp.Status
			res.bytes, res.err = io.Copy(io.Discard, resp.Body)
		}()
	}
	wg.Wait()

	for _, res := range results {
		if res.err != nil {
			fmt.Fprintf(w, "%s: error: %v\n", res.url, res.err)
			continue
		}
		fmt.Fprintf(w, "%s: %s (%d bytes)\n", res.url, res.status, res.bytes)
	}
}

func main() {}

package main

//go:generate wit-bindgen-go generate --world example --out gen ./wit

import (
	"io"
	"net/http"
	"strings"

	"go.wasmcloud.dev/component"
	"go.wasmcloud.dev/component/net/wasihttp"
)

var logger = component.DefaultLogger

func init() {
	// We can't use http.ServeMux yet ( only symbol linking is supported in 'init' )
	wasihttp.HandleFunc(entryHandler)
}

func entryHandler(w http.ResponseWriter, r *http.Request) {
	handlers := map[string]http.HandlerFunc{
		"/headers": headersHandler,
		"/error":   errorHandler,
		"/form":    formHandler,
		"/post":    postHandler,
	}

	logger.Info("Request received", "path", r.URL.Path)

	if handler, ok := handlers[r.URL.Path]; ok {
		handler(w, r)
		return
	}

	index := `
  /headers - echo your user agent back as a server side header
  /error - return a 500 error
  /form - echo the fields of a POST request
  /post - echo the body of a POST request
  `

	var keys []string
	for k := range handlers {
		keys = append(keys, k)
	}
	w.Header().Add("content-type", "text/plain")
	w.Header().Add("X-Requested-Path", r.URL.Path)
	w.Header().Add("X-Existing-Paths", strings.Join(keys, ","))
	w.Write([]byte(index))
}

func headersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Your-User-Agent", r.Header.Get("User-Agent"))

	w.Write([]byte("Check headers!"))
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Something went wrong", http.StatusInternalServerError)
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for key, values := range r.Form {
		w.Write([]byte(key + ": " + strings.Join(values, ",") + "\n"))
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.Copy(w, r.Body)
}

func main() {}
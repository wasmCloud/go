package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	// A lightweight, high performance HTTP request router
	"github.com/julienschmidt/httprouter"

	// Bindings for the wasi:keyvalue/store interface, generated from ./wit
	// by componentize-go.
	store "github.com/wasmCloud/go/examples/components/http-keyvalue-crud/wasi_keyvalue_store"

	// The wasmCloud wasihttp module enables us to write more idiomatic Go when using wasi:http.
	"go.wasmcloud.dev/component/net/wasihttp"
)

// Types for JSON validation.
type CheckRequest struct {
	Value string `json:"value"`
}

type CheckResponse struct {
	Valid   bool   `json:"valid"`
	Length  int    `json:"length,omitempty"`
	Message string `json:"message,omitempty"`
}

func init() {
	// Establishes the routes and methods for our key-value operations.
	router := httprouter.New()
	router.GET("/", indexHandler)
	router.POST("/crud/:key", postHandler)
	router.GET("/crud/:key", getHandler)
	router.DELETE("/crud/:key", deleteHandler)
	wasihttp.Handle(router)
}

func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintln(w, `{"message":"GET, POST, or DELETE to /crud/<key> (with JSON payload for POSTs)"}`)
}

// openBucket opens the default keyvalue bucket served by the host.
func openBucket() (*store.Bucket, error) {
	res := store.Open("default")
	if res.IsErr() {
		return nil, fmt.Errorf("failed to open bucket: %s", errString(res.Err()))
	}
	return res.Ok(), nil
}

func postHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Assigns the "key" parameter to the "key" variable.
	key := ps.ByName("key")

	// Checks the request for a valid JSON body and assigns it to the value variable.
	// The user will set the value via JSON payload:
	// curl -X POST 'localhost:8000/crud/key' -d '{"foo": "bar", "woo": "hoo"}'
	var req CheckRequest
	defer r.Body.Close()
	value, err := io.ReadAll(r.Body)
	if err != nil {
		errResponseJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := json.Unmarshal(value, &req); err != nil {
		errResponseJSON(w, http.StatusBadRequest, fmt.Sprintf("error with json input: %s", err.Error()))
		return
	}

	bucket, err := openBucket()
	if err != nil {
		errResponseJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Sets the value for the key in the current bucket and handles any errors.
	if kvSet := bucket.Set(key, value); kvSet.IsErr() {
		errResponseJSON(w, http.StatusBadRequest, errString(kvSet.Err()))
		return
	}

	// Confirms set, returning key and value in JSON body.
	fmt.Fprintf(w, `{"message":"Set %s", "value":"%s"}`+"\n", key, value)
}

func getHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Assigns the "key" parameter to the "key" variable.
	key := ps.ByName("key")

	bucket, err := openBucket()
	if err != nil {
		errResponseJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Gets the value for the defined key.
	kvGet := bucket.Get(key)
	if kvGet.IsErr() {
		errResponseJSON(w, http.StatusBadRequest, errString(kvGet.Err()))
		return
	}

	// Returns and reports that key does not exist if no value is found.
	value := kvGet.Ok()
	if value.IsNone() {
		errResponseJSON(w, http.StatusBadRequest, fmt.Sprintf("%s does not exist", key))
		return
	}

	// Returns key and value in JSON body.
	fmt.Fprintf(w, `{"message":"Got %s", "value":"%s"}`+"\n", key, value.Some())
}

func deleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Assigns the "key" parameter to the "key" variable.
	key := ps.ByName("key")

	bucket, err := openBucket()
	if err != nil {
		errResponseJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Returns and reports that key does not exist if no value is found.
	kvExists := bucket.Exists(key)
	if kvExists.IsErr() {
		errResponseJSON(w, http.StatusBadRequest, errString(kvExists.Err()))
		return
	}
	if !kvExists.Ok() {
		errResponseJSON(w, http.StatusBadRequest, fmt.Sprintf("%s does not exist", key))
		return
	}

	// Deletes the entry for the provided key.
	if kvDel := bucket.Delete(key); kvDel.IsErr() {
		errResponseJSON(w, http.StatusBadRequest, errString(kvDel.Err()))
		return
	}

	// Confirms delete in JSON body.
	fmt.Fprintf(w, `{"message":"Deleted %s"}`+"\n", key)
}

// errString renders a wasi:keyvalue/store error variant as text.
func errString(e store.Error) string {
	switch e.Tag() {
	case store.ErrorNoSuchStore:
		return "no such store"
	case store.ErrorAccessDenied:
		return "access denied"
	default:
		return e.Other()
	}
}

// JSON validation handling.
func errResponseJSON(w http.ResponseWriter, code int, message string) {
	msg, _ := json.Marshal(CheckResponse{Valid: false, Message: message})
	http.Error(w, string(msg), code)
	w.Header().Set("Content-Type", "application/json")
}

// Since we don't run this program like a CLI, the `main` function is empty. Instead,
// we call handler functions when an HTTP request is received.
func main() {}

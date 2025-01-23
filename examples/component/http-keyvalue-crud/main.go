//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --world component --out gen ./wit
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	// A lightweight, high performance HTTP request router
	"github.com/julienschmidt/httprouter"

	// For the keyvalue capability, we're using bindings for the wasi:keyvalue/store interface.
	store "github.com/wasmCloud/go/examples/component/http-keyvalue-crud/gen/wasi/keyvalue/store"

	// The cm module provides types and functions for interacting with the WebAssembly Component Model.
	"go.bytecodealliance.org/cm"

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

func postHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Assigns the "key" paramater to the "key" variable.
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

	// Opens the keyvalue bucket.
	kvStore := store.Open("default")
	if err := kvStore.Err(); err != nil {
		errResponseJSON(w, http.StatusInternalServerError, err.String())
		return
	}

	// Converts the value to a byte array.
	valueBytes := []byte(value)

	// Converts the byte array to the Component Model's cm.List type.
	valueList := cm.ToList(valueBytes)

	// Sets the value for the key in the current bucket and handles any errors.
	kvSet := store.Bucket.Set(*kvStore.OK(), key, valueList)
	if kvSet.IsErr() {
		errResponseJSON(w, http.StatusBadRequest, kvSet.Err().String())
		return
	}

	// Confirms set, returning key and value in JSON body.
	kvSetMessage := fmt.Sprintf("Set %s", key)
	kvSetResponse := fmt.Sprintf(`{"message":"%s", "value":"%s"}`, kvSetMessage, value)
	fmt.Fprintln(w, kvSetResponse)

}

func getHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Assigns the "key" paramater to the "key" variable.
	key := ps.ByName("key")

	// Opens the keyvalue bucket.
	kvStore := store.Open("default")
	if err := kvStore.Err(); err != nil {
		errResponseJSON(w, http.StatusInternalServerError, err.String())
		return
	}

	// Gets the value for the defined key.
	kvGet, kvGetErr, kvGetIsErr := store.Bucket.Get(*kvStore.OK(), key).Result()

	// Returns and reports that key does not exist if no value is found.
	if kvGet.Value().Len() == 0 {
		errResponseJSON(w, http.StatusBadRequest, fmt.Sprintf("%s does not exist", key))
		return
	}
	// Handles get errors other than non-existent key
	if kvGetIsErr {
		errResponseJSON(w, http.StatusBadRequest, kvGetErr.String())
		return
	}

	// Uses cm.LiftString to convert the byte value into a string, taking the data and len as arguments.
	kvGetJSON := cm.LiftString[string](kvGet.Value().Data(), kvGet.Value().Len())

	// Returns key and value in JSON body.
	kvGetMessage := fmt.Sprintf("Got %s", key)
	kvGetResponse := fmt.Sprintf(`{"message":"%s", "value":"%s"}`, kvGetMessage, kvGetJSON)
	fmt.Fprintln(w, kvGetResponse)

}

func deleteHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Assigns the "key" paramater to the "key" variable.
	key := ps.ByName("key")

	// Opens the keyvalue bucket.
	kvStore := store.Open("default")
	if err := kvStore.Err(); err != nil {
		errResponseJSON(w, http.StatusInternalServerError, err.String())
		return
	}
	// Returns and reports that key does not exist if no value is found.
	kvGet, _, _ := store.Bucket.Get(*kvStore.OK(), key).Result()
	if kvGet.Value().Len() == 0 {
		errResponseJSON(w, http.StatusBadRequest, fmt.Sprintf("%s does not exist", key))
		return
	}
	// Deletes the entry for the provided key.
	kvDel := store.Bucket.Delete(*kvStore.OK(), key)

	if kvDel.IsErr() {
		errResponseJSON(w, http.StatusBadRequest, kvDel.Err().String())
		return
	}
	// Confirms delete in JSON body.
	kvDelMessage := fmt.Sprintf("Deleted %s", key)
	kvDelResponse := fmt.Sprintf(`{"message":"%s"}`, kvDelMessage)
	fmt.Fprintln(w, kvDelResponse)

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

//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --world hello --out gen ./wit
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

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
	// Establishes the routes and methods for our keyvalue operations.
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/", indexHandler)
	router.HandlerFunc(http.MethodPost, "/post", postHandler)
	router.HandlerFunc(http.MethodGet, "/get", getHandler)
	router.HandlerFunc(http.MethodDelete, "/delete", deleteHandler)
	wasihttp.Handle(router)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Provide a key as query string to /post, /get, or /delete with the corresponding method\n")
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	// Checks for a valid key as query string.
	// For all routes and methods, the user will provide the key via a query string, for example:
	// curl 'localhost:8000/get?key'
	if len(r.URL.RawQuery) > 0 && regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(r.URL.RawQuery) {

		// Assigns the raw query string to the "key" variable.
		key := r.URL.RawQuery

		// Checks the request for a valid JSON body and assigns it to the value variable.
		// The user will set the value via JSON payload:
		// curl -X POST 'localhost:8000/post?key' -d '{"foo": "bar", "woo": "hoo"}'
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
			w.Write([]byte("Error: " + err.String()))
			return
		}

		// Converts the value to a byte array.
		valueBytes := []byte(value)

		// Converts the byte array to the Component Model's cm.List type.
		valueList := cm.ToList(valueBytes)

		// Sets the value for the key in the current bucket and handles any errors.
		kvSet := store.Bucket.Set(*kvStore.OK(), key, valueList)
		if err := kvSet.Err(); err != nil {
			w.Write([]byte("Error: " + err.String()))
			return
		} else {
			// Confirms that the key has been set.
			fmt.Fprintf(w, "Set %s to %s\n", key, value)
		}

	} else {

		// Prompts the user if no valid key is entered.
		fmt.Fprintf(w, "POST a key as an alphanumeric query with a JSON body\n")
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {

	// Checks for a valid key as query string.
	if len(r.URL.RawQuery) > 0 && regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(r.URL.RawQuery) {

		// Assigns the raw query string to the "key" variable.
		key := r.URL.RawQuery

		// Opens the keyvalue bucket.
		kvStore := store.Open("default")
		if err := kvStore.Err(); err != nil {
			w.Write([]byte("Error: " + err.String()))
			return
		}

		// Gets the value for the defined key.
		kvGet, kvGetErr, kvGetIsErr := store.Bucket.Get(*kvStore.OK(), key).Result()
		if err := kvGetErr; kvGetIsErr {
			w.Write([]byte("Error: " + err.String()))
			return
		} else if kvGet.Value().Len() == 0 {

			// Reports that key does not exist if no value is found.
			fmt.Fprintf(w, "%s does not exist\n", key)

		} else {

			// Uses cm.LiftString to convert the byte value into a string, taking the data and len as arguments.
			kvGetValue := cm.LiftString[string](kvGet.Value().Data(), kvGet.Value().Len())

			// Returns key and value.
			fmt.Fprintf(w, "Got %s value: %s\n", key, kvGetValue)
		}

	} else {

		// Prompts the user if no valid key is entered.
		fmt.Fprintf(w, "GET a key as an alphanumeric query\n")
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {

	// Checks for a valid key as query string.
	if len(r.URL.RawQuery) > 0 && regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(r.URL.RawQuery) {

		// Assigns the raw query string to the "key" variable.
		key := r.URL.RawQuery

		// Opens the keyvalue bucket.
		kvStore := store.Open("default")
		if err := kvStore.Err(); err != nil {
			w.Write([]byte("Error: " + err.String()))
			return
		}

		// Deletes the entry for the provided key.
		kvDel := store.Bucket.Delete(*kvStore.OK(), key)
		if err := kvDel.Err(); err != nil {
			w.Write([]byte("Error: " + err.String()))
			return
		} else {
			fmt.Fprintf(w, "Deleted %s\n", key)
		}

	} else {

		// Prompts the user if no valid key is entered.
		fmt.Fprintf(w, "DELETE a key as an alphanumeric query\n")
	}

}

// JSON validation handling.
func errResponseJSON(w http.ResponseWriter, code int, message string) {
	msg, _ := json.Marshal(CheckResponse{Valid: false, Message: message})
	http.Error(w, string(msg), code)
	w.Header().Set("Content-Type", "application/json")
}

// Since we don't run this program like a CLI, the `main` function is empty. Instead,
// we call the `handleRequest` function when an HTTP request is received.
func main() {}

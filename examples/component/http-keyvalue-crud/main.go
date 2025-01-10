//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --world hello --out gen ./wit
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

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
	// Establishes the endpoint for our keyvalue operations.
	// If we wanted to make this app more endpoint-driven (e.g., Create/Update at api/v1/crud/set),
	// we could create additional paths here. This example simply operates according to method.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/crud", handleRequest)
	wasihttp.Handle(mux)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {

	// If statements check the request method.
	// If the method is POST, attempts to perform a "set" against the keyvalue store (ie. Create and Update)
	if r.Method == http.MethodPost {

		// Checks for a valid key as query string.
		// This is how the user will provide the key for all methods -- a query string in a call like:
		// curl 'localhost:8000/api/v1/crud?key'
		if len(r.URL.RawQuery) > 0 && regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(r.URL.RawQuery) {

			// Assigns the raw query string to the "key" variable.
			key := r.URL.RawQuery

			// Checks the request for a valid JSON body and assigns it to the value variable.
			// This is how the user will set the value -- in a JSON payload like:
			// curl -X POST 'localhost:8000/api/v1/crud?key' -d '{"foo": "bar", "woo": "hoo"}'
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

		// If the method is GET, attempts to perform a "get" against the keyvalue store (ie. Read).
	} else if r.Method == http.MethodGet {

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

	} else if r.Method == http.MethodDelete {

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

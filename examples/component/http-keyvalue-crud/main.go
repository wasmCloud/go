//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --world hello --out gen ./wit
package main

import (
	store "crud/gen/wasi/keyvalue/store"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"encoding/json"

	"go.bytecodealliance.org/cm"
	"go.wasmcloud.dev/component/net/wasihttp"
)

type CheckRequest struct {
	Value string `json:"value"`
}

type CheckResponse struct {
	Valid   bool   `json:"valid"`
	Length  int    `json:"length,omitempty"`
	Message string `json:"message,omitempty"`
}

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/crud", handleRequest)
	wasihttp.Handle(mux)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {

	// Checks the request method
	if r.Method == http.MethodPost {

		// Checks for a valid key as query string
		if len(r.URL.RawQuery) > 0 && regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString(r.URL.RawQuery) {

			// Assigns the query string to the key variable
			key := r.URL.RawQuery

			// Checks the request for a valid JSON body and assign it to the value variable
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

			// Opens the keyvalue bucket
			kvStore := store.Open("default")
			if err := kvStore.Err(); err != nil {
				w.Write([]byte("Error: " + err.String()))
				return
			}

			// Converts the value to a byte array
			valueBytes := []byte(value)

			// Converts the byte array to the component model's cm.List type
			valueList := cm.ToList(valueBytes)

			// Sets the value for the key "foo" in the current bucket and handles any errors
			kvSet := store.Bucket.Set(*kvStore.OK(), key, valueList)
			if err := kvSet.Err(); err != nil {
				w.Write([]byte("Error: " + err.String()))
				return
			}

			fmt.Fprintf(w, "Set %s to %s\n", key, value)

		} else {

			fmt.Fprintf(w, "Please enter 1) a key as an alphanumeric query and 2) a JSON body\n")

		}

	}

}

func errResponseJSON(w http.ResponseWriter, code int, message string) {
	msg, _ := json.Marshal(CheckResponse{Valid: false, Message: message})
	http.Error(w, string(msg), code)
	w.Header().Set("Content-Type", "application/json")
}

// Since we don't run this program like a CLI, the `main` function is empty. Instead,
// we call the `handleRequest` function when an HTTP request is received.
func main() {}

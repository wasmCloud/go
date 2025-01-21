//go:generate go run go.bytecodealliance.org/cmd/wit-bindgen-go generate --world hello --out gen ./wit
package main

import (
	"fmt"
	"net/http"

	// A lightweight, high performance HTTP request router.
	"github.com/julienschmidt/httprouter"

	// For the keyvalue capability, we're using bindings for the wasi:keyvalue/store and /atomics interfaces.
	atomics "http-keyvalue-bus/gen/wasi/keyvalue/atomics"
	store "http-keyvalue-bus/gen/wasi/keyvalue/store"

	// The wasmcloud:bus/lattice interface enables us to set the operative link definition for an interface.
	lattice "http-keyvalue-bus/gen/wasmcloud/bus/lattice"

	// The cm module provides types and functions for interacting with the WebAssembly Component Model.
	cm "go.bytecodealliance.org/cm"

	// These wasmCloud modules enable us to write more idiomatic Go when using wasi:http and wasi:logging.
	"go.wasmcloud.dev/component/log/wasilog"
	"go.wasmcloud.dev/component/net/wasihttp"
)

func init() {
	// Establishes the routes for the application.
	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/:kv/:name", kvHandler)
	wasihttp.Handle(router)
}

func indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "Make a request to /redis/<name> or /nats/<name>\n")
}

func kvHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	logger := wasilog.ContextLogger("kvHandler")

	// Return if the kv parameter doesn't specify one of our KV providers.
	if (ps.ByName("kv") != "redis") && (ps.ByName("kv") != "nats") {
		fmt.Fprintf(w, "Make a request to /redis/<name> or /nats/<name>\n")
		return
	}

	// Since we don't give the NATS link a name in the manifest, its name is 'default' and the application would default
	// to using it the first time we made a call to keyvalue:store or keyvalue:atomics. We reset the link for every call,
	// however, in case the link has previously been set to Redis.

	kv := "default"
	if ps.ByName("kv") == "redis" {
		kv = "redis"
	}

	// We use the wasmcloud:bus/lattice interface (included among our component-go imports) to set the interface(s) for which
	// we want to select a link -- in this case, wasi:keyvalue/store and wasi:keyvalue/atomics. We place the interfaces in a
	// slice and then convert the slice to the cm.List type required by SetLinkName below.

	storeInterface := lattice.NewCallTargetInterface("wasi", "keyvalue", "store")
	atomicsInterface := lattice.NewCallTargetInterface("wasi", "keyvalue", "atomics")
	InterfacesList := cm.ToList([]lattice.CallTargetInterface{storeInterface, atomicsInterface})

	// Here we set the active link for the wasi:keyvalue/store and wasi:keyvalue/atomics interfaces to the value of kv.
	lattice.SetLinkName(kv, InterfacesList)

	// Set name value.
	name := ps.ByName("name")

	logger.Info("Greeting", "name", name)
	kvStore := store.Open("default")
	if err := kvStore.Err(); err != nil {
		w.Write([]byte("Error: " + err.String()))
		return
	}
	value := atomics.Increment(*kvStore.OK(), name, 1)
	if err := value.Err(); err != nil {
		w.Write([]byte("Error: " + err.String()))
		return
	}
	fmt.Fprintf(w, "Hello x%d, %s!\n", *value.OK(), name)
}

// Since we don't run this program like a CLI, the `main` function is empty. Instead,
// we call the `handleRequest` function when an HTTP request is received.
func main() {}

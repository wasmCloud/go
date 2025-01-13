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
	// Establishes the routes for Redis and NATS operations.
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/", indexHandler)
	router.HandlerFunc(http.MethodGet, "/redis", redisHandler)
	router.HandlerFunc(http.MethodGet, "/nats", natsHandler)
	wasihttp.Handle(router)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Make a request to the /redis or /nats endpoint with a name as a query string\n")
}

func natsHandler(w http.ResponseWriter, r *http.Request) {
	logger := wasilog.ContextLogger("natsHandler")
	name := "World"
	if len(r.FormValue("name")) > 0 {
		name = r.FormValue("name")
	}

	// We use the wasmcloud:bus/lattice interface (included among our component-go imports)
	// to set the interface(s) for which we want to select a link -- in this case,
	// wasi:keyvalue/store and wasi:keyvalue/atomics. For each, we convert the specification
	// to a slice and then the cm.List type required by SetLinkName below.

	storeInterface := lattice.NewCallTargetInterface("wasi", "keyvalue", "store")
	storeInterfaceSlice := []lattice.CallTargetInterface{storeInterface}
	storeInterfaceList := cm.ToList(storeInterfaceSlice)

	atomicsInterface := lattice.NewCallTargetInterface("wasi", "keyvalue", "atomics")
	atomicsInterfaceSlice := []lattice.CallTargetInterface{atomicsInterface}
	atomicsInterfaceList := cm.ToList(atomicsInterfaceSlice)

	// Here we set the operative link for the wasi:keyvalue/store and
	// wasi:keyvalue/atomics interfaces to the named link 'default'.
	// Since we don't give the NATS link a name in the manifest, its name is 'default' and
	// the application will default to using it the first time we make a call to keyvalue:store
	// or keyvalue:atomics. We reset the link for every call, however, in case the link has
	// previously been set to Redis.

	lattice.SetLinkName("default", storeInterfaceList)
	lattice.SetLinkName("default", atomicsInterfaceList)

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

func redisHandler(w http.ResponseWriter, r *http.Request) {
	logger := wasilog.ContextLogger("redisHandler")
	name := "World"
	if len(r.FormValue("name")) > 0 {
		name = r.FormValue("name")
	}
	// Our manifest includes a named link called 'redis'. We can use the
	// wasmcloud:bus/lattice interface (included among our component-go imports)
	// to set the interface(s) for which we want to select a link -- in this
	// case, wasi:keyvalue/store and wasi:keyvalue/atomics. For each, we convert
	// the specification to a slice and then the cm.List type required by
	// SetLinkName below.

	storeInterface := lattice.NewCallTargetInterface("wasi", "keyvalue", "store")
	storeInterfaceSlice := []lattice.CallTargetInterface{storeInterface}
	storeInterfaceList := cm.ToList(storeInterfaceSlice)

	atomicsInterface := lattice.NewCallTargetInterface("wasi", "keyvalue", "atomics")
	atomicsInterfaceSlice := []lattice.CallTargetInterface{atomicsInterface}
	atomicsInterfaceList := cm.ToList(atomicsInterfaceSlice)

	// Here we set the operative link for the wasi:keyvalue/store and
	// wasi:keyvalue/atomics interfaces to the named link 'redis'.

	lattice.SetLinkName("redis", storeInterfaceList)
	lattice.SetLinkName("redis", atomicsInterfaceList)

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

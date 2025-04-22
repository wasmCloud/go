//go:generate go tool wit-bindgen --cm go.bytecodealliance.org/cm --world example --out gen ./wit

package main

import (
	"github.com/wasmCloud/go/examples/component/invoke/gen/example/invoker/invoker"
	"go.wasmcloud.dev/component/log/wasilog"
)

const InvokeResponse = "Hello from the invoker!"

func init() {
	invoker.Exports.Call = invokerCall
}

func invokerCall() string {
	logger := wasilog.ContextLogger("Call")
	logger.Info("Invoking function")
	return InvokeResponse
}

func main() {}

//go:generate go tool wit-bindgen-go generate --world wasmcloud:invoker/component --out gen ./wit

package main

import (
	"go.wasmcloud.dev/component/log/wasilog"
	"invoke/gen/wasmcloud/invoker/invoker"
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

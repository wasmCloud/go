//go:generate go run go.wasmcloud.dev/wadge/cmd/wadge-bindgen-go -test
//go:generate ./scripts/ensure-harness.sh
// ^^^ This will build the test harness if it doesn't exist locally already. If you made a change to
// the harness, you can just `rm -f build/test-harness.wasm` and re-run `wash build`

package main

import (
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.wasmcloud.dev/wadge"
)

func init() {
	log.SetFlags(0)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})))
	// NOTE: This isn't using `embed` because we generate the test harness at build time. Because of
	// this, things like linters fail if run first because the file isn't there yet
	component, err := os.ReadFile("build/test-harness.wasm")
	if err != nil {
		log.Fatalf("failed to read test harness: %s", err)
	}
	instance, err := wadge.NewInstance(&wadge.Config{
		Wasm: component,
	})
	if err != nil {
		log.Fatalf("failed to construct new instance: %s", err)
	}
	wadge.SetInstance(instance)
}

func TestCall(t *testing.T) {
	wadge.RunTest(t, func() {
		buf := call()
		assert.Contains(t, buf, "SUCCESS: we selected a row")
	})
}

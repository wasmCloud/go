//go:generate go run go.wasmcloud.dev/wadge/cmd/wadge-bindgen-go -test

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	incominghandler "go.wasmcloud.dev/component/gen/wasi/http/incoming-handler"
	"go.wasmcloud.dev/wadge"
	"go.wasmcloud.dev/wadge/wadgehttp"
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
}

func TestIncomingHandler(t *testing.T) {
	wadge.RunTest(t, func() {
		req, err := http.NewRequest("POST", "/api/v1/check", bytes.NewReader([]byte(`{"value": "tes12345!"}`)))
		if err != nil {
			t.Fatalf("failed to create new HTTP request: %s", err)
		}
		resp, err := wadgehttp.HandleIncomingRequest(incominghandler.Exports.Handle, req)
		if err != nil {
			t.Fatalf("failed to handle incoming HTTP request: %s", err)
		}
		if want, got := http.StatusBadRequest, resp.StatusCode; want != got {
			t.Fatalf("unexpected status code: want %d, got %d", want, got)
		}
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read HTTP response body: %s", err)
		}
		defer resp.Body.Close()

		var respBody CheckResponse
		if err := json.Unmarshal(buf, &respBody); err != nil {
			t.Fatalf("failed to unmarshal response body: %s", err)
		}

		assert.False(t, respBody.Valid)
		assert.Contains(t, respBody.Message, "insecure password, try including more special characters, using uppercase letters or using a longer password")
	})
}

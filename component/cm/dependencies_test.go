//go:build !wasip1 && !wasip2 && !tinygo

package cm

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestDependencies(t *testing.T) {
	cmd := exec.Command("go", "list", "-f", "{{.Imports}}", "-tags", "module.std", ".")
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	err := cmd.Run()
	if err != nil {
		t.Error(err)
		return
	}

	got := strings.TrimSpace(stdout.String())
	// 	const want = "[structs unsafe]" // Should not include "encoding/json"
	const want = "[encoding/json structs unsafe]" // NOTE(lxf):wasmcloud vendored
	if got != want {
		t.Errorf("Expected dependencies %s, got %s", want, got)
	}
}

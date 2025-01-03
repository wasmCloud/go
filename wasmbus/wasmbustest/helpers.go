package wasmbustest

// NOTE(lxf): Avoid cyclical dependencies.
// This package should not import wadm/wasmbus/ctl/events.

import (
	"flag"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

// Run tests with '-show-output' flag to see the output of wash commands
var showOutput = flag.Bool("show-output", false, "show output of wash commands")

func CheckWadm() error {
	return Exec("wash", "app", "list")
}

func Exec(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	// NOTE(lxf): Uncomment the following lines to see the output of wash commands
	if *showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func WashDeploy(t *testing.T, path string) {
	t.Helper()
	if err := Exec("wash", "app", "deploy", path); err != nil {
		t.Fatalf("failed to deploy manifest: %v", err)
	}
}

func WithWash(t *testing.T) (*nats.Conn, func(*testing.T)) {
	t.Helper()

	if err := Exec("wash", "up", "-d"); err != nil {
		t.Fatalf("failed to start wash: %v", err)
		return nil, nil
	}

	maxTimeout := time.After(10 * time.Second)
	connected := false
	for !connected {
		select {
		case <-maxTimeout:
			t.Fatalf("timeout waiting for wash to start")
			return nil, nil
		case <-time.After(250 * time.Millisecond):
			if err := CheckWadm(); err == nil {
				connected = true
				continue
			} else {
				t.Logf("waiting for wash to start: %v", err)
			}
		}
	}

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("failed to connect to nats: %v", err)
		return nil, nil
	}

	return nc, func(*testing.T) {
		nc.Close()
		_ = Exec("wash", "down", "--purge-jetstream", "all")
	}
}

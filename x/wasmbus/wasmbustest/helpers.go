package wasmbustest

// NOTE(lxf): Avoid cyclical dependencies.
// This package should not import wadm/wasmbus/ctl/events.

import (
	"flag"
	"os"
	"os/exec"
	"time"

	"github.com/nats-io/nats.go"
)

// Run tests with '-wash-output' flag to see the output of wash commands
var showOutput = flag.Bool("wash-output", false, "show output of wash commands")

const (
	ValidComponent = "ghcr.io/wasmcloud/components/http-hello-world-rust:0.1.0"
	ValidProvider  = "ghcr.io/wasmcloud/http-client:0.12.0"
)

// interface that can be fulfilled by any testing.T implementation ( std, ginkgo, testify )
type TestingT interface {
	Helper()
	Fatalf(string, ...any)
	Logf(string, ...any)
}

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

func WashDeploy(t TestingT, path string) {
	t.Helper()
	if err := Exec("wash", "app", "deploy", path); err != nil {
		t.Fatalf("failed to deploy manifest: %v", err)
	}
}

func WithWash(t TestingT) (*nats.Conn, func(TestingT)) {
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

	return nc, func(TestingT) {
		nc.Close()
		if *showOutput {
			_ = Exec("wash", "get", "inventory")
		}
		_ = Exec("wash", "down", "--purge-jetstream", "all")
	}
}

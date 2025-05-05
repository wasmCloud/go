package provider

import (
	"fmt"
	"testing"

	"github.com/nats-io/nkeys"
)

// This test ensures that the LatticeTopics function returns the correct topics for wasmCloud 1.0 and 1.1+.
func TestLatticeTopics(t *testing.T) {
	xkey, err := nkeys.CreateCurveKeys()
	if err != nil {
		t.Errorf("Expected err to be nil, got: %v", err)
	}
	wasmCloudOneDotZero := HostData{ProviderKey: "providerfoo", LatticeRPCPrefix: "lattice123", ProviderXKeyPrivateKey: RedactedString(""), HostXKeyPublicKey: ""}
	OneDotZeroTopics := LatticeTopics(wasmCloudOneDotZero, xkey)

	// Test LatticeLinkGet
	expectedLinkGet := "wasmbus.rpc.lattice123.providerfoo.linkdefs.get"
	if OneDotZeroTopics.LatticeLinkGet != expectedLinkGet {
		t.Errorf("Expected LatticeLinkGet to be %q, got %q", expectedLinkGet, OneDotZeroTopics.LatticeLinkGet)
	}

	// Test LatticeLinkDel
	expectedLinkDel := "wasmbus.rpc.lattice123.providerfoo.linkdefs.del"
	if OneDotZeroTopics.LatticeLinkDel != expectedLinkDel {
		t.Errorf("Expected LatticeLinkDel to be %q, got %q", expectedLinkDel, OneDotZeroTopics.LatticeLinkDel)
	}

	// Test LatticeLinkPut
	expectedLinkPut := "wasmbus.rpc.lattice123.providerfoo.linkdefs.put"
	if OneDotZeroTopics.LatticeLinkPut != expectedLinkPut {
		t.Errorf("Expected LatticeLinkPut to be %q, got %q", expectedLinkPut, OneDotZeroTopics.LatticeLinkPut)
	}

	// Test LatticeShutdown
	expectedShutdown := "wasmbus.rpc.lattice123.providerfoo.default.shutdown"
	if OneDotZeroTopics.LatticeShutdown != expectedShutdown {
		t.Errorf("Expected LatticeShutdown to be %q, got %q", expectedShutdown, OneDotZeroTopics.LatticeShutdown)
	}

	// Test LatticeHealth
	expectedHealth := "wasmbus.rpc.lattice123.providerfoo.health"
	if OneDotZeroTopics.LatticeHealth != expectedHealth {
		t.Errorf("Expected LatticeHealth to be %q, got %q", expectedHealth, OneDotZeroTopics.LatticeHealth)
	}

	// Test secrets / wasmCloud 1.1 and later topics. All are the same as 1.0 except LatticeLinkPut
	xkeyPublicKey, err := xkey.PublicKey()
	if err != nil {
		t.Errorf("Expected err to be nil, got: %v", err)
	}
	xkeyPrivateKey, err := xkey.Seed()
	if err != nil {
		t.Errorf("Expected err to be nil, got: %v", err)
	}
	wasmCloudOneDotOne := HostData{ProviderKey: "providerfoo", LatticeRPCPrefix: "lattice123", ProviderXKeyPrivateKey: RedactedString(string(xkeyPrivateKey)), HostXKeyPublicKey: xkeyPublicKey}
	OneDotOneTopics := LatticeTopics(wasmCloudOneDotOne, xkey)

	// Test LatticeLinkGet
	if OneDotOneTopics.LatticeLinkGet != expectedLinkGet {
		t.Errorf("Expected LatticeLinkGet to be %q, got %q", expectedLinkGet, OneDotOneTopics.LatticeLinkGet)
	}

	// Test LatticeLinkDel
	if OneDotOneTopics.LatticeLinkDel != expectedLinkDel {
		t.Errorf("Expected LatticeLinkDel to be %q, got %q", expectedLinkDel, OneDotOneTopics.LatticeLinkDel)
	}

	// Test LatticeLinkPut
	if err != nil {
		t.Errorf("Expected err to be nil, got: %v", err)
	}
	expectedLinkPut = fmt.Sprintf("wasmbus.rpc.lattice123.%s.linkdefs.put", xkeyPublicKey)
	if OneDotOneTopics.LatticeLinkPut != expectedLinkPut {
		t.Errorf("Expected LatticeLinkPut to be %q, got %q", expectedLinkPut, OneDotOneTopics.LatticeLinkPut)
	}

	// Test LatticeShutdown
	if OneDotOneTopics.LatticeShutdown != expectedShutdown {
		t.Errorf("Expected LatticeShutdown to be %q, got %q", expectedShutdown, OneDotOneTopics.LatticeShutdown)
	}

	// Test LatticeHealth
	if OneDotOneTopics.LatticeHealth != expectedHealth {
		t.Errorf("Expected LatticeHealth to be %q, got %q", expectedHealth, OneDotOneTopics.LatticeHealth)
	}
}

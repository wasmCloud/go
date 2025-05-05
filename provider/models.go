package provider

import (
	"fmt"

	"github.com/nats-io/nkeys"
)

type Topics struct {
	LatticeLinkGet  string
	LatticeLinkDel  string
	LatticeLinkPut  string
	LatticeHealth   string
	LatticeShutdown string
}

func LatticeTopics(h HostData, providerXkey nkeys.KeyPair) Topics {
	// With secrets support in wasmCloud, links are delivered to the link put topic
	// where the topic segment is the XKey provider public key. On wasmCloud host
	// versions before secrets (<1.1.0), the topic segment is the provider key.
	// We can determine the topic segment based on the presence of the host xkey
	// public key and the provider xkey private key.
	var providerLinkPutKey string
	publicKey, err := providerXkey.PublicKey()
	if h.HostXKeyPublicKey == "" || h.ProviderXKeyPrivateKey == "" || err != nil {
		providerLinkPutKey = h.ProviderKey
	} else {
		providerLinkPutKey = publicKey
	}

	return Topics{
		LatticeLinkGet:  fmt.Sprintf("wasmbus.rpc.%s.%s.linkdefs.get", h.LatticeRPCPrefix, h.ProviderKey),
		LatticeLinkDel:  fmt.Sprintf("wasmbus.rpc.%s.%s.linkdefs.del", h.LatticeRPCPrefix, h.ProviderKey),
		LatticeLinkPut:  fmt.Sprintf("wasmbus.rpc.%s.%s.linkdefs.put", h.LatticeRPCPrefix, providerLinkPutKey),
		LatticeHealth:   fmt.Sprintf("wasmbus.rpc.%s.%s.health", h.LatticeRPCPrefix, h.ProviderKey),
		LatticeShutdown: fmt.Sprintf("wasmbus.rpc.%s.%s.default.shutdown", h.LatticeRPCPrefix, h.ProviderKey),
	}
}

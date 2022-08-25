package types

var (
	// VerifiableCredentialKey prefix for each key to a DidDocument
	VerifiableCredentialKey = []byte{0x63}
	// VcMetadataKey prefix for each key of a VcMetadata
	VcMetadataKey = []byte{0x64}
)

const (
	// ModuleName defines the module name
	ModuleName = "vc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// VcChainPrefix defines the vc prefix for this chain
	VcChainPrefix = "vc:cosmos:net:"

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "vc_mem_capability"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

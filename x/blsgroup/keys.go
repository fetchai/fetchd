package blsgroup

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "blsgroup"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
)

// KVStore keys
var (
	GroupPrefixKey = []byte{0x00}
)

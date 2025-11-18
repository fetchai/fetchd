package ica_migration

import (
	"fmt"
	"strings"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v10/modules/core/24-host"
)

const CapabilityModuleName = "capability"

var (
	// KeyPrefixIndexCapability defines a key prefix that stores index to capability
	// owners mappings.
	KeyPrefixIndexCapability = []byte("capability_index")

	// KeyMemInitialized defines the key that stores the initialized flag in the memory store
	KeyMemInitialized = []byte("mem_initialized")
)

type ICAMigrator struct {
	ICAstore           storetypes.KVStore
	IBCstore           storetypes.KVStore
	capabilitymemstore storetypes.KVStore
	capabilitystorekey storetypes.StoreKey
	cdc                codec.Codec
	capMap             map[uint64]*Capability
}

func NewICAMigrator(ctx sdk.Context, ICAstore storetypes.KVStore, IBCstore storetypes.KVStore, capabilitymemstore storetypes.KVStore, capabilitystorekey storetypes.StoreKey) ICAMigrator {
	ir := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)
	migrator := ICAMigrator{ICAstore: ICAstore, cdc: cdc, capabilitymemstore: capabilitymemstore, capabilitystorekey: capabilitystorekey, capMap: make(map[uint64]*Capability), IBCstore: IBCstore}
	migrator.InitMemStore(ctx)
	return migrator
}

// AssertChannelCapabilityMigrations checks that all channel capabilities generated using the interchain accounts controller port prefix
// are owned by the controller submodule and ibc.
func (m ICAMigrator) AssertChannelCapabilityMigrations(ctx sdk.Context) error {
	filteredChannels := m.GetAllChannelsWithPortPrefix(ctx, icatypes.ControllerPortPrefix)
	for _, ch := range filteredChannels {
		name := ChannelCapabilityPath(ch.PortId, ch.ChannelId)
		capability, found := m.GetCapability(ctx, name)
		if !found {
			ctx.Logger().Error(fmt.Sprintf("failed to find capability: %s", name))
			return fmt.Errorf("failed to find capability: %s", name)
		}

		isAuthenticated := m.AuthenticateCapability(ctx, capability, name)
		if !isAuthenticated {
			ctx.Logger().Error(fmt.Sprintf("expected capability owner: %s", icacontrollertypes.SubModuleName))
			return fmt.Errorf("expected capability owner: %s", icacontrollertypes.SubModuleName)
		}

		m.SetMiddlewareEnabled(ctx, ch.PortId, ch.ConnectionHops[0])
		ctx.Logger().Info("successfully migrated channel capability", "name", name)
	}
	return nil
}

// GetAllChannelsWithPortPrefix returns all channels with the specified port prefix. If an empty prefix is provided
// all channels will be returned.
func (m ICAMigrator) GetAllChannelsWithPortPrefix(ctx sdk.Context, portPrefix string) []types.IdentifiedChannel {
	if strings.TrimSpace(portPrefix) == "" {
		return m.GetAllChannels(ctx)
	}

	iterator := storetypes.KVStorePrefixIterator(m.IBCstore, types.FilteredPortPrefix(portPrefix))
	defer sdk.LogDeferred(ctx.Logger(), func() error { return iterator.Close() })

	var filteredChannels []types.IdentifiedChannel
	for ; iterator.Valid(); iterator.Next() {
		var channel types.Channel
		m.cdc.MustUnmarshal(iterator.Value(), &channel)

		portID, channelID := host.MustParseChannelPath(string(iterator.Key()))
		identifiedChannel := types.NewIdentifiedChannel(portID, channelID, channel)
		filteredChannels = append(filteredChannels, identifiedChannel)
	}
	return filteredChannels
}

func (m ICAMigrator) GetAllChannels(ctx sdk.Context) (channels []types.IdentifiedChannel) {
	m.IterateChannels(ctx, func(channel types.IdentifiedChannel) bool {
		channels = append(channels, channel)
		return false
	})
	return channels
}

// IterateChannels provides an iterator over all Channel objects. For each
// Channel, cb will be called. If the cb returns true, the iterator will close
// and stop.
func (m ICAMigrator) IterateChannels(ctx sdk.Context, cb func(types.IdentifiedChannel) bool) {

	iterator := storetypes.KVStorePrefixIterator(m.IBCstore, []byte(host.KeyChannelEndPrefix))

	defer sdk.LogDeferred(ctx.Logger(), func() error { return iterator.Close() })
	for ; iterator.Valid(); iterator.Next() {
		var channel types.Channel
		m.cdc.MustUnmarshal(iterator.Value(), &channel)

		portID, channelID := host.MustParseChannelPath(string(iterator.Key()))
		identifiedChannel := types.NewIdentifiedChannel(portID, channelID, channel)
		if cb(identifiedChannel) {
			break
		}
	}
}

const (
	KeyChannelCapabilityPrefix = "capabilities"
	KeyChannelPrefix           = "channels"
	KeyPortPrefix              = "ports"
)

func channelPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s/%s/%s", KeyPortPrefix, portID, KeyChannelPrefix, channelID)
}

// ChannelCapabilityPath defines the path under which capability keys associated
// with a channel are stored
func ChannelCapabilityPath(portID, channelID string) string {
	return fmt.Sprintf("%s/%s", KeyChannelCapabilityPrefix, channelPath(portID, channelID))
}

// GetCapability allows a module to fetch a capability which it previously claimed
// by name. The module is not allowed to retrieve capabilities which it does not
// own.
func (m ICAMigrator) GetCapability(ctx sdk.Context, name string) (*Capability, bool) {
	if strings.TrimSpace(name) == "" {
		return nil, false
	}

	key := RevCapabilityKey(CapabilityModuleName, name)
	indexBytes := m.capabilitymemstore.Get(key)
	index := sdk.BigEndianToUint64(indexBytes)

	if len(indexBytes) == 0 {
		// If a tx failed and NewCapability got reverted, it is possible
		// to still have the capability in the go map since changes to
		// go map do not automatically get reverted on tx failure,
		// so we delete here to remove unnecessary values in map
		// TODO: Delete index correctly from capMap by storing some reverse lookup
		// in-memory map. Issue: https://github.com/cosmos/cosmos-sdk/issues/7805

		return nil, false
	}

	capability := m.capMap[index]
	if capability == nil {
		panic("capability found in memstore is missing from map")
	}

	return capability, true
}

// IsInitialized returns true if the keeper is properly initialized, and false otherwise.
func (m ICAMigrator) IsInitialized(ctx sdk.Context) bool {
	return m.capabilitymemstore.Has(KeyMemInitialized)
}

// InitMemStore will assure that the module store is a memory store (it will panic if it's not)
// and willl initialize it. The function is safe to be called multiple times.
// InitMemStore must be called every time the app starts before the keeper is used (so
// `BeginBlock` or `InitChain` - whichever is first). We need access to the store so we
// can't initialize it in a constructor.
func (m ICAMigrator) InitMemStore(ctx sdk.Context) {
	memStoreType := m.capabilitymemstore.GetStoreType()

	if memStoreType != storetypes.StoreTypeMemory {
		panic(fmt.Errorf("invalid memory store type; got %s, expected: %s", memStoreType, storetypes.StoreTypeMemory))
	}

	// create context with no block gas meter to ensure we do not consume gas during local initialization logic.
	noGasCtx := ctx.WithBlockGasMeter(storetypes.NewInfiniteGasMeter()).WithGasMeter(storetypes.NewInfiniteGasMeter())

	// check if memory store has not been initialized yet by checking if initialized flag is nil.
	if !m.IsInitialized(noGasCtx) {
		prefixStore := prefix.NewStore(noGasCtx.KVStore(m.capabilitystorekey), KeyPrefixIndexCapability)
		iterator := storetypes.KVStorePrefixIterator(prefixStore, nil)

		// initialize the in-memory store for all persisted capabilities
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			index := IndexFromKey(iterator.Key())

			var capOwners CapabilityOwners

			m.cdc.MustUnmarshal(iterator.Value(), &capOwners)
			m.InitializeCapability(noGasCtx, index, capOwners)
		}

		// set the initialized flag so we don't rerun initialization logic
		m.capabilitymemstore.Set(KeyMemInitialized, []byte{1})
	}
}

// InitializeCapability takes in an index and an owners array. It creates the capability in memory
// and sets the fwd and reverse keys for each owner in the memstore.
// It is used during initialization from genesis.
func (m ICAMigrator) InitializeCapability(ctx sdk.Context, index uint64, owners CapabilityOwners) {

	capability := NewCapability(index)
	for _, owner := range owners.Owners {
		// Set the forward mapping between the module and capability tuple and the
		// capability name in the memKVStore
		m.capabilitymemstore.Set(FwdCapabilityKey(owner.Module, capability), []byte(owner.Name))

		// Set the reverse mapping between the module and capability name and the
		// index in the in-memory store. Since marshalling and unmarshalling into a store
		// will change memory address of capability, we simply store index as value here
		// and retrieve the in-memory pointer to the capability from our map
		m.capabilitymemstore.Set(RevCapabilityKey(owner.Module, owner.Name), sdk.Uint64ToBigEndian(index))

		// Set the mapping from index from index to in-memory capability in the go map
		m.capMap[index] = capability
	}
}

// NewCapability returns a reference to a new Capability to be used as an
// actual capability.
func NewCapability(index uint64) *Capability {
	return &Capability{Index: index}
}

// RevCapabilityKey returns a reverse lookup key for a given module and capability
// name.
func RevCapabilityKey(module, name string) []byte {
	return []byte(fmt.Sprintf("%s/rev/%s", module, name))
}

// FwdCapabilityKey returns a forward lookup key for a given module and capability
// reference.
func FwdCapabilityKey(module string, cap *Capability) []byte {
	return []byte(fmt.Sprintf("%s/fwd/%#016p", module, cap))
}

// IndexToKey returns bytes to be used as a key for a given capability index.
func IndexToKey(index uint64) []byte {
	return sdk.Uint64ToBigEndian(index)
}

// IndexFromKey returns an index from a call to IndexToKey for a given capability
// index.
func IndexFromKey(key []byte) uint64 {
	return sdk.BigEndianToUint64(key)
}

// AuthenticateCapability attempts to authenticate a given capability and name
// from a caller. It allows for a caller to check that a capability does in fact
// correspond to a particular name. The scoped keeper will lookup the capability
// from the internal in-memory store and check against the provided name. It returns
// true upon success and false upon failure.
//
// Note, the capability's forward mapping is indexed by a string which should
// contain its unique memory reference.
func (m ICAMigrator) AuthenticateCapability(ctx sdk.Context, cap *Capability, name string) bool {
	if strings.TrimSpace(name) == "" || cap == nil {
		return false
	}
	return m.GetCapabilityName(ctx, cap) == name
}

// GetCapabilityName allows a module to retrieve the name under which it stored a given
// capability given the capability
func (m ICAMigrator) GetCapabilityName(ctx sdk.Context, cap *Capability) string {
	if cap == nil {
		return ""
	}

	return string(m.capabilitymemstore.Get(FwdCapabilityKey(CapabilityModuleName, cap)))
}

// SetMiddlewareEnabled stores a flag to indicate that the underlying application callbacks should be enabled for the given port and connection identifier pair
func (m ICAMigrator) SetMiddlewareEnabled(ctx sdk.Context, portID, connectionID string) {
	m.ICAstore.Set(icatypes.KeyIsMiddlewareEnabled(portID, connectionID), icatypes.MiddlewareEnabled)
}

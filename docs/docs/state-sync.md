
# State-sync

State sync is a feature which allows you to quickly bootstrap a new node by allowing it to pull a *state snapshot* taken by other nodes.


## Configuring the new node

In order to instruct the node to sync itself using state sync, it need some configuration in the `~/.fetchd/config/config.toml` file.
Open this file in an editor and lookup the `statesync` section. By default, it should looks like this:

```yaml
#######################################################
###         State Sync Configuration Options        ###
#######################################################
[statesync]
# State sync rapidly bootstraps a new node by discovering, fetching, and restoring a state machine
# snapshot from peers instead of fetching and replaying historical blocks. Requires some peers in
# the network to take and serve state machine snapshots. State sync is not attempted if the node
# has any local state (LastBlockHeight > 0). The node will have a truncated block history,
# starting from the height of the snapshot.
enable = false

# RPC servers (comma-separated) for light client verification of the synced state machine and
# retrieval of state data for node bootstrapping. Also needs a trusted height and corresponding
# header hash obtained from a trusted source, and a period during which validators can be trusted.
#
# For Cosmos SDK-based chains, trust_period should usually be about 2/3 of the unbonding time (~2
# weeks) during which they can be financially punished (slashed) for misbehavior.
rpc_servers = ""
trust_height = 0
trust_hash = ""
trust_period = "168h0m0s"

# Time to spend discovering snapshots before initiating a restore.
discovery_time = "15s"

# Temporary directory for state sync snapshot chunks, defaults to the OS tempdir (typically /tmp).
# Will create a new, randomly named directory within, and remove it when done.
temp_dir = ""

# The timeout duration before re-requesting a chunk, possibly from a different
# peer (default: 1 minute).
chunk_request_timeout = "10s"

# The number of concurrent chunk fetchers to run (default: 1).
chunk_fetchers = "4"
```

A few changes are needed:

- First, set `enable = true` to activate the state sync engine.
- Then, **at least 2** rpc servers must be provided. A good place to find some is the [cosmos chain registry](https://github.com/cosmos/chain-registry/blob/master/fetchhub/chain.json#L62). Servers must be comma separated without space (ie: `rpc_servers = "https://rpc-fetchhub.fetch.ai:443,https://fetchapi.terminet.io"`)
- And last, a *recent* `trust_height` and `trust_hash` to start from are needed. Recent means it must be contained in the `trust_period` (~1 week old by default). These can be obtained from a RPC server **you trust to provide you correct data** (and the 2nd RPC server from `rpc_servers` will be charged of confirming that the data are correct). To retrieve the correct value for a fetch.ai RPC server, and the current network height, use:

```bash
curl https://rpc-fetchhub.fetch.ai:443/block | jq -r '{"trusted_hash": .result.block_id.hash, "trusted_height": .result.block.header.height}'
{
  "trusted_hash": "46868B76E6C814C35B2D109FCA177EBB70689AE3D46C65E4D75DE5363A86FF97",
  "trusted_height": "7041920"
}
```

and set these values in the config file.

Once this is set, make sure you have the correct genesis by downloading it from the RPC node:

```bash
wget https://storage.googleapis.com/fetch-ai-mainnet-v2-genesis/genesis-fetchhub4.json
mv genesis-fetchhub4.json ~/.fetchd/config/genesis.json
```

and start the node using the seeds from the chain-registry:

```bash
fetchd start --p2p.seeds=$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/fetchhub/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')
```

After the node initialized, it will start searching for available snapshots, and it should print log messages similar to:

```
8:22AM INF Discovered new snapshot format=1 hash="� ݫ/��\r�F#C(pD�<��\x066��\x1f��\x1f<i�ݝ" height=2000 module=statesync
8:22AM INF Discovered new snapshot format=1 hash="F�=\x05�Gh�{�|�����,�Q'�=]\x1a�$�b�ֿQ" height=1900 module=statesync
```

## Configure an existing node to provide snapshots

In order to provide new nodes snapshots they can start from, existing nodes need to be configure to create these snapshots.
This can be configured from the `~/.fetchd/config/app.toml` file, in the `state-sync` section.

```
###############################################################################
###                        State Sync Configuration                         ###
###############################################################################

# State sync snapshots allow other nodes to rapidly join the network without replaying historical
# blocks, instead downloading and applying a snapshot of the application state at a given height.
[state-sync]

# snapshot-interval specifies the block interval at which local state sync snapshots are
# taken (0 to disable). Must be a multiple of pruning-keep-every.
snapshot-interval = 0

# snapshot-keep-recent specifies the number of recent snapshots to keep and serve (0 to keep all).
snapshot-keep-recent = 2
```

Here `snapshot-interval` must be set to a number of blocks between each snapshot creation and it depends on your node prunning settings. The number of snapshots to keep can be set with `snapshot-keep-recent`.
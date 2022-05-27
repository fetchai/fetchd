# Chain State Snapshots

As blockchains get longer, the process of syncing from the genesis block begins to take many hours, or even days to complete.
In circumstances where a faster sync is required, various snapshots of the fetchd chain state data are available for download, to more quickly bootstrap a node.

Snapshots are available for both mainnet and the most recent testnet.  The URLs can be obtained from the [network page](../networks/).  
We aim to update snapshots on a daily basis.

The example below uses the pruned mainnet snapshot, but can be adapted as required for full or archive nodes.

## Using a snapshot

### Stop your node

If you are already running fetchd, it is important that you stop it before proceeding.  Instructions for this are highly installation dependent and beyond the scope of this document, but could be as simple as a Ctrl-C.  
If you have not already initialised your node, follow the instructions for [joining a testnet](../joining-a-testnet/) (modifying for mainnet as appropriate), then return to this page before starting fetchd.

### Reset your node

WARNING: This will irreversibly erase your node's state database.  Ensure you take whatever backups you deem appropriate before proceeding.

If using fetchd <= 0.10.3
`fetchd unsafe-reset-all`

If using fetchd >= 0.10.4
`fetchd tendermint reset-state`

### Download and install the snapshot

Many options here!  The example below assumes a bash-like environment, uses a single connection for downloading, confirms the md5sum of the downloaded data against that of the original, and does not land the original compressed data to disk.  This is a good starting point, but depending on your local environment you may wish to make adaptations that eg sacrifice disk space and extra md5sum complexity for the benefit of parallel downloads with aria2.  Entirely up to you... let us know how you get on!

```bash
# (optional) show the timestamp of the latest available snapshot
echo "Latest available snapshot timestamp : $(curl -s -I  https://storage.googleapis.com/fetch-ai-mainnet-snapshots/fetchhub-4-pruned.tgz | grep last-modified | cut -f3- -d' ')"

# download, decompress and extract state database
curl -v https://storage.googleapis.com/fetch-ai-mainnet-snapshots/fetchhub-4-pruned.tgz -o- 2>headers.out | tee >(md5sum > md5sum.out) | gunzip -c | tar -xvf - --directory=~/.fetchd

# (optional, but recommended) compare source md5 checksum provided in the headers by google, with the one calculated locally
[[ $(grep 'x-goog-hash: md5' headers.out | sed -z 's/^.*md5=\(.*\)/\1/g' | tr -d '\r' | base64 -d | od -An -vtx1 | tr -d ' \n') == $(awk '{ print $1 }' md5sum.out) ]] && echo "OK - md5sum match" || echo "ERROR - md5sum MISMATCH"

# (optional) show the creation date of the downloaded snapshot
echo "Downloaded snapshot timestamp: $(grep last-modified headers.out | cut -f3- -d' ')"
```

### Restart your node

Again, this entirely depends on your local installation, but a simple example for mainnet might be...

```bash
fetchd start --p2p.seeds 17693da418c15c95d629994a320e2c4f51a8069b@connect-fetchhub.fetch.ai:36456,a575c681c2861fe945f77cb3aba0357da294f1f2@connect-fetchhub.fetch.ai:36457,d7cda986c9f59ab9e05058a803c3d0300d15d8da@connect-fetchhub.fetch.ai:36458`.
```

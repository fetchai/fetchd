#!/bin/bash

set -e

# check the expected environment variable setup
if [ -z "${MNEMONIC}" ]; then echo "Missing MNEMONIC environment variable"; exit 1; fi
if [ -z "${PASSPHRASE}" ]; then echo "Missing PASSPHRASE environment variable"; exit 1; fi
if [ -z "${CHAINID}" ]; then echo "Missing CHAINID environment variable"; exit 1; fi
if [ -z "${MONIKER}" ]; then echo "Missing MONIKER environment variable"; exit 1; fi

# switch to the home holder
cd ~

echo -e "${MNEMONIC}\n${PASSPHRASE}\n${PASSPHRASE}\n" > mnemonic-setup.txt
echo -e "${PASSPHRASE}\n" > passphrase.txt
echo -e "${PASSPHRASE}\n${PASSPHRASE}\n${PASSPHRASE}\n${PASSPHRASE}\n" > passphrase4.txt

fetchd config chain-id "${CHAINID}"
fetchd config keyring-backend test

# setup the node with a default genesis
if [ ! -f "/root/.fetchd/config/genesis.json" ]; then
	echo 'Generating genesis file...'
    fetchd init "${MONIKER}" --chain-id "${CHAINID}" > /dev/null 2>&1
fi

# create the key if needed
fetchd keys show "${MONIKER}" < passphrase.txt > /dev/null 2>&1 || fetchd keys add "${MONIKER}" --recover < mnemonic-setup.txt > /dev/null 2>&1

# get the address
node_address=$(fetchd keys show ${MONIKER} -a < passphrase.txt)

# check to see if the final genesis configuration has been made
if [ ! -f /setup/genesis.json ]; then

	# create the validator file for the setup script
	if [ ! -f "/setup/${node_address}.validator" ]; then
			touch "/setup/${node_address}.validator"
	fi

	# publish node addresses
	if [ ! -f /setup/${node_address}.networkaddr ]; then
		node_id=$(fetchd tendermint show-node-id)
		echo "${node_id}@${MONIKER}:26656" > /setup/${node_address}.networkaddr
	fi

	# wait for the genesis file to be created
	while [ ! -f "/setup/genesis.intermediate.json" ]
	do
		echo "Node ${node_address} waiting for intermediate genesis configuration..."
		sleep 1
	done

	# copy the generated genesis file
	cp /setup/genesis.intermediate.json /root/.fetchd/config/genesis.json

	# generate the tx
	if [ ! -f "/setup/gentx-${node_address}.json" ]; then
		fetchd gentx ${MONIKER} 1000000000000000000atestfet --chain-id "${CHAINID}" --output-document /setup/gentx-${node_address}.json < passphrase4.txt
	fi
	
	# wait for the genesis file to be created
	while [ ! -f "/setup/genesis.json" ]
	do
		echo "Node ${node_address} waiting for final genesis configuration..."
		sleep 1
	done
fi

# copy the generated genesis file
cp /setup/genesis.json /root/.fetchd/config/genesis.json

# clean up all the temporary files
rm -f mnemonic-setup.txt passphrase.txt passphrase4.txt

# build up the arguments
args="--p2p.laddr tcp://0.0.0.0:26656 --rpc.laddr tcp://0.0.0.0:26657"

# calculate the persistent peers for the network
persistent_peers=$(ls -1 /setup/*.networkaddr | grep -v ${node_address} | xargs cat | awk -vORS=, '{ print $1 }' | sed 's/,$/\n/')
args="${args} --p2p.persistent_peers=${persistent_peers}"

# debug
echo "Moniker.....: ${MONIKER}"
echo "Chain ID....: ${CHAINID}"
echo "Node Address: ${node_address}"
echo "Args........: ${args}"

# run the node
exec fetchd start ${args} $@

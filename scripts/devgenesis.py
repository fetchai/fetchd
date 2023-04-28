#!/usr/bin/env python3

import argparse
import json
import os
import re
import sys
from pathlib import Path
from typing import Dict, Any, List
import subprocess
import bech32


STAKING_DENOM = "afet"
FUND_BALANCE = 10 ** 23

bonded_pool_address = "fetch1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3xxqtmq"
not_bonded_pool_address = "fetch1tygms3xhhs3yv487phx3dw4a95jn7t7ljxu6d5"
VOTING_PERIOD = "60s"





def _path(text: str) -> str:
    return os.path.abspath(text)


def _convert_to_valoper(address):
    hrp, data = bech32.bech32_decode(address)
    if hrp != 'fetch':
        print('Invalid address, expected normal fetch address')
        sys.exit(1)

    return bech32.bech32_encode('fetchvaloper', data)


def _from_coin_list(coins: List[Any]) -> Dict[str, int]:
    balances = {}
    for coin in coins:
        balances[str(coin['denom'])] = int(coin['amount'])
    return balances


def _to_coin_list(balances: Dict[str, int]) -> List[Any]:
    coins = []
    for denom in sorted(balances.keys()):
        amount = balances[denom]
        assert amount >= 0

        if amount == 0:
            continue

        coins.append({
            'denom': str(denom),
            'amount': str(amount)
        })
    return coins

def parse_commandline():
    parser = argparse.ArgumentParser()
    parser.add_argument('genesis_export', type=_path, help='The path to the genesis export')
    parser.add_argument('home_path', type=_path, help='The path to the local node data i.e. ~/.fetchd')
    parser.add_argument('validator_key_name', help='The name of the local key to use for the validator')
    return parser.parse_args()



def usage():
    script_name = os.path.basename(sys.argv[0])
    print(f"Usage: {script_name} path/to/exported/genesis.json [node_home_dir]")
    print()
    print("This script updates an exported genesis from a running chain")
    print("to be used to run on a single validator local node.")
    print("It will take the first validator and jail all the others")
    print("and replace the validator pubkey and the nodekey with the one")
    print("found in the node_home_dir folder.")
    print()
    print("If unspecified, node_home_dir default to the ~/.fetchd/ folder.")
    print(
        "This folder must exist and contain the files created by the 'fetchd init' command."
    )
    print()
    print(
        "The updated genesis will be written under node_home_dir/config/genesis.json, allowing"
    )
    print("the local chain to be started with:")
    print()
    print(
        "fetchd --home <node_home_dir> unsafe-reset-all && fetchd --home <node_home_dir> start"
    )
    print()
    sys.exit()

def main():
    if len(sys.argv) < 4:
        usage()

    args = parse_commandline()

    print('    Genesis Export:', args.genesis_export)
    print('  Fetchd Home Path:', args.home_path)
    print('Validator Key Name:', args.validator_key_name)

    # load up the local validator key
    local_validator_key_path = os.path.join(args.home_path, 'config', 'priv_validator_key.json')
    with open(local_validator_key_path, 'r') as input_file:
        local_validator_key = json.load(input_file)

    # extract the tendermint addresses
    cmd = ['fetchd', '--home', args.home_path, 'tendermint', 'show-address']
    validator_address = subprocess.check_output(cmd).decode().strip()
    validator_pubkey = local_validator_key['pub_key']['value']
    validator_hexaddr = local_validator_key['address']

    print(f"- new address: {validator_hexaddr}")
    print(f"- new pubkey: {validator_pubkey}")
    print(f"- new tendermint address: {validator_address}")

    # extract the address for the local validator key
    cmd = ['fetchd', '--home', args.home_path, 'keys', 'show', args.validator_key_name, '--output', 'json']
    key_data = json.loads(subprocess.check_output(cmd).decode())

    if key_data['type'] != 'local':
        print('Unable to use non-local key type')
        sys.exit(1)

    # extract the local address and convert into a valid validator operator address
    validator_operator_base_address = key_data['address']
    validator_operator_address = _convert_to_valoper(validator_operator_base_address)
    print(f'       {validator_operator_base_address}')
    print(validator_operator_address)



    # load the genesis up
    print('reading genesis export...')
    with open(args.genesis_export, 'r') as export_file:
        genesis = json.load(export_file)
    print('reading genesis export...complete')


    val_infos = genesis["app_state"]["staking"]["validators"][0]
    if not val_infos:
        print("Genesis file does not contain any validators")
        sys.exit(1)

    target_validator_address = val_infos["operator_address"]
    target_validator_public_key = val_infos["consensus_pubkey"]["key"]
    val_addr = [
        val
        for val in genesis["validators"]
        if val["pub_key"]["value"] == target_validator_public_key
    ][0]["address"]

    # Replace selected validator by current node one
    print(f"Replacing validator {target_validator_address}...")
    genesis = json.loads(re.sub(val_addr, validator_hexaddr, json.dumps(genesis)))
    genesis = json.loads(re.sub(target_validator_public_key, validator_pubkey, json.dumps(genesis)))

    # Set .app_state.slashing.signing_infos to contain only our validator signing infos
    print("Updating signing infos...")
    genesis["app_state"]["slashing"]["signing_infos"] = [
        {
            "address": validator_address,
            "validator_signing_info": {
                "address": validator_address,
                "index_offset": "0",
                "jailed_until": "1970-01-01T00:00:00Z",
                "missed_blocks_counter": "0",
                "start_height": "0",
                "tombstoned": False,
            },
        }
    ]

    # Update bonded and not bonded pool values to make invariant checks happy
    print("Updating bonded and not bonded token pool values...")

    val_tokens = int(val_infos["tokens"])
    val_power = int(val_tokens / (10**18))

    bonded_tokens = next(
        int(amount["amount"])
        for balance in genesis["app_state"]["bank"]["balances"]
        if balance["address"] == bonded_pool_address
        for amount in balance["coins"]
        if amount["denom"] == STAKING_DENOM
    )

    not_bonded_tokens = next(
        int(amount["amount"])
        for balance in genesis["app_state"]["bank"]["balances"]
        if balance["address"] == not_bonded_pool_address
        for amount in balance["coins"]
        if amount["denom"] == STAKING_DENOM
    )

    not_bonded_tokens = not_bonded_tokens + bonded_tokens - val_tokens

    for balance in genesis["app_state"]["bank"]["balances"]:
        if balance["address"] == bonded_pool_address:
            for amount in balance["coins"]:
                if amount["denom"] == STAKING_DENOM:
                    amount["amount"] = str(val_tokens)

        if balance["address"] == not_bonded_pool_address:
            for amount in balance["coins"]:
                if amount["denom"] == STAKING_DENOM:
                    amount["amount"] = str(not_bonded_tokens)

    # Create new account and fund it
    print("Creating new account and funding it...")
    # Add new balance to bank
    new_balance = {
        "address": validator_operator_base_address,
        "coins": [{"amount": str(FUND_BALANCE), "denom": STAKING_DENOM}],
    }
    genesis["app_state"]["bank"]["balances"].append(new_balance)

    # Add new account to auth
    last_account_number = int(
        genesis["app_state"]["auth"]["accounts"][-1]["account_number"]
    )
    new_account = {
        "@type": "/cosmos.auth.v1beta1.BaseAccount",
        "account_number": str(last_account_number + 1),
        "address": validator_operator_base_address,
        "pub_key": None,
        "sequence": "0",
    }
    genesis["app_state"]["auth"]["accounts"].append(new_account)

    # Update total supply
    for supply in genesis["app_state"]["bank"]["supply"]:
        if supply["denom"] == STAKING_DENOM:
            supply["amount"] = str(int(supply["amount"]) + FUND_BALANCE)

    # Remove all .validators but the one we work with
    print("Removing other validators from initchain...")
    genesis["validators"] = [
        val for val in genesis["validators"] if val["address"] == validator_hexaddr
    ]

    # Set .app_state.staking.last_validator_powers to contain only our validator
    print("Updating last voting power...")
    genesis["app_state"]["staking"]["last_validator_powers"] = [
        {"address": target_validator_address, "power": str(val_power)}
    ]

    # Jail everyone but our validator
    print("Jail other validators...")
    for validator in genesis["app_state"]["staking"]["validators"]:
        if validator["operator_address"] != target_validator_address:
            validator["status"] = "BOND_STATUS_UNBONDING"
            validator["jailed"] = True

    if "max_wasm_code_size" in genesis["app_state"]["wasm"]["params"]:
        print("Removing max_wasm_code_size...")
        del genesis["app_state"]["wasm"]["params"]["max_wasm_code_size"]

    print(f"Setting voting period to {VOTING_PERIOD}...")
    genesis["app_state"]["gov"]["voting_params"]["voting_period"] = VOTING_PERIOD

    print("Writing new genesis file...")
    with open(f"{args.home_path}/config/genesis.json", "w") as f:
        json.dump(genesis, f, indent=2)

    print(f"Done! Wrote new genesis at {args.home_path}/config/genesis.json")
    print("You can now start the chain:")
    print()
    print(
        f"fetchd --home {args.home_path} unsafe-reset-all && fetchd --home {args.home_path} start"
    )
    print()


if __name__ == "__main__":
    main()

#!/usr/bin/env python3

import argparse
import json
import os
import re
import subprocess
import sys
from pathlib import Path
from typing import Any, Dict, List

import bech32

DEFAULT_STAKING_DENOM = "afet"
DEFAULT_HOME_PATH = os.path.expanduser("~") + "/.fetchd"
DEFAULT_VALIDATOR_KEY_NAME = "validator"
FUND_BALANCE = 10**23
DEFAULT_VOTING_PERIOD = "60s"
NEW_DEFAULTS_CHAIN_ID = "test-1"


def parse_commandline():
    description = """This script updates an exported genesis from a running chain
to be used to run on a single validator local node.
It will take the first validator and jail all the others
and replace the validator pubkey and the nodekey with the one 
found in the node_home_dir folder

if unspecified, node_home_dir default to the ~/.fetchd/ folder. 
this folder must exists and contains the files created by the "fetchd init" command.

The updated genesis will be written under node_home_dir/config/genesis.json, allowing
the local chain to be started with:
    """

    parser = argparse.ArgumentParser(description=description)
    parser.add_argument(
        "genesis_export", type=_path, help="The path to the genesis export"
    )
    parser.add_argument(
        "--home_path",
        help="The path to the local node data i.e. ~/.fetchd",
        default=DEFAULT_HOME_PATH,
    )
    parser.add_argument(
        "--validator_key_name",
        help="The name of the local key to use for the validator",
        default=DEFAULT_VALIDATOR_KEY_NAME,
    )
    parser.add_argument(
        "--staking_denom", help="The staking denom", default=DEFAULT_STAKING_DENOM
    )
    parser.add_argument(
        "--chain_id", help="New chain ID to be set", default=NEW_DEFAULTS_CHAIN_ID
    )
    parser.add_argument(
        "--voting_period",
        help="The new voting period to be set",
        default=DEFAULT_VOTING_PERIOD,
    )

    return parser.parse_args()


def _path(text: str) -> str:
    return os.path.abspath(text)


def _convert_to_valoper(address):
    hrp, data = bech32.bech32_decode(address)
    if hrp != "fetch":
        print("Invalid address, expected normal fetch address")
        sys.exit(1)

    return bech32.bech32_encode("fetchvaloper", data)


def _ensure_account(genesis, address):
    for account in genesis["app_state"]["auth"]["accounts"]:
        if "address" in account and account["address"] == address:
            return

    # Add new account to auth
    last_account_number = int(
        genesis["app_state"]["auth"]["accounts"][-1]["account_number"]
    )
    new_account = {
        "@type": "/cosmos.auth.v1beta1.BaseAccount",
        "account_number": str(last_account_number + 1),
        "address": address,
        "pub_key": None,
        "sequence": "0",
    }
    genesis["app_state"]["auth"]["accounts"].append(new_account)
    return genesis


def _set_balance(genesis, address, new_balance, denom):
    account_found = False
    for balance in genesis["app_state"]["bank"]["balances"]:
        if balance["address"] == address:
            for amount in balance["coins"]:
                if amount["denom"] == denom:
                    amount["amount"] = str(new_balance)
                    account_found = True

    if not account_found:
        new_balance_entry = {
            "address": address,
            "coins": [{"amount": str(new_balance), "denom": denom}],
        }
        genesis["app_state"]["bank"]["balances"].append(new_balance_entry)
    return genesis


def _get_balance(genesis, address, denom):
    amount = 0
    for balance in genesis["app_state"]["bank"]["balances"]:
        if balance["address"] == address:
            for amount in balance["coins"]:
                if amount["denom"] == denom:
                    amount = int(amount["amount"])
                    break
            if amount is not 0:
                break
    return amount


def main():
    args = parse_commandline()

    print("    Genesis Export:", args.genesis_export)
    print("  Fetchd Home Path:", args.home_path)
    print("Validator Key Name:", args.validator_key_name)

    # load up the local validator key
    local_validator_key_path = os.path.join(
        args.home_path, "config", "priv_validator_key.json"
    )
    with open(local_validator_key_path, "r") as input_file:
        local_validator_key = json.load(input_file)

    # extract the tendermint addresses
    cmd = ["fetchd", "--home", args.home_path, "tendermint", "show-address"]
    validator_address = subprocess.check_output(cmd).decode().strip()
    validator_pubkey = local_validator_key["pub_key"]["value"]
    validator_hexaddr = local_validator_key["address"]

    # extract the address for the local validator key
    cmd = [
        "fetchd",
        "--home",
        args.home_path,
        "keys",
        "show",
        args.validator_key_name,
        "--output",
        "json",
    ]
    key_data = json.loads(subprocess.check_output(cmd).decode())

    if key_data["type"] != "local":
        print("Unable to use non-local key type")
        sys.exit(1)

    # extract the local address and convert into a valid validator operator address
    validator_operator_base_address = key_data["address"]
    validator_operator_address = _convert_to_valoper(validator_operator_base_address)
    print(f"       {validator_operator_base_address}")
    print(validator_operator_address)

    # load the genesis up
    print("reading genesis export...")
    with open(args.genesis_export, "r") as export_file:
        genesis = json.load(export_file)
    print("reading genesis export...complete")

    val_infos = genesis["app_state"]["staking"]["validators"][0]
    if not val_infos:
        print("Genesis file does not contain any validators")
        sys.exit(1)

    target_validator_operator_address = val_infos["operator_address"]
    target_validator_public_key = val_infos["consensus_pubkey"]["key"]
    val_addr = [
        val
        for val in genesis["validators"]
        if val["pub_key"]["value"] == target_validator_public_key
    ][0]["address"]

    # Replace selected validator by current node one
    print(f"Replacing validator {target_validator_operator_address}...")

    genesis_dump = json.dumps(genesis)
    genesis_dump = re.sub(val_addr, validator_hexaddr, genesis_dump)
    genesis_dump = re.sub(target_validator_public_key, validator_pubkey, genesis_dump)
    genesis_dump = re.sub(
        target_validator_operator_address, validator_operator_address, genesis_dump
    )
    genesis = json.loads(genesis_dump)

    # Update the chain id
    print(f"Updating chain id to {args.chain_id}...")
    genesis["chain_id"] = args.chain_id

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

    # Find the bonded and not bonded token pools
    print("Finding bonded and not bonded token pools...")

    bonded_pool_address = None
    not_bonded_pool_address = None
    for account in genesis["app_state"]["auth"]["accounts"]:
        if "name" in account:
            if account["name"] == "bonded_tokens_pool":
                bonded_pool_address = account["base_account"]["address"]
            elif account["name"] == "not_bonded_tokens_pool":
                not_bonded_pool_address = account["base_account"]["address"]

            if bonded_pool_address and not_bonded_pool_address:
                break

    # Update bonded and not bonded pool values to make invariant checks happy
    print("Updating bonded and not bonded token pool values...")

    val_tokens = int(val_infos["tokens"])
    val_power = int(val_tokens / (10**18))

    # Get current bonded and not bonded tokens
    bonded_tokens = _get_balance(genesis, bonded_pool_address, args.staking_denom)
    not_bonded_tokens = _get_balance(
        genesis, not_bonded_pool_address, args.staking_denom
    )

    new_not_bonded_tokens = not_bonded_tokens + bonded_tokens - val_tokens

    # Update bonded pool and not bonded pool balances
    _set_balance(genesis, bonded_pool_address, val_tokens, args.staking_denom)
    _set_balance(
        genesis, not_bonded_pool_address, new_not_bonded_tokens, args.staking_denom
    )

    # Create new account and fund it
    print(
        f"Creating new funded account for local validator {validator_operator_base_address}..."
    )

    # Add new balance to bank
    genesis = _set_balance(
        genesis, validator_operator_base_address, FUND_BALANCE, args.staking_denom
    )

    # Add new account to auth if not already there
    genesis = _ensure_account(genesis, validator_operator_base_address)

    # Update total supply
    for supply in genesis["app_state"]["bank"]["supply"]:
        if supply["denom"] == args.staking_denom:
            supply["amount"] = str(int(supply["amount"]) + FUND_BALANCE)

    # Remove all .validators but the one we work with
    print("Removing other validators from initchain...")
    genesis["validators"] = [
        val for val in genesis["validators"] if val["address"] == validator_hexaddr
    ]

    # Set .app_state.staking.last_validator_powers to contain only our validator
    print("Updating last voting power...")
    genesis["app_state"]["staking"]["last_validator_powers"] = [
        {"address": validator_operator_address, "power": str(val_power)}
    ]

    # Jail everyone but our validator
    print("Jail other validators...")
    for validator in genesis["app_state"]["staking"]["validators"]:
        if validator["operator_address"] != validator_operator_address:
            validator["status"] = "BOND_STATUS_UNBONDING"
            validator["jailed"] = True

    if "max_wasm_code_size" in genesis["app_state"]["wasm"]["params"]:
        print("Removing max_wasm_code_size...")
        del genesis["app_state"]["wasm"]["params"]["max_wasm_code_size"]

    # Set voting period
    print(f"Setting voting period to {args.voting_period}...")
    genesis["app_state"]["gov"]["voting_params"]["voting_period"] = args.voting_period

    print("Writing new genesis file...")
    with open(f"{args.home_path}/config/genesis.json", "w") as f:
        json.dump(genesis, f, indent=2)

    print(f"Done! Wrote new genesis at {args.home_path}/config/genesis.json")
    print("You can now start the chain:")
    print()
    print(
        f"fetchd --home {args.home_path} tendermint unsafe-reset-all && fetchd --home {args.home_path} start"
    )
    print()


if __name__ == "__main__":
    main()

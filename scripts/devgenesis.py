#!/usr/bin/env python3

import argparse
import json
import os
import re
import subprocess
import sys
from genesis_helpers import (
    get_balance,
    get_path,
    get_unjailed_validator,
    set_balance,
    ensure_account,
    convert_to_valoper,
    get_validator_info,
)


DEFAULT_STAKING_DENOM = "afet"
DEFAULT_HOME_PATH = os.path.expanduser("~") + "/.fetchd"
DEFAULT_VALIDATOR_KEY_NAME = "validator"
FUND_BALANCE = 10**23
DEFAULT_VOTING_PERIOD = "60s"


def jail_validators(genesis, validator_operator_address: None):
    for validator in genesis["app_state"]["staking"]["validators"]:
        if validator["operator_address"] != validator_operator_address:
            validator["status"] = "BOND_STATUS_UNBONDING"
            validator["jailed"] = True


def remove_max_wasm_code_size(genesis):
    if "max_wasm_code_size" in genesis["app_state"]["wasm"]["params"]:
        print("Removing max_wasm_code_size...")
        del genesis["app_state"]["wasm"]["params"]["max_wasm_code_size"]


def set_voting_period(genesis, voting_period):
    print(f"Setting voting period to {voting_period}...")
    genesis["app_state"]["gov"]["voting_params"]["voting_period"] = voting_period


def update_chain_id(genesis, chain_id):
    print(f"Updating chain id to {chain_id}...")
    genesis["chain_id"] = chain_id


def load_json_file(path) -> dict:
    with open(path, "r") as export_file:
        return json.load(export_file)


def replace_validator_from_key(
    genesis, src_validator_pubkey, dest_validator_pubkey
) -> str:
    val_info = get_validator_info(src_validator_pubkey)
    return replace_validator_with_info(genesis, val_info, dest_validator_pubkey)


def replace_validator_with_info(genesis, val_info, dest_validator_pubkey) -> str:
    src_validator_pubkey = val_info["consensus_pubkey"]["key"]

    src_operator_addr = val_info["operator_address"]
    print(f"Replacing validator {src_operator_addr}")

    val_addr = None
    val_info["consensus_pubkey"]["key"] = dest_validator_pubkey
    for val in genesis["validators"]:
        if val["pub_key"]["value"] == src_validator_pubkey:
            val["pub_key"]["value"] = dest_validator_pubkey
            val_addr = val["address"]
            break
    assert val_addr is not None, "Validator not found in genesis"




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
        "genesis_export", type=get_path, help="The path to the genesis export"
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
    parser.add_argument("--chain_id", help="New chain ID to be set", default=None)
    parser.add_argument(
        "--voting_period",
        help="The new voting period to be set",
        default=DEFAULT_VOTING_PERIOD,
    )

    return parser.parse_args()


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
    validator_operator_address = convert_to_valoper(validator_operator_base_address)
    print(f"       {validator_operator_base_address}")
    print(validator_operator_address)

    # load the genesis up
    print("reading genesis export...")
    genesis = load_json_file(args.genesis_export)

    print("reading genesis export...complete")

    val_info = get_unjailed_validator(genesis)
    if not val_info:
        print("Genesis file does not contain any validators")
        sys.exit(1)

    target_validator_operator_address = val_info["operator_address"]
    val_tokens = int(val_info["tokens"])
    val_power = int(val_tokens / (10**18))

    # Replace selected validator by current node one
    val_addr = replace_validator_with_info(genesis, val_info, validator_pubkey)

    genesis_dump = json.dumps(genesis)
    genesis_dump = re.sub(val_addr, validator_hexaddr, genesis_dump)
    genesis_dump = re.sub(
        target_validator_operator_address, validator_operator_address, genesis_dump
    )
    genesis = json.loads(genesis_dump)

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

    # Get current bonded and not bonded tokens
    bonded_tokens = get_balance(genesis, bonded_pool_address, args.staking_denom)
    not_bonded_tokens = get_balance(
        genesis, not_bonded_pool_address, args.staking_denom
    )

    new_not_bonded_tokens = not_bonded_tokens + bonded_tokens - val_tokens

    # Update bonded pool and not bonded pool balances
    set_balance(genesis, bonded_pool_address, val_tokens, args.staking_denom)
    set_balance(
        genesis, not_bonded_pool_address, new_not_bonded_tokens, args.staking_denom
    )

    # Create new account and fund it
    print(
        f"Creating new funded account for local validator {validator_operator_base_address}..."
    )

    # Add new balance to bank
    genesis = set_balance(
        genesis, validator_operator_base_address, FUND_BALANCE, args.staking_denom
    )

    # Add new account to auth if not already there
    genesis = ensure_account(genesis, validator_operator_base_address)

    # Update total supply of staking denom with new funds added
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
    jail_validators(genesis, validator_operator_address)

    remove_max_wasm_code_size(genesis)

    # Set voting period
    set_voting_period(genesis, args.voting_period)

    # Update the chain id if provided
    if args.chain_id:
        update_chain_id(genesis, args.chain_id)

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

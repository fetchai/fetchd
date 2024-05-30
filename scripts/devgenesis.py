#!/usr/bin/env python3

import argparse
import json
import os
import sys
from genesis_helpers import (
    get_balance,
    get_path,
    get_unjailed_validator,
    set_balance,
    ensure_account,
    convert_to_valoper,
    load_json_file,
    replace_validator_with_info,
    jail_validators,
    remove_max_wasm_code_size,
    set_voting_period,
    update_chain_id,
    get_local_key_data,
    hex_address_to_bech32,
    get_account_address_by_name,
)

DEFAULT_STAKING_DENOM = "afet"
DEFAULT_HOME_PATH = os.path.expanduser("~") + "/.fetchd"
DEFAULT_VALIDATOR_KEY_NAME = "validator"
FUND_BALANCE = 10**23
DEFAULT_VOTING_PERIOD = "60s"


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

    # extract the tendermint addresses
    local_validator_json = load_json_file(local_validator_key_path)
    validator_hexaddr = local_validator_json["address"]
    validator_address = hex_address_to_bech32(validator_hexaddr, "fetchvalcons")
    validator_pubkey = local_validator_json["pub_key"]["value"]

    # extract the address for the local validator key
    local_key_data = get_local_key_data(args.home_path, args.validator_key_name)

    # extract the local address and convert into a valid validator operator address
    local_validator_base_address = local_key_data["address"]
    local_validator_operator_address = convert_to_valoper(local_validator_base_address)
    print(f"{local_validator_base_address} {local_validator_operator_address}")

    # load the genesis up
    print("reading genesis export...")
    genesis = load_json_file(args.genesis_export)

    print("reading genesis export...complete")

    target_val_info = get_unjailed_validator(genesis)
    if not target_val_info:
        print("Genesis file does not contain any validators")
        sys.exit(1)

    val_tokens = int(target_val_info["tokens"])
    val_power = int(val_tokens / (10**18))

    # Replace selected validator by current node one
    replace_validator_with_info(
        genesis,
        target_val_info,
        validator_pubkey,
        validator_hexaddr,
        local_validator_operator_address,
    )

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

    bonded_pool_address = get_account_address_by_name(genesis, "bonded_tokens_pool")
    not_bonded_pool_address = get_account_address_by_name(
        genesis, "not_bonded_tokens_pool"
    )

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
        f"Creating new funded account for local validator {local_validator_base_address}..."
    )

    # Add new balance to bank
    set_balance(genesis, local_validator_base_address, FUND_BALANCE, args.staking_denom)

    # Add new account to auth if not already there
    ensure_account(genesis, local_validator_base_address)

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
        {"address": local_validator_operator_address, "power": str(val_power)}
    ]

    # Jail everyone but our validator
    print("Jail other validators...")
    jail_validators(genesis, local_validator_operator_address)

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

#!/usr/bin/env python3

import argparse as ap
import json
import os
import sys
from genesis_helpers import (
    get_balance,
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
    get_account,
    pubkey_to_bech32_address,
    ExpandPath, increase_balance,
)
from replace_validator import replace_validator_keys_recursive
from typing import Tuple


DEFAULT_HOME_PATH = os.path.expanduser("~/.fetchd")
DEFAULT_VALIDATOR_KEY_NAME = "validator"
FUND_BALANCE = 10**23
DEFAULT_VOTING_PERIOD = "60s"


def parse_commandline() -> Tuple[ap.Namespace, ap.ArgumentParser]:

    parser = ap.ArgumentParser(
        description="""
CLI for post-processing of `genesis.json` file to achieve desired changes.
The primary purpose of this CLI is for *testing* blockchain deployments.
It is not recommended to use this CLI for production grade deployments.""",
        epilog="""Example of usage:
python %(prog)s --home ~/.fetchd "my_genesis.json" reset_to_single_validator""",
        formatter_class=ap.RawTextHelpFormatter)
    parser.set_defaults(func=lambda *args: parser.print_help())

    parser.add_argument(
        "--home",
        help="The path to the local node data i.e. ~/.fetchd",
        default=DEFAULT_HOME_PATH,
        action=ExpandPath)
    parser.add_argument("genesis_file_path", type=str, help="The path to the genesis file", action=ExpandPath)

    subparsers = parser.add_subparsers(help='sub-command help')

    parser_single_validator = subparsers.add_parser(
        'reset_to_single_validator',
        help='Reset to single validator',
        description="""This script updates an exported genesis from a running chain
to be used to run on a single validator local node.
It will take the first validator which is not jailed, jail all remaining validators.
Then it will replace the validator pubkey & nodekey with the one found in the `node_home_dir` directory provided by the `--home node_home_dir` argument.
If unspecified, the `node_home_dir` value defaults to the "~/.fetchd" directory, this folder must exist and contain the directory structure created by the "fetchd init ..." command.

The updated genesis will be written under `node_home_dir`/config/genesis.json, allowing the local chain to be started with.""",
        epilog="""Example of usage:
python %(prog)s --home ~/.fetchd "my_genesis.json" reset_to_single_validator
python %(prog)s --home ~/.fetchd "my_genesis.json" reset_to_single_validator --validator_key_name "my_validator_key_name_in_fetchd_keyring" --voting_period 300s""",
        formatter_class=ap.RawTextHelpFormatter)
    parser_single_validator.add_argument(
        "--validator_key_name",
        help="The name of the local key to use for the validator",
        default=DEFAULT_VALIDATOR_KEY_NAME)
    parser_single_validator.add_argument("--chain_id", help="New chain ID to be set", default=None)
    parser_single_validator.add_argument(
        "--voting_period",
        help="The new voting period to be set",
        default=DEFAULT_VOTING_PERIOD)
    parser_single_validator.set_defaults(func=reset_to_single_validator)


    parser_replace_validator_keys = subparsers.add_parser(
        'replace_validator_keys',
        help='Replace consensus and operator keys of given validator',
        description="This script replaces a validator in the genesis file based on provided public keys and addresses.",
        epilog="""Example of usage:
    python %(prog)s --home ~/.fetchd "my_genesis.json" replace_validator_keys "Fd9qzmh+4ZfLwLw1obIN9jPcijh1O7ZwuVBQwbP7RaM=" "AtZLs0C20OK7BvwyBB8nkbo8NB05LwH1qyhkBNTD+M5i" "A2A07JmOtkK/rd/R1rhzj5sDzDJ+EbdGj7DY8ghVx0tq"
    python %(prog)s --home ~/.fetchd "my_genesis.json" replace_validator_keys "Fd9qzmh+4ZfLwLw1obIN9jPcijh1O7ZwuVBQwbP7RaM=" "AtZLs0C20OK7BvwyBB8nkbo8NB05LwH1qyhkBNTD+M5i" "A2A07JmOtkK/rd/R1rhzj5sDzDJ+EbdGj7DY8ghVx0tq" --output "my_resulting_genesis.json" """,
        formatter_class=ap.RawTextHelpFormatter)
    parser_replace_validator_keys.add_argument(
        "src_validator_pubkey",
        type=str,
        help="Source validator *consensus* public key in base64 format, for example: Fd9qzmh+4ZfLwLw1obIN9jPcijh1O7ZwuVBQwbP7RaM=")
    parser_replace_validator_keys.add_argument(
        "dest_validator_pubkey",
        type=str,
        help="Destination validator *consesnus* public key in base64 format, for example: AtZLs0C20OK7BvwyBB8nkbo8NB05LwH1qyhkBNTD+M5i")
    parser_replace_validator_keys.add_argument(
        "dest_validator_operator_pubkey",
        type=str,
        help="Destination validator *operator* public key in base64 format, for example: A2A07JmOtkK/rd/R1rhzj5sDzDJ+EbdGj7DY8ghVx0tq")
    parser_replace_validator_keys.add_argument(
        "--output",
        help="The path for modified genesis file",
        default="modified_genesis.json")
    parser_replace_validator_keys.set_defaults(func=replace_validator_keys)

    return parser.parse_args(), parser


def reset_to_single_validator(args: ap.Namespace):
    print("    Genesis Export:", args.genesis_file_path)
    print("  Fetchd Home Path:", args.home)
    print("Validator Key Name:", args.validator_key_name)

    # load up the local validator key
    local_validator_key_path = os.path.join(
        args.home, "config", "priv_validator_key.json"
    )

    # extract the tendermint addresses
    local_validator_json = load_json_file(local_validator_key_path)
    validator_hexaddr = local_validator_json["address"]
    validator_address = hex_address_to_bech32(validator_hexaddr, "fetchvalcons")
    validator_pubkey = local_validator_json["pub_key"]["value"]

    # extract the address for the local validator key
    local_key_data = get_local_key_data(args.home, args.validator_key_name)

    # extract the local address and convert into a valid validator operator address
    local_validator_base_address = local_key_data["address"]
    local_validator_operator_address = convert_to_valoper(local_validator_base_address)
    print(f"{local_validator_base_address} {local_validator_operator_address}")

    # load the genesis up
    print("reading genesis export...")
    genesis = load_json_file(args.genesis_file_path)
    print("reading genesis export...complete")

    staking_denom = genesis["app_state"]["staking"]["params"]["bond_denom"]
    print(f"Staking denom: {staking_denom}")

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
    bonded_tokens = get_balance(genesis, bonded_pool_address, staking_denom)
    not_bonded_tokens = get_balance(genesis, not_bonded_pool_address, staking_denom)

    new_not_bonded_tokens = not_bonded_tokens + bonded_tokens - val_tokens
    if new_not_bonded_tokens < 0:
        print(f"Invalid new_not_bonded_tokens amount: {new_not_bonded_tokens}")
        sys.exit(1)

    # Update bonded pool and not bonded pool balances
    set_balance(genesis, bonded_pool_address, val_tokens, staking_denom)
    set_balance(genesis, not_bonded_pool_address, new_not_bonded_tokens, staking_denom)

    # Create new account and fund it
    print(
        f"Creating new funded account for local validator {local_validator_base_address}..."
    )

    # Add new balance to bank
    increase_balance(genesis, local_validator_base_address, FUND_BALANCE, staking_denom)

    # Add new account to auth if not already there
    ensure_account(genesis, local_validator_base_address)

    # Update total supply of staking denom with new funds added
    for supply in genesis["app_state"]["bank"]["supply"]:
        if supply["denom"] == staking_denom:
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
    with open(f"{args.home}/config/genesis.json", "w") as f:
        json.dump(genesis, f)

    print(f"Done! Wrote new genesis at {args.home}/config/genesis.json")
    print("You can now start the chain:")
    print()
    print(
        f"fetchd --home {args.home} tendermint unsafe-reset-all && fetchd --home {args.home} start"
    )
    print()


def replace_validator_keys(args: ap.Namespace):
    print("       Genesis Path:", args.genesis_file_path)
    print("Source Validator PK:", args.src_validator_pubkey)
    print("Destination Validator PK:", args.dest_validator_pubkey)
    print("Destination Operator PK:", args.dest_validator_operator_pubkey)

    # Load the genesis file
    print("Reading genesis file...")
    genesis = load_json_file(args.genesis_file_path)
    print("Reading genesis file...complete")

    # TODO(pb): Whole this check can be dropped, since it does not have any effect (code will continue disregard):
    # Ensure that operator is not already registered in auth module:
    dest_operator_base_address = pubkey_to_bech32_address(
        args.dest_validator_operator_pubkey, "fetch"
    )
    new_operator_has_account = get_account(genesis, dest_operator_base_address)
    if new_operator_has_account:
        print(
            "New operator account already existed before - it is recommended to generate new operator key"
        )

    replace_validator_keys_recursive(
        genesis=genesis,
        src_validator_pubkey=args.src_validator_pubkey,
        dest_validator_pubkey=args.dest_validator_pubkey,
        dest_validator_operator_pubkey=args.dest_validator_operator_pubkey)

    # Save the modified genesis file
    output_genesis_path = args.output
    print(f"Writing modified genesis file to {output_genesis_path}...")
    with open(output_genesis_path, "w") as f:
        json.dump(genesis, f)
    print("Modified genesis file written successfully.")


def main():
    args, _ = parse_commandline()
    args.func(args)


if __name__ == "__main__":
    main()

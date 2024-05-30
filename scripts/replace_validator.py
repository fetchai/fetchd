import argparse
import json
import os

from genesis_helpers import (
    load_json_file,
    get_staking_validator_info,
    replace_validator_with_info,
    validator_pubkey_to_valcons_address,
    replace_validator_slashing,
    validator_pubkey_to_hex_address,
)


def parse_commandline():
    description = "This script replaces a validator in the genesis file based on provided public keys and addresses."

    parser = argparse.ArgumentParser(description=description)
    parser.add_argument("genesis", type=str, help="The path to the genesis file")
    parser.add_argument(
        "src_validator_pubkey",
        type=str,
        help="Source validator public key in base64 format, f.e. Fd9qzmh+4ZfLwLw1obIN9jPcijh1O7ZwuVBQwbP7RaM=",
    )
    parser.add_argument(
        "dest_validator_pubkey",
        type=str,
        help="Destination validator public key in base64 format, f.e. Fd9qzmh+4ZfLwLw1obIN9jPcijh1O7ZwuVBQwbP7RaM=",
    )
    parser.add_argument(
        "dest_validator_operator_address",
        type=str,
        help="Destination validator operator address, f.e. fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml",
    )

    parser.add_argument(
        "--output",
        help="The path for modified genesis file",
        default="modified_genesis.json",
    )

    return parser.parse_args()


def main():
    args = parse_commandline()

    print("       Genesis Path:", args.genesis)
    print("Source Validator PK:", args.src_validator_pubkey)
    print("Destination Validator PK:", args.dest_validator_pubkey)
    print("Destination Operator Address:", args.dest_validator_operator_address)

    # Load the genesis file
    print("Reading genesis file...")
    genesis = load_json_file(args.genesis)
    print("Reading genesis file...complete")

    target_val_info = get_staking_validator_info(genesis, args.src_validator_pubkey)
    target_consensus_address = validator_pubkey_to_valcons_address(
        target_val_info["consensus_pubkey"]["key"]
    )

    dest_validator_hex_addr = validator_pubkey_to_hex_address(
        args.dest_validator_pubkey
    )
    dest_consensus_address = validator_pubkey_to_valcons_address(
        dest_validator_hex_addr
    )

    # Replace validator slashing module entry
    replace_validator_slashing(
        genesis, target_consensus_address, dest_consensus_address
    )

    # Replace the validator in the genesis file
    replace_validator_with_info(
        genesis,
        target_val_info,
        args.dest_validator_pubkey,
        dest_validator_hex_addr,
        args.dest_validator_operator_address,
    )

    # Save the modified genesis file
    output_genesis_path = args.output
    print(f"Writing modified genesis file to {output_genesis_path}...")
    with open(output_genesis_path, "w") as f:
        json.dump(genesis, f)
    print("Modified genesis file written successfully.")


if __name__ == "__main__":
    main()

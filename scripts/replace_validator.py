import argparse
import json
import re

from genesis_helpers import (
    load_json_file,
    get_staking_validator_info,
    validator_pubkey_to_valcons_address,
    validator_pubkey_to_hex_address,
    pubkey_to_bech32_address,
    convert_to_valoper,
    convert_to_base,
    get_validator_info,
    get_account,
    find_key_path,
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
        "dest_validator_operator_pubkey",
        type=str,
        help="Destination validator operator public key in base64 format, f.e. Fd9qzmh+4ZfLwLw1obIN9jPcijh1O7ZwuVBQwbP7RaM=",
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
    print("Destination Operator PK:", args.dest_validator_operator_pubkey)

    # Load the genesis file
    print("Reading genesis file...")
    genesis = load_json_file(args.genesis)
    print("Reading genesis file...complete")

    # Input values #
    target_staking_val_info = get_staking_validator_info(
        genesis, args.src_validator_pubkey
    )
    target_val_info = get_validator_info(genesis, args.src_validator_pubkey)

    target_val_pubkey = target_staking_val_info["consensus_pubkey"]["key"]
    target_consensus_address = validator_pubkey_to_valcons_address(target_val_pubkey)

    target_operator_address = target_staking_val_info["operator_address"]
    target_operator_base_address = convert_to_base(target_operator_address)

    target_operator_pubkey = get_account(genesis, target_operator_base_address)[
        "pub_key"
    ]["key"]

    # Output values #
    dest_validator_hex_addr = validator_pubkey_to_hex_address(
        args.dest_validator_pubkey
    )
    dest_consensus_address = validator_pubkey_to_valcons_address(
        args.dest_validator_pubkey
    )

    dest_operator_base_address = pubkey_to_bech32_address(
        args.dest_validator_operator_pubkey, "fetch"
    )
    dest_operator_valoper_address = convert_to_valoper(dest_operator_base_address)

    # Replacements

    # Replace validator hex address and pubkey
    target_val_info["address"] = dest_validator_hex_addr
    target_val_info["pub_key"]["value"] = args.dest_validator_pubkey
    target_staking_val_info["consensus_pubkey"]["key"] = args.dest_validator_pubkey

    # Ensure that operator is not already registered in auth module:
    new_operator_has_account = False
    for account in genesis["app_state"]["auth"]["accounts"]:
        if "address" in account and account["address"] == dest_operator_base_address:
            new_operator_has_account = True
            break

    if new_operator_has_account:
        print(
            "New operator account already existed before - it is recommended to generate new operator key"
        )

    # Replace operator account pubkey
    if not new_operator_has_account:
        for account in genesis["app_state"]["auth"]["accounts"]:
            if (
                "pub_key" in account
                and account["pub_key"]
                and "key" in account["pub_key"]
                and account["pub_key"]["key"] == target_operator_pubkey
            ):
                account["pub_key"]["key"] = args.dest_validator_operator_pubkey

    # Brute force replacement of all remaining occurrences
    genesis_dump = json.dumps(genesis)

    # Convert validator valcons address
    genesis_dump = re.sub(
        target_consensus_address, dest_consensus_address, genesis_dump
    )

    # Convert operator valoper address
    genesis_dump = re.sub(
        target_operator_address,
        dest_operator_valoper_address,
        genesis_dump,
    )

    # Convert operator base account address
    genesis_dump = re.sub(
        target_operator_base_address,
        dest_operator_base_address,
        genesis_dump,
    )

    genesis = json.loads(genesis_dump)

    # Save the modified genesis file
    output_genesis_path = args.output
    print(f"Writing modified genesis file to {output_genesis_path}...")
    with open(output_genesis_path, "w") as f:
        json.dump(genesis, f)
    print("Modified genesis file written successfully.")


if __name__ == "__main__":
    main()

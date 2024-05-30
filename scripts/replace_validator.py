import argparse
import json
import os
from genesis_helpers import load_json_file, replace_validator_from_pubkey


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
        "dest_validator_hexaddr",
        type=str,
        help="Destination validator hex address, f.e. 758F13BB838F48DEE6D6E611F5A90B66CBF8BDB7",
    )
    parser.add_argument(
        "dest_validator_operator_address, f.e. fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml",
        type=str,
        help="Destination validator operator address",
    )

    return parser.parse_args()


def main():
    args = parse_commandline()

    print("       Genesis Path:", args.genesis)
    print("Source Validator PK:", args.src_validator_pubkey)
    print("Destination Validator PK:", args.dest_validator_pubkey)
    print("Destination Hex Address:", args.dest_validator_hexaddr)
    print("Destination Operator Address:", args.dest_validator_operator_address)

    # Load the genesis file
    print("Reading genesis file...")
    genesis = load_json_file(args.genesis)
    print("Reading genesis file...complete")

    # Replace the validator in the genesis file
    replace_validator_from_pubkey(
        genesis,
        args.src_validator_pubkey,
        args.dest_validator_pubkey,
        args.dest_validator_hexaddr,
        args.dest_validator_operator_address,
    )

    # Save the modified genesis file
    output_genesis_path = f"{os.path.dirname(args.genesis)}/modified_genesis.json"
    print(f"Writing modified genesis file to {output_genesis_path}...")
    with open(output_genesis_path, "w") as f:
        json.dump(genesis, f, indent=2)
    print("Modified genesis file written successfully.")


if __name__ == "__main__":
    main()

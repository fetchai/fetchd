import argparse
import copy
import json
import os
from collections import OrderedDict


def load_json_file(filename):
    """Load the entire JSON file into memory."""
    with open(filename, 'r') as file:
        data = json.load(file)
    return data


def save_json_file(filename, data):
    """Save a dictionary to a JSON file."""
    with open(filename, 'w') as file:
        json.dump(data, file)


def parse_commandline():
    description = """This script prunes the genesis file by removing unused codes
and updating the code IDs in the contracts. It takes the input genesis file
and outputs the pruned genesis file.
    """

    parser = argparse.ArgumentParser(description=description)
    parser.add_argument(
        "genesis_export", type=_path, help="The path to the genesis export"
    )
    parser.add_argument(
        "--output_file",
        help="The path to save the pruned genesis file",
        default="out.json",
    )

    return parser.parse_args()


def _path(text: str) -> str:
    return os.path.abspath(text)


def main():
    # Parse command-line arguments
    args = parse_commandline()

    # Load genesis file
    print("Opening genesis file")
    genesis = load_json_file(args.genesis_export)

    wasm = genesis["app_state"]["wasm"]
    codes = wasm["codes"]
    contracts = wasm["contracts"]

    print("Building code hashes map")
    # Create maps with code hashes and IDs
    original_code_hash_to_code = OrderedDict()
    original_code_id_to_hash = {}
    for code in codes:
        code_hash = code["code_info"]["code_hash"]
        original_code_hash_to_code[code_hash] = code
        original_code_id_to_hash[code["code_id"]] = code_hash

    # Create pruned code list
    new_codes = []
    new_code_hash_to_code = {}

    next_code_id = 1
    for code_hash, code in original_code_hash_to_code.items():
        new_code = copy.deepcopy(code)
        new_code["code_id"] = str(next_code_id)
        new_code_hash_to_code[code["code_info"]["code_hash"]] = new_code

        new_codes.append(new_code)
        next_code_id += 1

    # Replace original code list with pruned one
    wasm["codes"] = new_codes

    print("Replacing code IDs")
    # Replace old code IDs with new ones
    for contract in contracts:
        code_id = contract["contract_info"]["code_id"]

        # Get code hash
        code_hash = original_code_id_to_hash[code_id]

        # Convert old code ID to new one
        new_code_id = new_code_hash_to_code[code_hash]["code_id"]

        contract["contract_info"]["code_id"] = new_code_id

    # Replace code_id sequence
    sequences = wasm["sequences"]
    for sequence in sequences:
        if sequence["id_key"] == "BGxhc3RDb2RlSWQ=":
            sequence["value"] = str(next_code_id)

    # Store pruned genesis file
    print("Writing output json")
    save_json_file(args.output_file, genesis)


if __name__ == "__main__":
    main()

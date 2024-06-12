#import json
#import re
from typing import Optional, Mapping, Any
from genesis_helpers import (
    #load_json_file,
    get_staking_validator_info,
    validator_pubkey_to_valcons_address,
    validator_pubkey_to_hex_address,
    pubkey_to_bech32_address,
    convert_to_valoper,
    convert_to_base,
    get_validator_info,
    get_account,
)


#def replace_validator_keys(genesis_file_path: str,
#                           src_validator_pubkey: str,
#                           dest_validator_pubkey: str,
#                           dest_validator_operator_pubkey: str,
#                           output_genesis_file_path: Optional[str] = None,
#                           ):
#    print("       Genesis Path:", genesis_file_path)
#    print("Source Validator PK:", src_validator_pubkey)
#    print("Destination Validator PK:", dest_validator_pubkey)
#    print("Destination Operator PK:", dest_validator_operator_pubkey)
#
#    # Load the genesis file
#    print("Reading genesis file...")
#    genesis = load_json_file(genesis_file_path)
#    print("Reading genesis file...complete")
#
#    # Input values #
#    target_staking_val_info = get_staking_validator_info(
#        genesis, src_validator_pubkey
#    )
#    target_val_info = get_validator_info(genesis, src_validator_pubkey)
#
#    target_val_pubkey = target_staking_val_info["consensus_pubkey"]["key"]
#    target_consensus_address = validator_pubkey_to_valcons_address(target_val_pubkey)
#
#    target_operator_address = target_staking_val_info["operator_address"]
#    target_operator_base_address = convert_to_base(target_operator_address)
#
#    # Output values #
#    dest_validator_hex_addr = validator_pubkey_to_hex_address(
#        dest_validator_pubkey
#    )
#    dest_consensus_address = validator_pubkey_to_valcons_address(
#        dest_validator_pubkey
#    )
#
#    dest_operator_base_address = pubkey_to_bech32_address(
#        dest_validator_operator_pubkey, "fetch"
#    )
#    dest_operator_valoper_address = convert_to_valoper(dest_operator_base_address)
#
#    # Replacements
#
#    # Replace validator hex address and pubkey
#    target_val_info["address"] = dest_validator_hex_addr
#    target_val_info["pub_key"]["value"] = dest_validator_pubkey
#    target_staking_val_info["consensus_pubkey"]["key"] = dest_validator_pubkey
#
#    # Ensure that operator is not already registered in auth module:
#    new_operator_has_account = get_account(genesis, dest_operator_base_address)
#
#    if new_operator_has_account:
#        print(
#            "New operator account already existed before - it is recommended to generate new operator key"
#        )
#
#    # Replace operator account pubkey
#    if not new_operator_has_account:
#        target_operator_account = get_account(genesis, target_operator_base_address)
#        # Replace pubkey if present
#        if target_operator_account["pub_key"]:
#            new_pubkey = {
#                "@type": "/cosmos.crypto.secp256k1.PubKey",
#                "key": dest_validator_operator_pubkey,
#            }
#            target_operator_account["pub_key"] = new_pubkey
#
#    # Brute force replacement of all remaining occurrences
#    genesis_dump = json.dumps(genesis)
#
#    # Convert validator valcons address
#    genesis_dump = re.sub(
#        target_consensus_address, dest_consensus_address, genesis_dump
#    )
#
#    # Convert operator valoper address
#    genesis_dump = re.sub(
#        target_operator_address,
#        dest_operator_valoper_address,
#        genesis_dump,
#    )
#
#    # Convert operator base account address
#    genesis_dump = re.sub(
#        target_operator_base_address,
#        dest_operator_base_address,
#        genesis_dump,
#    )
#
#    genesis = json.loads(genesis_dump)
#
#    # Save the modified genesis file
#    if output_genesis_file_path is None:
#        output_genesis_file_path = genesis_file_path
#    print(f"Writing modified genesis file to {output_genesis_file_path}...")
#    with open(output_genesis_file_path, "w") as f:
#        json.dump(genesis, f)
#    print("Modified genesis file written successfully.")


#def iterate_json(json_obj: Mapping[str, Any], change_value_predicate: Callable[[Union[str, int], Any], Optional[Any]]):
#    if isinstance(json_obj, dict):
#        for key, value in json_obj.items():
#            res_val = change_value_predicate(key, value)
#            if res_val is not None:
#                value = res_val
#                json_obj[key] = value
#            iterate_json(value, change_value_predicate)
#    elif isinstance(json_obj, list):
#        for idx, value in enumerate(json_obj):
#            res_val = change_value_predicate(idx, value)
#            if res_val is not None:
#                value = res_val
#                json_obj[idx] = value
#            iterate_json(value, change_value_predicate)


def iterate_json_generator(json_obj: Mapping[str, Any]):
    if isinstance(json_obj, dict):
        for key, value in json_obj.items():
            yield json_obj, key, value
            yield from iterate_json_generator(value)
    elif isinstance(json_obj, list):
        for idx, value in enumerate(json_obj):
            yield json_obj, idx, value
            yield from iterate_json_generator(value)


def replace_validator_keys_recur(genesis: Mapping[str, Any],
                                 src_validator_pubkey: str,
                                 dest_validator_pubkey: str,
                                 dest_validator_operator_pubkey: str,
                                 ):
    # Input values #
    target_staking_val_info = get_staking_validator_info(
        genesis, src_validator_pubkey
    )
    target_val_info = get_validator_info(genesis, src_validator_pubkey)

    target_val_pubkey = target_staking_val_info["consensus_pubkey"]["key"]
    target_consensus_address = validator_pubkey_to_valcons_address(target_val_pubkey)

    target_operator_address = target_staking_val_info["operator_address"]
    target_operator_base_address = convert_to_base(target_operator_address)

    # Output values #
    dest_validator_hex_addr = validator_pubkey_to_hex_address(
        dest_validator_pubkey
    )
    dest_consensus_address = validator_pubkey_to_valcons_address(
        dest_validator_pubkey
    )

    dest_operator_base_address = pubkey_to_bech32_address(
        dest_validator_operator_pubkey, "fetch"
    )
    dest_operator_valoper_address = convert_to_valoper(dest_operator_base_address)

    # Replacements

    # Replace validator hex address and pubkey
    target_val_info["address"] = dest_validator_hex_addr
    target_val_info["pub_key"]["value"] = dest_validator_pubkey
    target_staking_val_info["consensus_pubkey"]["key"] = dest_validator_pubkey

    # Ensure that operator is not already registered in auth module:
    new_operator_has_account = get_account(genesis, dest_operator_base_address)

    #if new_operator_has_account:
    #    print(
    #        "New operator account already existed before - it is recommended to generate new operator key"
    #    )

    # Replace operator account pubkey
    if not new_operator_has_account:
        target_operator_account = get_account(genesis, target_operator_base_address)
        # Replace pubkey if present
        if target_operator_account["pub_key"]:
            new_pubkey = {
                "@type": "/cosmos.crypto.secp256k1.PubKey",
                "key": dest_validator_operator_pubkey,
            }
            target_operator_account["pub_key"] = new_pubkey

    for js_obj, key, value in iterate_json_generator(genesis):
        if isinstance(value, str):
            # Convert validator valcons address
            if value == target_consensus_address:
                js_obj[key] = dest_consensus_address
            # Convert operator valoper address
            elif value == target_operator_address:
                js_obj[key] = dest_operator_valoper_address
            # Convert operator base account address
            elif value == target_operator_base_address:
                js_obj[key] = dest_operator_base_address

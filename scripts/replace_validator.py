from typing import Mapping, Any
from genesis_helpers import (
    get_staking_validator_info,
    validator_pubkey_to_valcons_address,
    validator_pubkey_to_hex_address,
    pubkey_to_bech32_address,
    convert_to_valoper,
    convert_to_base,
    get_validator_info,
    get_account,
)


def iterate_json_generator(json_obj: Mapping[str, Any]):
    if isinstance(json_obj, dict):
        for key, value in json_obj.items():
            yield json_obj, key, value
            yield from iterate_json_generator(value)
    elif isinstance(json_obj, list):
        for idx, value in enumerate(json_obj):
            yield json_obj, idx, value
            yield from iterate_json_generator(value)


def replace_validator_keys_recursive(genesis: Mapping[str, Any],
                                     src_validator_pubkey: str,
                                     dest_validator_pubkey: str,
                                     dest_validator_operator_pubkey: str):
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

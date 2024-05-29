import json
import os
import re
import subprocess

import bech32
import sys


def get_path(text: str) -> str:
    return os.path.abspath(text)


def to_bech32(prefix: str, data: bytes) -> str:
    data_base5 = bech32.convertbits(data, 8, 5, True)
    if data_base5 is None:
        raise RuntimeError("Unable to parse address")  # pragma: no cover
    return bech32.bech32_encode(prefix, data_base5)


def convert_to_valoper(address):
    hrp, data = bech32.bech32_decode(address)
    if hrp != "fetch":
        print("Invalid address, expected normal fetch address")
        sys.exit(1)

    return bech32.bech32_encode("fetchvaloper", data)


def ensure_account(genesis, address):
    for account in genesis["app_state"]["auth"]["accounts"]:
        if "address" in account and account["address"] == address:
            return

    # Add new account to auth
    last_account_number = int(
        genesis["app_state"]["auth"]["accounts"][-1]["account_number"]
    )

    # Ensure unique account number
    new_account = {
        "@type": "/cosmos.auth.v1beta1.BaseAccount",
        "account_number": str(last_account_number + 1),
        "address": address,
        "pub_key": None,
        "sequence": "0",
    }
    genesis["app_state"]["auth"]["accounts"].append(new_account)


def set_balance(genesis, address, new_balance, denom):
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


def get_balance(genesis, address, denom):
    res_amount = 0
    for balance in genesis["app_state"]["bank"]["balances"]:
        if balance["address"] == address:
            for amount in balance["coins"]:
                if amount["denom"] == denom:
                    res_amount = int(amount["amount"])
                    break
            if res_amount != 0:
                break
    return res_amount


def get_unjailed_validator(genesis) -> dict:
    for val_info in genesis["app_state"]["staking"]["validators"]:
        if not val_info["jailed"]:
            return val_info


def get_validator_info(genesis, validator_pubkey) -> dict:
    for val_info in genesis["app_state"]["staking"]["validators"]:
        if val_info["consensus_pubkey"]["key"] == validator_pubkey:
            return val_info
    return None


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
    genesis,
    src_validator_pubkey,
    dest_validator_pubkey,
    dest_validator_hexaddr,
    dest_validator_operator_address,
):
    val_info = get_validator_info(src_validator_pubkey)
    replace_validator_with_info(
        genesis,
        val_info,
        dest_validator_pubkey,
        dest_validator_hexaddr,
        dest_validator_operator_address,
    )


def replace_validator_with_info(
    genesis,
    val_info,
    dest_validator_pubkey,
    dest_validator_hexaddr,
    dest_validator_operator_address,
):
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

    genesis_dump = json.dumps(genesis)
    genesis_dump = re.sub(val_addr, dest_validator_hexaddr, genesis_dump)
    genesis_dump = re.sub(
        src_operator_addr,
        dest_validator_operator_address,
        genesis_dump,
    )

    # Replace genesis
    genesis.clear()
    genesis.update(json.loads(genesis_dump))


def get_local_key_data(home_path, validator_key_name) -> dict:
    # extract the address for the local validator key
    cmd = [
        "fetchd",
        "--home",
        home_path,
        "keys",
        "show",
        validator_key_name,
        "--output",
        "json",
    ]
    local_key_data = json.loads(subprocess.check_output(cmd).decode())

    if local_key_data["type"] != "local":
        print("Unable to use non-local key type")
        sys.exit(1)
    return local_key_data


def hex_address_to_bech32(hex_address, prefix="fetchvalcons") -> str:
    binary_address = bytes.fromhex(hex_address)
    return to_bech32(prefix, binary_address)

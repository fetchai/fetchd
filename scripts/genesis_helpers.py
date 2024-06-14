import base64
import hashlib
import json
import os
import re
import subprocess
import bech32
import sys
from Crypto.Hash import RIPEMD160  # type: ignore # nosec
from argparse import Action


class ExpandPath(Action):
    def __call__(self, parser, namespace, values, option_string=None):
        path = os.path.abspath(os.path.expanduser(values))
        setattr(namespace, self.dest, path)


def sha256(contents: bytes) -> bytes:
    """
    Get sha256 hash.

    :param contents: bytes contents.

    :return: bytes sha256 hash.
    """
    h = hashlib.sha256()
    h.update(contents)
    return h.digest()


def ripemd160(contents: bytes) -> bytes:
    """
    Get ripemd160 hash using PyCryptodome.

    :param contents: bytes contents.

    :return: bytes ripemd160 hash.
    """
    h = RIPEMD160.new()
    h.update(contents)
    return h.digest()


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


def convert_to_base(address):
    hrp, data = bech32.bech32_decode(address)
    if hrp != "fetchvaloper":
        print("Invalid address, expected fetchvaloper address")
        sys.exit(1)

    return bech32.bech32_encode("fetch", data)


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


def get_account(genesis, address) -> dict:
    for account in genesis["app_state"]["auth"]["accounts"]:
        if "address" in account and account["address"] == address:
            return account
    return None


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


def increase_balance(genesis, address, new_balance, denom):
    account_found = False
    denom_found = False

    for balance in genesis["app_state"]["bank"]["balances"]:
        if balance["address"] == address:
            account_found = True
            for amount in balance["coins"]:
                if amount["denom"] == denom:
                    amount["amount"] = str(int(amount["amount"]) + new_balance)
                    denom_found = True
                    break

            if not denom_found:
                balance["coins"].append({"amount": str(new_balance), "denom": denom})


    if not account_found:
        new_balance_entry = {
            "address": address,
            "coins": [{"amount": str(new_balance), "denom": denom}],
        }
        genesis["app_state"]["bank"]["balances"].append(new_balance_entry)


def get_balance(genesis, address, denom, ensure=False):
    res_amount = 0
    for balance in genesis["app_state"]["bank"]["balances"]:
        if balance["address"] == address:
            for amount in balance["coins"]:
                if amount["denom"] == denom:
                    res_amount = int(amount["amount"])
                    break
            if res_amount != 0:
                break
    if ensure and res_amount == 0:
        print(f"Address {address} has no amount of {denom}")
        sys.exit(1)
    return res_amount


def get_unjailed_validator(genesis) -> dict:
    for val_info in genesis["app_state"]["staking"]["validators"]:
        if not val_info["jailed"]:
            return val_info


def get_staking_validator_info(genesis, validator_pubkey) -> dict:
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


def get_validator_info(genesis, validator_pubkey):
    for val in genesis["validators"]:
        if val["pub_key"]["value"] == validator_pubkey:
            return val

    print("Validator not found in genesis")
    sys.exit(1)


def replace_validator_from_pubkey(
    genesis,
    src_validator_pubkey,
    dest_validator_pubkey,
    dest_validator_hexaddr,
    dest_validator_operator_address,
):
    val_info = get_staking_validator_info(src_validator_pubkey)
    replace_validator_with_info(
        genesis,
        val_info,
        dest_validator_pubkey,
        dest_validator_hexaddr,
        dest_validator_operator_address,
    )


def replace_validator_with_info(
    genesis,
    val_staking_info,
    dest_validator_pubkey,
    dest_validator_hexaddr,
    dest_validator_operator_address,
):
    src_validator_pubkey = val_staking_info["consensus_pubkey"]["key"]
    src_operator_addr = val_staking_info["operator_address"]
    print(f"Replacing validator {src_operator_addr}")

    # Update genesis["validators"] data
    val_info = get_validator_info(genesis, src_validator_pubkey)
    val_info["pub_key"]["value"] = dest_validator_pubkey
    val_info["address"] = dest_validator_hexaddr

    # Update staking module data
    val_staking_info["consensus_pubkey"]["key"] = dest_validator_pubkey

    # Search and replace addresses
    genesis_dump = json.dumps(genesis)

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


def bech32_to_hex_address(address) -> str:
    hrp, data = bech32.bech32_decode(address)
    decoded_data = bytes(bech32.convertbits(data, 5, 8, False))
    return decoded_data.hex().upper()


def get_account_address_by_name(genesis, account_name) -> str:
    for account in genesis["app_state"]["auth"]["accounts"]:
        if "name" in account:
            if account["name"] == account_name:
                return account["base_account"]["address"]


def _validator_pubkey_to_binary_address(validator_pubkey):
    return sha256(base64.b64decode(validator_pubkey))[:20]


def _pubkey_to_binary_address(pubkey):
    return ripemd160(sha256(base64.b64decode(pubkey)))


def pubkey_to_bech32_address(pubkey, prefix="fetch"):
    return to_bech32(prefix, _pubkey_to_binary_address(pubkey))


def validator_pubkey_to_valcons_address(validator_pubkey):
    return to_bech32(
        "fetchvalcons", _validator_pubkey_to_binary_address(validator_pubkey)
    )


def validator_pubkey_to_hex_address(validator_pubkey):
    return _validator_pubkey_to_binary_address(validator_pubkey).hex().upper()


def replace_validator_slashing(genesis, source_addr, dest_addr):
    found = False
    for info in genesis["app_state"]["slashing"]["signing_infos"]:
        if info["address"] == source_addr:
            info["address"] = dest_addr
            info["validator_signing_info"]["address"] = dest_addr
            found = True
            break

    if not found:
        print("Validator not found in slashing")

    # Brute force replacement of all remaining occurrences
    genesis_dump = json.dumps(genesis)
    genesis_dump = re.sub(source_addr, dest_addr, genesis_dump)

    # Replace genesis
    genesis.clear()
    genesis.update(json.loads(genesis_dump))


def find_key_path(json_dict, target):
    paths = []

    def _find_key_path(current, target, path):
        if isinstance(current, dict):
            for key, value in current.items():
                new_path = path + [key]
                _find_key_path(value, target, new_path)
        elif isinstance(current, list):
            for index, item in enumerate(current):
                new_path = path + [index]
                _find_key_path(item, target, new_path)
        else:
            if current == target:
                paths.append(path)
        return None

    _find_key_path(json_dict, target, [])
    return paths

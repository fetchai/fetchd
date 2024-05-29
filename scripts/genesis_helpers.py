import os
import bech32
import sys


def get_path(text: str) -> str:
    return os.path.abspath(text)


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
    return genesis


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
    return genesis


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

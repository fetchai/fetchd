#!/usr/bin/env python3

import json
import os
import re
import sys
from pathlib import Path


def usage():
    script_name = os.path.basename(sys.argv[0])
    print(f"Usage: {script_name} path/to/exported/genesis.json [node_home_dir]")
    print()
    print("This script updates an exported genesis from a running chain")
    print("to be used to run on a single validator local node.")
    print("It will take the first validator and jail all the others")
    print("and replace the validator pubkey and the nodekey with the one")
    print("found in the node_home_dir folder.")
    print()
    print("If unspecified, node_home_dir default to the ~/.fetchd/ folder.")
    print(
        "This folder must exist and contain the files created by the 'fetchd init' command."
    )
    print()
    print(
        "The updated genesis will be written under node_home_dir/config/genesis.json, allowing"
    )
    print("the local chain to be started with:")
    print()
    print(
        "fetchd --home <node_home_dir> unsafe-reset-all && fetchd --home <node_home_dir> start"
    )
    print()
    sys.exit()


staking_denom = "afet"
bonded_pool_address = "fetch1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3xxqtmq"
not_bonded_pool_address = "fetch1tygms3xhhs3yv487phx3dw4a95jn7t7ljxu6d5"
voting_period = "60s"

account_to_fund = "fetch1mrf5yyjnnlpy0egvpk2pvjdk9667j2gtu8kpfy"
fund_balance = 10**23


def main():
    if len(sys.argv) < 2:
        usage()

    gen_file = sys.argv[1]
    if not os.path.isfile(gen_file):
        usage()

    out_homedir = os.path.expanduser(sys.argv[2] if len(sys.argv) > 2 else "~/.fetchd/")

    if not os.path.isfile(f"{out_homedir}/config/priv_validator_key.json"):
        print(f"Cannot find file {out_homedir}/config/priv_validator_key.json")
        sys.exit(1)

    print(f"Found {out_homedir}/config/priv_validator_key.json")

    with open(gen_file, "r") as f:
        gen_content = json.load(f)

    with open(f"{out_homedir}/config/priv_validator_key.json", "r") as f:
        priv_validator_key = json.load(f)

    new_hexaddr = priv_validator_key["address"]
    print(f"- new address: {new_hexaddr}")
    new_pubkey = priv_validator_key["pub_key"]["value"]
    print(f"- new pubkey: {new_pubkey}")
    new_tmaddr = (
        os.popen(f"fetchd --home {out_homedir} tendermint show-address").read().strip()
    )
    print(f"- new tendermint address: {new_tmaddr}")

    val_infos = gen_content["app_state"]["staking"]["validators"][0]
    if not val_infos:
        print("Genesis file does not contain any validators")
        sys.exit(1)

    val_opaddr = val_infos["operator_address"]
    val_pubkey = val_infos["consensus_pubkey"]["key"]
    val_addr = [
        val
        for val in gen_content["validators"]
        if val["pub_key"]["value"] == val_pubkey
    ][0]["address"]

    # Replace selected validator by current node one
    print(f"Replacing validator {val_opaddr}...")
    gen_content = json.loads(re.sub(val_addr, new_hexaddr, json.dumps(gen_content)))
    gen_content = json.loads(re.sub(val_pubkey, new_pubkey, json.dumps(gen_content)))

    # Set .app_state.slashing.signing_infos to contain only our validator signing infos
    print("Updating signing infos...")
    gen_content["app_state"]["slashing"]["signing_infos"] = [
        {
            "address": new_tmaddr,
            "validator_signing_info": {
                "address": new_tmaddr,
                "index_offset": "0",
                "jailed_until": "1970-01-01T00:00:00Z",
                "missed_blocks_counter": "0",
                "start_height": "0",
                "tombstoned": False,
            },
        }
    ]

    # Update bonded and not bonded pool values to make invariant checks happy
    print("Updating bonded and not bonded token pool values...")

    val_tokens = int(val_infos["tokens"])
    val_power = int(val_tokens / (10**18))

    bonded_tokens = next(
        int(amount["amount"])
        for balance in gen_content["app_state"]["bank"]["balances"]
        if balance["address"] == bonded_pool_address
        for amount in balance["coins"]
        if amount["denom"] == staking_denom
    )

    not_bonded_tokens = next(
        int(amount["amount"])
        for balance in gen_content["app_state"]["bank"]["balances"]
        if balance["address"] == not_bonded_pool_address
        for amount in balance["coins"]
        if amount["denom"] == staking_denom
    )

    not_bonded_tokens = not_bonded_tokens + bonded_tokens - val_tokens

    for balance in gen_content["app_state"]["bank"]["balances"]:
        if balance["address"] == bonded_pool_address:
            for amount in balance["coins"]:
                if amount["denom"] == staking_denom:
                    amount["amount"] = str(val_tokens)

        if balance["address"] == not_bonded_pool_address:
            for amount in balance["coins"]:
                if amount["denom"] == staking_denom:
                    amount["amount"] = str(not_bonded_tokens)

    # Create new account and fund it
    print("Creating new account and funding it...")
    new_balance = {
        "address": account_to_fund,
        "coins": [{"amount": str(fund_balance), "denom": staking_denom}],
    }
    gen_content["app_state"]["bank"]["balances"].append(new_balance)

    last_account_number = int(
        gen_content["app_state"]["auth"]["accounts"][-1]["account_number"]
    )
    new_account = {
        "@type": "/cosmos.auth.v1beta1.BaseAccount",
        "account_number": str(last_account_number + 1),
        "address": account_to_fund,
        "pub_key": None,
        "sequence": "0",
    }
    gen_content["app_state"]["auth"]["accounts"].append(new_account)

    # Update total supply
    for supply in gen_content["app_state"]["bank"]["supply"]:
        if supply["denom"] == staking_denom:
            supply["amount"] = str(int(supply["amount"]) + fund_balance)

    # Remove all .validators but the one we work with
    print("Removing other validators from initchain...")
    gen_content["validators"] = [
        val for val in gen_content["validators"] if val["address"] == new_hexaddr
    ]

    # Set .app_state.staking.last_validator_powers to contain only our validator
    print("Updating last voting power...")
    gen_content["app_state"]["staking"]["last_validator_powers"] = [
        {"address": val_opaddr, "power": str(val_power)}
    ]

    # Jail everyone but our validator
    print("Jail other validators...")
    for validator in gen_content["app_state"]["staking"]["validators"]:
        if validator["operator_address"] != val_opaddr:
            validator["status"] = "BOND_STATUS_UNBONDING"
            validator["jailed"] = True

    if "max_wasm_code_size" in gen_content["app_state"]["wasm"]["params"]:
        print("Removing max_wasm_code_size...")
        del gen_content["app_state"]["wasm"]["params"]["max_wasm_code_size"]

    print(f"Setting voting period to {voting_period}...")
    gen_content["app_state"]["gov"]["voting_params"]["voting_period"] = voting_period

    with open(f"{out_homedir}/config/genesis.json", "w") as f:
        json.dump(gen_content, f, indent=2)

    print(f"Done! Wrote new genesis at {out_homedir}/config/genesis.json")
    print("You can now start the chain:")
    print()
    print(
        f"fetchd --home {out_homedir} unsafe-reset-all && fetchd --home {out_homedir} start"
    )
    print()


if __name__ == "__main__":
    main()

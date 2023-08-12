#!/usr/bin/env python3
import os
import re
import sys
import time
import shutil
import subprocess
import json

DENOM = 'atestfet'
FETCHD_CONFIG_ROOT = '/root/.fetchd/config'
GENESIS_PATH = os.path.join(FETCHD_CONFIG_ROOT, 'genesis.json')
GENTX_PATH = os.path.join(FETCHD_CONFIG_ROOT, 'gentx')


def create_genesis(chain_id: str):
    cmd = ['fetchd', 'init', 'setup-node', '--chain-id', chain_id]
    subprocess.check_call(cmd)
    replace_denom_cmd = ['sed', '-i', 's/stake/'+DENOM+'/g', GENESIS_PATH]
    subprocess.check_call(replace_denom_cmd)

def get_validators():
    validators = set()
    for item in os.listdir('/setup'):
        match = re.match(r'^(fetch1[a-z0-9]+)\.validator$', item)
        if match is not None:
            validators.add(match.group(1))
    return validators


def get_gentxs():
    gentxs = set()
    for item in os.listdir('/setup'):
        match = re.match(r'^gentx-fetch1[a-z0-9]+\.json$', item)
        if match is not None:
            path = os.path.join('/setup', item)
            gentxs.add((item, path))
    return gentxs


def main():
    MUNICIPAL_INFL_TARGET_ADDRESS = 'MUNICIPAL_INFL_TARGET_ADDRESS'

    for name in ('CHAINID', 'NUM_VALIDATORS'):
        if name not in os.environ:
            print('{} environment variable not present'.format(name))
            sys.exit(1)

    # extract the environment variables
    chain_id = os.environ['CHAINID']
    num_validators = int(os.environ['NUM_VALIDATORS'])
    municipal_infl_target_address = os.environ.get(MUNICIPAL_INFL_TARGET_ADDRESS, None)

    # create the initial genesis file
    if os.path.exists(GENESIS_PATH):
        os.remove(GENESIS_PATH)
    create_genesis(chain_id)

    with open(GENESIS_PATH, 'r+') as f:
        genesis = json.load(f)
        genesis["app_state"]["staking"]["params"]["max_validators"] = 10
        genesis["app_state"]["staking"]["params"]["max_entries"] = 10
        municipal_infl_genesis_conf = genesis["app_state"]["mint"]["minter"]["municipal_inflation"]
        if municipal_infl_target_address:
            municipal_infl_genesis_conf.extend([
                {"denom": "nanomobx", "inflation": {"target_address": municipal_infl_target_address, "value": "0.03"}},
                {"denom": "denom005", "inflation": {"target_address": municipal_infl_target_address, "value": "0.05"}},
                {"denom": "denom100", "inflation": {"target_address": municipal_infl_target_address, "value": "1.0"}},
                {"denom": "denom010", "inflation": {"target_address": municipal_infl_target_address, "value": "0.1"}},
                {"denom": "denom050", "inflation": {"target_address": municipal_infl_target_address, "value": "0.5"}},
                {"denom": "denom020", "inflation": {"target_address": municipal_infl_target_address, "value": "0.2"}}
                ])

        f.seek(0)
        json.dump(genesis, f, indent=4)
        f.truncate()

    # collect up the validators identities
    validators = get_validators()
    while len(validators) != num_validators:
        print('Waiting for validators to be setup...')
        time.sleep(1)
        validators = get_validators()

    for validator in validators:
        cmd = ['fetchd', 'add-genesis-account',
               validator, '200000000000000000000atestfet']
        subprocess.check_call(cmd)

    if municipal_infl_target_address:
        token_list = ["200000000000000000000atestfet"]

        for infl in municipal_infl_genesis_conf:
            token_list.append(f'{10**18}{infl["denom"]}')
        tokens = ','.join(token_list)

        cmd = ['fetchd', 'add-genesis-account',
               municipal_infl_target_address, tokens]

        subprocess.check_call(cmd)

    # copy the generated genesis file
    shutil.copy(GENESIS_PATH, '/setup/genesis.intermediate.json')

    # collect up the gentxs
    gentxs = get_gentxs()
    while len(gentxs) != num_validators:
        print('Waiting for validators to gentxs...')
        time.sleep(1)
        gentxs = get_gentxs()

    # copy all the gentxs into place
    os.makedirs('/root/.fetchd/config/gentx')
    for item, path in gentxs:
        shutil.copy(path, os.path.join(GENTX_PATH, item))

    # collect up the txs
    cmd = ['fetchd', 'collect-gentxs']
    subprocess.check_call(cmd)

    # generate the final genesis configuration
    shutil.copy(GENESIS_PATH, '/setup/genesis.json')


if __name__ == '__main__':
    main()

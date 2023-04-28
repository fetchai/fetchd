#!/usr/bin/env python3
import argparse
import json
import os
import subprocess
import sys
from pprint import pprint
from typing import Dict, Any, List

import bech32

STAKING_DENOM = 'afet'


def _path(text: str) -> str:
    return os.path.abspath(text)


def _convert_to_valoper(address):
    hrp, data = bech32.bech32_decode(address)
    if hrp != 'fetch':
        print('Invalid address, expected normal fetch address')
        sys.exit(1)

    return bech32.bech32_encode('fetchvaloper', data)


def _from_coin_list(coins: List[Any]) -> Dict[str, int]:
    balances = {}
    for coin in coins:
        balances[str(coin['denom'])] = int(coin['amount'])
    return balances


def _to_coin_list(balances: Dict[str, int]) -> List[Any]:
    coins = []
    for denom in sorted(balances.keys()):
        amount = balances[denom]
        assert amount >= 0

        if amount == 0:
            continue

        coins.append({
            'denom': str(denom),
            'amount': str(amount)
        })
    return coins


def parse_commandline():
    parser = argparse.ArgumentParser()
    parser.add_argument('genesis_export', type=_path, help='The path to the genesis export')
    parser.add_argument('home_path', type=_path, help='The path to the local node data i.e. ~/.fetchd')
    parser.add_argument('validator_key_name', help='The name of the local key to use for the validator')
    return parser.parse_args()


def main():
    args = parse_commandline()

    print('    Genesis Export:', args.genesis_export)
    print('  Fetchd Home Path:', args.home_path)
    print('Validator Key Name:', args.validator_key_name)

    # load up the local validator key
    local_validator_key_path = os.path.join(args.home_path, 'config', 'priv_validator_key.json')
    with open(local_validator_key_path, 'r') as input_file:
        local_validator_key = json.load(input_file)

    # extract the tendermint addresses
    cmd = ['fetchd', '--home', args.home_path, 'tendermint', 'show-address']
    validator_address = subprocess.check_output(cmd).decode().strip()
    validator_pubkey = local_validator_key['pub_key']['value']
    validator_hexaddr = local_validator_key['address']

    # extract the address for the local validtor key
    cmd = ['fetchd', '--home', args.home_path, 'keys', 'show', args.validator_key_name, '--output', 'json']
    key_data = json.loads(subprocess.check_output(cmd).decode())

    if key_data['type'] != 'local':
        print('Unable to use non-local key type')
        sys.exit(1)

    # extract the local address and convert into a valid validator operator address
    validator_operator_base_address = key_data['address']
    validator_operator_address = _convert_to_valoper(validator_operator_base_address)
    print(f'       {validator_operator_base_address}')
    print(validator_operator_address)

    # load the genesis up
    print('reading genesis export...')
    with open(args.genesis_export, 'r') as export_file:
        genesis = json.load(export_file)
    print('reading genesis export...complete')

    # target validator
    target_validator_address = genesis['app_state']['staking']['validators'][0]['operator_address']
    target_validator_public_key = genesis['app_state']['staking']['validators'][0]['consensus_pubkey']['key']

    # lookup the target validator in the validators list and update the data
    print('updating target validator...')
    updated = False
    for n in range(len(genesis['validators'])):
        if genesis['validators'][n]['pub_key']['value'] == target_validator_public_key:
            genesis['validators'][n]['pub_key']['value'] = validator_pubkey
            genesis['validators'][n]['address'] = validator_hexaddr
            updated = True
            break
    if not updated:
        print('Failed to locate validator in validator list')
        sys.exit(1)

    # update the signing infos
    print('updating signing infos...')
    genesis['app_state']['slashing']['signing_infos'] = [{
        'address': validator_address,
        'validator_signing_info': {
            'address': validator_address,
            'index_offset': '0',
            'jailed_until': '1970-01-01T00:00:00Z',
            'missed_blocks_counter': '0',
            'start_height': '0',
        }
    }]

    assert len(genesis['app_state']['bank']['balances']) >= 2

    print('balance shenanigans...')
    # this part of the script taken from the original script, I assume some additional things
    # were meant to be done here
    target_validator_tokens = int(genesis['app_state']['staking']['validators'][0]['tokens'])

    # find an address which
    updated = False
    for n in range(len(genesis['app_state']['bank']['balances'])):
        balances = _from_coin_list(genesis['app_state']['bank']['balances'][n]['coins'])
        staking_denom_balance = balances.get(STAKING_DENOM, 0)

        if staking_denom_balance >= target_validator_tokens:

            delta_balance = staking_denom_balance - target_validator_tokens
            balances[STAKING_DENOM] = target_validator_tokens
            genesis['app_state']['bank']['balances'][n]['coins'] = _to_coin_list(balances)

            # balance the books!
            alt = 0 if n != 0 else 1 # another balance entry
            alt_balances = _from_coin_list(genesis['app_state']['bank']['balances'][alt]['coins'])
            alt_balances[STAKING_DENOM] = alt_balances.get(STAKING_DENOM, 0) + delta_balance
            genesis['app_state']['bank']['balances'][alt]['coins'] = _to_coin_list(alt_balances)

            updated = True
            break
    assert updated, "unable to update balances"

    # filter out all the validators except our sole one
    print('dropping other validators...')
    genesis['validators'] = list(
        filter(
            lambda x: x['address'] == validator_hexaddr,
            genesis['validators'],
        )
    )

    # set last voting power
    print('set last voting power...')
    genesis['app_state']['staking']['last_validator_powers'] = [{
      'address': target_validator_address,
      'power': str(target_validator_tokens // (10 ** 18)),
    }]

    # jail the other validators
    print('jail other validators...')
    for n in range(len(genesis['app_state']['staking']['validators'])):
        if genesis['app_state']['staking']['validators'][n]['operator_address'] != target_validator_address:
            genesis['app_state']['staking']['validators'][n]['status'] = 'BOND_STATUS_UNBONDING'
            genesis['app_state']['staking']['validators'][n]['jailed'] = True

    # debug
    # genesis['app_state']['auth']['accounts'] = [
    #     genesis['app_state']['auth']['accounts'][0]
    # ]

    # genesis['app_state']['bank']['balances'] = [
    #     genesis['app_state']['bank']['balances'][0]
    # ]

    # genesis['app_state']['distribution']['delegator_starting_infos'] = [
    #     genesis['app_state']['distribution']['delegator_starting_infos'][0]
    # ]

    # genesis['app_state']['distribution']['outstanding_rewards'] = [
    #     genesis['app_state']['distribution']['outstanding_rewards'][0]
    # ]

    # genesis['app_state']['distribution']['validator_accumulated_commissions'] = [
    #     genesis['app_state']['distribution']['validator_accumulated_commissions'][0]
    # ]

    # genesis['app_state']['distribution']['validator_current_rewards'] = [
    #     genesis['app_state']['distribution']['validator_current_rewards'][0]
    # ]

    # genesis['app_state']['distribution']['validator_historical_rewards'] = [
    #     genesis['app_state']['distribution']['validator_historical_rewards'][0]
    # ]

    # genesis['app_state']['distribution']['validator_slash_events'] = [
    #     genesis['app_state']['distribution']['validator_slash_events'][0]
    # ]

    # dump the genesis out
    print('dumping genesis file...')
    local_genesis_path = os.path.join(args.home_path, 'config', 'genesis.json')
    with open(local_genesis_path, 'w') as output_file:
        json.dump(genesis, output_file, indent=2, sort_keys=True)
    print('dumping genesis file...complete')

    # pprint(genesis['app_state']['staking']['validators'][0])
    # print()


# NEW_HEXADDR=$(jq -r '.address' "${OUT_HOMEDIR}/config/priv_validator_key.json")
# echo "- new address: ${NEW_HEXADDR}"
# NEW_PUBKEY=$(jq -r '.pub_key.value' "${OUT_HOMEDIR}/config/priv_validator_key.json")
# echo "- new pubkey: ${NEW_PUBKEY}"
# NEW_TMADDR=$(fetchd --home "${OUT_HOMEDIR}" tendermint show-address)
# echo "- new tendermint address: ${NEW_TMADDR}"

# if [ ! -f "${OUT_HOMEDIR}/config/priv_validator_key.json" ]; then
#   echo "cannot find file ${OUT_HOMEDIR}/config/priv_validator_key.json"
#   exit 1
# fi

# echo "Found ${OUT_HOMEDIR}/config/priv_validator_key.json"


if __name__ == '__main__':
    main()
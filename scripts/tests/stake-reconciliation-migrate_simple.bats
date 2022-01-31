#!/usr/bin/env bats

load "node_modules/bats-support/load"
load "node_modules/bats-assert/load"

DIR="$(dirname "$(realpath "$BATS_TEST_FILENAME")")"
GENESIS_IN_PATH="$DIR/test_genesis.json"
GENESIS_OUT_PATH="$DIR/actual_genesis.json"

migrate() {
  fetchd stake-reconciliation-migrate --skip-validate "$GENESIS_IN_PATH" -s "$DIR/test_staked_export.csv" -r "$DIR/test_registrations.json" -o "$GENESIS_OUT_PATH"
}

@test "migration checks that the new account exists (i.e. has a balance)" {
  run migrate
  assert_output --partial "\"fetch1yq8xg4whn2mzdjpuafn78spz2ppzht2vvcemay\" ineligible for reason: unable to find new account with address \"fetch1hffw4sfztgdud2z4eldq9h86cg8rcmhmxf3mc3\""

  OLD_ACCOUNT_BALANCE="$(jq -r ".app_state.bank.balances[5].coins[0].amount" "$GENESIS_OUT_PATH")"
  assert_equal "$OLD_ACCOUNT_BALANCE" "60000000000000000"
}

@test "migration checks that the old account has a non-zero sequence number" {
  run migrate

  assert_output --partial "\"fetch13rhthqhve78m2rn3uzd55f4m72cjulv5xcl9hn\" ineligible for reason: sequence number must be 0"

  OLD_ACCOUNT_BALANCE=$(jq -r ".app_state.bank.balances[4].coins[0].amount" "$GENESIS_OUT_PATH")
  assert_equal "$OLD_ACCOUNT_BALANCE" "50000000000000000"

  NEW_ACCOUNT_BALANCE=$(jq -r ".app_state.bank.balances[10].coins[0].amount" "$GENESIS_OUT_PATH")
  assert_equal "$NEW_ACCOUNT_BALANCE" "505555555555555555"
}

@test "migration checks that the old account balance matches staked export amount" {
  run migrate

  assert_output --partial "\"fetch1ayhm8yfucqvlrknnqcslcz4ll4j003x6z0yvfr\" ineligible for reason: old account balance must match staked export amount"

  OLD_ACCOUNT_BALANCE=$(jq -r ".app_state.bank.balances[3].coins[0].amount" "$GENESIS_OUT_PATH")
  assert_equal "$OLD_ACCOUNT_BALANCE" "40000000000000001"

  NEW_ACCOUNT_BALANCE=$(jq -r ".app_state.bank.balances[9].coins[0].amount" "$GENESIS_OUT_PATH")
  assert_equal "$NEW_ACCOUNT_BALANCE" "404444444444444444"
}

@test "migration updates new account balances and zeros out old migrated account balances" {
  run migrate

  OLD_ACCOUNT_BALANCE_1=$(jq -r ".app_state.bank.balances[0].coins" "$GENESIS_OUT_PATH")
  OLD_ACCOUNT_BALANCE_2=$(jq -r ".app_state.bank.balances[1].coins" "$GENESIS_OUT_PATH")
  assert_equal "$OLD_ACCOUNT_BALANCE_1" "[]"
  assert_equal "$OLD_ACCOUNT_BALANCE_2" "[]"

  NEW_ACCOUNT_BALANCE_1=$(jq -r ".app_state.bank.balances[6].coins[0].amount" "$GENESIS_OUT_PATH")
  NEW_ACCOUNT_BALANCE_2=$(jq -r ".app_state.bank.balances[7].coins[0].amount" "$GENESIS_OUT_PATH")
  assert_equal "$NEW_ACCOUNT_BALANCE_1" "111111111111111111"
  assert_equal "$NEW_ACCOUNT_BALANCE_2" "222222222222222222"
}

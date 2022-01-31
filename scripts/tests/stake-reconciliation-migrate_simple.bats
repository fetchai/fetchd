#!/usr/bin/env bats

load "node_modules/bats-support/load"
load "node_modules/bats-assert/load"

DIR="$(dirname "$(realpath "$BATS_TEST_FILENAME")")"
GENESIS_IN_PATH="$DIR/test_genesis.json"

migrate() {
  fetchd stake-reconciliation-migrate --skip-validate "$GENESIS_IN_PATH" -s "$DIR/test_staked_export.csv" -r "$DIR/test_registrations.json" "$@"
}

@test "migration checks that the new account exists (i.e. has a balance)" {
  run migrate -d

  assert_success
  assert_output --partial "\"fetch1yq8xg4whn2mzdjpuafn78spz2ppzht2vvcemay\" ineligible for reason: unable to find new account with address \"fetch1hffw4sfztgdud2z4eldq9h86cg8rcmhmxf3mc3\""
}

@test "migration checks that the old account has a non-zero sequence number" {
  run migrate -d

  assert_success
  assert_output --partial "\"fetch13rhthqhve78m2rn3uzd55f4m72cjulv5xcl9hn\" ineligible for reason: sequence number must be 0"
}

@test "migration checks that the old account balance matches staked export amount" {
  run migrate -d

  assert_success
  assert_output --partial "\"fetch1ayhm8yfucqvlrknnqcslcz4ll4j003x6z0yvfr\" ineligible for reason: old account balance must match staked export amount"
}

@test "migration updates new account balances and zeros out old migrated account balances" {
  run migrate
  assert_success

  SUM_COINS_MASK='[{"amount": "xxxxxxxxxxxxxxxxxx", "denom": "afet"}]'
  NEW_COINS_MASK='[{"amount": "x0xxxxxxxxxxxxxxxx", "denom": "afet"}]'
  OLD_COINS_MASK='[{"amount": "x0000000000000000", "denom": "afet"}]'
  MISMATCH_COINS_MASK='[{"amount": "x0000000000000001", "denom": "afet"}]'
  EXPECTED_COINS="[
    {
      \"old\": [],
      \"new\": $SUM_COINS_MASK
    },
    {
      \"old\": [],
      \"new\": $SUM_COINS_MASK
    },
    {
      \"old\": $OLD_COINS_MASK,
      \"new\": $NEW_COINS_MASK
    },
    {
      \"old\": $MISMATCH_COINS_MASK,
      \"new\": $NEW_COINS_MASK
    },
    {
      \"old\": $OLD_COINS_MASK,
      \"new\": $NEW_COINS_MASK
    },
    {
      \"old\": $OLD_COINS_MASK,
      \"new\": null
    }
  ]"
  for i in $(seq 0 5); do
  OLD_ACCOUNT_BALANCE=$(jq -r ".app_state.bank.balances[$i].coins" <(echo "$output"))
  assert_equal "$OLD_ACCOUNT_BALANCE" "$(jq ".[$i].old" <(echo "$EXPECTED_COINS") | sed "s,x,$((i+1)),g")"

  # NB: New account for this index does not exist.
  if [ "$i" = 5 ]; then
    continue
  fi

  NEW_ACCOUNT_BALANCE=$(jq -r ".app_state.bank.balances[$((i + 6))].coins" <(echo "$output"))
  assert_equal "$NEW_ACCOUNT_BALANCE" "$(jq ".[$i].new" <(echo "$EXPECTED_COINS") | sed "s,x,$((i+1)),g")"
  done
}

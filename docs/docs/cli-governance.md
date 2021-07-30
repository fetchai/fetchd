# Governance Proposals

In order to change any attribute of a network, a governance proposal must be submitted. This could be a simple poll, a software update or a governing parameter change. 

### Parameter change
This is an example of the process in which network parameters may be changed through the use of a governance proposal.

The values within this code can be changed in order to alter the minimum deposited fund threshold for a proposal to enter the voting phase - alongside the length of the deposit stage in which the minimum deposit threshold must be met.
```
# A JSON file containing the following code should be created to instantiate the proposal.
# The two variables of interest are the "amount" which is set from 10000000stake to 1000stake
# and the "max_deposit_period" which is changed from the default value to 7200000000000
# equal to 2 hours, instead of the standard 2 days (in nanoseconds).

{
  "title": "Deposit Value Proposal",
  "description": "Update min proposal threshold and deposit period",
  "changes": [
    {
      "subspace": "deposit_params",
      "key": "min_deposit",
      "value": [
        {
          "denom":"stake",
          "amount":"1000"
        }
      ]
    },
    {
      "subspace": "deposit_params",
      "key": "max_deposit_period",
      "value": "7200000000000"
    }
  ]
}
```
```
# Create initial proposal by uploading the JSON file
# this is signed by a key 'proposer' that provides a portion of the current threshold deposit
fetchd tx gov submit-proposal --proposal ~/json_path/proposal.json --deposit <deposit_value> --from proposer

# In order to later refer to this proposal, the proposal_id can be determined
fetchd query gov proposals
```

### Proposal voting and querying
After the deposit period has passed, there are two outcomes: either the current minimum threshold is met, or the value is not met and the funds are burned. In the first case this proposal is submitted and then voted on, returning a tally at the end of the voting period.

At any point of the deposit stage, the deposit value can be queried.
```
# This command returns a text representation of the current total deposit value of a proposal
fetchd query gov deposits <proposal_id>

# Other users may contribute to funding the proposal using
fetchd tx gov deposit <proposal_id> <deposit_amount> --from contributer
```

In order to submit a vote on a proposal that has passed into the voting phase, all users except the proposer may do so using this command.
```
# Submit a vote from a key 'voter' with the desired outcome of the voter
fetchd tx gov vote <proposal_id> <yes|no|no_with_veto|abstain> --from voter
```

After this deposit phase, the current voting turnout and tally can be queried, which displays a list of all voters and their choice.
```
# The current voting statistics can be printed using
fetchd query gov votes <proposal_id>
```
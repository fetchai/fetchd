# Governance Proposals

In order to change any attribute of a network, a governance proposal must be submitted. This could be a simple poll, a software update or a governing parameter change. 

### Parameter change
This is an example of the process in which network parameters may be changed through the use of a governance proposal.

The values within this code can be changed in order to alter the minimum deposited fund threshold for a proposal to enter the voting phase, and the length of the deposit stage in which the minimum deposit threshold must be met.
```
# A JSON file containing the following code should be created to instantiate the proposal.
# The two variables of interest are the "amount" which is set from 10000000stake to 1000stake
# and the "max_deposit_period" which is changed from the default value to 7200000000000
# equal to 2 hours, instead of the standard 2 days (in nanoseconds).

{
  "title": "Staking Param Change",
  "description": "Update max validators",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 105
    }
  ],
  "deposit": "1000000000000000000atestfet"
}
```
```
# Create initial proposal by uploading the JSON file
# this is signed by a key 'proposer' that provides a portion of the current threshold deposit
fetchd tx gov submit-proposal --proposal ~/json_path/proposal.json --from proposer

# In order to later refer to this proposal, the proposal_id can be determined
fetchd query gov proposals
```

### Proposal deposit phase
The characteristics of the deposit phase are described by a set of network governance parameters, where the deposit period is two days from the initial proposal deposit until expiration, and a minimum threshold of 10000000denom as default. The minimum threshold must be met during this deposit period in order to proceed to the voting phase. The proposer may provide all of this threshold, or just some. In which case, supporters of the proposal may donate additional funding towards the goal of meeting the threshold.

At any point of the deposit stage, the deposit pot can be queried.

```
# To get the proposal ID, use the txhash obtained when the proposal was submitted and run the following command:
fetchd query tx <txhash>

# This command returns a text representation of the current total deposit value of a proposal
fetchd query gov deposits <proposal_id>

# Other users may contribute to funding the proposal using
fetchd tx gov deposit <proposal_id> <deposit_amount> --from contributer
```

[This](https://docs.cosmos.network/master/modules/gov/01_concepts.html#proposal-submission) documentation provides a more detailed explanation of the deposit funding period.

### Proposal voting and querying
After the deposit period has passed, there are two outcomes: either the current minimum threshold is met, or the value is not met and the funds are returned. In the first case this proposal is submitted and to be voted on, returning a tally at the end of the voting period. 

In order to submit a vote on a proposal that has passed into the voting phase, all staked users except the proposer may do so using this command.
```
# Submit a vote from a key 'voter' with the desired outcome of the voter
fetchd tx gov vote <proposal_id> <yes|no|no_with_veto|abstain> --from voter
```

The current voting turnout and tally can be queried, which displays a list of all voters and their choice.
```
# The current voting statistics can be printed using
fetchd query gov votes <proposal_id>
```

#### Example output
```
votes:
- option: VOTE_OPTION_YES
  proposal_id: "1"
  voter: fetch1dmehhhvul8y7slqs3zu2z3fede9kzlnyupd9rr
- option: VOTE_OPTION_NO
  proposal_id: "1"
  voter: fetch1064endj5ne5e868asnf0encctwlga4y2jf3h28
- option: VOTE_OPTION_YES
  proposal_id: "1"
  voter: fetch1k3ee923osju93jm03fkfmewnal39fjdbakje1x
```

### Voting outcome
After the voting period has ended, the results are used to determine the next step of the proposal. The potential outcomes include:

- **Majority *yes* vote**
	-	The proposal passes through and the users act according to the proposal type - e.g. A Software update proposal passes, and users begin uptake of the new version
- **Majority *no* vote**
	- The funds deposited to pass into the voting stage are returned, and there is no governance change

- **Majority *no_with_veto* vote**
	- This outcome is indicative of a proposal which may undermine the current governance system, e.g. a proposal to set the deposit threshold or voting period to an absurd value
	- All funds deposited in the proposal are to be burned subject to this outcome, and there is no governance change
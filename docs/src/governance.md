# Governance

In order to be able to take part in the governance you either need to be running a full validator node or you need to have have delegated stake to an existing validator

## Stake Delegation

In order to delegate stake to a validator the following command should be used:

```bash
fetchcli tx staking delegate <VALOPER_ADDRESS> <AMOUNT> --from <KEY_NAME>
```

Where the `<VALOPER_ADDRESS>` begins with the prefix `fetchvaloper1...` and the `<AMOUNT>` field contains the currency denomination. For example:

```bash
fetchcli tx staking delegate fetchvaloper1cct4fhhksplu9m9wjljuthjqhjj93z0s97p3g7 1000atestfet --from agent
```

## Proposals Overview

There are three types of proposal:

- **Text Proposals**: These are the most basic type of proposal. They can be used to get the opinion from participants of the network on a given topic.
- **Parameter Proposals**: These proposals are used to update the value of an existing software parameter of the network.
- **Software Upgrade Proposals**: These are used to propose an upgrade of the `fetchd` software, particularly in cases where the software changes might not necessary be backwards compatible or in some way present a major update to the network.

## The Proposal Process

Any FET holder can submit a proposal. In order for the proposal to be open for voting, it needs to come with a deposit that is greater than a parameter called *minDeposit*. The deposit need not be provided in its entirety by the submitter. If the initial proposer's deposit is not sufficient, the proposal enters the **deposit period** status. Then, any FET holder can increase the deposit by sending a *depositTx* transaction to the network.

Once the deposit reaches *minDeposit*, the proposal enters the **voting period**, which lasts 2 weeks. Any bonded FET holder can then cast a vote on this proposal. The user has the following options for voting:

* Yes
* No
* NoWithVeto
* Abstain

At the end of the voting period, the proposal is accepted if there are more than 50% Yes votes (excluding Abstain votes) and less than 33.33% of NoWithVeto votes (excluding Abstain votes).


## Generating Proposals

When creating a proposal, the user will create a proposal JSON file with all the relevant information. An example of a text proposal is shown below:

```json
{
  "title": "Switch to semantic commit messages for fetchd",
  "description": "This proposal is advocating a switch to sematic commit messages\nYou can find the full discussion here: https://github.com/fetchai/fetchd/issues/231",
  "type": "Text",
  "deposit": "10000000000000000000atestfet"
}
```

It is always recommended that the description of a text proposal has a link to a Github issue with the full proposal text along with the discussions about it.

Once the user has created the JSON file, to generate the text propsal on chain run the following command:

`fetchcli tx gov submit-proposal --proposal proposal.json --from <name of signing key>`

## Increasing the deposit for a proposal

If a user wants to increase the deposit of a proposal they would run the following command:

`fetchcli tx gov deposit <proposalID> 100atestfet --from <key name>`

For example:

`fetchcli tx gov deposit 2 100atestfet --from validator`

## Listing current proposals

Current proposals are visible from the block explorer and using the CLI.

To get the list of current proposals and their corresponding *proposal-ids* the run the following command:

`fetchcli query gov proposals`

## Voting on a proposal

To vote for a proposal run the following command

`fetchcli tx gov vote <proposalID> <option> --from <delegatorKeyName>`

For example:

`fetchcli tx gov vote 5 yes --from validator`

<div class="admonition note">
  <p class="admonition-title">Note</p>
  <p>When using CLI commands make sure that your CLI is pointing at the correct network. See the <a href="../cli-introduction/">CLI introduction documentation</a> for more details</p>
</div>

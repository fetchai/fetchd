# Validators Overview

## Introduction

The Fetch.ai Ledger relies on a set of validators that are responsible for committing new blocks in the blockchain. These validators participate in the consensus protocol by broadcasting votes which contain cryptographic signatures signed by each validator's private key.

Validator candidates can bond their own FET and have FET delegated, or staked, to them by token holders. The validators are determined by who has the most stake delegated to them. The top N validator candidates with the most stake will become the active validators.

Validators and their delegators will earn FET as block provisions and tokens as transaction fees through execution of the consensus protocol. Transaction fees will be paid in FET.

If validators double sign, are frequently offline or do not participate in governance, their staked FET (including FET of users that delegated to them) can be slashed. The penalty depends on the severity of the violation.

## Validator benefits

Validators will produce blocks on the Fetch.ai mainnet, in exchange for rewards, but also experience a number of other benefits:

- Greater say in the future direction of the protocol through their role in governance.
- Reputation as a valued member of the community.
- Faster submission of transactions and direct access to the network operation.
- Access to more and faster information on the current state of the network.
- Access to other business models (as they emerge, including oracles or hosted agent applications).

## Validator Revenue

Similarly to most other Proof-of-Stake blockchains, the Fetch ledger provides revenue to validators in the form of block rewards and transaction fees. Smaller holders of FET tokens are able to delegate their stake to validators in exchange for a share of the rewards, as determined by the validator.

- **Block rewards (FET)**: Rewards are provided for every block that is produced by a validator. The inflation rate is set at an annual rate of 3% during the first three years of the networks’ operation.

- **Transaction Fees**: All transactions that are submitted to the chain are charged a transaction fee denominated in FET. This is payable for simple transfers, smart contract deployments, contract calls, governance and any other type of transaction. The usage of state and computational resources involves a fee proportional to the quantity of resources used.

- **Agent and ML services**: Validators also an opportunity to support AI and agent-based services on the Fetch.ai blockchain. These include hosting datasets for machine learning applications, operating agent search-and-discovery services and operating agent networks that provide oracle services.

## Who typically acts as a validator in a network?

- Community members who are enthusiastic about the Fetch.ai protocol.
- Those wishing to take part directly in governance or network operations.
- Professional staking operations.
- Blockchain founders and developers.

## What are the costs?

It costs time and money to operate a node. If your node is offline (unavailable) some of its stake will be slashed (i.e. deducted).

## Hardware

The hardware resources for running a validator node largely depend on the network load. As a recommended configuration we suggest the following requirements

- 2 x CPU, either Intel or AMD, with the SSE4.1, SSE4.2 and AVX flags (use lscpu to verify)
- 8 GB RAM
- 500 GB SSD
- 100 Mbit/s always-on internet connection
- Linux OS (Ubuntu 18.04 or 20.04 recommended) / MacOS

Uptime in incredibly important for being a validator. It is expected that validators will have appriopriate redundancies for compute, power, connectivity etc. While the blockchain itself it highly replicated it is also expected that validators will perform local storage backups in order to minimise validator down time.

## Set Up a Website

Set up a dedicated validator's website and signal your intention to become a validator on our [Discord](https://discord.gg/UDzpBFa) server. This is important since delegators will want to have information about the entity they are delegating their FET to.

Strictly speaking this is not necessary, however, it is recommended. As a validator on the network you will want to get other community users to delegate stake to your validator. The more combined stake that a validate has then the great share of the block rewards they will take.

## Seek Legal Advice

Seek legal advice if you intend to run a Validator.

## Community

We highly recommdend to check out the validator community on the discord channel for more information and to see that latest announcements about becoming a validator.

* [Discord](https://discord.gg/UDzpBFa)

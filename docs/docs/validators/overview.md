# Validators Overview

## Introduction

The Fetch.ai Ledger relies on a set of validators that are responsible for committing new blocks in the blockchain. These validators participate in the consensus protocol by broadcasting votes which contain cryptographic signatures signed by each validator's private key.

Validator candidates can bond their own FET and have FET delegated, or staked, to them by token holders. The validators are determined by who has the most stake delegated to them. The top N validator candidates with the most stake will become the active validators.

Validators and their delegators will earn FET as block provisions and tokens as transaction fees through execution of the consensus protocol. Transaction fees will be paid in FET.

If validators double sign, are frequently offline or do not participate in governance, their staked FET (including FET of users that delegated to them) can be slashed. The penalty depends on the severity of the violation.

## Hardware

The hardware resources for running a validator node largely depend on the network load. As a recommended configuration we suggest the following requirements

- 2 x CPU, either Intel or AMD, with the SSE4.1, SSE4.2 and AVX flags (use lscpu to verify)
- 4 GB RAM
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

# Mainnet V2 is here

Fetchhub, our mainnet v2 is here! You will find all information to setup your client and connect on the [network](./networks/) and [cli-introduction](./cli-introduction/) pages. Previous test networks are still available as well for your tests and experiments.

This documentation covers some of the things you need to know in order to prepare yourself and develop for this new network:

## Test Networks

The starting point for most developers will be our [test network](https://explore-stargateworld.fetch.ai), since this is primarily used for agent development and testing. The test net is **fully supported** by our Agent Framework from version 0.6 and above. If you are developing autonomous economic agents using our latest framework, then moving to the newest test net is relatively straightforward, unless smart contracts are involved.

The newest test net has a new address format. Here is a typical example:

```text
fetch1almpjpf769p23k0v4m5eglvzr4jupsjs66vxf4
```

<div class="admonition note">
  <p class="admonition-title">Note</p>
  <p>These new addresses are not compatible with mainnet v1 or testnet v1. It is expected that users will create new address for agents on the network</p>
</div>

## Key links and information

Block explorer and token tap: [https://explore-agent-land.fetch.ai/](https://explore-stargateworld.fetch.ai)

Understanding building block relationships: [https://docs.fetch.ai/aea/oef-ledger/](https://docs.fetch.ai/aea/oef-ledger/)

For more detailed information, have a look at our [network](./networks/) information page.

## Roadmap for Smart Contracts

The v2.0 main-net and our test-nets support a webassembly (WASM) virtual machine VM. When mainnet V2 is released, smart contracts will be developed using Cosmwasm, which will include the ability to develop your contracts in a variety of languages depending on your preferences, such as Rust, Go or Javascript. For more information on Cosmwasm, and Rust, which is the first language to be supported by the Cosmwasm VM, you can go to [https://www.cosmwasm.com/](https://www.cosmwasm.com/) and [https://github.com/CosmWasm/cosmwasm-template](https://github.com/CosmWasm/cosmwasm-template).

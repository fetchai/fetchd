# CLI - Managing Keys

Managing your keys is an essential part of working with the Ledger, since all interactions are authenticated with these keys.

## Adding keys

To create a new local key you need run the following command:

```bash
fetchcli keys add <your_key_name>
```

<div class="admonition note">
  <p class="admonition-title">Note</p>
  <p>These keys are stored locally on your system. By default, these keys will be stored in the OS level keychain, however, in general these keys are considered less secure than using a hardware device</p>
</div>

After running the command `fetchcli` will print out a summary of the new key. An example of this output is shown below:

```text
- name: test
  type: local
  address: fetch142tawq2rj397mctc3jtw9dfzf03ns0ze4swat0
  pubkey: fetchpub1addwnpepqvtmze0ekffynnjx9n85g6sexzl49ze2vpgc2f52fteyyghjtvvqw682nkx
  mnemonic: ""
  threshold: 0
  pubkeys: []
```

This will be followed by a 24-word mnemonic that can be used to re-generate the private key and address for the account (keep this safe, if ever used to control main-net tokens).

## Looking up an address

A common operation that you will want to do is to lookup the address for a specified key. This can be done quickly using the following command:

```bash
fetchcli keys show -a <name of key>
```

An example of the expected output is shown below:

```bash
fetch142tawq2rj397mctc3jtw9dfzf03ns0ze4swat0
```

A less common operation, but still useful, would be to lookup the public key for a specified key. The can be achieved with the following command:

```bash
fetchcli keys show -p <name of the key>
```

An example of the expected output is shown below:

```bash
fetchpub1addwnpepqvtmze0ekffynnjx9n85g6sexzl49ze2vpgc2f52fteyyghjtvvqw682nkx
```

## Listing keys

To lookup more detailed information for all keys on your system use the following command:

```bash
fetchcli keys list
```

This will output all of your keys information in a yaml format that is similar to the one generated when you first created the key.

```bash
- name: test
  type: local
  address: fetch142tawq2rj397mctc3jtw9dfzf03ns0ze4swat0
  pubkey: fetchpub1addwnpepqvtmze0ekffynnjx9n85g6sexzl49ze2vpgc2f52fteyyghjtvvqw682nkx
  mnemonic: ""
  threshold: 0
  pubkeys: []
```

## Hardware Wallets

### Setup

We recommend hardware wallets as a solution for managing private keys. The Fetch v2.0 ledger is compatible with Ledger Nano devices. To use your Ledger Nano you will need to complete the following steps:

1. Set-up your wallet by creating a PIN and passphrase, which must be stored securely to enable recovery if the device is lost or damaged.
2. Connect your device to your PC and update the firmware to the latest version using the Ledger Live application.
3. Install the Cosmos application using the software manager (Manager > Cosmos > Install).

### Adding a new key

In order to use the hardware wallet address with the cli, the user must first add it via `fetchcli`. This process only records the public information about the key.

To import the key first plug in the device and enter the device pin. Once you have unlocked the device navigate to the Cosmos app on the device and open it.

To add the key use the following command:

```bash
fetchcli keys add <name for the key> --ledger --index 0
```

<div class="admonition note">
  <p class="admonition-title">Note</p>
  <p>The <code>--ledger</code> flag tells the command line tool to talk to the ledger device and the <code>--index</code> flag selects which HD index should be used.</p>
</div>

When running this command, the Ledger device will prompt you to verify the genereated address. Once you have done this you will get an output in the following form:

```bash
- name: hw1
  type: ledger
  address: fetch1xqqftqp8ranv2taxsx8h594xprfw3qxl7j3ra2
  pubkey: fetchpub1addwnpepq2dulyd9mly3xqnvfgdsjkqlqzsxldpdhd6cnpm67sx90zhfw2ragk9my5h
  mnemonic: ""
  threshold: 0
  pubkeys: []
```

### Importing a private key generated from the Agent Framework

It is also possible to import private keys from the agent framework. For example the private key that you generated for your agent as part of the AW-1 challenge.

To import this key into the local keychain simply run the following command:

```bash
fetchcli keys import --agent-raw-key <name of the new key> <path to the key file>
```

For example:

```bash
fetchcli keys import --agent-raw-key agent /home/user/fetchai_private_key.txt
```

<div class="admonition note">
  <p class="admonition-title">Note</p>
  <p>The <code>--agent-raw-key</code> flag is very important here. Also note that this flag only allows the user to import scep265k1 private keys</p>
</div>

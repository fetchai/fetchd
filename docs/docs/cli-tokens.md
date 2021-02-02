# CLI - Managing Tokens

## Querying your balance

Once the wasm is configured for the desired [network](../cli-introduction/). The user can query there balance using the following command:

```bash
fetchcli query account fetch1akvyhle79nts4rwn075t85xrwmp5ysuqynxcn4
```

If the address exists on the network then the user will expect to see an output in the following form:

```text
  address: fetch1akvyhle79nts4rwn075t85xrwmp5ysuqynxcn4
  coins:
  - denom: atestfet
    amount: "1000000000000000000"
  public_key: ""
  account_number: 20472
  sequence: 0
```


## Sending funds

To send funds from one address to another address then you would use the `tx send` subcommand. As shown below:

```bash
./build/fetchcli tx send <from address or key name> <target address> <amount>
```

In a more concrete example if the user wanted to send `100atestfet` from `main` key/address to `fetch106vm9q6ezu9va7v7e0cvq0nedc54egjm692fcp` then the following command would be used.

```bash
./build/fetchcli tx send main fetch106vm9q6ezu9va7v7e0cvq0nedc54egjm692fcp 100atestfet
```

When you run the command you will get a similar output and prompt. The user can check the details of the transfer and then press 'y' to confirm the transfer.

```text
{"chain_id":"agent-land","account_number":"20472","sequence":"0","fee":{"amount":[],"gas":"200000"},"msgs":[{"type":"cosmos-sdk/MsgSend","value":{"from_address":"fetch1akvyhle79nts4rwn075t85xrwmp5ysuqynxcn4","to_address":"fetch106vm9q6ezu9va7v7e0cvq0nedc54egjm692fcp","amount":[{"denom":"atestfet","amount":"100"}]}}],"memo":""}

confirm transaction before signing and broadcasting [y/N]: y
```

Once the transfer has been made a summary is presented to the user. An example is shown below:

```text
height: 0
txhash: CA7C2C842F8F577E9621C2B23A016D93B979AC1A45015807799C5AD959503FA4
codespace: ""
code: 0
data: ""
rawlog: '[]'
logs: []
info: ""
gaswanted: 0
gasused: 0
tx: null
timestamp: ""
```
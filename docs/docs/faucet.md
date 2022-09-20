# Token Faucet

For our test networks, we have a simple token faucet implemented to allow users of the network to get started quickly. You can send it an account address, and it will transfer some test token on it.

Token Faucets are network specific, depending on the network type they may or may not be deployed. Please check the [networks](../live-networks/) page for specific details.

The Token Faucet itself is available from the network block explorer (`GET FUNDS` button on the homepage).

Enter your `fetch...` address in the popup and click *Add funds* button. Wait a few blocks for the transaction to be processed, and you should see it appear along with some funds on your account.

## Add funds to Wallet using faucet APIs:

You can also request and get testnet tokens in your wallet using the APIs.

### Get some atestfet

```bash
curl -X POST -H 'Content-Type: application/json' -d '{"address":"<address>"}' https://faucet-dorado.fetch.ai/api/v3/claims
```

### Get some nanomobx

```bash
curl -X POST -H 'Content-Type: application/json' -d '{"address":"<address>"}' https://faucet-mobx-dorado.fetch.ai/api/v3/claims
```

### Get some ulrn

```bash
curl -X POST -H 'Content-Type: application/json' -d '{"address":"<address>"}' https://faucet-lrn-dorado.fetch.ai/api/v3/claims
```

### Sample response for fund request to faucet

```text
{"status":"ok","uuid":"<uuid>","target":"<address>"}
```

### Check the wallet balance

```bash
fetchd query bank balances <address>
```

```text
balances:
- amount: "<balance>"
  denom: atestfet
pagination:
  next_key: null
  total: "0"
```

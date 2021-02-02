## Python examples using Agent Land

In this section, we look at some Python that you can use to create addresses, fund them, and submit transactions to the network. It's super-simple to do:

```python3
from aea.crypto.fetchai import FetchAICrypto, FetchAIApi, FetchAIFaucetApi

# Get a new address:
fetch_crypto = FetchAICrypto()
address = fetch_crypto

# This is relatively slow, so if you need to stock up a number of
# addresses, it’s best to get it ONCE, then distribute yourself:
# for rapid tests:
FetchAIFaucetApi().get_wealth(address)
balance = FetchAIApi().get_balance(address)
print(f”Our address {address} has a balance of {balance}”)
```

As you can see, it’s pretty straightforward. The above works in standalone code, and is supported in the agent framework. Here is a code fragment that shows the construction, signing and submission of a transaction:

```python3
from aea.crypto.fetchai import FetchAICrypto, FetchAIApi, FetchAIFaucetApi

# ... code, etc...

def test_construct_sign_and_submit_transfer_transaction():
    """Test the construction, signing and submitting of a transfer transaction."""
    account = FetchAICrypto()
    balance = get_wealth(account.address)
    assert balance > 0, "Failed to fund account."
    fc2 = FetchAICrypto()
    fetchai_api = FetchAIApi(**FETCHAI_TESTNET_CONFIG)
    amount = 10000
    assert amount < balance, "Not enough funds."
    transfer_transaction = fetchai_api.get_transfer_transaction(
        sender_address=account.address,
        destination_address=fc2.address,
        amount=amount,
        tx_fee=1000,
        tx_nonce="something",
    )
    assert (
        isinstance(transfer_transaction, dict) and len(transfer_transaction) == 6
    ), "Incorrect transfer_transaction constructed."
    signed_transaction = account.sign_transaction(transfer_transaction)
    assert (
        isinstance(signed_transaction, dict)
        and len(signed_transaction["tx"]) == 4
        and isinstance(signed_transaction["tx"]["signatures"], list)
    ), "Incorrect signed_transaction constructed."
    transaction_digest = fetchai_api.send_signed_transaction(signed_transaction)
    assert transaction_digest is not None, "Failed to submit transfer transaction!"

        # Now let's wait around for a while for this transaction to go through"
    not_settled = True
    elapsed_time = 0
    while not_settled and elapsed_time < 20:
        elapsed_time += 1
        time.sleep(2)
        transaction_receipt = fetchai_api.get_transaction_receipt(transaction_digest)
        if transaction_receipt is None:
            continue
        is_settled = fetchai_api.is_transaction_settled(transaction_receipt)
        not_settled = not is_settled
    assert transaction_receipt is not None, "Failed to retrieve transaction receipt."
    assert is_settled, "Failed to verify tx!"
    tx = fetchai_api.get_transaction(transaction_digest)
    is_valid = fetchai_api.is_transaction_valid(
        tx, fc2.address, account.address, "", amount
    )
    assert is_valid, "Failed to settle tx correctly!"
    assert tx == transaction_receipt, "Should be same!"
```
# Block Explorer

Each of the [networks](./networks/) has a dedicated block explorer web site associated with it. This is a useful tool for monitoring network activity.

## Logging in with the Ledger Nano

Return to the block explorer landing page and click on the key button in the top right corner. You'll then be prompted to "Sign in With Ledger". You must accept this request on your ledger nano device. After completing this process, the key button will be replaced by a person icon with a link to your personal address page, which keeps track of the activity that you have performed on the test-net.

## Getting Testnet Tokens from the Faucet

For networks that support it, you can obtain tokens for your account by copying the address and pasting it into the token faucet. Then, return to the main page, press the "Get Funds" button and paste your address in the pop-up. Afterwards you can return to your address page (via the person icon) and should observe that you have been allocated 1 TESTFET.

## Transferring Tokens to another Address

After receiving tokens, you can send these to another address using the purple Transfer button on your address page. This will trigger a pop-up that prompts you to specify the destination address and the amount you wish to transfer. After filling in this information, you will be asked to sign the transaction using your ledger nano. The confirmation that the transaction has been broadcast gives two links that can be used to check that the transaction has been executed on the blockchain using either the transaction hash or your account page.

<div class="admonition note">
  <p class="admonition-title"><b>Note:</b></p>
  <p>The transaction format includes a memo field that can be used to check the transaction information on the ledger nano display.</p>
</div>

## Delegating Stake to a Validator

You can delegate your test-net tokens to a validator who is operating the network by clicking on the _Validators_ tab, and selecting one of the validators that you wish to delegate stake towards. In the _Voting Power_ panel there is an option to `DELEGATE` tokens. Pressing this button will trigger a pop-up that prompts you to select a delegation amount and then sign the transaction with your Ledger Nano device.  

After delegating tokens, buttons labelled with `REDELEGATE` and `UNDELEGATE` will appear. The delegation of tokens to a validator provides you with a reward for helping to secure the network. It is also possible to delegate your tokens to a different validator using a `REDELEGATE` transaction. You can return any bonded tokens to your address by submitting an `UNDELEGATE` request, which will trigger the tokens to be returned after 21 days have elapsed. The rewards that you receive from delegating tokens to a validator are shown in the account page. These can be sent to your address by sending a `WITHDRAW` transaction. 
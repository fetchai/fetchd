# Messages

In this section we describe the processing of the staking messages and the corresponding updates to the state. All created/modified state objects specified by each message are defined within the [state](./02_state_transitions.md) section.

## MsgCreateVerifiableCredential

A verifiable credential is created using the `MsgCreateVerifiableCredential` service message.

+++ https://github.com/allinbits/cosmos-cash/blob/main/proto/verifiable-credential-service/tx.proto#L14

+++ https://github.com/allinbits/cosmos-cash/blob/main/proto/verifiable-credential-service/tx.proto#L18-L28

This service message is expected to fail if:

- another verifiable credential with the same id is already registered

This service message creates and stores the `VerifiableCredential` object at appropriate indexes.



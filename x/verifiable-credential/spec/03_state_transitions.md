# State Transitions

This document describes the state transitions pertaining to:

1. [VerifiableCredential](./02_state.md#verifiable)
2. [CredentialSubject](./02_state.md#credentialsubject)
3. [Proof](./02_state.md#proof)

## VerifiableCredential 

### Setting a decentralized identifier (DID) in the store
+++ https://github.com/allinbits/cosmos-cash/blob/main/x/verifiable-credential-service/keeper/verifiable-credential.go#L8-L10

### Adding an identifier
+++ https://github.com/allinbits/cosmos-cash/blob/main/x/verifiable-credential-service/keeper/msg_server.go#L23-L45

## CredentialSubject
CredentialSubjects are to a VerifiableCredential DidDocuments, when the credential is created.

## Proof
An issuer creates and attach a proof to a verifiable credential when the VerifiableCredential data structure is created.


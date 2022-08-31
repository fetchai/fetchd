# State

This document describes the state pertaining to:

1. [VerifiableCredential](./02_state.md#verifiable)
2. [CredentialSubject](./02_state.md#credentialsubject)
3. [Proof](./02_state.md#proof)


Three data structues represent a VerifiableCredential:

- VerifiableCredential
- CredentialSubject
- Proof

## VerifiableCredential
VerifiableCredentials are stored in the state under the 0x61 key and are stored using their ids. VerifiableCredentials allow credentials to be issued to users. [[more_info]](https://www.w3.org/TR/vc-data-model/)

- VerifiableCredential: `0x61 | VerifiableCredential.Id -> ProtocolBuffer(VerifiableCredential)`

### Structure
+++ https://github.com/allinbits/cosmos-cash/blob/main/proto/verifiable-credential-service/verifiable-credential.proto#L44-L52


## CredentialSubject
CredentialSubject is stored as a field under in the VerifiableCredential data structure. A CredentialSubject is used to attach abritary data to a service. [[more_info]](https://www.w3.org/TR/vc-data-model/#credential-subject)

### Structure
+++ https://github.com/allinbits/cosmos-cash/blob/main/proto/verifiable-credential-service/verifiable-credential.proto#L54-L57

## Proof
A Proof is stored as as a field under in the VerifiableCredential data structure. Proofs are used to validate that the credential was issued by the correct issuer. [[more_info]](https://www.w3.org/TR/vc-data-model/#proofs-signatures)

### Structure
+++ https://github.com/allinbits/cosmos-cash/blob/main/proto/verifiable-credential-service/verifiable-credential.proto#L59-L65

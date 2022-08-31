package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/verifiable-credential module sentinel errors
var (
	ErrVerifiableCredentialNotFound = sdkerrors.Register(ModuleName, 1100, "vc not found")
	ErrVerifiableCredentialFound    = sdkerrors.Register(ModuleName, 1101, "vc found")
	ErrDidDocumentDoesNotExist      = sdkerrors.Register(ModuleName, 1102, "did does not exist in the store")
	ErrVerifiableCredentialIssuer   = sdkerrors.Register(ModuleName, 1103, "provided verifiable credential and did public key do not match")
	ErrMessageSigner                = sdkerrors.Register(ModuleName, 1104, "message signer does not match provided did")
)

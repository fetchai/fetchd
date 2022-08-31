package types

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// VerificationMaterialType encode the verification material type
type VerificationMaterialType string

// Verification method material types
const (
	DIDVMethodTypeEcdsaSecp256k1VerificationKey2019 VerificationMaterialType = "EcdsaSecp256k1VerificationKey2019"
	DIDVMethodTypeEd25519VerificationKey2018        VerificationMaterialType = "Ed25519VerificationKey2018"
	DIDVMethodTypeCosmosAccountAddress              VerificationMaterialType = "CosmosAccountAddress"
	DIDVMethodTypeX25519KeyAgreementKey2019         VerificationMaterialType = "X25519KeyAgreementKey2019"
)

// String return string name for the Verification Method type
func (p VerificationMaterialType) String() string {
	return string(p)
}

type VerificationMaterial interface {
	EncodeToString() string
	Type() VerificationMaterialType
}

// BlockchainAccountID formats an account address as per the CAIP-10 Account ID specification.
// https://w3c.github.io/did-spec-registries/#blockchainaccountid
// https://github.com/ChainAgnostic/CAIPs/blob/master/CAIPs/caip-10.md
type BlockchainAccountID string

// EncodeToString returns the string representation of a blockchain account id
func (baID BlockchainAccountID) EncodeToString() string {
	return string(baID)
}

// Type returns the string representation of a blockchain account id
func (baID BlockchainAccountID) Type() VerificationMaterialType {
	return DIDVMethodTypeCosmosAccountAddress
}

// MatchAddress check if a blockchain id address matches another address
// the match ignore the chain ID
func (baID BlockchainAccountID) MatchAddress(address string) bool {
	return baID.GetAddress() == address
}

// GetAddress get the address from a blockchain account id
// TODO: this function shall return an error for invalid addresses
func (baID BlockchainAccountID) GetAddress() string {
	addrStart := strings.LastIndex(string(baID), ":")
	if addrStart < 0 {
		return ""
	}
	return string(baID)[addrStart+1:]
}

// NewBlockchainAccountID build a new blockchain account ID struct
func NewBlockchainAccountID(chainID, account string) BlockchainAccountID {
	return BlockchainAccountID(fmt.Sprint("cosmos:", chainID, ":", account))
}

// PublicKeyMultibase formats an account address as per the CAIP-10 Account ID specification.
// https://w3c.github.io/did-spec-registries/#publickeymultibase
// https://datatracker.ietf.org/doc/html/draft-multiformats-multibase-03#appendix-B.1
type PublicKeyMultibase struct {
	data   []byte
	vmType VerificationMaterialType
}

// EncodeToString returns the string representation of the key in hex format. F is the hex format prefix
// https://datatracker.ietf.org/doc/html/draft-multiformats-multibase-03#appendix-B.1
func (pkh PublicKeyMultibase) EncodeToString() string {
	return string(fmt.Sprint("F", hex.EncodeToString(pkh.data)))
}

// Type the verification material type
// https://datatracker.ietf.org/doc/html/draft-multiformats-multibase-03#appendix-B.1
func (pkh PublicKeyMultibase) Type() VerificationMaterialType {
	return pkh.vmType
}

// NewPublicKeyMultibase build a new blockchain account ID struct
func NewPublicKeyMultibase(pubKey []byte, vmType VerificationMaterialType) PublicKeyMultibase {
	return PublicKeyMultibase{
		data:   pubKey,
		vmType: vmType,
	}
}

// NewPublicKeyMultibaseFromHex build a new blockchain account ID struct
func NewPublicKeyMultibaseFromHex(pubKeyHex string, vmType VerificationMaterialType) (pkm PublicKeyMultibase, err error) {
	pkb, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return
	}
	// TODO: shall we check if it is conform to the verification material? probably
	pkm = PublicKeyMultibase{
		data:   pkb,
		vmType: vmType,
	}
	return
}

// PublicKeyHex formats an account public key as hex string.
// https://w3c.github.io/did-spec-registries/#publickeyhex
// Note that this property is deprecated in favor of publicKeyMultibase or publicKeyJwk,
// but is maintained for compatibility with legacy implementations
type PublicKeyHex struct {
	data   []byte
	vmType VerificationMaterialType
}

// EncodeToString returns the string representation of the key in hex format.
// https://datatracker.ietf.org/doc/html/draft-multiformats-multibase-03#appendix-B.1
func (pkh PublicKeyHex) EncodeToString() string {
	return string(hex.EncodeToString(pkh.data))
}

// Type the verification material type
// https://datatracker.ietf.org/doc/html/draft-multiformats-multibase-03#appendix-B.1
func (pkh PublicKeyHex) Type() VerificationMaterialType {
	return pkh.vmType
}

// NewPublicKeyHex build a new public key hex struct
func NewPublicKeyHex(pubKey []byte, vmType VerificationMaterialType) PublicKeyHex {
	return PublicKeyHex{
		data:   pubKey,
		vmType: vmType,
	}
}

// NewPublicKeyHexFromString build a new blockchain account ID struct
func NewPublicKeyHexFromString(pubKeyHex string, vmType VerificationMaterialType) (pkh PublicKeyHex, err error) {
	pkb, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return
	}
	// TODO: shall we check if it is conform to the verification material? probably
	pkh = PublicKeyHex{
		data:   pkb,
		vmType: vmType,
	}
	return
}

//// PublicKeyJwk formats an account public key as hex string.
//// https://w3c.github.io/did-spec-registries/#publickeyhex
//// Note that this property is deprecated in favor of publicKeyMultibase or publicKeyJwk,
//// but is maintained for compatibility with legacy implementations
//type PublicKeyJwk struct {
//	data   []byte
//	vmType VerificationMaterialType
//}
//
//// EncodeToString returns the string representation of the key in hex format.
//// https://datatracker.ietf.org/doc/html/draft-multiformats-multibase-03#appendix-B.1
//func (pkh PublicKeyJwk) EncodeToString() string {
//	return string(hex.EncodeToString(pkh.data))
//}
//
//// Type the verification material type
//// https://datatracker.ietf.org/doc/html/draft-multiformats-multibase-03#appendix-B.1
//func (pkh PublicKeyJwk) Type() VerificationMaterialType {
//	return pkh.vmType
//}
//
//// NewPublicKeyJwk build a new public key hex struct
//func NewPublicKeyJwk(pubKey []byte, vmType VerificationMaterialType) PublicKeyJwk {
//	return PublicKeyJwk{
//		data:   pubKey,
//		vmType: vmType,
//	}
//}
//
//// NewPublicKeyJwkFromString build a new blockchain account ID struct
//func NewPublicKeyJwkFromString(pubKeyHex string, vmType VerificationMaterialType) (pkh PublicKeyJwk, err error) {
//	pkb, err := hex.DecodeString(pubKeyHex)
//	if err != nil {
//		return
//	}
//	// TODO: shall we check if it is conform to the verification material? probably
//	pkh = PublicKeyJwk{
//		data:   pkb,
//		vmType: vmType,
//	}
//	return
//}

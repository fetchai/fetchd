package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/fetchai/fetchd/x/verifiable-credential/crypto/accumulator"
	"github.com/fetchai/fetchd/x/verifiable-credential/crypto/anonymouscredential"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	didtypes "github.com/fetchai/fetchd/x/did/types"
)

// Defines the accepted credential types
const (
	IdentityCredential        = "IdentityCredential"
	UserCredential            = "UserCredential"
	RegistrationCredential    = "RegistrationCredential"
	AnonymousCredentialSchema = "AnonymousCredentialSchema"
)

// IsValidCredentialType tells if a credential type is valid (accepted)
func IsValidCredentialType(credential string) bool {
	switch credential {
	case IdentityCredential,
		UserCredential,
		RegistrationCredential,
		AnonymousCredentialSchema:
		return true
	default:
		return false
	}
}

func NewChainVcId(chainName, id string) string {
	return fmt.Sprint(VcChainPrefix, chainName, ":", id)
}

// NewAnonymousCredentialSchema constructs a new VerifiableCredential instance
func NewAnonymousCredentialSchema(
	id string,
	issuer string,
	issuanceDate time.Time,
	credentialSubject VerifiableCredential_AnonCredSchema,
) VerifiableCredential {
	return VerifiableCredential{
		Context:           []string{"https://www.w3.org/TR/vc-data-model/"},
		Id:                id,
		Type:              []string{"VerifiableCredential", AnonymousCredentialSchema},
		Issuer:            issuer,
		IssuanceDate:      &issuanceDate,
		CredentialSubject: &credentialSubject,
		Proof:             nil,
	}
}

// NewUserVerifiableCredential constructs a new VerifiableCredential instance
func NewUserVerifiableCredential(
	id string,
	issuer string,
	issuanceDate time.Time,
	credentialSubject VerifiableCredential_UserCred,
) VerifiableCredential {
	return VerifiableCredential{
		Context:           []string{"https://www.w3.org/TR/vc-data-model/"},
		Id:                id,
		Type:              []string{"VerifiableCredential", UserCredential},
		Issuer:            issuer,
		IssuanceDate:      &issuanceDate,
		CredentialSubject: &credentialSubject,
		Proof:             nil,
	}
}

// NewRegistrationVerifiableCredential constructs a new VerifiableCredential instance
func NewRegistrationVerifiableCredential(
	id string,
	issuer string,
	issuanceDate time.Time,
	credentialSubject VerifiableCredential_RegistrationCred,
) VerifiableCredential {
	return VerifiableCredential{
		Context:           []string{"https://www.w3.org/TR/vc-data-model/"},
		Id:                id,
		Type:              []string{"VerifiableCredential", RegistrationCredential},
		Issuer:            issuer,
		IssuanceDate:      &issuanceDate,
		CredentialSubject: &credentialSubject,
		Proof:             nil,
	}
}

// NewUserCredentialSubject create a new credential subject
func NewAnonymousCredentialSchemaSubject(
	subId string,
	subType []string,
	subContext []string,
	publicParams *anonymouscredential.PublicParameters,
) VerifiableCredential_AnonCredSchema {
	return VerifiableCredential_AnonCredSchema{
		&AnonymousCredentialSchemaSubject{
			Id:           subId,
			Type:         subType,
			Context:      subContext,
			PublicParams: publicParams,
		},
	}
}

// NewUserCredentialSubject create a new credential subject
func NewUserCredentialSubject(
	id string,
	root string,
	isVerified bool,
) VerifiableCredential_UserCred {
	return VerifiableCredential_UserCred{
		&UserCredentialSubject{
			Id:         id,
			Root:       root,
			IsVerified: isVerified,
		},
	}
}

// NewRegistrationCredentialSubject create a new registration credential subject
// TODO: placeholder implementation, refactor it
func NewRegistrationCredentialSubject(
	id string,
	country string,
	shortName string,
	longName string,
) VerifiableCredential_RegistrationCred {
	return VerifiableCredential_RegistrationCred{
		&RegistrationCredentialSubject{
			Id: id,
			Address: &Address{
				Country: country,
			},
			LegalPersons: []*LegalPerson{
				{
					Names: []*Name{
						{
							Type: "SN",
							Name: shortName,
						},
						{
							Type: "LN",
							Name: longName,
						},
					},
				},
			},
			Ids: []*Id{
				{
					Id:   "529900W6B9NEA233DS71",
					Type: "LEIX",
				},
			},
		},
	}
}

// NewProof create a new proof for a verifiable credential
func NewProof(
	proofType string,
	created string,
	proofPurpose string,
	verificationMethod string,
	signature string,
) Proof {
	return Proof{
		Type:               proofType,
		Created:            created,
		ProofPurpose:       proofPurpose,
		VerificationMethod: verificationMethod,
		Signature:          signature,
	}
}

// Validate validates a verifiable credential against a provided public key
func (vc VerifiableCredential) Validate(
	pk cryptotypes.PubKey,
) error {
	s, err := base64.StdEncoding.DecodeString(vc.Proof.Signature)
	if err != nil {
		return err
	}

	// reset the proof
	vc.Proof = nil

	// TODO: this is an expensive operation, could lead to DDOS
	// TODO: we can hash this and make this less expensive
	isCorrectPubKey := pk.VerifySignature(
		vc.GetBytes(),
		s,
	)

	if !isCorrectPubKey {
		return fmt.Errorf("failed to verify verificable credential proof")
	}

	return nil
}

// Sign signs a credential with a provided private key
func (vc VerifiableCredential) Sign(
	keyring keyring.Keyring,
	address sdk.Address,
	verificationMethodID string,
) (VerifiableCredential, error) {
	tm := time.Now()
	// reset the proof
	vc.Proof = nil
	// TODO: this could be expensive review this signing method
	// TODO: we can hash this an make this less expensive
	signature, pubKey, err := keyring.SignByAddress(address, vc.GetBytes())
	if err != nil {
		return vc, err
	}

	p := NewProof(
		pubKey.Type(),
		tm.Format(time.RFC3339),
		// TODO: define proof purposes
		"assertionMethod",
		verificationMethodID,
		base64.StdEncoding.EncodeToString(signature),
	)
	vc.Proof = &p
	return vc, nil
}

func (vc VerifiableCredential) Hash() string {
	// TODO: implement the hashing of creds for signing
	return "TODO"
}

// HasType tells whenever a credential has a specific type
func (vc VerifiableCredential) HasType(vcType string) bool {
	for _, vct := range vc.Type {
		if vct == vcType {
			return true
		}
	}
	return false
}

// GetSubjectDID return the credential DID subject, that is the holder
// of the credentials
func (vc VerifiableCredential) GetSubjectDID() didtypes.DID {
	switch subj := vc.CredentialSubject.(type) {
	case *VerifiableCredential_RegistrationCred:
		return didtypes.DID(subj.RegistrationCred.Id)
	case *VerifiableCredential_UserCred:
		return didtypes.DID(subj.UserCred.Id)
	case *VerifiableCredential_AnonCredSchema:
		return didtypes.DID(subj.AnonCredSchema.Id)
	default:
		// TODO, not great
		return didtypes.DID("")
	}
}

// GetIssuerDID returns the did of the issuer
func (vc VerifiableCredential) GetIssuerDID() didtypes.DID {
	return didtypes.DID(vc.Issuer)
}

// GetBytes is a helper for serializing
func (vc VerifiableCredential) GetBytes() []byte {
	dAtA, _ := vc.Marshal()
	return dAtA
}

// SetMembershipState sets a new membership state
func (vc VerifiableCredential) UpdateAccumulatorState(
	state *accumulator.State,
) (VerifiableCredential, error) {
	sub, ok := vc.GetCredentialSubject().(*VerifiableCredential_AnonCredSchema)
	if !ok {
		return VerifiableCredential{}, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "not an anonymous credential")
	}
	if sub.AnonCredSchema.PublicParams.AccumulatorPublicParams == nil {
		return VerifiableCredential{}, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "anonymous credential scheme does not have accumulator")
	}
	states := sub.AnonCredSchema.PublicParams.AccumulatorPublicParams.States
	states = append(states, state)
	sub.AnonCredSchema.PublicParams.AccumulatorPublicParams.States = states
	return vc, nil
}

func (vc VerifiableCredential) UpdatePublicParameters(pp *anonymouscredential.PublicParameters) (VerifiableCredential, error) {
	sub, ok := vc.GetCredentialSubject().(*VerifiableCredential_AnonCredSchema)
	if !ok {
		return VerifiableCredential{}, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "not an anonymous credential")
	}
	sub.AnonCredSchema.PublicParams = pp
	return vc, nil
}

// NewVcMetadata returns a VcMetadata struct that has equals created and updated date,
// and with deactivated field set to false
func NewVcMetadata(versionData []byte, created time.Time) VcMetadata {
	m := VcMetadata{
		Created: &created,
	}
	UpdateVcMetadata(&m, versionData, created, false)
	return m
}

// UpdateVcMetadata updates a VC metadata time and version id
func UpdateVcMetadata(meta *VcMetadata, versionData []byte, updated time.Time, deactivated bool) {
	txH := sha256.Sum256(versionData)
	meta.VersionId = hex.EncodeToString(txH[:])
	meta.Updated = &updated
	meta.Deactivated = deactivated
}

// IsEmpty tells if the trimmed input is empty
func IsEmpty(input string) bool {
	return strings.TrimSpace(input) == ""
}

// IsValidVcMetadata tells if a vc metadata is valid,
// that is if it has a non empty versionId and a non-zero create date
func IsValidVcMetadata(vcMeta *VcMetadata) bool {
	if vcMeta == nil {
		return false
	}
	if IsEmpty(vcMeta.VersionId) {
		return false
	}
	if vcMeta.Created == nil || vcMeta.Created.IsZero() {
		return false
	}
	return true
}

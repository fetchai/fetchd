package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type VerificationRelationship int

// A verification relationship expresses the relationship between the DID subject and a verification method.
// This enum is used to
// cfr. https://www.w3.org/TR/did-core/#verification-relationships
const (
	Authentication       = "authentication"       // https://www.w3.org/TR/did-core/#authentication
	AssertionMethod      = "assertionMethod"      // https://www.w3.org/TR/did-core/#assertion
	KeyAgreement         = "keyAgreement"         // https://www.w3.org/TR/did-core/#key-agreement
	CapabilityInvocation = "capabilityInvocation" // https://www.w3.org/TR/did-core/#capability-invocation
	CapabilityDelegation = "capabilityDelegation" // https://www.w3.org/TR/did-core/#capability-delegation
)

const (
	authentication VerificationRelationship = iota
	assertionMethod
	keyAgreement
	capabilityInvocation
	capabilityDelegation
)

// VerificationRelationships are the supported list of verification relationships
var VerificationRelationships = map[string]VerificationRelationship{
	Authentication:       authentication,
	AssertionMethod:      assertionMethod,
	KeyAgreement:         keyAgreement,
	CapabilityInvocation: capabilityInvocation,
	CapabilityDelegation: capabilityDelegation,
}

// verificationRelationships retrieve the pointer to the verification relationship
// if it exists, otherwise returns nil
func (didDoc *DidDocument) getRelationships(rel VerificationRelationship) *[]string {
	switch rel {
	case authentication:
		return &didDoc.Authentication
	case assertionMethod:
		return &didDoc.AssertionMethod
	case keyAgreement:
		return &didDoc.KeyAgreement
	case capabilityInvocation:
		return &didDoc.CapabilityInvocation
	case capabilityDelegation:
		return &didDoc.CapabilityDelegation
	default:
		return nil
	}
}

// parseRelationshipLabels parse relationships labels to a slice of VerificationRelationship
// making sure that the relationsips are not repeated
func parseRelationshipLabels(relNames ...string) (vrs []VerificationRelationship, err error) {
	names := distinct(relNames)
	vrs = make([]VerificationRelationship, len(names))
	for i, vrn := range distinct(relNames) {
		vr, validName := VerificationRelationships[vrn]
		if !validName {
			err = sdkerrors.Wrapf(ErrInvalidInput, "unsupported verification relationship %s", vrn)
			return
		}
		vrs[i] = vr
	}
	return
}

/**
Regexp generated using this ABNF specs and using https://abnf.msweet.org/index.php

did-url            = did path-abempty [ "?" query ] [ "#" fragment ]
did                = "did:" method-name ":" method-specific-id
method-name        = 1*method-char
method-char        = %x61-7A / DIGIT
method-specific-id = *( *idchar ":" ) 1*idchar
idchar             = ALPHA / DIGIT / "." / "-" / "_" / pct-encoded
pct-encoded        = "%" HEXDIG HEXDIG
query              = *( pchar / "/" / "?" )
fragment           = *( pchar / "/" / "?" )
path-abempty       = *( "/" segment )
segment            = *pchar
unreserved         = ALPHA / DIGIT / "-" / "." / "_" / "~"
pchar              = unreserved / pct-encoded / sub-delims / ":" / "@"
sub-delims         = "!" / "$" / "&" / "'" / "(" / ")"
                 / "*" / "+" / "," / ";" / "="
*/

const (
	contextDIDBase         = "https://www.w3.org/ns/did/v1"
	didValidationRegexpStr = `^did\:[a-z0-9]+\:(([A-Z.a-z0-9]|\-|_|%[0-9A-Fa-f][0-9A-Fa-f])*\:)*([A-Z.a-z0-9]|\-|_|%[0-9A-Fa-f][0-9A-Fa-f])+$`
	didURLValidationRegexpStr = `^did\:[a-z0-9]+\:(([A-Z.a-z0-9]|\-|_|%[0-9A-Fa-f][0-9A-Fa-f])*\:)*([A-Z.a-z0-9]|\-|_|%[0-9A-Fa-f][0-9A-Fa-f])+(/(([-A-Z._a-z0-9]|~)|%[0-9A-Fa-f][0-9A-Fa-f]|(\!|\$|&|'|\(|\)|\*|\+|,|;|\=)|\:|@)*)*(\?(((([-A-Z._a-z0-9]|~)|%[0-9A-Fa-f][0-9A-Fa-f]|(\!|\$|&|'|\(|\)|\*|\+|,|;|\=)|\:|@)|/|\?)*))?(#(((([-A-Z._a-z0-9]|~)|%[0-9A-Fa-f][0-9A-Fa-f]|(\!|\$|&|'|\(|\)|\*|\+|,|;|\=)|\:|@)|/|\?)*))?$`
	rfc3986RegexpStr          = `^(([^:/?#]+):)?(//([^/?#]*))?([^?#]*)(\?([^#]*))?(#(.*))?$`
)

var (
	didValidationRegexp    = regexp.MustCompile(didValidationRegexpStr)
	didURLValidationRegexp = regexp.MustCompile(didURLValidationRegexpStr)
	rfc3986Regexp          = regexp.MustCompile(rfc3986RegexpStr)
)

// DID as typed string
type DID string

// NewChainDID format a DID from a method specific did
// cfr.https://www.w3.org/TR/did-core/#did
func NewChainDID(chainName, didID string) DID {
	return DID(fmt.Sprint(DidChainPrefix, chainName, ":", didID))
}

// NewKeyDID format a DID of type key
func NewKeyDID(account string) DID {
	return DID(fmt.Sprint(DidKeyPrefix, account))
}

// String return the string representation of the did
func (did DID) String() string {
	return string(did)
}

// NewVerificationMethodID compose a verification method id from an account address
func (did DID) NewVerificationMethodID(vmID string) string {
	return fmt.Sprint(did, "#", vmID)
}

// IsValidDID validate the input string according to the
// did specification (cfr. https://www.w3.org/TR/did-core/#did-syntax ).
func IsValidDID(input string) bool {
	return didValidationRegexp.MatchString(input)
}

// IsValidDIDURL validate the input string according to the
// did url specification (cfr. https://www.w3.org/TR/did-core/#did-url-syntax  ).
func IsValidDIDURL(input string) bool {
	return didURLValidationRegexp.MatchString(input)
}

// IsValidRFC3986Uri checks if the input string is a valid RFC3986 URI
// (cfr https://datatracker.ietf.org/doc/html/rfc3986#page-50)
func IsValidRFC3986Uri(input string) bool {
	return rfc3986Regexp.MatchString(input)
}

// IsValidDIDDocument tells if a DID document is valid,
// that is if it has the default context and a valid subject
func IsValidDIDDocument(didDoc *DidDocument) bool {
	if didDoc == nil {
		return false
	}

	if !IsValidDID(didDoc.Id) {
		return false
	}

	for _, element := range didDoc.Controller {
		if !IsValidDIDKeyFormat(element) {
			return false
		}
	}

	for _, c := range didDoc.Context {
		if c == contextDIDBase {
			return true
		}
	}

	return false
}

// IsValidDIDKeyFormat verify that a did is compliant with the did:cosmos:key format
// that is the ID must be a bech32 address no longer than 255 bytes
func IsValidDIDKeyFormat(did string) bool {
	if _, err := sdk.AccAddressFromBech32(strings.TrimPrefix(did, DidKeyPrefix)); err != nil {
		return false
	}
	return true
}

// IsValidDIDMetadata tells if a DID metadata is valid,
// that is if it has a non empty versionId and a non-zero create date
func IsValidDIDMetadata(didMeta *DidMetadata) bool {
	if didMeta == nil {
		return false
	}
	if IsEmpty(didMeta.VersionId) {
		return false
	}
	if didMeta.Created == nil || didMeta.Created.IsZero() {
		return false
	}
	return true
}

// ValidateVerification perform basic validation on a verification struct
// optionally validating the validation method controller against a list
// of allowed controllers.
// in case of error returns an cosmos-sdk wrapped error
// XXX: this pattern creates a ambiguous semantic (but maybe is not too severe (use WithCredentials and array of credentials))
func ValidateVerification(v *Verification, allowedControllers ...string) (err error) {
	if v == nil {
		err = sdkerrors.Wrap(ErrInvalidInput, "verification is not defined")
		return
	}
	// verify that the method id is correct
	if !IsValidDIDURL(v.Method.Id) {
		err = sdkerrors.Wrapf(ErrInvalidDIDURLFormat, "verification method id: %v", v.Method.Id)
		return
	}

	// if the controller is not set return error
	if !IsValidDID(v.Method.Controller) {
		err = sdkerrors.Wrapf(ErrInvalidDIDFormat, "verification method controller %v", v.Method.Controller)
		return
	}

	// check for empty method type
	if IsEmpty(v.Method.Type) {
		err = sdkerrors.Wrapf(ErrInvalidInput, "verification method type not set for verification method %s", v.Method.Id)
		return
	}

	// check the verification material
	switch x := v.Method.VerificationMaterial.(type) {
	case *VerificationMethod_BlockchainAccountID:
		if IsEmpty(x.BlockchainAccountID) {
			err = sdkerrors.Wrapf(ErrInvalidInput, "verification material blockchain account id invalid for verification method %s", v.Method.Id)
			return
		}
	case *VerificationMethod_PublicKeyMultibase:
		if IsEmpty(x.PublicKeyMultibase) {
			err = sdkerrors.Wrapf(ErrInvalidInput, "verification material multibase pubkey invalid for verification method %s", v.Method.Id)
			return
		}
	case *VerificationMethod_PublicKeyHex:
		if IsEmpty(x.PublicKeyHex) {
			err = sdkerrors.Wrapf(ErrInvalidInput, "verification material pubkey invalid for verification method %s", v.Method.Id)
			return
		}
	default:
		err = sdkerrors.Wrapf(ErrInvalidInput, "verification material not set for verification method %s", v.Method.Id)
		return
	}

	// check for empty publickey
	if v.Method.VerificationMaterial.Size() == 0 {
		err = sdkerrors.Wrapf(ErrInvalidInput, "verification material not set for verification method %s", v.Method.Id)
		return
	}

	// check that there is at least a relationship
	if len(v.Relationships) == 0 {
		err = sdkerrors.Wrap(ErrEmptyRelationships, "at least a verification relationship is required")
		return
	}
	return
}

// ValidateService performs basic on a service struct
func ValidateService(s *Service) (err error) {
	if s == nil {
		err = sdkerrors.Wrap(ErrInvalidInput, "service is not defined")
		return
	}
	// verify that the id is not empty and is a valid url (according to RFC3986)
	if IsEmpty(s.Id) {
		err = sdkerrors.Wrap(ErrInvalidInput, "service id cannot be empty")
		return
	}

	if !IsValidRFC3986Uri(s.Id) {
		err = sdkerrors.Wrapf(ErrInvalidRFC3986UriFormat, "service id %s is not a valid RFC3986 uri", s.Id)
		return
	}

	// verify that the endpoint is not empty and is a valid url (according to RFC3986)
	if IsEmpty(s.ServiceEndpoint) {
		err = sdkerrors.Wrap(ErrInvalidInput, "service endpoint cannot be empty;")
		return
	}

	if !IsValidRFC3986Uri(s.ServiceEndpoint) {
		err = sdkerrors.Wrapf(ErrInvalidRFC3986UriFormat, "service endpoint %s is not a valid RFC3986 uri", s.ServiceEndpoint)
		return
	}

	// check that the service type is not empty
	if IsEmpty(s.Type) {
		err = sdkerrors.Wrap(ErrInvalidInput, "service type cannot be empty")
		return
	}

	return
}

// IsEmpty tells if the trimmed input is empty
func IsEmpty(input string) bool {
	return strings.TrimSpace(input) == ""
}

// DidDocumentOption implements variadic pattern for optional did document fields
type DidDocumentOption func(*DidDocument) error

// WithVerifications add optional verifications
func WithVerifications(verifications ...*Verification) DidDocumentOption {
	return func(did *DidDocument) error {
		return did.AddVerifications(verifications...)
	}
}

//WithServices add optional services
func WithServices(services ...*Service) DidDocumentOption {
	return func(did *DidDocument) error {
		return did.AddServices(services...)
	}
}

// WithControllers add optional did controller
func WithControllers(controllers ...string) DidDocumentOption {
	return func(did *DidDocument) (err error) {
		return did.AddControllers(controllers...)
	}
}

// NewDidDocument constructs a new DidDocument
func NewDidDocument(id string, options ...DidDocumentOption) (did DidDocument, err error) {

	if !IsValidDID(id) {
		err = sdkerrors.Wrapf(ErrInvalidDIDFormat, "did %s", id)
		return
	}

	did = DidDocument{
		Context: []string{contextDIDBase},
		Id:      id,
	}
	// apply all the options
	for _, fn := range options {
		if err = fn(&did); err != nil {
			return
		}
	}
	return
}

// AddControllers add a controller to a did document if not exists
func (didDoc *DidDocument) AddControllers(controllers ...string) error {
	if len(controllers) == 0 {
		return nil
	}
	// join the exiting controllers with the new ones
	dc := distinct(append(didDoc.Controller, controllers...))
	for _, c := range dc {
		if !IsValidDID(c) {
			return sdkerrors.Wrapf(ErrInvalidDIDFormat, "did document controller validation error '%s'", c)
		}
		if !IsValidDIDKeyFormat(c) {
			// TODO: link to the documentation for the error
			return sdkerrors.Wrapf(ErrInvalidInput, "did document controller '%s' must be of type key", c)
		}
	}

	// remove duplicates
	didDoc.Controller = dc
	return nil
}

// DeleteControllers delete controllers from a did document
func (didDoc *DidDocument) DeleteControllers(controllers ...string) error {
	if len(controllers) == 0 {
		return nil
	}
	dc := distinct(controllers)
	for _, c := range dc {
		if !IsValidDID(c) {
			return sdkerrors.Wrapf(ErrInvalidDIDFormat, "did document controller validation error '%s'", c)
		}
	}
	// remove existing
	didDoc.Controller = subtraction(didDoc.Controller, controllers)
	return nil
}

// AddVerifications add one or more verification method and relations to a did document
func (didDoc *DidDocument) AddVerifications(verifications ...*Verification) (err error) {
	// verify that there are no duplicates in method ids
	index := make(map[string]struct{}, len(didDoc.VerificationMethod))
	// load existing verifications if any
	for _, v := range didDoc.VerificationMethod {
		index[v.Id] = struct{}{}
	}

	// loop through the verifications and look for problems
	for _, v := range verifications {
		// perform base validation checks
		if err = ValidateVerification(v); err != nil {
			return
		}

		// verify that there are no duplicates in method ids
		if _, found := index[v.Method.Id]; found {
			err = sdkerrors.Wrapf(ErrInvalidInput, "duplicated verification method id %s", v.Method.Id)
			return
		}
		index[v.Method.Id] = struct{}{}

		// first add the method to the list of methods
		didDoc.VerificationMethod = append(didDoc.VerificationMethod, v.GetMethod())

		// now add the relationships
		vrs, err := parseRelationshipLabels(v.Relationships...)
		if err != nil {
			return err
		}
		didDoc.setRelationships(v.Method.Id, vrs...)

		// update context
		didDoc.Context = union(didDoc.Context, v.Context)

	}
	return
}

// RevokeVerification revoke a verification method
// and all relationships associated with it
func (didDoc *DidDocument) RevokeVerification(methodID string) error {

	del := func(x int) {
		lastIdx := len(didDoc.VerificationMethod) - 1
		switch lastIdx {
		case 0:
			didDoc.VerificationMethod = nil
		case x:
			didDoc.VerificationMethod = didDoc.VerificationMethod[:lastIdx]
		default:
			didDoc.VerificationMethod[x] = didDoc.VerificationMethod[lastIdx]
			didDoc.VerificationMethod = didDoc.VerificationMethod[:lastIdx]
		}
	}

	// remove relationships
	didDoc.setRelationships(methodID)

	// now remove the method
	for i, vm := range didDoc.VerificationMethod {
		if vm.Id == methodID {
			del(i)
			return nil
		}
	}
	return sdkerrors.Wrapf(ErrVerificationMethodNotFound, "verification method id: %v", methodID)
}

// SetVerificationRelationships for a did document
func (didDoc *DidDocument) SetVerificationRelationships(methodID string, relationships ...string) error {
	// verify that the method id is correct
	if !IsValidDIDURL(methodID) {
		return sdkerrors.Wrapf(ErrInvalidDIDURLFormat, "verification method id: %v", methodID)
	}
	// check that the methodID exists
	hasVM := false
	for _, vm := range didDoc.VerificationMethod {
		if vm.Id == methodID {
			hasVM = true
			break
		}
	}
	if !hasVM {
		return sdkerrors.Wrapf(ErrVerificationMethodNotFound, "verification method %v not found", methodID)
	}
	// check that there is at least a relationship
	if len(relationships) == 0 {
		return sdkerrors.Wrap(ErrEmptyRelationships, "at least a verification relationship is required")
	}
	// check that the provided relationships are valid
	vrs, err := parseRelationshipLabels(relationships...)
	if err != nil {
		return err
	}
	// update the relationships
	didDoc.setRelationships(methodID, vrs...)
	return nil
}

// setRelationships overwrite relationships for a did document
func (didDoc *DidDocument) setRelationships(methodID string, relationships ...VerificationRelationship) {

	// first remove existing relationships
	for _, vr := range VerificationRelationships {
		vrs := didDoc.getRelationships(vr)
		for i, vmID := range *vrs {
			if vmID == methodID {
				lastIdx := len(*vrs) - 1 // get the last index of the current relationship list
				switch lastIdx {
				case 0: // remove the relationships since there is no elements left
					*vrs = nil
				case i: // if it's at the last position, just drop the last position
					*vrs = (*vrs)[:lastIdx]
				default: // swap and drop last position
					(*vrs)[i] = (*vrs)[lastIdx]
					(*vrs) = (*vrs)[:lastIdx]
				}
			}
		}
	}

	// then assign the new ones
	for _, vr := range relationships {
		vrs := didDoc.getRelationships(vr)
		*vrs = append(*vrs, methodID)
	}
}

// GetVerificationMethodBlockchainAddress returns the verification method cosmos blockchain address of a verification method.
// it fails if the verification method is not supported or if the verification method is not found
func (didDoc DidDocument) GetVerificationMethodBlockchainAddress(methodID string) (address string, err error) {
	for _, vm := range didDoc.VerificationMethod {
		if vm.Id == methodID {
			switch k := vm.VerificationMaterial.(type) {
			case *VerificationMethod_BlockchainAccountID:
				address = BlockchainAccountID(k.BlockchainAccountID).GetAddress()
			case *VerificationMethod_PublicKeyMultibase:
				address, err = toAddress(k.PublicKeyMultibase[1:])
			case *VerificationMethod_PublicKeyHex:
				address, err = toAddress(k.PublicKeyHex)
			}
			return
		}
	}
	err = ErrVerificationMethodNotFound
	return
}

// GetVerificationMethodPublicKey returns the public key in verification method cosmos blockchain address of a verification method.
// it fails if the verification method is not supported or if the public key is not found
func (didDoc DidDocument) GetVerificationMethodPublicKey(methodID string) (pk cryptotypes.PubKey, err error) {
	pk = nil
	for _, vm := range didDoc.VerificationMethod {
		if vm.Id == methodID {
			switch k := vm.VerificationMaterial.(type) {
			case *VerificationMethod_BlockchainAccountID:
				err = fmt.Errorf("public key not found in verification method: %s", methodID)
			case *VerificationMethod_PublicKeyMultibase:
				pk, err = toPublicKey(k.PublicKeyMultibase[1:])
			case *VerificationMethod_PublicKeyHex:
				pk, err = toPublicKey(k.PublicKeyHex)
			}
			return pk, err
		}
	}
	err = ErrVerificationMethodNotFound
	return pk, err
}

// GetVerificationRelationships returns the relationships associated with the
// verification method id.
func (didDoc DidDocument) GetVerificationRelationships(methodID string) []string {
	relationships := []string{}
	for vrn, vr := range VerificationRelationships {
		for _, vmID := range *didDoc.getRelationships(vr) {
			if vmID == methodID {
				relationships = append(relationships, vrn)
			}
		}
	}
	return relationships
}

// HasRelationship verifies if a controller did
// exist for at least one of the relationships in the did document
func (didDoc DidDocument) HasRelationship(
	signer BlockchainAccountID,
	relationships ...string,
) bool {
	// first check if the controller exists
	for _, vm := range didDoc.VerificationMethod {
		switch k := vm.VerificationMaterial.(type) {
		case *VerificationMethod_BlockchainAccountID:
			if k.BlockchainAccountID != signer.EncodeToString() {
				continue
			}
		case *VerificationMethod_PublicKeyMultibase:
			addr, err := toAddress(k.PublicKeyMultibase[1:])
			if err != nil || !signer.MatchAddress(addr) {
				continue
			}
		case *VerificationMethod_PublicKeyHex:
			addr, err := toAddress(k.PublicKeyHex)
			if err != nil || !signer.MatchAddress(addr) {
				continue
			}
		}
		vrs := didDoc.GetVerificationRelationships(vm.Id)
		if len(intersection(vrs, relationships)) > 0 {
			return true
		}
	}
	return false
}

// HasPublicKey validates if a public key is contained in a DidDocument
func (didDoc DidDocument) HasPublicKey(pubkey cryptotypes.PubKey) bool {
	for _, vm := range didDoc.VerificationMethod {
		switch key := vm.VerificationMaterial.(type) {
		case *VerificationMethod_BlockchainAccountID:
			address := sdk.MustBech32ifyAddressBytes(
				sdk.GetConfig().GetBech32AccountAddrPrefix(),
				pubkey.Address().Bytes(),
			)
			if BlockchainAccountID(key.BlockchainAccountID).MatchAddress(address) {
				return true
			}
		case *VerificationMethod_PublicKeyMultibase:
			if key.PublicKeyMultibase == fmt.Sprint("F", hex.EncodeToString(pubkey.Bytes())) {
				return true
			}

		case *VerificationMethod_PublicKeyHex:
			if key.PublicKeyHex == hex.EncodeToString(pubkey.Bytes()) {
				return true
			}
		}
	}
	return false
}

// HasController returns true if the DID document has the input DID as a controller, false otherwise
func (didDoc *DidDocument) HasController(controller DID) bool {
	ctrl := controller.String()
	for _, c := range didDoc.Controller {
		if c == ctrl {
			return true
		}
	}
	return false
}

// AddServices add services to a did document
func (didDoc *DidDocument) AddServices(services ...*Service) (err error) {
	if didDoc.Service == nil {
		didDoc.Service = []*Service{}
	}

	// used to check duplicates
	index := make(map[string]struct{}, len(didDoc.Service))

	// load existing services
	for _, s := range didDoc.Service {
		index[s.Id] = struct{}{}
	}

	// services must be unique
	for _, s := range services {
		if err = ValidateService(s); err != nil {
			return
		}

		// verify that there are no duplicates in method ids
		if _, found := index[s.Id]; found {
			err = sdkerrors.Wrapf(ErrInvalidInput, "duplicated verification method id %s", s.Id)
			return
		}
		index[s.Id] = struct{}{}

		didDoc.Service = append(didDoc.Service, s)
	}
	return
}

// DeleteService delete an existing service from a did document
func (didDoc *DidDocument) DeleteService(serviceID string) {
	del := func(x int) {
		lastIdx := len(didDoc.Service) - 1
		switch lastIdx {
		case 0: // remove the relationships since there is no elements left
			didDoc.Service = nil
		case x: // if it's at the last position, just drop the last position
			didDoc.Service = didDoc.Service[:lastIdx]
		default: // swap and drop last position
			didDoc.Service[x] = didDoc.Service[lastIdx]
			didDoc.Service = didDoc.Service[:lastIdx]
		}
	}

	for i, s := range didDoc.Service {
		if s.Id == serviceID {
			del(i)
			break
		}
	}
}

// GetBytes is a helper for serializing
func (didDoc DidDocument) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&didDoc))
}

// Verifications is a list of verification
type Verifications []*Verification

// NewVerification build a new verification to be
// attached to a did document
func NewVerification(
	method VerificationMethod,
	relationships []string,
	contexts []string,
) *Verification {
	return &Verification{
		Context:       contexts,
		Method:        &method,
		Relationships: relationships,
	}
}

// NewAccountVerification is a shortcut to create a verification based on comsos address
func NewAccountVerification(did DID, chainID, accountAddress string, verificationMethods ...string) *Verification {
	return NewVerification(
		NewVerificationMethod(
			fmt.Sprint(did.String(), "#", accountAddress),
			did,
			NewBlockchainAccountID(chainID, accountAddress),
		),
		verificationMethods,
		nil,
	)
}

// NewVerificationMethod build a new verification method
func NewVerificationMethod(id string, controller DID, vmr VerificationMaterial) VerificationMethod {
	vm := VerificationMethod{
		Id:         id,
		Controller: controller.String(),
		Type:       string(vmr.Type()),
	}
	switch vmr.(type) {
	case BlockchainAccountID:
		vm.VerificationMaterial = &VerificationMethod_BlockchainAccountID{vmr.EncodeToString()}
	case PublicKeyMultibase:
		vm.VerificationMaterial = &VerificationMethod_PublicKeyMultibase{vmr.EncodeToString()}
	case PublicKeyHex:
		vm.VerificationMaterial = &VerificationMethod_PublicKeyHex{vmr.EncodeToString()}
	}
	return vm
}

// GetBytes is a helper for serializing
func (did Verification) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&did))
}

// Services are a list of services
type Services []*Service

// NewService creates a new service
func NewService(id string, serviceType string, serviceEndpoint string) *Service {
	return &Service{
		Id:              id,
		Type:            serviceType,
		ServiceEndpoint: serviceEndpoint,
	}
}

// NewDidMetadata returns a DidMetadata struct that has equals created and updated date,
// and with deactivated field set to false
func NewDidMetadata(versionData []byte, created time.Time) DidMetadata {
	m := DidMetadata{
		Created:     &created,
		Deactivated: false,
	}
	UpdateDidMetadata(&m, versionData, created)
	return m
}

// UpdateDidMetadata updates a DID metadata time and version id
func UpdateDidMetadata(meta *DidMetadata, versionData []byte, updated time.Time) {
	txH := sha256.Sum256(versionData)
	meta.VersionId = hex.EncodeToString(txH[:])
	meta.Updated = &updated
}

// ResolveAccountDID generates a DID document from an address
func ResolveAccountDID(did, chainID string) (didDoc DidDocument, didMeta DidMetadata, err error) {
	if !IsValidDIDKeyFormat(did) {
		err = ErrInvalidDidMethodFormat
		return
	}
	account := strings.TrimPrefix(did, DidKeyPrefix)
	accountDID := DID(did)
	// compose the metadata
	didMeta = NewDidMetadata([]byte(account), time.Now())
	// compose the did document
	didDoc, err = NewDidDocument(did, WithVerifications(
		NewVerification(
			NewVerificationMethod(
				accountDID.NewVerificationMethodID(account),
				accountDID, // the controller is the same as the did subject
				NewBlockchainAccountID(chainID, account),
			),
			[]string{
				Authentication,
				KeyAgreement,
				AssertionMethod,
				CapabilityInvocation,
				CapabilityDelegation,
			},
			nil,
		),
	))
	return
}

// toAddress encode a kexKey string to cosmos based address
func toAddress(hexKey string) (addr string, err error) {
	// decode the hex string
	pkb, err := hex.DecodeString(hexKey)
	if err != nil {
		return
	}
	// check the size of the decoded byte slice, otherwise the pk.Address will panic
	if len(pkb) != secp256k1.PubKeySize {
		err = fmt.Errorf("invalid public key size")
		return
	}
	// load the public key
	pk := &secp256k1.PubKey{Key: pkb}
	// generate the address
	addr, err = sdk.Bech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), pk.Address())
	return
}

func toPublicKey(hexKey string) (cryptotypes.PubKey, error) {
	// decode the hex string
	pkb, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	// check the size of the decoded byte slice, otherwise the pk.Address will panic
	if len(pkb) != secp256k1.PubKeySize {
		err = fmt.Errorf("invalid public key size")
		return nil, err
	}
	// load the public key
	pk := &secp256k1.PubKey{Key: pkb}
	return pk, nil
}

// union perform union, distinct amd sort operation between two slices
// duplicated element in list a are
func union(a, b []string) []string {
	if len(b) == 0 {
		return a
	}
	m := make(map[string]struct{})
	for _, item := range a {
		m[item] = struct{}{}
	}
	for _, item := range b {
		if _, ok := m[item]; !ok {
			m[item] = struct{}{}
		}
	}
	u := make([]string, 0, len(m))
	for k := range m {
		u = append(u, k)
	}
	sort.Strings(u)
	return u
}

func intersection(a, b []string) []string {
	m := make(map[string]struct{})
	for _, item := range a {
		m[item] = struct{}{}
	}
	var i []string
	for _, item := range distinct(b) {
		if _, ok := m[item]; ok {
			i = append(i, item)
		}
	}
	sort.Strings(i)
	return i
}

// distinct remove duplicates and sorts from a list of strings
func distinct(a []string) []string {
	m := make(map[string]struct{})
	for _, item := range a {
		m[item] = struct{}{}
	}
	d := make([]string, 0, len(m))
	for k := range m {
		d = append(d, k)
	}
	sort.Strings(d)
	return d
}

// subtraction remove set b from a
func subtraction(a, b []string) []string {
	m := make(map[string]struct{})
	for _, item := range a {
		m[item] = struct{}{}
	}
	var s []string
	for _, item := range distinct(b) {
		if _, ok := m[item]; !ok {
			s = append(s, item)
		}
	}
	sort.Strings(s)
	return s
}

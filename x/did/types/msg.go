package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// msg types
const (
	TypeMsgCreateDidDocument = "create-did"
)

var _ sdk.Msg = &MsgCreateDidDocument{}

// NewMsgCreateDidDocument creates a new MsgCreateDidDocument instance
func NewMsgCreateDidDocument(
	id string,
	verifications []*Verification,
	services []*Service,
	signerAccount string,
) *MsgCreateDidDocument {
	return &MsgCreateDidDocument{
		Id:            id,
		Verifications: verifications,
		Services:      services,
		Signer:        signerAccount,
	}
}

// Route implements sdk.Msg
func (MsgCreateDidDocument) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgCreateDidDocument) Type() string {
	return TypeMsgCreateDidDocument
}

func (msg MsgCreateDidDocument) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgCreateDidDocument) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// --------------------------
// UPDATE IDENTIFIER
// --------------------------

// msg types
const (
	TypeMsgUpdateDidDocument = "update-did"
)

func NewMsgUpdateDidDocument(
	didDoc *DidDocument,
	signerAccount string,
) *MsgUpdateDidDocument {
	return &MsgUpdateDidDocument{
		Doc:    didDoc,
		Signer: signerAccount,
	}
}

// Route implements sdk.Msg
func (MsgUpdateDidDocument) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgUpdateDidDocument) Type() string {
	return TypeMsgUpdateDidDocument
}

func (msg MsgUpdateDidDocument) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgUpdateDidDocument) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// --------------------------
// ADD VERIFICATION
// --------------------------
// msg types
const (
	TypeMsgAddVerification = "add-verification"
)

var _ sdk.Msg = &MsgAddVerification{}

// NewMsgAddVerification creates a new MsgAddVerification instance
func NewMsgAddVerification(
	id string,
	verification *Verification,
	signerAccount string,
) *MsgAddVerification {
	return &MsgAddVerification{
		Id:           id,
		Verification: verification,
		Signer:       signerAccount,
	}
}

// Route implements sdk.Msg
func (MsgAddVerification) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgAddVerification) Type() string {
	return TypeMsgAddVerification
}

func (msg MsgAddVerification) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgAddVerification) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// --------------------------
// REVOKE VERIFICATION
// --------------------------

// msg types
const (
	TypeMsgRevokeVerification = "revoke-verification"
)

var _ sdk.Msg = &MsgRevokeVerification{}

// NewMsgRevokeVerification creates a new MsgRevokeVerification instance
func NewMsgRevokeVerification(
	id string,
	methodID string,
	signerAccount string,
) *MsgRevokeVerification {
	return &MsgRevokeVerification{
		Id:       id,
		MethodId: methodID,
		Signer:   signerAccount,
	}
}

// Route implements sdk.Msg
func (MsgRevokeVerification) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgRevokeVerification) Type() string {
	return TypeMsgRevokeVerification
}

func (msg MsgRevokeVerification) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgRevokeVerification) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// --------------------------
// SET VERIFICATION RELATIONSHIPS
// --------------------------
// msg types
const (
	TypeMsgSetVerificationRelationships = "set-verification-relationships"
)

func NewMsgSetVerificationRelationships(
	id string,
	methodID string,
	relationships []string,
	signerAccount string,
) *MsgSetVerificationRelationships {
	return &MsgSetVerificationRelationships{
		Id:            id,
		MethodId:      methodID,
		Relationships: relationships,
		Signer:        signerAccount,
	}
}

// Route implements sdk.Msg
func (MsgSetVerificationRelationships) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgSetVerificationRelationships) Type() string {
	return TypeMsgSetVerificationRelationships
}

func (msg MsgSetVerificationRelationships) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgSetVerificationRelationships) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// --------------------------
// ADD SERVICE
// --------------------------

// msg types
const (
	TypeMsgAddService = "add-service"
)

var _ sdk.Msg = &MsgAddService{}

// NewMsgAddService creates a new MsgAddService instance
func NewMsgAddService(
	id string,
	service *Service,
	signerAccount string,
) *MsgAddService {
	return &MsgAddService{
		Id:          id,
		ServiceData: service,
		Signer:      signerAccount,
	}
}

// Route implements sdk.Msg
func (MsgAddService) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgAddService) Type() string {
	return TypeMsgAddService
}

func (msg MsgAddService) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgAddService) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// --------------------------
// DELETE SERVICE
// --------------------------

// msg types
const (
	TypeMsgDeleteService = "delete-service"
)

func NewMsgDeleteService(
	id string,
	serviceID string,
	signerAccount string,
) *MsgDeleteService {
	return &MsgDeleteService{
		Id:        id,
		ServiceId: serviceID,
		Signer:    signerAccount,
	}
}

// Route implements sdk.Msg
func (MsgDeleteService) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgDeleteService) Type() string {
	return TypeMsgDeleteService
}

func (msg MsgDeleteService) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgDeleteService) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// --------------------------
// ADD CONTROLLER
// --------------------------

// msg types
const (
	TypeMsgAddController = "add-controller"
)

func NewMsgAddController(
	id string,
	controllerDID string,
	signerAccount string,
) *MsgAddController {
	return &MsgAddController{
		Id:            id,
		ControllerDid: controllerDID,
		Signer:        signerAccount,
	}
}

// Route implements sdk.Msg
func (MsgAddController) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgAddController) Type() string {
	return TypeMsgAddController
}

func (msg MsgAddController) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgAddController) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

// --------------------------
// DELETE CONTROLLER
// --------------------------

// msg types
const (
	TypeMsgDeleteController = "delete-controller"
)

func NewMsgDeleteController(
	id string,
	controllerDID string,
	signerAccount string,
) *MsgDeleteController {
	return &MsgDeleteController{
		Id:            id,
		ControllerDid: controllerDID,
		Signer:        signerAccount,
	}
}

// Route implements sdk.Msg
func (MsgDeleteController) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgDeleteController) Type() string {
	return TypeMsgDeleteController
}

func (msg MsgDeleteController) GetSignBytes() []byte {
	panic("IBC messages do not support amino")
}

// GetSigners implements sdk.Msg
func (msg MsgDeleteController) GetSigners() []sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{accAddr}
}

package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	didtypes "github.com/fetchai/fetchd/x/did/types"
	"github.com/stretchr/testify/assert"
)

func TestMsgCreateDidDocument_Route(t *testing.T) {
	assert.Equalf(t, didtypes.ModuleName, didtypes.MsgCreateDidDocument{}.Route(), "Route()")
}

func TestMsgCreateDidDocument_Type(t *testing.T) {
	assert.Equalf(t, didtypes.TypeMsgCreateDidDocument, didtypes.MsgCreateDidDocument{}.Type(), "Type()")
}

func TestMsgCreateDidDocument_GetSignBytes(t *testing.T) {
	assert.Panicsf(t, func() { didtypes.MsgCreateDidDocument{}.GetSignBytes() }, "GetSignBytes()")
}

func TestMsgCreateDidDocument_GetSigners(t *testing.T) {
	a, err := sdk.AccAddressFromBech32("cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
	assert.NoError(t, err)
	assert.Equal(t,
		didtypes.MsgCreateDidDocument{Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"}.GetSigners(),
		[]sdk.AccAddress{a},
	)
	assert.Panics(t, func() { didtypes.MsgCreateDidDocument{Signer: "invalid"}.GetSigners() })
}

func TestMsgUpdateDidDocument_Route(t *testing.T) {
	assert.Equalf(t, didtypes.ModuleName, didtypes.MsgUpdateDidDocument{}.Route(), "Route()")
}

func TestMsgUpdateDidDocument_Type(t *testing.T) {
	assert.Equalf(t, didtypes.TypeMsgUpdateDidDocument, didtypes.MsgUpdateDidDocument{}.Type(), "Type()")
}

func TestMsgUpdateDidDocument_GetSignBytes(t *testing.T) {
	assert.Panicsf(t, func() { didtypes.MsgUpdateDidDocument{}.GetSignBytes() }, "GetSignBytes()")
}

func TestMsgUpdateDidDocument_GetSigners(t *testing.T) {
	a, err := sdk.AccAddressFromBech32("cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
	assert.NoError(t, err)
	assert.Equal(t,
		didtypes.MsgUpdateDidDocument{Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"}.GetSigners(),
		[]sdk.AccAddress{a},
	)
	assert.Panics(t, func() { didtypes.MsgUpdateDidDocument{Signer: "invalid"}.GetSigners() })
}

func TestMsgAddVerification_Route(t *testing.T) {
	assert.Equalf(t, didtypes.ModuleName, didtypes.MsgAddVerification{}.Route(), "Route()")
}

func TestMsgAddVerification_Type(t *testing.T) {
	assert.Equalf(t, didtypes.TypeMsgAddVerification, didtypes.MsgAddVerification{}.Type(), "Type()")
}

func TestMsgAddVerification_GetSignBytes(t *testing.T) {
	assert.Panicsf(t, func() { didtypes.MsgAddVerification{}.GetSignBytes() }, "GetSignBytes()")
}

func TestMsgAddVerification_GetSigners(t *testing.T) {
	a, err := sdk.AccAddressFromBech32("cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
	assert.NoError(t, err)
	assert.Equal(t,
		didtypes.MsgAddVerification{Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"}.GetSigners(),
		[]sdk.AccAddress{a},
	)
	assert.Panics(t, func() { didtypes.MsgAddVerification{Signer: "invalid"}.GetSigners() })
}

func TestMsgRevokeVerification_Route(t *testing.T) {
	assert.Equalf(t, didtypes.ModuleName, didtypes.MsgRevokeVerification{}.Route(), "Route()")
}

func TestMsgRevokeVerification_Type(t *testing.T) {
	assert.Equalf(t, didtypes.TypeMsgRevokeVerification, didtypes.MsgRevokeVerification{}.Type(), "Type()")
}

func TestMsgRevokeVerification_GetSignBytes(t *testing.T) {
	assert.Panicsf(t, func() { didtypes.MsgRevokeVerification{}.GetSignBytes() }, "GetSignBytes()")
}

func TestMsgRevokeVerification_GetSigners(t *testing.T) {
	a, err := sdk.AccAddressFromBech32("cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
	assert.NoError(t, err)
	assert.Equal(t,
		didtypes.MsgRevokeVerification{Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"}.GetSigners(),
		[]sdk.AccAddress{a},
	)
	assert.Panics(t, func() { didtypes.MsgRevokeVerification{Signer: "invalid"}.GetSigners() })
}

func TestMsgSetVerificationRelationships_Route(t *testing.T) {
	assert.Equalf(t, didtypes.ModuleName, didtypes.MsgSetVerificationRelationships{}.Route(), "Route()")
}

func TestMsgSetVerificationRelationships_Type(t *testing.T) {
	assert.Equalf(t, didtypes.TypeMsgSetVerificationRelationships, didtypes.MsgSetVerificationRelationships{}.Type(), "Type()")
}

func TestMsgSetVerificationRelationships_GetSignBytes(t *testing.T) {
	assert.Panicsf(t, func() { didtypes.MsgSetVerificationRelationships{}.GetSignBytes() }, "GetSignBytes()")
}

func TestMsgSetVerificationRelationships_GetSigners(t *testing.T) {
	a, err := sdk.AccAddressFromBech32("cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
	assert.NoError(t, err)
	assert.Equal(t,
		didtypes.MsgSetVerificationRelationships{Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"}.GetSigners(),
		[]sdk.AccAddress{a},
	)
	assert.Panics(t, func() { didtypes.MsgSetVerificationRelationships{Signer: "invalid"}.GetSigners() })
}

func TestMsgDeleteService_Route(t *testing.T) {
	assert.Equalf(t, didtypes.ModuleName, didtypes.MsgDeleteService{}.Route(), "Route()")
}

func TestMsgDeleteService_Type(t *testing.T) {
	assert.Equalf(t, didtypes.TypeMsgDeleteService, didtypes.MsgDeleteService{}.Type(), "Type()")
}

func TestMsgDeleteService_GetSignBytes(t *testing.T) {
	assert.Panicsf(t, func() { didtypes.MsgDeleteService{}.GetSignBytes() }, "GetSignBytes()")
}

func TestMsgDeleteService_GetSigners(t *testing.T) {
	a, err := sdk.AccAddressFromBech32("cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
	assert.NoError(t, err)
	assert.Equal(t,
		didtypes.MsgDeleteService{Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"}.GetSigners(),
		[]sdk.AccAddress{a},
	)
	assert.Panics(t, func() { didtypes.MsgDeleteService{Signer: "invalid"}.GetSigners() })
}

func TestMsgAddService_Route(t *testing.T) {
	assert.Equalf(t, didtypes.ModuleName, didtypes.MsgAddService{}.Route(), "Route()")
}

func TestMsgAddService_Type(t *testing.T) {
	assert.Equalf(t, didtypes.TypeMsgAddService, didtypes.MsgAddService{}.Type(), "Type()")
}

func TestMsgAddService_GetSignBytes(t *testing.T) {
	assert.Panicsf(t, func() { didtypes.MsgAddService{}.GetSignBytes() }, "GetSignBytes()")
}

func TestMsgAddService_GetSigners(t *testing.T) {
	a, err := sdk.AccAddressFromBech32("cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
	assert.NoError(t, err)
	assert.Equal(t,
		didtypes.MsgAddService{Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"}.GetSigners(),
		[]sdk.AccAddress{a},
	)
	assert.Panics(t, func() { didtypes.MsgAddService{Signer: "invalid"}.GetSigners() })
}

func TestMsgAddController_Route(t *testing.T) {
	assert.Equalf(t, didtypes.ModuleName, didtypes.MsgAddController{}.Route(), "Route()")
}

func TestMsgAddController_Type(t *testing.T) {
	assert.Equalf(t, didtypes.TypeMsgAddController, didtypes.MsgAddController{}.Type(), "Type()")
}

func TestMsgAddController_GetSignBytes(t *testing.T) {
	assert.Panicsf(t, func() { didtypes.MsgAddController{}.GetSignBytes() }, "GetSignBytes()")
}

func TestMsgAddController_GetSigners(t *testing.T) {
	a, err := sdk.AccAddressFromBech32("cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
	assert.NoError(t, err)
	assert.Equal(t,
		didtypes.MsgAddController{Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"}.GetSigners(),
		[]sdk.AccAddress{a},
	)
	assert.Panics(t, func() { didtypes.MsgAddController{Signer: "invalid"}.GetSigners() })
}

func TestMsgDeleteController_Route(t *testing.T) {
	assert.Equalf(t, didtypes.ModuleName, didtypes.MsgDeleteController{}.Route(), "Route()")
}

func TestMsgDeleteController_Type(t *testing.T) {
	assert.Equalf(t, didtypes.TypeMsgDeleteController, didtypes.MsgDeleteController{}.Type(), "Type()")
}

func TestMsgDeleteController_GetSignBytes(t *testing.T) {
	assert.Panicsf(t, func() { didtypes.MsgDeleteController{}.GetSignBytes() }, "GetSignBytes()")
}

func TestMsgDeleteController_GetSigners(t *testing.T) {
	a, err := sdk.AccAddressFromBech32("cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
	assert.NoError(t, err)
	assert.Equal(t,
		didtypes.MsgDeleteController{Signer: "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"}.GetSigners(),
		[]sdk.AccAddress{a},
	)
	assert.Panics(t, func() { didtypes.MsgDeleteController{Signer: "invalid"}.GetSigners() })
}

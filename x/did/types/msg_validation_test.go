package types_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"

	didtypes "github.com/fetchai/fetchd/x/did/types"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateDidDocument(t *testing.T) {
	tests := []struct {
		id            string
		verifications didtypes.Verifications
		services      didtypes.Services
		owner         string
		expectPass    bool
	}{
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:auth:whatever",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					[]string{},
				},
			},
			didtypes.Services{},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			true,
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						didtypes.DIDVMethodTypeCosmosAccountAddress.String(),
						"did:auth:whatever",
						&didtypes.VerificationMethod_BlockchainAccountID{""},
					},
					[]string{},
				},
			},
			didtypes.Services{},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // empty pub key
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						"",
						"did:auth:whatever",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					[]string{},
				},
			},
			didtypes.Services{},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // emtpy verification method type
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:auth:whatever",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					[]string{},
				},
			},
			didtypes.Services{},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // empty relationships
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#/asd 123",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:auth:whatever",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					[]string{},
				},
			},
			didtypes.Services{},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // invalid method id
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:auth:whatever",
						&didtypes.VerificationMethod_PublicKeyHex{""},
					},
					[]string{},
				},
			},
			didtypes.Services{},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // empty verification key
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						"",
						"did:auth:whatever",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					[]string{},
				},
			},
			didtypes.Services{},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // empty verification method type
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						didtypes.DIDVMethodTypeCosmosAccountAddress.String(),
						"",
						&didtypes.VerificationMethod_BlockchainAccountID{"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"},
					},
					[]string{},
				},
			},
			didtypes.Services{},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // invalid verification method controller
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{},
			didtypes.Services{},
			"owner",
			false, // empty verifications
		},

		{
			"invalid did",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"cont",
						&didtypes.VerificationMethod_BlockchainAccountID{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					[]string{},
				},
			},
			didtypes.Services{},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // invalid did
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:auth:whatever",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					[]string{},
				},
			},
			didtypes.Services{
				&didtypes.Service{
					"the:agent:service",
					"DIDCommMessaging",
					"https://agent.xyz/agent/123",
				},
			},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			true,
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:auth:whatever",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					[]string{},
				},
			},
			didtypes.Services{
				&didtypes.Service{
					"the:agent:service",
					"",
					"https://agent.xyz/agent/123",
				},
			},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // empty service type
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:auth:whatever",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					[]string{},
				},
			},
			didtypes.Services{
				&didtypes.Service{
					"",
					"DIDCommMessaging",
					"https://agent.xyz/agent/123",
				},
			},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // service id is not valid
		},
		{
			"did:auth:whatever",
			didtypes.Verifications{
				&didtypes.Verification{
					[]string{string(didtypes.Authentication)},
					&didtypes.VerificationMethod{
						"did:auth:whatever#1",
						didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
						"did:auth:whatever",
						&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
					},
					[]string{},
				},
			},
			didtypes.Services{
				&didtypes.Service{
					"this:is:fine",
					"DIDCommMessaging",
					"",
				},
			},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // service id is not valid
		},
	}

	for i, tc := range tests {
		msg := didtypes.NewMsgCreateDidDocument(
			tc.id,
			tc.verifications,
			tc.services,
			tc.owner,
		)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: TestMsgCreateDidDocument#%v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: TestMsgCreateDidDocument#%v", i)
		}
	}
}

func TestMsgUpdateDidDocument(t *testing.T) {
	tests := []struct {
		id          string
		controllers []string
		signer      string
		expectPass  bool
	}{
		{
			"did:cash:subject",
			[]string{"did:cash:controller-1"},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			true,
		},
		{
			"did:cash:subject",
			[]string{},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			true,
		},
		{
			// FIXME: duplicated controller
			"did:cash:subject",
			[]string{"did:cash:controller-1", "did:cash:controller-1"},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			true,
		},
		{
			"invalid did",
			[]string{"did:cash:controller-1"},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // invalid did
		},
		{
			"did:cash:subject",
			[]string{"invalid:controller"},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // invalid controller
		},
		{
			"did:cash:subject",
			[]string{"did:cash:controller-1", "did:cash:controller-2", ""},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // invalid controller
		},
	}

	for i, tc := range tests {
		msg := didtypes.NewMsgUpdateDidDocument(
			&didtypes.DidDocument{Id: tc.id, Controller: tc.controllers},
			tc.signer,
		)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: TestMsgUpdateDidDocument#%v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: TestMsgUpdateDidDocument#%v", i)
		}
	}
}

func TestMsgAddVerification(t *testing.T) {
	tests := []struct {
		id         string
		auth       didtypes.Verification
		owner      string
		expectPass bool
	}{
		{
			"did:cash:subject",
			didtypes.Verification{
				[]string{string(didtypes.Authentication)},
				&didtypes.VerificationMethod{
					"did:cash:subject#1",
					didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
					"did:cash:subject",
					&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
				},
				[]string{},
			},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			true,
		},
		{
			"something not right",
			didtypes.Verification{
				[]string{string(didtypes.Authentication)},
				&didtypes.VerificationMethod{
					"did:cash:subject#1",
					didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019.String(),
					"did:cash:subject",
					&didtypes.VerificationMethod_PublicKeyHex{"03dfd0a469806d66a23c7c948f55c129467d6d0974a222ef6e24a538fa6882f3d7"},
				},
				[]string{},
			},
			"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
			false, // invalid did
		},
	}

	for i, tc := range tests {
		msg := didtypes.NewMsgAddVerification(
			tc.id,
			&tc.auth,
			tc.owner,
		)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: TestMsgAddVerification#%v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: TestMsgAddVerification#%v", i)
		}
	}
}

func TestMsgRevokeVerification(t *testing.T) {
	tests := []struct {
		id         string
		key        string
		signer     string
		expectPass bool
	}{
		{
			"did:cash:subject",
			"did:cash:subject#key-method-1",
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			true,
		},
		{
			"invalid did",
			"did:cash:subject#key-method-1",
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // invalid did
		},
		{
			"did:cash:subject",
			"did:cash:subject  #   key-method-1",
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // invalid method id
		},
		{
			"did:cash:subject",
			"did:cash:subject#key-method-1",
			"",
			true, // empty signer
		},
	}

	for i, tc := range tests {
		msg := didtypes.NewMsgRevokeVerification(
			tc.id,
			tc.key,
			tc.signer,
		)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: TestMsgRevokeVerification#%v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: TestMsgRevokeVerification#%v", i)
		}
	}
}

func TestMsgSetVerificationRelationships(t *testing.T) {
	tests := []struct {
		id            string
		key           string
		relationships []string
		signer        string
		expectPass    bool
	}{
		{
			"did:cash:subject",
			"did:cash:subject#key-method-1",
			[]string{"authorization", "keyExchange"},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			true,
		},
		{
			"did:cash:subject",
			"did:cash:subject#key-method-1",
			[]string{"authorization"},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			true,
		},
		{
			"did:cash:subject",
			"did:cash:subject  #   key-method-1",
			[]string{"authorization", "keyExchange"},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // invalid method id
		},
		{
			"invalid did",
			"did:cash:subject#key-method-1",
			[]string{"authorization", "keyExchange"},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // invalid did
		},
		{
			"did:cash:subject",
			"did:cash:subject#key-method-1",
			[]string{},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // empty relationship
		},
		{
			"did:cash:subject",
			"did:cash:subject#key-method-1",
			[]string{"authorization", "keyExchange"},
			"",
			true, // empty signer
		},
	}

	for i, tc := range tests {
		t.Logf("TestMsgRevokeVerification#%d", i)
		msg := didtypes.NewMsgSetVerificationRelationships(
			tc.id,
			tc.key,
			tc.relationships,
			tc.signer,
		)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: TestMsgSetVerificationRelationships#%v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: TestMsgSetVerificationRelationships#%v", i)
		}
	}
}

func TestMsgAddService(t *testing.T) {
	tests := []struct {
		id         string
		service    *didtypes.Service
		signer     string
		expectPass bool
	}{
		{
			"did:cash:subject",
			&didtypes.Service{
				Id:              "a:valid:url",
				Type:            "DIDCommMessaging",
				ServiceEndpoint: "https://agent.xyz/validate",
			},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			true,
		},
		{
			"invalid did",
			&didtypes.Service{
				Id:              "my:agent",
				Type:            "DIDCommMessaging",
				ServiceEndpoint: "https://agent.xyz/validate",
			},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // invalid did
		},
		{
			"did:cash:subject",
			&didtypes.Service{
				Id:              "",
				Type:            "DIDCommMessaging",
				ServiceEndpoint: "https://agent.xyz/validate",
			},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // invalid agent id
		},
		{
			"did:cash:subject",
			&didtypes.Service{
				Id:              "my:agent",
				Type:            "",
				ServiceEndpoint: "https://agent.xyz/validate",
			},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // empty type
		},
		{
			"did:cash:subject",
			&didtypes.Service{
				Id:              "my:agent",
				Type:            "DIDCommMessaging",
				ServiceEndpoint: "",
			},
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // empty service endpoint
		},
	}

	for i, tc := range tests {
		t.Logf("TestMsgRevokeVerification#%d", i)
		msg := didtypes.NewMsgAddService(
			tc.id,
			tc.service,
			tc.signer,
		)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: TestMsgAddService#%v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: TestMsgAddService#%v", i)
		}
	}
}

func TestMsgDeleteService(t *testing.T) {
	tests := []struct {
		id         string
		serviceID  string
		signer     string
		expectPass bool
	}{
		{
			"did:cash:subject",
			"my:service:uri",
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			true,
		},
		{
			"invalid did",
			"my:service:uri",
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, //invalid did
		},
		{
			"did:cash:subject",
			"",
			"cosmos1uam3kpjdx3wksx46lzq6y628wwyzv0guuren75",
			false, // empty service id
		},
	}

	for i, tc := range tests {
		t.Logf("TestMsgRevokeVerification#%d", i)
		msg := didtypes.NewMsgDeleteService(
			tc.id,
			tc.serviceID,
			tc.signer,
		)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: TestMsgDeleteService#%v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: TestMsgDeleteService#%v", i)
		}
	}
}

func TestMsgAddController_ValidateBasic(t *testing.T) {
	type fields struct {
		Id            string
		ControllerDid string
		Signer        string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			"PASS: controller is valid",
			fields{
				Id:            "did:cosmos:net:foochain:12345",
				ControllerDid: "did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
				Signer:        "cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
			},
			nil,
		},
		{
			"FAIL: invalid did",
			fields{
				Id:            "not a did",
				ControllerDid: "did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
				Signer:        "cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
			},
			sdkerrors.Wrap(didtypes.ErrInvalidDIDFormat, "not a did"),
		},
		{
			"FAIL: invalid controller did",
			fields{
				Id:            "did:cosmos:net:foochain:12345",
				ControllerDid: "did:cosmos:net:foochain:whatever",
				Signer:        "cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
			},
			sdkerrors.Wrap(didtypes.ErrInvalidDIDFormat, "did:cosmos:net:foochain:whatever"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := didtypes.MsgAddController{
				Id:            tt.fields.Id,
				ControllerDid: tt.fields.ControllerDid,
				Signer:        tt.fields.Signer,
			}

			err := msg.ValidateBasic()
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
		})
	}
}

func TestMsgDeleteController_ValidateBasic(t *testing.T) {
	type fields struct {
		Id            string
		ControllerDid string
		Signer        string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			"PASS: controller is valid",
			fields{
				Id:            "did:cosmos:net:foochain:12345",
				ControllerDid: "did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
				Signer:        "cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
			},
			nil,
		},
		{
			"FAIL: invalid did",
			fields{
				Id:            "not a did",
				ControllerDid: "did:cosmos:key:cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
				Signer:        "cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
			},
			sdkerrors.Wrap(didtypes.ErrInvalidDIDFormat, "not a did"),
		},
		{
			"FAIL: invalid controller did",
			fields{
				Id:            "did:cosmos:net:foochain:12345",
				ControllerDid: "not a did",
				Signer:        "cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
			},
			sdkerrors.Wrap(didtypes.ErrInvalidDIDFormat, "not a did"),
		}, // TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := didtypes.MsgDeleteController{
				Id:            tt.fields.Id,
				ControllerDid: tt.fields.ControllerDid,
				Signer:        tt.fields.Signer,
			}
			err := msg.ValidateBasic()
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
		})
	}
}

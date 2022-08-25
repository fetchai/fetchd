package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/fetchai/fetchd/x/did/types"
	"github.com/fetchai/fetchd/x/did/keeper"
)

func (s *KeeperTestSuite) TestHandleMsgCreateDidDocument() {
	var (
		req    types.MsgCreateDidDocument
		errExp error
	)

	server := keeper.NewMsgServerImpl(s.app.DidKeeper)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"Pass: can create a an did",
			func() {
				req = *types.NewMsgCreateDidDocument("did:cosmos:cash:subject", nil, nil, "subject")
				errExp = nil
			},
		},
		{
			"FAIL: did doc validation fails",
			func() {
				req = *types.NewMsgCreateDidDocument("invalid did", nil, nil, "subject")
				errExp = sdkerrors.Wrapf(types.ErrInvalidDIDFormat, "did %s", "invalid did")
			},
		},
		{
			"FAIL: did already exists",
			func() {
				did := "did:cosmos:cash:subject"
				didDoc, _ := types.NewDidDocument(did)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				req = *types.NewMsgCreateDidDocument(did, nil, nil, "subject")
				errExp = sdkerrors.Wrapf(types.ErrDidDocumentFound, "a document with did %s already exists", did)
			},
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			tc.malleate()
			_, err := server.CreateDidDocument(s.ctx, &req)
			if errExp == nil {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(errExp.Error(), err.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestHandleMsgUpdateDidDocument() {
	var (
		req    types.MsgUpdateDidDocument
		errExp error
	)

	server := keeper.NewMsgServerImpl(s.app.DidKeeper)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"FAIL: not found",
			func() {
				req = *types.NewMsgUpdateDidDocument(&types.DidDocument{Id: "did:cosmos:cash:subject"}, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrDidDocumentNotFound, "did document at %s not found", "did:cosmos:cash:subject")
			},
		},
		{
			"FAIL: unauthorized",
			func() {

				did := "did:cosmos:cash:subject"
				didDoc, _ := types.NewDidDocument(did)
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)

				req = *types.NewMsgUpdateDidDocument(&types.DidDocument{Id: didDoc.Id, Controller: []string{"did:cosmos:cash:controller"}}, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrUnauthorized, "signer account %s not authorized to update the target did document at %s", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8", did)

			},
		},
		{
			"PASS: replace did document",
			func() {

				did := "did:cosmos:cash:subject"
				didDoc, _ := types.NewDidDocument(did, types.WithVerifications(
					types.NewVerification(
						types.NewVerificationMethod(
							"did:cosmos:cash:subject#key-1",
							"did:cosmos:cash:subject",
							types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{types.Authentication},
						nil,
					),
				))
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))

				newDidDoc, err := types.NewDidDocument(did)
				s.Require().Nil(err)

				req = *types.NewMsgUpdateDidDocument(&newDidDoc, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = nil
			},
		},
		{
			"FAIL: invalid controllers",
			func() {
				didDoc, _ := types.NewDidDocument("did:cosmos:cash:subject", types.WithVerifications(
					types.NewVerification(
						types.NewVerificationMethod(
							"did:cosmos:cash:subject#key-1",
							"did:cosmos:cash:subject",
							types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
						),
						[]string{types.Authentication},
						nil,
					),
				))
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)

				controllers := []string{
					"did:cosmos:cash:controller-1",
					"did:cosmos:cash:controller-2",
					"invalid",
				}

				req = *types.NewMsgUpdateDidDocument(&types.DidDocument{Id: didDoc.Id, Controller: controllers}, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrInvalidDIDFormat, "invalid did document")
			},
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			tc.malleate()

			_, err := server.UpdateDidDocument(sdk.WrapSDKContext(s.sdkCtx), &req)

			if errExp == nil {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(errExp.Error(), err.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestHandleMsgAddVerification() {
	var (
		req    types.MsgAddVerification
		errExp error
	)

	server := keeper.NewMsgServerImpl(s.app.DidKeeper)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"FAIL: can not add verification, did does not exist",
			func() {
				req = *types.NewMsgAddVerification("did:cosmos:cash:subject", nil, "subject")
				errExp = sdkerrors.Wrapf(types.ErrDidDocumentNotFound, "did document at %s not found", "did:cosmos:cash:subject")
			},
		},
		{
			"FAIL: can not add verification, unauthorized",
			func() {
				// setup
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.CapabilityInvocation},
							nil,
						),
					),
				)
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				// actual test
				v := types.NewVerification(
					types.NewVerificationMethod(
						"did:cosmos:cash:subject#key-2",
						"did:cosmos:cash:subject",
						types.NewBlockchainAccountID("foochainid", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
					),
					[]string{types.Authentication},
					nil,
				)
				req = *types.NewMsgAddVerification(didDoc.Id, v, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrUnauthorized, "signer account %s not authorized to update the target did document at %s", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8", didDoc.Id)
			},
		},
		{
			"FAIL: can not add verification, unauthorized, key mismatch",
			func() {
				// setup
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				// actual test
				v := types.NewVerification(
					types.NewVerificationMethod(
						"did:cosmos:cash:subject#key-2",
						"did:cosmos:cash:subject",
						types.NewBlockchainAccountID("foochainid", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
					),
					[]string{types.Authentication},
					nil,
				)
				req = *types.NewMsgAddVerification(didDoc.Id, v, "cash1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2")
				errExp = sdkerrors.Wrapf(types.ErrUnauthorized, "signer account %s not authorized to update the target did document at %s", "cash1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2", didDoc.Id)
			},
		},
		{
			"FAIL: can not add verification, invalid verification",
			func() {
				// setup
				//signer := "subject"
				signer := "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				// actual test
				v := types.NewVerification(
					types.NewVerificationMethod(
						"",
						"did:cosmos:cash:subject",
						types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
					),
					[]string{types.Authentication},
					nil,
				)
				req = *types.NewMsgAddVerification(didDoc.Id, v, signer)
				errExp = sdkerrors.Wrapf(types.ErrInvalidDIDURLFormat, "verification method id: %v", "")
			},
		},
		{
			"PASS: can add verification to did document",
			func() {
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)

				v := types.NewVerification(
					types.NewVerificationMethod(
						"did:cosmos:cash:subject#key-2",
						"did:cosmos:cash:subject",
						types.NewBlockchainAccountID("foochainid", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"),
					),
					[]string{types.Authentication},
					nil,
				)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))
				req = *types.NewMsgAddVerification(didDoc.Id, v, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = nil
			},
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			tc.malleate()

			_, err := server.AddVerification(s.ctx, &req)

			if errExp == nil {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(errExp.Error(), err.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestHandleMsgSetVerificationRelationships() {
	var (
		req    types.MsgSetVerificationRelationships
		errExp error
	)

	server := keeper.NewMsgServerImpl(s.app.DidKeeper)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"FAIL: can not add verification relationship, did does not exist",
			func() {
				req = *types.NewMsgSetVerificationRelationships(
					"did:cosmos:cash:subject",
					"did:cosmos:cash:subject#key-1",
					[]string{types.Authentication},
					"subject",
				)
				errExp = sdkerrors.Wrapf(types.ErrDidDocumentNotFound, "did document at %s not found", "did:cosmos:cash:subject")
			},
		},
		{
			"FAIL: can not add verification relationship, unauthorized",
			func() {
				// setup
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.CapabilityInvocation},
							nil,
						),
					),
				)
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				// actual test
				req = *types.NewMsgSetVerificationRelationships(
					"did:cosmos:cash:subject",
					"did:cosmos:cash:subject#key-1",
					[]string{types.Authentication},
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
				)

				errExp = sdkerrors.Wrapf(types.ErrUnauthorized, "signer account %s not authorized to update the target did document at %s", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8", "did:cosmos:cash:subject")
			},
		},
		{
			"FAIL: can not add verification relationship, invalid relationship provided",
			func() {
				// setup
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								//"did:cosmos:cash:subject#cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				// actual test
				req = *types.NewMsgSetVerificationRelationships(
					"did:cosmos:cash:subject",
					"did:cosmos:cash:subject#key-1",
					nil,
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
				)
				errExp = sdkerrors.Wrap(types.ErrEmptyRelationships, "at least a verification relationship is required")
			},
		},
		{
			"FAIL: verification method does not exist ",
			func() {
				// setup
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				// actual test
				req = *types.NewMsgSetVerificationRelationships(
					"did:cosmos:cash:subject",
					"did:cosmos:cash:subject#key-does-not-exists",
					[]string{types.Authentication, types.CapabilityInvocation},
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
				)
				errExp = sdkerrors.Wrapf(types.ErrVerificationMethodNotFound, "verification method %v not found", "did:cosmos:cash:subject#key-does-not-exists")
			},
		},
		{
			"PASS: add a new relationship",
			func() {
				// setup
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))
				// actual test
				req = *types.NewMsgSetVerificationRelationships(
					"did:cosmos:cash:subject",
					"did:cosmos:cash:subject#key-1",
					[]string{types.Authentication, types.CapabilityInvocation},
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
				)
				errExp = nil
			},
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			tc.malleate()

			_, err := server.SetVerificationRelationships(s.ctx, &req)

			if errExp == nil {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(errExp.Error(), err.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestHandleMsgRevokeVerification() {
	var (
		req    types.MsgRevokeVerification
		errExp error
	)

	server := keeper.NewMsgServerImpl(s.app.DidKeeper)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"FAIL: can not revoke verification, did does not exist",
			func() {
				req = *types.NewMsgRevokeVerification("did:cosmos:cash:2222", "service-id", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrDidDocumentNotFound, "did document at %s not found", "did:cosmos:cash:2222")
			},
		},
		{
			"FAIL: can not revoke verification, not found",
			func() {
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				req = *types.NewMsgRevokeVerification(didDoc.Id, "did:cosmos:cash:subject#not-existent", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrVerificationMethodNotFound, "verification method id: %v", "did:cosmos:cash:subject#not-existent")
			},
		},
		{
			"FAIL: can not revoke verification, unauthorized",
			func() {
				signer := "controller-1"
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.CapabilityDelegation},
							nil,
						),
					),
				)

				vmID := "did:cosmos:cash:subject#key-1"

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				// controller-1 does not exists
				req = *types.NewMsgRevokeVerification(didDoc.Id, vmID, signer)

				errExp = sdkerrors.Wrapf(types.ErrUnauthorized, "signer account %s not authorized to update the target did document at %s", signer, didDoc.Id)
			},
		},
		{
			"PASS: can revoke verification",
			func() {
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))
				req = *types.NewMsgRevokeVerification(didDoc.Id,
					"did:cosmos:cash:subject#key-1",
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8",
				)
				errExp = nil
			},
		},
	}
	for i, tc := range testCases {
		s.Run(fmt.Sprintf("TestHandleMsgRevokeVerification#%v", i), func() {
			tc.malleate()

			_, err := server.RevokeVerification(s.ctx, &req)

			if errExp == nil {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(errExp.Error(), err.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestHandleMsgAddService() {
	var (
		req    types.MsgAddService
		errExp error
	)

	server := keeper.NewMsgServerImpl(s.app.DidKeeper)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"FAIL: can not add service, did does not exist",
			func() {
				service := types.NewService(
					"service-id",
					"NonUserCredential",
					"cash/multihash",
				)
				req = *types.NewMsgAddService("did:cosmos:cash:subject", service, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrDidDocumentNotFound, "did document at %s not found", "did:cosmos:cash:subject")
			},
		},
		{
			"FAIL: can not add service, service not defined",
			func() {
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)
				// create the did
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				// try adding a service
				req = *types.NewMsgAddService("did:cosmos:cash:subject", nil, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrap(types.ErrInvalidInput, "service is not defined")
			},
		},
		{
			"FAIL: cannot add service to did document (unauthorized, wrong relationship)",
			func() {
				signer := "subject"
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.CapabilityInvocation, types.CapabilityDelegation},
							nil,
						),
					),
				)

				service := types.NewService(
					"service-id",
					"UserCredential",
					"cash/multihash",
				)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				req = *types.NewMsgAddService(didDoc.Id, service, signer)

				errExp = sdkerrors.Wrapf(types.ErrUnauthorized, "signer account %s not authorized to update the target did document at %s", signer, didDoc.Id)
			},
		},
		{
			"FAIL: cannot add service to did document with an empty type",
			func() {
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)

				service := types.NewService(
					"service-id",
					"",
					"cash/multihash",
				)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				req = *types.NewMsgAddService(didDoc.Id, service, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrap(types.ErrInvalidInput, "service type cannot be empty")
			},
		},
		{
			"FAIL: duplicated service",
			func() {
				//signer := "subject"
				signer := "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8"
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
					types.WithServices(
						types.NewService(
							"service-id",
							"UserCredential",
							"cash/multihash",
						),
					),
				)

				service := types.NewService(
					"service-id",
					"UserCredential",
					"cash/multihash",
				)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				req = *types.NewMsgAddService(didDoc.Id, service, signer)
				errExp = sdkerrors.Wrapf(types.ErrInvalidInput, "duplicated verification method id %s", "service-id")
			},
		},
		{
			"PASS: can add service to did document",
			func() {
				signer := "subject"
				didDoc, err := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewBlockchainAccountID("foochainid", signer),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)

				if err != nil {
					s.FailNow("test setup failed: ", err)
				}

				service := types.NewService(
					"service-id",
					"UserCredential",
					"cash/multihash",
				)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))

				req = *types.NewMsgAddService(didDoc.Id, service, signer)
				errExp = nil
			},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.malleate()

			_, err := server.AddService(s.ctx, &req)

			if errExp == nil {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(errExp.Error(), err.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestHandleMsgDeleteService() {
	var (
		req    types.MsgDeleteService
		errExp error
	)

	server := keeper.NewMsgServerImpl(s.app.DidKeeper)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"FAIL: can not delete service, did does not exist",
			func() {
				req = *types.NewMsgDeleteService("did:cosmos:cash:2222", "service-id", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrDidDocumentNotFound, "did document at %s not found", "did:cosmos:cash:2222")
			},
		},
		{

			"PASS: can delete service from did document",
			func() {
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
					types.WithServices(
						types.NewService(
							"service-id",
							"UserCredential",
							"cash/multihash",
						),
					),
				)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))
				req = *types.NewMsgDeleteService(didDoc.Id, "service-id", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = nil
			},
		},
		{
			"FAIL: cannot remove an invalid serviceID",
			func() {

				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)

				serviceID := ""

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				req = *types.NewMsgDeleteService(didDoc.Id, serviceID, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrInvalidState, "the did document doesn't have services associated")
			},
		},
		{
			"FAIL: unauthorized (wrong relationship)",
			func() {
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								"did:cosmos:cash:subject",
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.CapabilityInvocation},
							nil,
						),
					),
				)

				serviceID := "service-id"

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				req = *types.NewMsgDeleteService(didDoc.Id, serviceID, "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrUnauthorized, "signer account %s not authorized to update the target did document at %s", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8", didDoc.Id)
			},
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf(tc.name), func() {
			tc.malleate()

			_, err := server.DeleteService(s.ctx, &req)

			if errExp == nil {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(errExp.Error(), err.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestHandleMsgAddController() {
	var (
		req    types.MsgAddController
		errExp error
	)

	server := keeper.NewMsgServerImpl(s.app.DidKeeper)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"FAIL: cannot add controller, did doesn't exist",
			func() {
				req = *types.NewMsgAddController(
					"did:cosmos:cash:subject",
					"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrDidDocumentNotFound, "did document at %s not found", "did:cosmos:cash:subject")
			},
		},
		{
			"FAIL: controller is not a valid did",
			func() {
				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)
				// create the did
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				// try adding a service
				req = *types.NewMsgAddController("did:cosmos:cash:subject", "", "cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrap(types.ErrInvalidDIDFormat, "did document controller validation error ''")
			},
		},
		{
			"FAIL: signer not authorized to change controller",
			func() {

				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.CapabilityInvocation, types.CapabilityDelegation},
							nil,
						),
					),
				)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)

				req = *types.NewMsgAddController(
					"did:cosmos:cash:subject",
					"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
					"cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2", // does not match the pub key (it's the new controller)
				)

				errExp = sdkerrors.Wrapf(types.ErrUnauthorized, "signer account %s not authorized to update the target did document at %s", "cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2", didDoc.Id)
			},
		},
		{
			"FAIL: controller is not type key",
			func() {

				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)

				req = *types.NewMsgAddController(
					"did:cosmos:cash:subject",
					"did:cosmos:net:foochain:whatever",
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8", // does not match the pub key (it's the new controller)
				)

				errExp = sdkerrors.Wrapf(types.ErrInvalidInput, "did document controller 'did:cosmos:net:foochain:whatever' must be of type key")
			},
		},
		{
			"PASS: can add controller (via authentication relationship)",
			func() {
				didDoc, err := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)

				if err != nil {
					s.FailNow("test setup failed: ", err)
				}

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))

				req = *types.NewMsgAddController(
					"did:cosmos:cash:subject",
					"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = nil
			},
		},
		{
			"PASS: can add controller (via controller)",
			func() {
				didDoc, err := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.AssertionMethod},
							nil,
						),
					),
					types.WithControllers("did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2"),
				)

				if err != nil {
					s.FailNow("test setup failed: ", err)
				}

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))

				req = *types.NewMsgAddController(
					"did:cosmos:cash:subject",
					"did:cosmos:key:cosmos17t8t3t6a6vpgk69perfyq930593sa8dn4kzsdf",
					"cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2") // this is the controller

				errExp = nil
			},
		},
		{
			"PASS: controller already added (duplicated)",
			func() {
				didDoc, err := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)

				if err != nil {
					s.FailNow("test setup failed: ", err)
				}

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))

				req = *types.NewMsgAddController(
					"did:cosmos:cash:subject",
					"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = nil
			},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.malleate()

			_, err := server.AddController(s.ctx, &req)

			if errExp == nil {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(errExp.Error(), err.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestHandleMsgDeleteController() {
	var (
		req    types.MsgDeleteController
		errExp error
	)

	server := keeper.NewMsgServerImpl(s.app.DidKeeper)

	// FAIL: cannot delete controller, did doesn't exist
	// FAIL: signer not authorized to change controller
	// PASS: controller removed
	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"FAIL: cannot delete controller, did doesn't exist",
			func() {
				req = *types.NewMsgDeleteController(
					"did:cosmos:cash:subject",
					"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = sdkerrors.Wrapf(types.ErrDidDocumentNotFound, "did document at %s not found", "did:cosmos:cash:subject")
			},
		},
		{
			"FAIL: signer not authorized to change controller",
			func() {

				didDoc, _ := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.CapabilityInvocation, types.CapabilityDelegation},
							nil,
						),
					),
				)

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)

				req = *types.NewMsgDeleteController(
					"did:cosmos:cash:subject",
					"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
					"cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2", // does not match the pub key (it's the new controller)
				)

				errExp = sdkerrors.Wrapf(types.ErrUnauthorized, "signer account %s not authorized to update the target did document at %s", "cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2", didDoc.Id)
			},
		},
		{
			"PASS: can delete controller (via authentication relationship)",
			func() {
				didDoc, err := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.Authentication},
							nil,
						),
					),
				)

				if err != nil {
					s.FailNow("test setup failed: ", err)
				}

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))

				req = *types.NewMsgDeleteController(
					"did:cosmos:cash:subject",
					"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
					"cosmos1sl48sj2jjed7enrv3lzzplr9wc2f5js5tzjph8")
				errExp = nil
			},
		}, {
			"PASS: can delete controller (via controller)",
			func() {
				didDoc, err := types.NewDidDocument(
					"did:cosmos:cash:subject",
					types.WithVerifications(
						types.NewVerification(
							types.NewVerificationMethod(
								"did:cosmos:cash:subject#key-1",
								types.DID("did:cosmos:cash:subject"),
								types.NewPublicKeyMultibase([]byte{3, 223, 208, 164, 105, 128, 109, 102, 162, 60, 124, 148, 143, 85, 193, 41, 70, 125, 109, 9, 116, 162, 34, 239, 110, 36, 165, 56, 250, 104, 130, 243, 215}, types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
							),
							[]string{types.AssertionMethod},
							nil,
						),
					),
					types.WithControllers(
						"did:cosmos:key:cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2",
						"did:cosmos:key:cosmos17t8t3t6a6vpgk69perfyq930593sa8dn4kzsdf",
					),
				)

				if err != nil {
					s.FailNow("test setup failed: ", err)
				}

				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
				s.app.DidKeeper.SetDidMetadata(s.sdkCtx, []byte(didDoc.Id), types.NewDidMetadata([]byte{1}, time.Now()))

				req = *types.NewMsgDeleteController(
					"did:cosmos:cash:subject",
					"did:cosmos:key:cosmos17t8t3t6a6vpgk69perfyq930593sa8dn4kzsdf",
					"cosmos1lvl2s8x4pta5f96appxrwn3mypsvumukvk7ck2", // this is the controller
				)

				errExp = nil
			},
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.malleate()

			_, err := server.DeleteController(s.ctx, &req)

			if errExp == nil {
				s.Require().NoError(err)
			} else {
				s.Require().Error(err)
				s.Require().Equal(errExp.Error(), err.Error())
			}
		})
	}
}

package keeper_test

import (
	"fmt"
	"time"

	"github.com/fetchai/fetchd/x/verifiable-credential/crypto/accumulator"
	"github.com/fetchai/fetchd/x/verifiable-credential/crypto/anonymouscredential"
	"github.com/fetchai/fetchd/x/verifiable-credential/crypto/bbsplus"

	didtypes "github.com/fetchai/fetchd/x/did/types"
	"github.com/fetchai/fetchd/x/verifiable-credential/keeper"
	"github.com/fetchai/fetchd/x/verifiable-credential/types"
)

func (s *KeeperTestSuite) TestMsgSeverIssueRegistrationCredential() {
	server := keeper.NewMsgServerImpl(s.app.VcKeeper)
	var req types.MsgIssueRegistrationCredential

	testCases := []struct {
		msg       string
		malleate  func()
		expectErr error
	}{
		{
			msg:       "PASS: issuer can issue registration credential for alice",
			expectErr: nil,
			malleate: func() {
				var vc types.VerifiableCredential
				issuerDid := didtypes.DID("did:cosmos:net:test:issuer")
				aliceDid := didtypes.DID("did:cosmos:net:test:alice")
				issuerAddress := s.GetIssuerAddress()
				vc = types.NewRegistrationVerifiableCredential(
					"alice-registraion-credential",
					issuerDid.String(),
					time.Now(),
					types.NewRegistrationCredentialSubject(
						aliceDid.String(),
						"EU",
						"emti",
						"E-Money Token Issuer",
					),
				)
				vc, _ = vc.Sign(
					s.keyring, s.GetIssuerAddress(),
					issuerDid.NewVerificationMethodID(issuerAddress.String()),
				)
				req = types.MsgIssueRegistrationCredential{
					Credential: &vc,
					Owner:      issuerAddress.String(),
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			didResp, err := server.IssueRegistrationCredential(s.ctx, &req)
			if tc.expectErr == nil {
				s.NoError(err)
				s.NotNil(didResp)
			} else {
				s.Require().Error(err)
				s.Assert().Contains(err.Error(), tc.expectErr.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgSeverIssueUserCredential() {
	server := keeper.NewMsgServerImpl(s.app.VcKeeper)
	var req types.MsgIssueUserCredential

	testCases := []struct {
		msg       string
		malleate  func()
		expectErr error
	}{
		{
			msg:       "PASS: issuer can issue user credential for alice",
			expectErr: nil,
			malleate: func() {
				var vc types.VerifiableCredential
				issuerDid := didtypes.DID("did:cosmos:net:test:issuer")
				aliceDid := didtypes.DID("did:cosmos:net:test:alice")
				issuerAddress := s.GetIssuerAddress()
				vc = types.NewUserVerifiableCredential(
					"alice-registration-credential",
					issuerDid.String(),
					time.Now(),
					types.NewUserCredentialSubject(
						aliceDid.String(),
						"b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
						true,
					),
				)
				vc, _ = vc.Sign(
					s.keyring, s.GetIssuerAddress(),
					issuerDid.NewVerificationMethodID(issuerAddress.String()),
				)
				req = types.MsgIssueUserCredential{
					Credential: &vc,
					Owner:      issuerAddress.String(),
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			didResp, err := server.IssueUserCredential(s.ctx, &req)
			if tc.expectErr == nil {
				s.NoError(err)
				s.NotNil(didResp)
			} else {
				s.Require().Error(err)
				s.Assert().Contains(err.Error(), tc.expectErr.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgSeverIssueAnonymousCredentialSchema() {
	server := keeper.NewMsgServerImpl(s.app.VcKeeper)
	var req types.MsgIssueAnonymousCredentialSchema

	testCases := []struct {
		msg       string
		malleate  func()
		expectErr error
	}{
		{
			msg:       "PASS: issuer can issue anonymous credential schema",
			expectErr: nil,
			malleate: func() {
				var vc types.VerifiableCredential
				issuerDid := didtypes.DID("did:cosmos:net:test:issuer")
				issuerAddress := s.GetIssuerAddress()

				vc = types.NewAnonymousCredentialSchema(
					"vc:cosmos:net:test:anonymous-credential-schema-2022",
					issuerDid.String(),
					time.Now(),
					types.NewAnonymousCredentialSchemaSubject(
						issuerDid.String(),
						[]string{"BBS+", "Accumulator"},
						[]string{
							"https://eprint.iacr.org/2016/663.pdf",
							"https://eprint.iacr.org/2020/777.pdf",
							"https://github.com/coinbase/kryptology",
							"https://github.com/kitounliu/kryptology/tree/combine",
						},
						&anonymouscredential.PublicParameters{
							BbsPlusPublicParams: &bbsplus.PublicParameters{
								5,
								[]byte("placeholder for bbs+ public key"),
							},
							AccumulatorPublicParams: &accumulator.PublicParameters{
								[]byte("placeholder for accumulator public key"),
								nil,
							},
						},
					),
				)

				vc, _ = vc.Sign(
					s.keyring, s.GetIssuerAddress(),
					issuerDid.NewVerificationMethodID(issuerAddress.String()),
				)
				req = types.MsgIssueAnonymousCredentialSchema{
					Credential: &vc,
					Owner:      issuerAddress.String(),
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			vcResp, err := server.IssueAnonymousCredentialSchema(s.ctx, &req)
			if tc.expectErr == nil {
				s.NoError(err)
				s.NotNil(vcResp)
			} else {
				s.Require().Error(err)
				s.Assert().Contains(err.Error(), tc.expectErr.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgSeverUpdateAccumulatorState() {
	server := keeper.NewMsgServerImpl(s.app.VcKeeper)
	// create an anonymous credential scheme first

	// update the anonymous credential schema
	var req *types.MsgUpdateAccumulatorState

	testCases := []struct {
		msg       string
		malleate  func()
		expectErr error
	}{
		{
			msg:       "PASS: issuer can update accumulator state",
			expectErr: nil,
			malleate: func() {
				// create a vc first
				var vc types.VerifiableCredential
				issuerDid := didtypes.DID("did:cosmos:net:test:issuer")
				issuerAddress := s.GetIssuerAddress()

				vc = types.NewAnonymousCredentialSchema(
					"vc:cosmos:net:test:anonymous-credential-schema-for-testing-update-accumulator-state",
					issuerDid.String(),
					time.Now(),
					types.NewAnonymousCredentialSchemaSubject(
						issuerDid.String(),
						[]string{"BBS+", "Accumulator"},
						[]string{
							"https://eprint.iacr.org/2016/663.pdf",
							"https://eprint.iacr.org/2020/777.pdf",
							"https://github.com/coinbase/kryptology",
							"https://github.com/kitounliu/kryptology/tree/combine",
						},
						&anonymouscredential.PublicParameters{
							BbsPlusPublicParams: &bbsplus.PublicParameters{
								5,
								[]byte("placeholder for bbs+ public key"),
							},
							AccumulatorPublicParams: &accumulator.PublicParameters{
								[]byte("placeholder for accumulator public key"),
								nil,
							},
						},
					),
				)

				vc, _ = vc.Sign(
					s.keyring, s.GetIssuerAddress(),
					issuerDid.NewVerificationMethodID(issuerAddress.String()),
				)
				vcReq := types.MsgIssueAnonymousCredentialSchema{
					Credential: &vc,
					Owner:      issuerAddress.String(),
				}
				vcResp, err := server.IssueAnonymousCredentialSchema(s.ctx, &vcReq)
				s.NoError(err)
				s.NotNil(vcResp)

				// update the accumulator state
				// clean the proof
				vc.Proof = nil
				now := time.Now()
				vc.IssuanceDate = &now
				newState := accumulator.State{AccValue: []byte("placeholder for new accumulator state")}
				vc, err = vc.UpdateAccumulatorState(&newState)
				s.NoError(err)
				// update proof
				vc, _ = vc.Sign(
					s.keyring, s.GetIssuerAddress(),
					issuerDid.NewVerificationMethodID(issuerAddress.String()),
				)
				req = types.NewMsgUpdateAccumulatorState(vc.Id, vc.IssuanceDate, &newState, vc.Proof, issuerAddress.String())
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			vcResp, err := server.UpdateAccumulatorState(s.ctx, req)
			if tc.expectErr == nil {
				s.NoError(err)
				s.NotNil(vcResp)
			} else {
				s.Require().Error(err)
				s.Assert().Contains(err.Error(), tc.expectErr.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgSeverUpdateVerifiableCredential() {
	server := keeper.NewMsgServerImpl(s.app.VcKeeper)
	// create an anonymous credential scheme first

	// update the anonymous credential schema
	var req types.MsgUpdateVerifiableCredential

	testCases := []struct {
		msg       string
		malleate  func()
		expectErr error
	}{
		{
			msg:       "PASS: issuer can update anonymous credential schema",
			expectErr: nil,
			malleate: func() {
				// create a vc first
				var vc types.VerifiableCredential
				issuerDid := didtypes.DID("did:cosmos:net:test:issuer")
				issuerAddress := s.GetIssuerAddress()

				vc = types.NewAnonymousCredentialSchema(
					"vc:cosmos:net:test:anonymous-credential-schema-for-testing-update-verifiable-credential",
					issuerDid.String(),
					time.Now(),
					types.NewAnonymousCredentialSchemaSubject(
						issuerDid.String(),
						[]string{"BBS+", "Accumulator"},
						[]string{
							"https://eprint.iacr.org/2016/663.pdf",
							"https://eprint.iacr.org/2020/777.pdf",
							"https://github.com/coinbase/kryptology",
							"https://github.com/kitounliu/kryptology/tree/combine",
						},
						&anonymouscredential.PublicParameters{
							BbsPlusPublicParams: &bbsplus.PublicParameters{
								5,
								[]byte("placeholder for bbs+ public key"),
							},
							AccumulatorPublicParams: &accumulator.PublicParameters{
								[]byte("placeholder for accumulator public key"),
								nil,
							},
						},
					),
				)

				vc, _ = vc.Sign(
					s.keyring, s.GetIssuerAddress(),
					issuerDid.NewVerificationMethodID(issuerAddress.String()),
				)
				vcReq := types.MsgIssueAnonymousCredentialSchema{
					Credential: &vc,
					Owner:      issuerAddress.String(),
				}
				vcResp, err := server.IssueAnonymousCredentialSchema(s.ctx, &vcReq)
				s.NoError(err)
				s.NotNil(vcResp)

				// update public parameters
				newPp := &anonymouscredential.PublicParameters{
					BbsPlusPublicParams: &bbsplus.PublicParameters{
						10,
						[]byte("placeholder for new bbs+ public key"),
					},
					AccumulatorPublicParams: &accumulator.PublicParameters{
						[]byte("placeholder for new accumulator public key"),
						nil,
					},
				}
				vc, err = vc.UpdatePublicParameters(newPp)
				s.NoError(err)
				// clean the proof
				vc.Proof = nil
				// update proof
				vc, _ = vc.Sign(
					s.keyring, s.GetIssuerAddress(),
					issuerDid.NewVerificationMethodID(issuerAddress.String()),
				)
				req = types.MsgUpdateVerifiableCredential{
					Credential: &vc,
					Owner:      issuerAddress.String(),
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			vcResp, err := server.UpdateVerifiableCredential(s.ctx, &req)
			if tc.expectErr == nil {
				s.NoError(err)
				s.NotNil(vcResp)
			} else {
				s.Require().Error(err)
				s.Assert().Contains(err.Error(), tc.expectErr.Error())
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgSeveRevokeVerifableCredential() {
	server := keeper.NewMsgServerImpl(s.app.VcKeeper)
	var req types.MsgRevokeCredential

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		//{
		//	"PASS: correctly deletes vc",
		//	func() {
		//		// NEED ACCOUNTS HERE
		//		vc := types.NewUserVerifiableCredential(
		//			"new-verifiable-cred-3",
		//			didDoc.Id,
		//			time.Now(),
		//			types.NewUserCredentialSubject(
		//				"accAddr",
		//				"root",
		//				true,
		//			),
		//		)
		//		suite.keeper.SetVerifiableCredential(suite.ctx, []byte(vc.Id), vc)
		//
		//		req = *types.NewMsgRevokeVerifiableCredential(vc.Id, "cosmos1m26ukcnpme38enptw85w2twcr8gllnj8anfy6a")
		//	},
		//	true,
		//},
		{
			"FAIL: vc issuer and did id do not match",
			func() {
				did := "did:cosmos:cash:subject"
				didDoc, _ := didtypes.NewDidDocument(did, didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cosmos:cash:subject#key-1",
							"did:cosmos:cash:subject",
							didtypes.NewBlockchainAccountID(s.sdkCtx.ChainID(), "cosmos1m26ukcnpme38enptw85w2twcr8gllnj8anfy6a"),
						),
						[]string{didtypes.Authentication},
						nil,
					),
				))
				cs := types.NewUserCredentialSubject(
					"accAddr",
					"root",
					true,
				)

				vc := types.NewUserVerifiableCredential(
					"new-verifiable-cred-3",
					"did:cosmos:cash:noone",
					time.Now(),
					cs,
				)
				s.app.VcKeeper.SetVerifiableCredential(s.sdkCtx, []byte(vc.Id), vc)
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)

				req = *types.NewMsgRevokeVerifiableCredential(vc.Id, "cosmos1m26ukcnpme38enptw85w2twcr8gllnj8anfy6a")
			},
			false,
		},
		{
			"FAIL: vc does not exist",
			func() {
				did := "did:cosmos:cash:subject"
				didDoc, _ := didtypes.NewDidDocument(did, didtypes.WithVerifications(
					didtypes.NewVerification(
						didtypes.NewVerificationMethod(
							"did:cosmos:cash:subject#key-1",
							"did:cosmos:cash:subject",
							didtypes.NewBlockchainAccountID(s.sdkCtx.ChainID(), "cosmos1m26ukcnpme38enptw85w2twcr8gllnj8anfy6a"),
						),
						[]string{didtypes.Authentication},
						nil,
					),
				))
				s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(didDoc.Id), didDoc)
			},
			false,
		},
		{
			"FAIL: did does not exists",
			func() {
				req = *types.NewMsgRevokeVerifiableCredential(
					"new-verifiable-cred-3",
					"did:cash:1111",
				)
			},
			false,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			vcResp, err := server.RevokeCredential(s.ctx, &req)
			if tc.expPass {
				s.NoError(err)
				s.NotNil(vcResp)

			} else {
				s.Require().Error(err)
			}
		})
	}
}

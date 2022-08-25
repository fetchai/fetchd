package keeper_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fetchai/fetchd/testutil"

	"github.com/cosmos/cosmos-sdk/crypto/hd"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	ct "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	didtypes "github.com/fetchai/fetchd/x/did/types"
	"github.com/fetchai/fetchd/x/verifiable-credential/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/fetchai/fetchd/app"
)

// Keeper test suit enables the keeper package to be tested
type KeeperTestSuite struct {
	suite.Suite

	app    *app.App
	sdkCtx sdk.Context
	ctx    context.Context

	queryClient types.QueryClient
	blockTime   time.Time

	keyring keyring.Keyring
}

// SetupTest creates a test suite to test the did
func (s *KeeperTestSuite) SetupTest() {
	app := testutil.Setup(s.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	s.blockTime = time.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: s.blockTime})

	s.app = app
	s.sdkCtx = ctx
	s.ctx = sdk.WrapSDKContext(ctx)

	interfaceRegistry := ct.NewInterfaceRegistry()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)
	types.RegisterQueryServer(queryHelper, s.app.VcKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	s.queryClient = queryClient

	s.keyring = keyring.NewInMemory(s.app.AppCodec())
	// helper func to register accounts in the keychain and the account keeper
	registerAccount := func(uid string, withPubKey bool) {
		i, _, _ := s.keyring.NewMnemonic(uid, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		iAddr, err := i.GetAddress()
		s.Require().NoError(err)
		a := s.app.AccountKeeper.NewAccountWithAddress(ctx, iAddr)
		if withPubKey {
			pk, err := i.GetPubKey()
			a.SetPubKey(pk)
			s.Require().NoError(err)
		}
		s.app.AccountKeeper.SetAccount(ctx, s.app.AccountKeeper.NewAccount(ctx, a))
	}

	registerAccount("issuer", true)
	registerAccount("alice", true)
	registerAccount("bob", false)

	// create did for issuer and alice
	issuerDid := didtypes.NewChainDID("test", "issuer")
	issuerVmId := issuerDid.NewVerificationMethodID(s.GetIssuerAddress().String())
	issuerInfo, err := s.keyring.Key("issuer")
	s.NoError(err)
	issuerPk, err := issuerInfo.GetPubKey()
	s.NoError(err)
	issuerDidDoc, _ := didtypes.NewDidDocument(issuerDid.String(), didtypes.WithVerifications(
		didtypes.NewVerification(
			didtypes.NewVerificationMethod(
				issuerVmId,
				issuerDid,
				didtypes.NewPublicKeyMultibase(issuerPk.Bytes(), didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
			),
			[]string{didtypes.Authentication},
			nil,
		),
	))
	s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(issuerDidDoc.Id), issuerDidDoc)

	aliceDid := didtypes.NewChainDID("test", "alice")
	aliceVmId := aliceDid.NewVerificationMethodID(s.GetAliceAddress().String())
	aliceInfo, err := s.keyring.Key("alice")
	s.NoError(err)
	alicePk, err := aliceInfo.GetPubKey()
	s.NoError(err)
	aliceDidDoc, _ := didtypes.NewDidDocument(aliceDid.String(), didtypes.WithVerifications(
		didtypes.NewVerification(
			didtypes.NewVerificationMethod(
				aliceVmId,
				aliceDid,
				didtypes.NewPublicKeyMultibase(alicePk.Bytes(), didtypes.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019),
			),
			[]string{didtypes.Authentication},
			nil,
		),
	))
	s.app.DidKeeper.SetDidDocument(s.sdkCtx, []byte(aliceDidDoc.Id), aliceDidDoc)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s KeeperTestSuite) GetAliceAddress() sdk.Address {
	return s.GetKeyAddress("alice")
}

func (s KeeperTestSuite) GetBobAddress() sdk.Address {
	return s.GetKeyAddress("bob")
}

func (s KeeperTestSuite) GetIssuerAddress() sdk.Address {
	return s.GetKeyAddress("issuer")
}

func (s KeeperTestSuite) GetKeyAddress(uid string) sdk.Address {
	i, _ := s.keyring.Key(uid)
	addr, _ := i.GetAddress()
	return addr
}

func (s *KeeperTestSuite) TestGenericKeeperSetAndGet() {
	testCases := []struct {
		msg string
		did types.VerifiableCredential
		// TODO: add mallate func and clean up test
		expPass bool
	}{
		//{
		//	"data stored successfully",
		//	types.NewUserVerifiableCredential(
		//		"did:cash:1111",
		//		"",
		//		time.Now(),
		//		types.NewUserCredentialSubject("", "root", true),
		//	),
		//	true,
		//},
	}
	for _, tc := range testCases {
		s.app.VcKeeper.Set(s.sdkCtx,
			[]byte(tc.did.Id),
			[]byte{0x01},
			tc.did,
			s.app.VcKeeper.MarshalVerifiableCredential,
		)
		s.app.VcKeeper.Set(s.sdkCtx,
			[]byte(tc.did.Id+"1"),
			[]byte{0x01},
			tc.did,
			s.app.VcKeeper.MarshalVerifiableCredential,
		)
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			if tc.expPass {
				_, found := s.app.VcKeeper.Get(
					s.sdkCtx,
					[]byte(tc.did.Id),
					[]byte{0x01},
					s.app.VcKeeper.UnmarshalVerifiableCredential,
				)
				s.Require().True(found)

				iterator := s.app.VcKeeper.GetAll(
					s.sdkCtx,
					[]byte{0x01},
				)
				defer iterator.Close()

				var array []interface{}
				for ; iterator.Valid(); iterator.Next() {
					array = append(array, iterator.Value())
				}
				s.Require().Equal(2, len(array))
			} else {
				// TODO write failure cases
				s.Require().False(tc.expPass)
			}
		})
	}
}

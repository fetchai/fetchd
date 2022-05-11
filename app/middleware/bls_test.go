package middleware_test

import (
	"encoding/hex"
	"errors"
	"sync"
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestdata "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authmiddleware "github.com/cosmos/cosmos-sdk/x/auth/middleware"
	"github.com/stretchr/testify/require"

	"github.com/fetchai/fetchd/app/middleware"
	"github.com/fetchai/fetchd/crypto/keys/bls12381"
	"github.com/fetchai/fetchd/testutil/testdata"
)

type mockBlsPubKeyValidationFunc struct {
	Counter       int
	ExpectedError error
	lock          sync.Mutex
}

func (m *mockBlsPubKeyValidationFunc) Validate(pk *bls12381.PubKey) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.Counter++

	return m.ExpectedError
}

func (s *MWTestSuite) TestBlsPubKeyValidation() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	require := s.Require()

	validationMock := &mockBlsPubKeyValidationFunc{Counter: 0}

	txHandler := authmiddleware.ComposeMiddlewares(
		noopTxHandler,
		middleware.BlsPubKeyValidationMiddleware(s.app.AccountKeeper, validationMock.Validate),
		authmiddleware.SetPubKeyMiddleware(s.app.AccountKeeper),
	)

	// keys and addresses
	priv1, pub1, addr1 := testdata.KeyTestPubAddrBls12381()
	priv2, pub2, addr2 := testdata.KeyTestPubAddrBls12381()
	priv3, pub3, addr3 := sdktestdata.KeyTestPubAddr()
	priv4, pub4, addr4 := sdktestdata.KeyTestPubAddrSecp256R1(require)

	addrs := []sdk.AccAddress{addr1, addr2, addr3, addr4}
	pubs := []cryptotypes.PubKey{pub1, pub2, pub3, pub4}

	msgs := make([]sdk.Msg, len(addrs))
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(acc.SetAccountNumber(uint64(i)))
		s.app.AccountKeeper.SetAccount(ctx, acc)
		msgs[i] = sdktestdata.NewTestMsg(addr)
	}
	require.NoError(txBuilder.SetMsgs(msgs...))
	txBuilder.SetFeeAmount(sdktestdata.NewTestFeeAmount())
	txBuilder.SetGasLimit(sdktestdata.NewTestGasLimit())

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1, priv2, priv3, priv4}, []uint64{0, 1, 2, 3}, []uint64{0, 0, 0, 0}
	testTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	require.NoError(err)

	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})
	require.NoError(err)

	// BLS Validation func must only be called twice
	require.Equal(validationMock.Counter, 2)

	// Require that all accounts have pubkey set after middleware runs
	for i, addr := range addrs {
		pk, err := s.app.AccountKeeper.GetPubKey(ctx, addr)
		require.NoError(err, "Error on retrieving pubkey from account")
		require.True(pubs[i].Equals(pk),
			"Wrong Pubkey retrieved from AccountKeeper, idx=%d\nexpected=%s\n     got=%s", i, pubs[i], pk)
	}

	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})
	require.NoError(err)

	// BLS Validation func must not be called anymore after pubkey have been set on accounts
	require.Equal(validationMock.Counter, 2)
}

func (s *MWTestSuite) TestBlsPubKeyValidationErrors() {
	ctx := s.SetupTest(true) // setup
	txBuilder := s.clientCtx.TxConfig.NewTxBuilder()
	require := s.Require()

	validationMock := &mockBlsPubKeyValidationFunc{
		Counter:       0,
		ExpectedError: errors.New("test"),
	}

	txHandler := authmiddleware.ComposeMiddlewares(
		noopTxHandler,
		middleware.BlsPubKeyValidationMiddleware(s.app.AccountKeeper, validationMock.Validate),
		authmiddleware.SetPubKeyMiddleware(s.app.AccountKeeper),
	)

	// keys and addresses
	priv1, _, addr1 := testdata.KeyTestPubAddrBls12381()

	addrs := []sdk.AccAddress{addr1}

	msgs := make([]sdk.Msg, len(addrs))
	// set accounts and create msg for each address
	for i, addr := range addrs {
		acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		require.NoError(acc.SetAccountNumber(uint64(i)))
		s.app.AccountKeeper.SetAccount(ctx, acc)
		msgs[i] = sdktestdata.NewTestMsg(addr)
	}
	require.NoError(txBuilder.SetMsgs(msgs...))
	txBuilder.SetFeeAmount(sdktestdata.NewTestFeeAmount())
	txBuilder.SetGasLimit(sdktestdata.NewTestGasLimit())

	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}
	testTx, _, err := s.createTestTx(txBuilder, privs, accNums, accSeqs, ctx.ChainID())
	require.NoError(err)

	_, err = txHandler.DeliverTx(sdk.WrapSDKContext(ctx), tx.Request{Tx: testTx})

	require.Equal(validationMock.Counter, 1)
	require.Equal(validationMock.ExpectedError, err)
}

func TestDefaultBlsPubkKeyValidationFunc(t *testing.T) {
	_, validPk, _ := testdata.KeyTestPubAddrBls12381()

	// value taken from https://github.com/ethereum/consensus-spec-tests/blob/master/tests/general/phase0/bls/verify/small/verify_infinity_pubkey_and_infinity_signature/data.yaml
	infinityPkBytes, err := hex.DecodeString("c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)
	infinityPubkey := &bls12381.PubKey{Key: infinityPkBytes}

	testCases := []struct {
		desc        string
		pk          *bls12381.PubKey
		expectError bool
	}{
		{
			desc:        "valid pubkey",
			pk:          validPk.(*bls12381.PubKey),
			expectError: false,
		},
		{
			desc:        "infinity pubkey",
			pk:          infinityPubkey,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			err := middleware.DefaultBlsPubkKeyValidationFunc(tc.pk)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

}

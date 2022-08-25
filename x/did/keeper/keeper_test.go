package keeper_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	ct "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fetchai/fetchd/x/did/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/fetchai/fetchd/app"
	"github.com/fetchai/fetchd/testutil"
)

// Keeper test suit enables the keeper package to be tested
type KeeperTestSuite struct {
	suite.Suite

	app       *app.App
	sdkCtx    sdk.Context
	ctx       context.Context
	blockTime time.Time

	queryClient types.QueryClient
}

// SetupTest creates a test suite to test the did
func (s *KeeperTestSuite) SetupTest() {
	app := testutil.Setup(s.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{ChainID: "foochainid"})

	s.blockTime = time.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: s.blockTime})

	s.app = app
	s.sdkCtx = ctx
	s.ctx = sdk.WrapSDKContext(ctx)

	interfaceRegistry := ct.NewInterfaceRegistry()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)
	types.RegisterQueryServer(queryHelper, s.app.DidKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	s.queryClient = queryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestGenericKeeperSetAndGet() {
	testCases := []struct {
		msg     string
		didFn   func() types.DidDocument
		expPass bool
	}{
		{
			"data stored successfully",
			func() types.DidDocument {
				dd, _ := types.NewDidDocument(
					"did:cash:subject",
				)
				return dd
			},
			true,
		},
	}
	for _, tc := range testCases {
		dd := tc.didFn()
		s.app.DidKeeper.Set(s.sdkCtx,
			[]byte(dd.Id),
			[]byte{0x01},
			dd,
			s.app.DidKeeper.Marshal,
		)
		s.app.DidKeeper.Set(s.sdkCtx,
			[]byte(dd.Id+"1"),
			[]byte{0x01},
			dd,
			s.app.DidKeeper.Marshal,
		)
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			if tc.expPass {
				_, found := s.app.DidKeeper.Get(
					s.sdkCtx,
					[]byte(dd.Id),
					[]byte{0x01},
					s.app.DidKeeper.UnmarshalDidDocument,
				)
				s.Require().True(found)

				iterator := s.app.DidKeeper.GetAll(
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

func (s *KeeperTestSuite) TestGenericKeeperDelete() {
	testCases := []struct {
		msg     string
		didFn   func() types.DidDocument
		expPass bool
	}{
		{
			"data stored successfully",
			func() types.DidDocument {
				dd, _ := types.NewDidDocument(
					"did:cash:subject",
				)
				return dd
			},
			true,
		},
	}
	for _, tc := range testCases {
		dd := tc.didFn()
		s.app.DidKeeper.Set(s.sdkCtx,
			[]byte(dd.Id),
			[]byte{0x01},
			dd,
			s.app.DidKeeper.Marshal,
		)
		s.app.DidKeeper.Set(s.sdkCtx,
			[]byte(dd.Id+"1"),
			[]byte{0x01},
			dd,
			s.app.DidKeeper.Marshal,
		)
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			if tc.expPass {
				s.app.DidKeeper.Delete(
					s.sdkCtx,
					[]byte(dd.Id),
					[]byte{0x01},
				)

				_, found := s.app.DidKeeper.Get(
					s.sdkCtx,
					[]byte(dd.Id),
					[]byte{0x01},
					s.app.DidKeeper.UnmarshalDidDocument,
				)
				s.Require().False(found)

			} else {
				// TODO write failure cases
				s.Require().False(tc.expPass)
			}
		})
	}
}

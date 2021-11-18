package testsuite

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12381"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/suite"

	"github.com/fetchai/fetchd/types"
	servermodule "github.com/fetchai/fetchd/types/module/server"
	"github.com/fetchai/fetchd/types/testutil"
	"github.com/fetchai/fetchd/x/group"
	"github.com/fetchai/fetchd/x/group/testdata"
)

type IntegrationTestSuite struct {
	suite.Suite

	fixtureFactory *servermodule.FixtureFactory
	fixture        testutil.Fixture

	ctx              context.Context
	sdkCtx           sdk.Context
	genesisCtx       types.Context
	msgClient        group.MsgClient
	queryClient      group.QueryClient
	addr1            sdk.AccAddress
	addr2            sdk.AccAddress
	addr3            sdk.AccAddress
	addr4            sdk.AccAddress
	addr5            sdk.AccAddress
	addr6            sdk.AccAddress
	addrBls1         sdk.AccAddress
	addrBls2         sdk.AccAddress
	addrBls3         sdk.AccAddress
	addrBls4         sdk.AccAddress
	addrBls5         sdk.AccAddress
	addrBls6         sdk.AccAddress
	groupAccountAddr sdk.AccAddress
	groupID          uint64

	skBls1 cryptotypes.PrivKey
	skBls2 cryptotypes.PrivKey
	skBls3 cryptotypes.PrivKey
	skBls4 cryptotypes.PrivKey
	skBls5 cryptotypes.PrivKey
	skBls6 cryptotypes.PrivKey

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	mintKeeper    mintkeeper.Keeper

	blockTime time.Time
}

func NewIntegrationTestSuite(
	fixtureFactory *servermodule.FixtureFactory,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.BaseKeeper,
	mintKeeper mintkeeper.Keeper) *IntegrationTestSuite {

	return &IntegrationTestSuite{
		fixtureFactory: fixtureFactory,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		mintKeeper:     mintKeeper,
	}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.fixture = s.fixtureFactory.Setup()

	s.blockTime = time.Now().UTC()

	// TODO clean up once types.Context merged upstream into sdk.Context
	sdkCtx := s.fixture.Context().(types.Context).WithBlockTime(s.blockTime)
	s.sdkCtx, _ = sdkCtx.CacheContext()
	s.ctx = types.Context{Context: s.sdkCtx}

	s.genesisCtx = types.Context{Context: sdkCtx}
	s.bankKeeper.SetSupply(sdkCtx, banktypes.NewSupply(sdk.Coins{}))
	s.Require().NoError(s.bankKeeper.MintCoins(s.sdkCtx, minttypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("test", 400000000))))

	s.accountKeeper.SetParams(sdkCtx, authtypes.DefaultParams())
	s.bankKeeper.SetParams(sdkCtx, banktypes.DefaultParams())

	s.msgClient = group.NewMsgClient(s.fixture.TxConn())
	s.queryClient = group.NewQueryClient(s.fixture.QueryConn())

	s.Require().GreaterOrEqual(len(s.fixture.Signers()), 6)
	s.addr1 = s.fixture.Signers()[0]
	s.addr2 = s.fixture.Signers()[1]
	s.addr3 = s.fixture.Signers()[2]
	s.addr4 = s.fixture.Signers()[3]
	s.addr5 = s.fixture.Signers()[4]
	s.addr6 = s.fixture.Signers()[5]

	s.addrBls1 = s.fixture.SignersBls()[0]
	s.addrBls2 = s.fixture.SignersBls()[1]
	s.addrBls3 = s.fixture.SignersBls()[2]
	s.addrBls4 = s.fixture.SignersBls()[3]
	s.addrBls5 = s.fixture.SignersBls()[4]
	s.addrBls6 = s.fixture.SignersBls()[5]

	s.skBls1 = s.fixture.SksBls()[0]
	s.skBls2 = s.fixture.SksBls()[1]
	s.skBls3 = s.fixture.SksBls()[2]
	s.skBls4 = s.fixture.SksBls()[3]
	s.skBls5 = s.fixture.SksBls()[4]
	s.skBls6 = s.fixture.SksBls()[5]

	accBls1 := s.accountKeeper.NewAccountWithAddress(s.sdkCtx, s.addrBls1)
	accBls2 := s.accountKeeper.NewAccountWithAddress(s.sdkCtx, s.addrBls2)
	accBls3 := s.accountKeeper.NewAccountWithAddress(s.sdkCtx, s.addrBls3)
	accBls4 := s.accountKeeper.NewAccountWithAddress(s.sdkCtx, s.addrBls4)
	accBls5 := s.accountKeeper.NewAccountWithAddress(s.sdkCtx, s.addrBls5)
	accBls6 := s.accountKeeper.NewAccountWithAddress(s.sdkCtx, s.addrBls6)

	pkBls1 := s.skBls1.PubKey()
	pkBls2 := s.skBls2.PubKey()
	pkBls3 := s.skBls3.PubKey()
	pkBls4 := s.skBls4.PubKey()
	pkBls5 := s.skBls5.PubKey()
	pkBls6 := s.skBls6.PubKey()

	err := accBls1.SetPubKey(pkBls1)
	s.Require().NoError(err)
	err = accBls2.SetPubKey(pkBls2)
	s.Require().NoError(err)
	err = accBls3.SetPubKey(pkBls3)
	s.Require().NoError(err)
	err = accBls4.SetPubKey(pkBls4)
	s.Require().NoError(err)
	err = accBls5.SetPubKey(pkBls5)
	s.Require().NoError(err)
	err = accBls6.SetPubKey(pkBls6)
	s.Require().NoError(err)

	err = accBls1.SetPopValid(true)
	s.Require().NoError(err)
	err = accBls2.SetPopValid(true)
	s.Require().NoError(err)
	err = accBls3.SetPopValid(true)
	s.Require().NoError(err)
	err = accBls4.SetPopValid(true)
	s.Require().NoError(err)
	err = accBls5.SetPopValid(true)
	s.Require().NoError(err)
	err = accBls6.SetPopValid(true)
	s.Require().NoError(err)

	s.accountKeeper.SetAccount(s.sdkCtx, accBls1)
	s.accountKeeper.SetAccount(s.sdkCtx, accBls2)
	s.accountKeeper.SetAccount(s.sdkCtx, accBls3)
	s.accountKeeper.SetAccount(s.sdkCtx, accBls4)
	s.accountKeeper.SetAccount(s.sdkCtx, accBls5)
	s.accountKeeper.SetAccount(s.sdkCtx, accBls6)

	// Initial group, group account and balance setup
	members := []group.Member{
		{Address: s.addr5.String(), Weight: "1"}, {Address: s.addr2.String(), Weight: "2"},
	}
	groupRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    s.addr1.String(),
		Members:  members,
		Metadata: nil,
	})
	s.Require().NoError(err)
	s.groupID = groupRes.GroupId

	policy := group.NewThresholdDecisionPolicy(
		"2",
		gogotypes.Duration{Seconds: 1},
	)
	accountReq := &group.MsgCreateGroupAccount{
		Admin:    s.addr1.String(),
		GroupId:  s.groupID,
		Metadata: nil,
	}
	err = accountReq.SetDecisionPolicy(policy)
	s.Require().NoError(err)
	accountRes, err := s.msgClient.CreateGroupAccount(s.ctx, accountReq)
	s.Require().NoError(err)
	addr, err := sdk.AccAddressFromBech32(accountRes.Address)
	s.Require().NoError(err)
	s.groupAccountAddr = addr
	s.Require().NoError(fundAccount(s.bankKeeper, s.sdkCtx, s.groupAccountAddr, sdk.Coins{sdk.NewInt64Coin("test", 10000)}))
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.fixture.Teardown()
}

// TODO: https://github.com/cosmos/cosmos-sdk/issues/9346
func fundAccount(bankKeeper bankkeeper.Keeper, ctx sdk.Context, addr sdk.AccAddress, amounts sdk.Coins) error {
	if err := bankKeeper.MintCoins(ctx, minttypes.ModuleName, amounts); err != nil {
		return err
	}
	return bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, amounts)
}

func (s *IntegrationTestSuite) TestCreateGroup() {
	members := []group.Member{{
		Address:  s.addr5.String(),
		Weight:   "1",
		Metadata: nil,
	}, {
		Address:  s.addr6.String(),
		Weight:   "2",
		Metadata: nil,
	}}

	blsMembers := []group.Member{{
		Address:  s.addrBls1.String(),
		Weight:   "3",
		Metadata: nil,
	}, {
		Address:  s.addrBls2.String(),
		Weight:   "5",
		Metadata: nil,
	}}

	mixedMembers := []group.Member{{
		Address:  s.addr5.String(),
		Weight:   "2",
		Metadata: nil,
	}, {
		Address:  s.addrBls1.String(),
		Weight:   "3",
		Metadata: nil,
	}}

	expGroups := []*group.GroupInfo{
		{
			GroupId:     s.groupID,
			Version:     1,
			Admin:       s.addr1.String(),
			TotalWeight: "3",
			Metadata:    nil,
		},
		{
			GroupId:     2,
			Version:     1,
			Admin:       s.addr1.String(),
			TotalWeight: "3",
			Metadata:    nil,
		},
		{
			GroupId:     2,
			Version:     1,
			Admin:       s.addr1.String(),
			TotalWeight: "8",
			Metadata:    nil,
			BlsOnly:     true,
		},
		{
			GroupId:     2,
			Version:     1,
			Admin:       s.addr1.String(),
			TotalWeight: "5",
			Metadata:    nil,
		},
	}

	specs := map[string]struct {
		req             *group.MsgCreateGroup
		expErr          bool
		expGroups       []*group.GroupInfo
		expectedMembers []group.Member
	}{
		"all good": {
			req: &group.MsgCreateGroup{
				Admin:    s.addr1.String(),
				Members:  members,
				Metadata: nil,
			},
			expGroups:       expGroups[0:2],
			expectedMembers: members,
		},
		"all good with bls members": {
			req: &group.MsgCreateGroup{
				Admin:    s.addr1.String(),
				Members:  blsMembers,
				Metadata: nil,
				BlsOnly:  true,
			},
			expGroups: []*group.GroupInfo{
				expGroups[0],
				expGroups[2],
			},
			expectedMembers: blsMembers,
		},
		"all good with mixed members": {
			req: &group.MsgCreateGroup{
				Admin:    s.addr1.String(),
				Members:  mixedMembers,
				Metadata: nil,
				BlsOnly:  false,
			},
			expGroups: []*group.GroupInfo{
				expGroups[0],
				expGroups[3],
			},
			expectedMembers: mixedMembers,
		},
		"mixed members not allowed when bls only": {
			req: &group.MsgCreateGroup{
				Admin:    s.addr1.String(),
				Members:  mixedMembers,
				Metadata: nil,
				BlsOnly:  true,
			},
			expErr:          true,
			expectedMembers: mixedMembers,
		},
		"zero member weight": {
			req: &group.MsgCreateGroup{
				Admin: s.addr1.String(),
				Members: []group.Member{{
					Address:  s.addr3.String(),
					Weight:   "0",
					Metadata: nil,
				}},
				Metadata: nil,
			},
			expErr: true,
		},
	}

	var seq uint32 = 1
	for msg, spec := range specs {
		tc := spec
		s.Run(msg, func() {
			sdkCtx, _ := s.sdkCtx.CacheContext()
			ctx := types.Context{Context: sdkCtx}

			res, err := s.msgClient.CreateGroup(ctx, tc.req)
			if tc.expErr {
				s.Require().Error(err)
				_, err := s.queryClient.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: uint64(seq + 1)})
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			s.Assert().Equal(uint64(2), res.GroupId)

			// then all data persisted
			loadedGroupRes, err := s.queryClient.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: res.GroupId})
			s.Require().NoError(err)
			s.Assert().Equal(tc.req.Admin, loadedGroupRes.Info.Admin)
			s.Assert().Equal(tc.req.Metadata, loadedGroupRes.Info.Metadata)
			s.Assert().Equal(res.GroupId, loadedGroupRes.Info.GroupId)
			s.Assert().Equal(uint64(1), loadedGroupRes.Info.Version)

			// and members are stored as well
			membersRes, err := s.queryClient.GroupMembers(ctx, &group.QueryGroupMembersRequest{GroupId: res.GroupId})
			s.Require().NoError(err)
			loadedMembers := membersRes.Members
			s.Require().Equal(len(tc.expectedMembers), len(loadedMembers), "want %#v, got %#v", tc.expectedMembers, membersRes)
			// we reorder members by address to be able to compare them
			sort.Slice(tc.expectedMembers, func(i, j int) bool {
				addri, err := sdk.AccAddressFromBech32(tc.expectedMembers[i].Address)
				s.Require().NoError(err)
				addrj, err := sdk.AccAddressFromBech32(tc.expectedMembers[j].Address)
				s.Require().NoError(err)
				return bytes.Compare(addri, addrj) < 0
			})
			for i := range loadedMembers {
				s.Assert().Equal(tc.expectedMembers[i].Metadata, loadedMembers[i].Member.Metadata)
				s.Assert().Equal(tc.expectedMembers[i].Address, loadedMembers[i].Member.Address)
				s.Assert().Equal(tc.expectedMembers[i].Weight, loadedMembers[i].Member.Weight)
				s.Assert().Equal(res.GroupId, loadedMembers[i].GroupId)
			}

			// query groups by admin
			groupsRes, err := s.queryClient.GroupsByAdmin(ctx, &group.QueryGroupsByAdminRequest{Admin: s.addr1.String()})
			s.Require().NoError(err)
			loadedGroups := groupsRes.Groups
			s.Require().Equal(len(tc.expGroups), len(loadedGroups))
			for i := range loadedGroups {
				s.Assert().Equal(tc.expGroups[i].Metadata, loadedGroups[i].Metadata)
				s.Assert().Equal(tc.expGroups[i].Admin, loadedGroups[i].Admin)
				s.Assert().Equal(tc.expGroups[i].TotalWeight, loadedGroups[i].TotalWeight)
				s.Assert().Equal(tc.expGroups[i].GroupId, loadedGroups[i].GroupId)
				s.Assert().Equal(tc.expGroups[i].Version, loadedGroups[i].Version)
				s.Assert().Equal(tc.expGroups[i].BlsOnly, loadedGroups[i].BlsOnly)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateGroupAdmin() {
	members := []group.Member{{
		Address:  s.addr1.String(),
		Weight:   "1",
		Metadata: nil,
	}}
	oldAdmin := s.addr2.String()
	newAdmin := s.addr3.String()
	groupRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    oldAdmin,
		Members:  members,
		Metadata: nil,
	})
	s.Require().NoError(err)
	groupID := groupRes.GroupId

	specs := map[string]struct {
		req       *group.MsgUpdateGroupAdmin
		expStored *group.GroupInfo
		expErr    bool
	}{
		"with correct admin": {
			req: &group.MsgUpdateGroupAdmin{
				GroupId:  groupID,
				Admin:    oldAdmin,
				NewAdmin: newAdmin,
			},
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       newAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     2,
			},
		},
		"with wrong admin": {
			req: &group.MsgUpdateGroupAdmin{
				GroupId:  groupID,
				Admin:    s.addr4.String(),
				NewAdmin: newAdmin,
			},
			expErr: true,
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       oldAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
			},
		},
		"with unknown groupID": {
			req: &group.MsgUpdateGroupAdmin{
				GroupId:  999,
				Admin:    oldAdmin,
				NewAdmin: newAdmin,
			},
			expErr: true,
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       oldAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
			},
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			_, err := s.msgClient.UpdateGroupAdmin(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// then
			res, err := s.queryClient.GroupInfo(s.ctx, &group.QueryGroupInfoRequest{GroupId: groupID})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expStored, res.Info)
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateGroupMetadata() {
	oldAdmin := s.addr1.String()
	groupID := s.groupID

	specs := map[string]struct {
		req       *group.MsgUpdateGroupMetadata
		expErr    bool
		expStored *group.GroupInfo
	}{
		"with correct admin": {
			req: &group.MsgUpdateGroupMetadata{
				GroupId:  groupID,
				Admin:    oldAdmin,
				Metadata: []byte{1, 2, 3},
			},
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       oldAdmin,
				Metadata:    []byte{1, 2, 3},
				TotalWeight: "3",
				Version:     2,
			},
		},
		"with wrong admin": {
			req: &group.MsgUpdateGroupMetadata{
				GroupId:  groupID,
				Admin:    s.addr3.String(),
				Metadata: []byte{1, 2, 3},
			},
			expErr: true,
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       oldAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
			},
		},
		"with unknown groupid": {
			req: &group.MsgUpdateGroupMetadata{
				GroupId:  999,
				Admin:    oldAdmin,
				Metadata: []byte{1, 2, 3},
			},
			expErr: true,
			expStored: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       oldAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
			},
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx, _ := s.sdkCtx.CacheContext()
			ctx := types.Context{Context: sdkCtx}
			_, err := s.msgClient.UpdateGroupMetadata(ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// then
			res, err := s.queryClient.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: groupID})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expStored, res.Info)
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateGroupMembers() {
	member1 := s.addr5.String()
	member2 := s.addr6.String()
	members := []group.Member{{
		Address:  member1,
		Weight:   "1",
		Metadata: nil,
	}}

	myAdmin := s.addr4.String()
	groupRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    myAdmin,
		Members:  members,
		Metadata: nil,
	})
	s.Require().NoError(err)
	groupID := groupRes.GroupId

	specs := map[string]struct {
		req        *group.MsgUpdateGroupMembers
		expErr     bool
		expGroup   *group.GroupInfo
		expMembers []*group.GroupMember
	}{
		"add new member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  member2,
					Weight:   "2",
					Metadata: nil,
				}},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "3",
				Version:     2,
			},
			expMembers: []*group.GroupMember{
				{
					Member: &group.Member{
						Address:  member2,
						Weight:   "2",
						Metadata: nil,
					},
					GroupId: groupID,
				},
				{
					Member: &group.Member{
						Address:  member1,
						Weight:   "1",
						Metadata: nil,
					},
					GroupId: groupID,
				},
			},
		},
		"update member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  member1,
					Weight:   "2",
					Metadata: []byte{1, 2, 3},
				}},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "2",
				Version:     2,
			},
			expMembers: []*group.GroupMember{
				{
					GroupId: groupID,
					Member: &group.Member{
						Address:  member1,
						Weight:   "2",
						Metadata: []byte{1, 2, 3},
					},
				},
			},
		},
		"update member with same data": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address: member1,
					Weight:  "1",
				}},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     2,
			},
			expMembers: []*group.GroupMember{
				{
					GroupId: groupID,
					Member: &group.Member{
						Address: member1,
						Weight:  "1",
					},
				},
			},
		},
		"replace member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{
					{
						Address:  member1,
						Weight:   "0",
						Metadata: nil,
					},
					{
						Address:  member2,
						Weight:   "1",
						Metadata: nil,
					},
				},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     2,
			},
			expMembers: []*group.GroupMember{{
				GroupId: groupID,
				Member: &group.Member{
					Address:  member2,
					Weight:   "1",
					Metadata: nil,
				},
			}},
		},
		"remove existing member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  member1,
					Weight:   "0",
					Metadata: nil,
				}},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "0",
				Version:     2,
			},
			expMembers: []*group.GroupMember{},
		},
		"remove unknown member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  s.addr4.String(),
					Weight:   "0",
					Metadata: nil,
				}},
			},
			expErr: true,
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
			},
			expMembers: []*group.GroupMember{{
				GroupId: groupID,
				Member: &group.Member{
					Address:  member1,
					Weight:   "1",
					Metadata: nil,
				},
			}},
		},
		"with wrong admin": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupID,
				Admin:   s.addr3.String(),
				MemberUpdates: []group.Member{{
					Address:  member1,
					Weight:   "2",
					Metadata: nil,
				}},
			},
			expErr: true,
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
			},
			expMembers: []*group.GroupMember{{
				GroupId: groupID,
				Member: &group.Member{
					Address: member1,
					Weight:  "1",
				},
			}},
		},
		"with unknown groupID": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: 999,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  member1,
					Weight:   "2",
					Metadata: nil,
				}},
			},
			expErr: true,
			expGroup: &group.GroupInfo{
				GroupId:     groupID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "1",
				Version:     1,
			},
			expMembers: []*group.GroupMember{{
				GroupId: groupID,
				Member: &group.Member{
					Address: member1,
					Weight:  "1",
				},
			}},
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx, _ := s.sdkCtx.CacheContext()
			ctx := types.Context{Context: sdkCtx}
			_, err := s.msgClient.UpdateGroupMembers(ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// then
			res, err := s.queryClient.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: groupID})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expGroup, res.Info)

			// and members persisted
			membersRes, err := s.queryClient.GroupMembers(ctx, &group.QueryGroupMembersRequest{GroupId: groupID})
			s.Require().NoError(err)
			loadedMembers := membersRes.Members
			s.Require().Equal(len(spec.expMembers), len(loadedMembers))
			// we reorder group members by address to be able to compare them
			sort.Slice(spec.expMembers, func(i, j int) bool {
				addri, err := sdk.AccAddressFromBech32(spec.expMembers[i].Member.Address)
				s.Require().NoError(err)
				addrj, err := sdk.AccAddressFromBech32(spec.expMembers[j].Member.Address)
				s.Require().NoError(err)
				return bytes.Compare(addri, addrj) < 0
			})
			for i := range loadedMembers {
				s.Assert().Equal(spec.expMembers[i].Member.Metadata, loadedMembers[i].Member.Metadata)
				s.Assert().Equal(spec.expMembers[i].Member.Address, loadedMembers[i].Member.Address)
				s.Assert().Equal(spec.expMembers[i].Member.Weight, loadedMembers[i].Member.Weight)
				s.Assert().Equal(spec.expMembers[i].GroupId, loadedMembers[i].GroupId)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateGroupMembersBls() {
	member := s.addr3.String()
	memberBls1 := s.addrBls1.String()
	memberBls2 := s.addrBls2.String()
	memberBls3 := s.addrBls3.String()
	membersBls := []group.Member{
		{
			Address:  memberBls1,
			Weight:   "5",
			Metadata: nil,
		},
		{
			Address:  memberBls2,
			Weight:   "7",
			Metadata: nil,
		},
	}

	myAdmin := s.addr4.String()
	groupBlsRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    myAdmin,
		Members:  membersBls,
		Metadata: nil,
		BlsOnly:  true,
	})
	s.Require().NoError(err)
	groupBlsID := groupBlsRes.GroupId

	specs := map[string]struct {
		req        *group.MsgUpdateGroupMembers
		expErr     bool
		expGroup   *group.GroupInfo
		expMembers []*group.GroupMember
	}{
		"add new bls member": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupBlsID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  memberBls3,
					Weight:   "3",
					Metadata: nil,
				}},
			},
			expGroup: &group.GroupInfo{
				GroupId:     groupBlsID,
				Admin:       myAdmin,
				Metadata:    nil,
				TotalWeight: "15",
				Version:     2,
				BlsOnly:     true,
			},
			expMembers: []*group.GroupMember{
				{
					Member: &group.Member{
						Address:  memberBls1,
						Weight:   "5",
						Metadata: nil,
					},
					GroupId: groupBlsID,
				},
				{
					Member: &group.Member{
						Address:  memberBls2,
						Weight:   "7",
						Metadata: nil,
					},
					GroupId: groupBlsID,
				},
				{
					Member: &group.Member{
						Address:  memberBls3,
						Weight:   "3",
						Metadata: nil,
					},
					GroupId: groupBlsID,
				},
			},
		},
		"add new non-bls member not allowed": {
			req: &group.MsgUpdateGroupMembers{
				GroupId: groupBlsID,
				Admin:   myAdmin,
				MemberUpdates: []group.Member{{
					Address:  member,
					Weight:   "1",
					Metadata: nil,
				}},
			},
			expErr: true,
		},
	}

	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx, _ := s.sdkCtx.CacheContext()
			ctx := types.Context{Context: sdkCtx}
			_, err := s.msgClient.UpdateGroupMembers(ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// then
			res, err := s.queryClient.GroupInfo(ctx, &group.QueryGroupInfoRequest{GroupId: groupBlsID})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expGroup, res.Info)

			// and members persisted
			membersRes, err := s.queryClient.GroupMembers(ctx, &group.QueryGroupMembersRequest{GroupId: groupBlsID})
			s.Require().NoError(err)
			loadedMembers := membersRes.Members
			s.Require().Equal(len(spec.expMembers), len(loadedMembers))
			// we reorder group members by address to be able to compare them
			sort.Slice(spec.expMembers, func(i, j int) bool {
				addri, err := sdk.AccAddressFromBech32(spec.expMembers[i].Member.Address)
				s.Require().NoError(err)
				addrj, err := sdk.AccAddressFromBech32(spec.expMembers[j].Member.Address)
				s.Require().NoError(err)
				return bytes.Compare(addri, addrj) < 0
			})
			sort.Slice(loadedMembers, func(i, j int) bool {
				addri, err := sdk.AccAddressFromBech32(loadedMembers[i].Member.Address)
				s.Require().NoError(err)
				addrj, err := sdk.AccAddressFromBech32(loadedMembers[j].Member.Address)
				s.Require().NoError(err)
				return bytes.Compare(addri, addrj) < 0
			})
			for i := range loadedMembers {
				s.Assert().Equal(spec.expMembers[i].Member.Metadata, loadedMembers[i].Member.Metadata)
				s.Assert().Equal(spec.expMembers[i].Member.Address, loadedMembers[i].Member.Address)
				s.Assert().Equal(spec.expMembers[i].Member.Weight, loadedMembers[i].Member.Weight)
				s.Assert().Equal(spec.expMembers[i].GroupId, loadedMembers[i].GroupId)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestCreateGroupAccount() {
	groupRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    s.addr1.String(),
		Members:  nil,
		Metadata: nil,
	})
	s.Require().NoError(err)
	myGroupID := groupRes.GroupId

	specs := map[string]struct {
		req    *group.MsgCreateGroupAccount
		policy group.DecisionPolicy
		expErr bool
	}{
		"all good": {
			req: &group.MsgCreateGroupAccount{
				Admin:    s.addr1.String(),
				Metadata: nil,
				GroupId:  myGroupID,
			},
			policy: group.NewThresholdDecisionPolicy(
				"1",
				gogotypes.Duration{Seconds: 1},
			),
		},
		"decision policy threshold > total group weight": {
			req: &group.MsgCreateGroupAccount{
				Admin:    s.addr1.String(),
				Metadata: nil,
				GroupId:  myGroupID,
			},
			policy: group.NewThresholdDecisionPolicy(
				"10",
				gogotypes.Duration{Seconds: 1},
			),
		},
		"group id does not exists": {
			req: &group.MsgCreateGroupAccount{
				Admin:    s.addr1.String(),
				Metadata: nil,
				GroupId:  9999,
			},
			policy: group.NewThresholdDecisionPolicy(
				"1",
				gogotypes.Duration{Seconds: 1},
			),
			expErr: true,
		},
		"admin not group admin": {
			req: &group.MsgCreateGroupAccount{
				Admin:    s.addr4.String(),
				Metadata: nil,
				GroupId:  myGroupID,
			},
			policy: group.NewThresholdDecisionPolicy(
				"1",
				gogotypes.Duration{Seconds: 1},
			),
			expErr: true,
		},
	}

	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			err := spec.req.SetDecisionPolicy(spec.policy)
			s.Require().NoError(err)

			res, err := s.msgClient.CreateGroupAccount(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			addr := res.Address

			// then all data persisted
			groupAccountRes, err := s.queryClient.GroupAccountInfo(s.ctx, &group.QueryGroupAccountInfoRequest{Address: addr})
			s.Require().NoError(err)

			groupAccount := groupAccountRes.Info
			s.Assert().Equal(addr, groupAccount.Address)
			s.Assert().Equal(myGroupID, groupAccount.GroupId)
			s.Assert().Equal(spec.req.Admin, groupAccount.Admin)
			s.Assert().Equal(spec.req.Metadata, groupAccount.Metadata)
			s.Assert().Equal(uint64(1), groupAccount.Version)
			s.Assert().Equal(spec.policy.(*group.ThresholdDecisionPolicy), groupAccount.GetDecisionPolicy())
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateGroupAccountAdmin() {
	admin, newAdmin := s.addr1, s.addr2
	groupAccountAddr, myGroupID, policy, derivationKey := createGroupAndGroupAccount(admin, s)

	specs := map[string]struct {
		req             *group.MsgUpdateGroupAccountAdmin
		expGroupAccount *group.GroupAccountInfo
		expErr          bool
	}{
		"with wrong admin": {
			req: &group.MsgUpdateGroupAccountAdmin{
				Admin:    s.addr5.String(),
				Address:  groupAccountAddr,
				NewAdmin: newAdmin.String(),
			},
			expGroupAccount: &group.GroupAccountInfo{
				Admin:          admin.String(),
				Address:        groupAccountAddr,
				GroupId:        myGroupID,
				Metadata:       nil,
				Version:        2,
				DecisionPolicy: nil,
				DerivationKey:  derivationKey,
			},
			expErr: true,
		},
		"with wrong group account": {
			req: &group.MsgUpdateGroupAccountAdmin{
				Admin:    admin.String(),
				Address:  s.addr5.String(),
				NewAdmin: newAdmin.String(),
			},
			expGroupAccount: &group.GroupAccountInfo{
				Admin:          admin.String(),
				Address:        groupAccountAddr,
				GroupId:        myGroupID,
				Metadata:       nil,
				Version:        2,
				DecisionPolicy: nil,
				DerivationKey:  derivationKey,
			},
			expErr: true,
		},
		"correct data": {
			req: &group.MsgUpdateGroupAccountAdmin{
				Admin:    admin.String(),
				Address:  groupAccountAddr,
				NewAdmin: newAdmin.String(),
			},
			expGroupAccount: &group.GroupAccountInfo{
				Admin:          newAdmin.String(),
				Address:        groupAccountAddr,
				GroupId:        myGroupID,
				Metadata:       nil,
				Version:        2,
				DecisionPolicy: nil,
				DerivationKey:  derivationKey,
			},
			expErr: false,
		},
	}

	for msg, spec := range specs {
		spec := spec
		err := spec.expGroupAccount.SetDecisionPolicy(policy)
		s.Require().NoError(err)

		s.Run(msg, func() {
			_, err := s.msgClient.UpdateGroupAccountAdmin(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			res, err := s.queryClient.GroupAccountInfo(s.ctx, &group.QueryGroupAccountInfoRequest{
				Address: groupAccountAddr,
			})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expGroupAccount, res.Info)
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateGroupAccountMetadata() {
	admin := s.addr1
	groupAccountAddr, myGroupID, policy, derivationKey := createGroupAndGroupAccount(admin, s)

	specs := map[string]struct {
		req             *group.MsgUpdateGroupAccountMetadata
		expGroupAccount *group.GroupAccountInfo
		expErr          bool
	}{
		"with wrong admin": {
			req: &group.MsgUpdateGroupAccountMetadata{
				Admin:    s.addr5.String(),
				Address:  groupAccountAddr,
				Metadata: []byte("hello"),
			},
			expGroupAccount: &group.GroupAccountInfo{},
			expErr:          true,
		},
		"with wrong group account": {
			req: &group.MsgUpdateGroupAccountMetadata{
				Admin:    admin.String(),
				Address:  s.addr5.String(),
				Metadata: []byte("hello"),
			},
			expGroupAccount: &group.GroupAccountInfo{},
			expErr:          true,
		},
		"with comment too long": {
			req: &group.MsgUpdateGroupAccountMetadata{
				Admin:    admin.String(),
				Address:  s.addr5.String(),
				Metadata: []byte(strings.Repeat("a", 256)),
			},
			expGroupAccount: &group.GroupAccountInfo{},
			expErr:          true,
		},
		"correct data": {
			req: &group.MsgUpdateGroupAccountMetadata{
				Admin:    admin.String(),
				Address:  groupAccountAddr,
				Metadata: []byte("hello"),
			},
			expGroupAccount: &group.GroupAccountInfo{
				Admin:          admin.String(),
				Address:        groupAccountAddr,
				GroupId:        myGroupID,
				Metadata:       []byte("hello"),
				Version:        2,
				DecisionPolicy: nil,
				DerivationKey:  derivationKey,
			},
			expErr: false,
		},
	}

	for msg, spec := range specs {
		spec := spec
		err := spec.expGroupAccount.SetDecisionPolicy(policy)
		s.Require().NoError(err)

		s.Run(msg, func() {
			_, err := s.msgClient.UpdateGroupAccountMetadata(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			res, err := s.queryClient.GroupAccountInfo(s.ctx, &group.QueryGroupAccountInfoRequest{
				Address: groupAccountAddr,
			})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expGroupAccount, res.Info)
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateGroupAccountDecisionPolicy() {
	admin := s.addr1
	groupAccountAddr, myGroupID, policy, derivationKey := createGroupAndGroupAccount(admin, s)

	specs := map[string]struct {
		req             *group.MsgUpdateGroupAccountDecisionPolicy
		policy          group.DecisionPolicy
		expGroupAccount *group.GroupAccountInfo
		expErr          bool
	}{
		"with wrong admin": {
			req: &group.MsgUpdateGroupAccountDecisionPolicy{
				Admin:   s.addr5.String(),
				Address: groupAccountAddr,
			},
			policy:          policy,
			expGroupAccount: &group.GroupAccountInfo{},
			expErr:          true,
		},
		"with wrong group account": {
			req: &group.MsgUpdateGroupAccountDecisionPolicy{
				Admin:   admin.String(),
				Address: s.addr5.String(),
			},
			policy:          policy,
			expGroupAccount: &group.GroupAccountInfo{},
			expErr:          true,
		},
		"correct data": {
			req: &group.MsgUpdateGroupAccountDecisionPolicy{
				Admin:   admin.String(),
				Address: groupAccountAddr,
			},
			policy: group.NewThresholdDecisionPolicy(
				"2",
				gogotypes.Duration{Seconds: 2},
			),
			expGroupAccount: &group.GroupAccountInfo{
				Admin:          admin.String(),
				Address:        groupAccountAddr,
				GroupId:        myGroupID,
				Metadata:       nil,
				Version:        2,
				DecisionPolicy: nil,
				DerivationKey:  derivationKey,
			},
			expErr: false,
		},
	}

	for msg, spec := range specs {
		spec := spec
		err := spec.expGroupAccount.SetDecisionPolicy(spec.policy)
		s.Require().NoError(err)

		err = spec.req.SetDecisionPolicy(spec.policy)
		s.Require().NoError(err)

		s.Run(msg, func() {
			_, err := s.msgClient.UpdateGroupAccountDecisionPolicy(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			res, err := s.queryClient.GroupAccountInfo(s.ctx, &group.QueryGroupAccountInfoRequest{
				Address: groupAccountAddr,
			})
			s.Require().NoError(err)
			s.Assert().Equal(spec.expGroupAccount, res.Info)
		})
	}
}

func (s *IntegrationTestSuite) TestGroupAccountsByAdminOrGroup() {
	admin := s.addr2
	groupRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    admin.String(),
		Members:  nil,
		Metadata: nil,
	})
	s.Require().NoError(err)
	myGroupID := groupRes.GroupId

	policies := []group.DecisionPolicy{
		group.NewThresholdDecisionPolicy(
			"1",
			gogotypes.Duration{Seconds: 1},
		),
		group.NewThresholdDecisionPolicy(
			"5",
			gogotypes.Duration{Seconds: 1},
		),
		group.NewThresholdDecisionPolicy(
			"10",
			gogotypes.Duration{Seconds: 1},
		),
	}

	count := 3
	expectAccs := make([]*group.GroupAccountInfo, count)
	for i := range expectAccs {
		req := &group.MsgCreateGroupAccount{
			Admin:    admin.String(),
			Metadata: nil,
			GroupId:  myGroupID,
		}
		err := req.SetDecisionPolicy(policies[i])
		s.Require().NoError(err)
		res, err := s.msgClient.CreateGroupAccount(s.ctx, req)
		s.Require().NoError(err)

		expectAcc := &group.GroupAccountInfo{
			Address:  res.Address,
			Admin:    admin.String(),
			Metadata: nil,
			GroupId:  myGroupID,
			Version:  uint64(1),
		}
		err = expectAcc.SetDecisionPolicy(policies[i])
		s.Require().NoError(err)
		expectAccs[i] = expectAcc
	}
	// we reorder accounts by address to be able to compare them
	sort.Slice(expectAccs, func(i, j int) bool {
		addri, err := sdk.AccAddressFromBech32(expectAccs[i].Address)
		s.Require().NoError(err)
		addrj, err := sdk.AccAddressFromBech32(expectAccs[j].Address)
		s.Require().NoError(err)
		return bytes.Compare(addri, addrj) < 0
	})

	// query group account by group
	accountsByGroupRes, err := s.queryClient.GroupAccountsByGroup(s.ctx, &group.QueryGroupAccountsByGroupRequest{
		GroupId: myGroupID,
	})
	s.Require().NoError(err)
	accounts := accountsByGroupRes.GroupAccounts
	s.Require().Equal(len(accounts), count)
	for i := range accounts {
		s.Assert().Equal(accounts[i].Address, expectAccs[i].Address)
		s.Assert().Equal(accounts[i].GroupId, expectAccs[i].GroupId)
		s.Assert().Equal(accounts[i].Admin, expectAccs[i].Admin)
		s.Assert().Equal(accounts[i].Metadata, expectAccs[i].Metadata)
		s.Assert().Equal(accounts[i].Version, expectAccs[i].Version)
		s.Assert().Equal(accounts[i].GetDecisionPolicy(), expectAccs[i].GetDecisionPolicy())
	}

	// query group account by admin
	accountsByAdminRes, err := s.queryClient.GroupAccountsByAdmin(s.ctx, &group.QueryGroupAccountsByAdminRequest{
		Admin: admin.String(),
	})
	s.Require().NoError(err)
	accounts = accountsByAdminRes.GroupAccounts
	s.Require().Equal(len(accounts), count)
	for i := range accounts {
		s.Assert().Equal(accounts[i].Address, expectAccs[i].Address)
		s.Assert().Equal(accounts[i].GroupId, expectAccs[i].GroupId)
		s.Assert().Equal(accounts[i].Admin, expectAccs[i].Admin)
		s.Assert().Equal(accounts[i].Metadata, expectAccs[i].Metadata)
		s.Assert().Equal(accounts[i].Version, expectAccs[i].Version)
		s.Assert().Equal(accounts[i].GetDecisionPolicy(), expectAccs[i].GetDecisionPolicy())
	}
}

func (s *IntegrationTestSuite) TestCreateProposal() {
	myGroupID := s.groupID
	accountAddr := s.groupAccountAddr

	msgSend := &banktypes.MsgSend{
		FromAddress: s.groupAccountAddr.String(),
		ToAddress:   s.addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}

	accountReq := &group.MsgCreateGroupAccount{
		Admin:    s.addr1.String(),
		GroupId:  myGroupID,
		Metadata: nil,
	}
	policy := group.NewThresholdDecisionPolicy(
		"100",
		gogotypes.Duration{Seconds: 1},
	)
	err := accountReq.SetDecisionPolicy(policy)
	s.Require().NoError(err)
	bigThresholdRes, err := s.msgClient.CreateGroupAccount(s.ctx, accountReq)
	s.Require().NoError(err)
	bigThresholdAddr := bigThresholdRes.Address

	defaultProposal := group.Proposal{
		Status: group.ProposalStatusSubmitted,
		Result: group.ProposalResultUnfinalized,
		VoteState: group.Tally{
			YesCount:     "0",
			NoCount:      "0",
			AbstainCount: "0",
			VetoCount:    "0",
		},
		ExecutorResult: group.ProposalExecutorResultNotRun,
	}

	specs := map[string]struct {
		req         *group.MsgCreateProposal
		msgs        []sdk.Msg
		expProposal group.Proposal
		expErr      bool
		postRun     func(sdkCtx sdk.Context)
	}{
		"all good with minimal fields set": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{s.addr2.String()},
			},
			expProposal: defaultProposal,
			postRun:     func(sdkCtx sdk.Context) {},
		},
		"all good with good msg payload": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{s.addr2.String()},
			},
			msgs: []sdk.Msg{&banktypes.MsgSend{
				FromAddress: accountAddr.String(),
				ToAddress:   s.addr2.String(),
				Amount:      sdk.Coins{sdk.NewInt64Coin("token", 100)},
			}},
			expProposal: defaultProposal,
			postRun:     func(sdkCtx sdk.Context) {},
		},
		"group account required": {
			req: &group.MsgCreateProposal{
				Metadata:  nil,
				Proposers: []string{s.addr2.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"existing group account required": {
			req: &group.MsgCreateProposal{
				Address:   s.addr1.String(),
				Proposers: []string{s.addr2.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"impossible case: decision policy threshold > total group weight": {
			req: &group.MsgCreateProposal{
				Address:   bigThresholdAddr,
				Proposers: []string{s.addr2.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"only group members can create a proposal": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{s.addr4.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"all proposers must be in group": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{s.addr2.String(), s.addr4.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"proposers must not be empty": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{s.addr2.String(), ""},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"admin that is not a group member can not create proposal": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Metadata:  nil,
				Proposers: []string{s.addr1.String()},
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"reject msgs that are not authz by group account": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Metadata:  nil,
				Proposers: []string{s.addr2.String()},
			},
			msgs:    []sdk.Msg{&testdata.MsgAuthenticated{Signers: []sdk.AccAddress{s.addr1}}},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"with try exec": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{s.addr2.String()},
				Exec:      group.Exec_EXEC_TRY,
			},
			msgs: []sdk.Msg{msgSend},
			expProposal: group.Proposal{
				Status: group.ProposalStatusClosed,
				Result: group.ProposalResultAccepted,
				VoteState: group.Tally{
					YesCount:     "2",
					NoCount:      "0",
					AbstainCount: "0",
					VetoCount:    "0",
				},
				ExecutorResult: group.ProposalExecutorResultSuccess,
			},
			postRun: func(sdkCtx sdk.Context) {
				fromBalances := s.bankKeeper.GetAllBalances(sdkCtx, accountAddr)
				s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("test", 9900)}, fromBalances)
				toBalances := s.bankKeeper.GetAllBalances(sdkCtx, s.addr2)
				s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("test", 100)}, toBalances)
			},
		},
		"with try exec, not enough yes votes for proposal to pass": {
			req: &group.MsgCreateProposal{
				Address:   accountAddr.String(),
				Proposers: []string{s.addr5.String()},
				Exec:      group.Exec_EXEC_TRY,
			},
			msgs: []sdk.Msg{msgSend},
			expProposal: group.Proposal{
				Status: group.ProposalStatusSubmitted,
				Result: group.ProposalResultUnfinalized,
				VoteState: group.Tally{
					YesCount:     "1",
					NoCount:      "0",
					AbstainCount: "0",
					VetoCount:    "0",
				},
				ExecutorResult: group.ProposalExecutorResultNotRun,
			},
			postRun: func(sdkCtx sdk.Context) {},
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			err := spec.req.SetMsgs(spec.msgs)
			s.Require().NoError(err)

			res, err := s.msgClient.CreateProposal(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			id := res.ProposalId

			// then all data persisted
			proposalRes, err := s.queryClient.Proposal(s.ctx, &group.QueryProposalRequest{ProposalId: id})
			s.Require().NoError(err)
			proposal := proposalRes.Proposal

			s.Assert().Equal(accountAddr.String(), proposal.Address)
			s.Assert().Equal(spec.req.Metadata, proposal.Metadata)
			s.Assert().Equal(spec.req.Proposers, proposal.Proposers)

			submittedAt, err := gogotypes.TimestampFromProto(&proposal.SubmittedAt)
			s.Require().NoError(err)
			s.Assert().Equal(s.blockTime, submittedAt)

			s.Assert().Equal(uint64(1), proposal.GroupVersion)
			s.Assert().Equal(uint64(1), proposal.GroupAccountVersion)
			s.Assert().Equal(spec.expProposal.Status, proposal.Status)
			s.Assert().Equal(spec.expProposal.Result, proposal.Result)
			s.Assert().Equal(spec.expProposal.VoteState, proposal.VoteState)
			s.Assert().Equal(spec.expProposal.ExecutorResult, proposal.ExecutorResult)

			timeout, err := gogotypes.TimestampFromProto(&proposal.Timeout)
			s.Require().NoError(err)
			s.Assert().Equal(s.blockTime.Add(time.Second), timeout)

			if spec.msgs == nil { // then empty list is ok
				s.Assert().Len(proposal.GetMsgs(), 0)
			} else {
				s.Assert().Equal(spec.msgs, proposal.GetMsgs())
			}

			spec.postRun(s.sdkCtx)
		})
	}
}

func (s *IntegrationTestSuite) TestVote() {
	members := []group.Member{
		{Address: s.addr4.String(), Weight: "1"},
		{Address: s.addr3.String(), Weight: "2"},
	}
	groupRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    s.addr1.String(),
		Members:  members,
		Metadata: nil,
	})
	s.Require().NoError(err)
	myGroupID := groupRes.GroupId

	policy := group.NewThresholdDecisionPolicy(
		"2",
		gogotypes.Duration{Seconds: 1},
	)
	accountReq := &group.MsgCreateGroupAccount{
		Admin:    s.addr1.String(),
		GroupId:  myGroupID,
		Metadata: nil,
	}
	err = accountReq.SetDecisionPolicy(policy)
	s.Require().NoError(err)
	accountRes, err := s.msgClient.CreateGroupAccount(s.ctx, accountReq)
	s.Require().NoError(err)
	accountAddr := accountRes.Address
	groupAccount, err := sdk.AccAddressFromBech32(accountAddr)
	s.Require().NoError(err)
	s.Require().NotNil(groupAccount)

	s.Require().NoError(fundAccount(s.bankKeeper, s.sdkCtx, groupAccount, sdk.Coins{sdk.NewInt64Coin("test", 10000)}))

	req := &group.MsgCreateProposal{
		Address:   accountAddr,
		Metadata:  nil,
		Proposers: []string{s.addr4.String()},
		Msgs:      nil,
	}
	err = req.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
		FromAddress: accountAddr,
		ToAddress:   s.addr5.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}})
	s.Require().NoError(err)

	proposalRes, err := s.msgClient.CreateProposal(s.ctx, req)
	s.Require().NoError(err)
	myProposalID := proposalRes.ProposalId

	// proposals by group account
	proposalsRes, err := s.queryClient.ProposalsByGroupAccount(s.ctx, &group.QueryProposalsByGroupAccountRequest{
		Address: accountAddr,
	})
	s.Require().NoError(err)
	proposals := proposalsRes.Proposals
	s.Require().Equal(len(proposals), 1)
	s.Assert().Equal(req.Address, proposals[0].Address)
	s.Assert().Equal(req.Metadata, proposals[0].Metadata)
	s.Assert().Equal(req.Proposers, proposals[0].Proposers)

	submittedAt, err := gogotypes.TimestampFromProto(&proposals[0].SubmittedAt)
	s.Require().NoError(err)
	s.Assert().Equal(s.blockTime, submittedAt)

	s.Assert().Equal(uint64(1), proposals[0].GroupVersion)
	s.Assert().Equal(uint64(1), proposals[0].GroupAccountVersion)
	s.Assert().Equal(group.ProposalStatusSubmitted, proposals[0].Status)
	s.Assert().Equal(group.ProposalResultUnfinalized, proposals[0].Result)
	s.Assert().Equal(group.Tally{
		YesCount:     "0",
		NoCount:      "0",
		AbstainCount: "0",
		VetoCount:    "0",
	}, proposals[0].VoteState)

	specs := map[string]struct {
		srcCtx            sdk.Context
		expVoteState      group.Tally
		req               *group.MsgVote
		doBefore          func(ctx context.Context)
		postRun           func(sdkCtx sdk.Context)
		expProposalStatus group.Proposal_Status
		expResult         group.Proposal_Result
		expExecutorResult group.Proposal_ExecutorResult
		expErr            bool
	}{
		"vote yes": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_YES,
			},
			expVoteState: group.Tally{
				YesCount:     "1",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expResult:         group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"with try exec": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr3.String(),
				Choice:     group.Choice_CHOICE_YES,
				Exec:       group.Exec_EXEC_TRY,
			},
			expVoteState: group.Tally{
				YesCount:     "2",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusClosed,
			expResult:         group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
			postRun: func(sdkCtx sdk.Context) {
				fromBalances := s.bankKeeper.GetAllBalances(sdkCtx, groupAccount)
				s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("test", 9900)}, fromBalances)
				toBalances := s.bankKeeper.GetAllBalances(sdkCtx, s.addr2)
				s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("test", 100)}, toBalances)
			},
		},
		"with try exec, not enough yes votes for proposal to pass": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_YES,
				Exec:       group.Exec_EXEC_TRY,
			},
			expVoteState: group.Tally{
				YesCount:     "1",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expResult:         group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"vote no": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			expVoteState: group.Tally{
				YesCount:     "0",
				NoCount:      "1",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expResult:         group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"vote abstain": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_ABSTAIN,
			},
			expVoteState: group.Tally{
				YesCount:     "0",
				NoCount:      "0",
				AbstainCount: "1",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expResult:         group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"vote veto": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_VETO,
			},
			expVoteState: group.Tally{
				YesCount:     "0",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "1",
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expResult:         group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"apply decision policy early": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr3.String(),
				Choice:     group.Choice_CHOICE_YES,
			},
			expVoteState: group.Tally{
				YesCount:     "2",
				NoCount:      "0",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusClosed,
			expResult:         group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun:           func(sdkCtx sdk.Context) {},
		},
		"reject new votes when final decision is made already": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_YES,
			},
			doBefore: func(ctx context.Context) {
				_, err := s.msgClient.Vote(ctx, &group.MsgVote{
					ProposalId: myProposalID,
					Voter:      s.addr3.String(),
					Choice:     group.Choice_CHOICE_VETO,
				})
				s.Require().NoError(err)
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"existing proposal required": {
			req: &group.MsgVote{
				ProposalId: 999,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"empty choice": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"invalid choice": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     5,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"voter must be in group": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr2.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"voter must not be empty": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      "",
				Choice:     group.Choice_CHOICE_NO,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"voters must not be nil": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Choice:     group.Choice_CHOICE_NO,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"admin that is not a group member can not vote": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr1.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"on timeout": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			srcCtx:  s.sdkCtx.WithBlockTime(s.blockTime.Add(time.Second)),
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"closed already": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			doBefore: func(ctx context.Context) {
				_, err := s.msgClient.Vote(ctx, &group.MsgVote{
					ProposalId: myProposalID,
					Voter:      s.addr3.String(),
					Choice:     group.Choice_CHOICE_YES,
				})
				s.Require().NoError(err)
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"voted already": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			doBefore: func(ctx context.Context) {
				_, err := s.msgClient.Vote(ctx, &group.MsgVote{
					ProposalId: myProposalID,
					Voter:      s.addr4.String(),
					Choice:     group.Choice_CHOICE_YES,
				})
				s.Require().NoError(err)
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"with group modified": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			doBefore: func(ctx context.Context) {
				_, err = s.msgClient.UpdateGroupMetadata(ctx, &group.MsgUpdateGroupMetadata{
					GroupId:  myGroupID,
					Admin:    s.addr1.String(),
					Metadata: []byte{1, 2, 3},
				})
				s.Require().NoError(err)
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
		"with policy modified": {
			req: &group.MsgVote{
				ProposalId: myProposalID,
				Voter:      s.addr4.String(),
				Choice:     group.Choice_CHOICE_NO,
			},
			doBefore: func(ctx context.Context) {
				m, err := group.NewMsgUpdateGroupAccountDecisionPolicyRequest(
					s.addr1,
					groupAccount,
					&group.ThresholdDecisionPolicy{
						Threshold: "1",
						Timeout:   gogotypes.Duration{Seconds: 1},
					},
				)
				s.Require().NoError(err)

				_, err = s.msgClient.UpdateGroupAccountDecisionPolicy(ctx, m)
				s.Require().NoError(err)
			},
			expErr:  true,
			postRun: func(sdkCtx sdk.Context) {},
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx := s.sdkCtx
			if !spec.srcCtx.IsZero() {
				sdkCtx = spec.srcCtx
			}
			sdkCtx, _ = sdkCtx.CacheContext()
			ctx := types.Context{Context: sdkCtx}

			if spec.doBefore != nil {
				spec.doBefore(ctx)
			}
			_, err := s.msgClient.Vote(ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			s.Require().NoError(err)
			// vote is stored and all data persisted
			res, err := s.queryClient.VoteByProposalVoter(ctx, &group.QueryVoteByProposalVoterRequest{
				ProposalId: spec.req.ProposalId,
				Voter:      spec.req.Voter,
			})
			s.Require().NoError(err)
			loaded := res.Vote
			s.Assert().Equal(spec.req.ProposalId, loaded.ProposalId)
			s.Assert().Equal(spec.req.Voter, loaded.Voter)
			s.Assert().Equal(spec.req.Choice, loaded.Choice)
			s.Assert().Equal(spec.req.Metadata, loaded.Metadata)
			submittedAt, err := gogotypes.TimestampFromProto(&loaded.SubmittedAt)
			s.Require().NoError(err)
			s.Assert().Equal(s.blockTime, submittedAt)

			// query votes by proposal
			votesByProposalRes, err := s.queryClient.VotesByProposal(ctx, &group.QueryVotesByProposalRequest{
				ProposalId: spec.req.ProposalId,
			})
			s.Require().NoError(err)
			votesByProposal := votesByProposalRes.Votes
			s.Require().Equal(1, len(votesByProposal))
			vote := votesByProposal[0]
			s.Assert().Equal(spec.req.ProposalId, vote.ProposalId)
			s.Assert().Equal(spec.req.Voter, vote.Voter)
			s.Assert().Equal(spec.req.Choice, vote.Choice)
			s.Assert().Equal(spec.req.Metadata, vote.Metadata)
			submittedAt, err = gogotypes.TimestampFromProto(&vote.SubmittedAt)
			s.Require().NoError(err)
			s.Assert().Equal(s.blockTime, submittedAt)

			// query votes by voter
			voter := spec.req.Voter
			votesByVoterRes, err := s.queryClient.VotesByVoter(ctx, &group.QueryVotesByVoterRequest{
				Voter: voter,
			})
			s.Require().NoError(err)
			votesByVoter := votesByVoterRes.Votes
			s.Require().Equal(1, len(votesByVoter))
			s.Assert().Equal(spec.req.ProposalId, votesByVoter[0].ProposalId)
			s.Assert().Equal(voter, votesByVoter[0].Voter)
			s.Assert().Equal(spec.req.Choice, votesByVoter[0].Choice)
			s.Assert().Equal(spec.req.Metadata, votesByVoter[0].Metadata)
			submittedAt, err = gogotypes.TimestampFromProto(&votesByVoter[0].SubmittedAt)
			s.Require().NoError(err)
			s.Assert().Equal(s.blockTime, submittedAt)

			// and proposal is updated
			proposalRes, err := s.queryClient.Proposal(ctx, &group.QueryProposalRequest{
				ProposalId: spec.req.ProposalId,
			})
			s.Require().NoError(err)
			proposal := proposalRes.Proposal
			s.Assert().Equal(spec.expVoteState, proposal.VoteState)
			s.Assert().Equal(spec.expResult, proposal.Result)
			s.Assert().Equal(spec.expProposalStatus, proposal.Status)
			s.Assert().Equal(spec.expExecutorResult, proposal.ExecutorResult)

			spec.postRun(sdkCtx)
		})
	}
}

// todo: add test for timeout
func (s *IntegrationTestSuite) TestVoteAgg() {
	members := []group.Member{
		{Address: s.addrBls1.String(), Weight: "1"},
		{Address: s.addrBls2.String(), Weight: "2"},
		{Address: s.addrBls3.String(), Weight: "3"},
		{Address: s.addrBls4.String(), Weight: "4"},
		{Address: s.addrBls5.String(), Weight: "5"},
	}
	groupRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    s.addrBls6.String(),
		Members:  members,
		Metadata: nil,
		BlsOnly:  true,
	})
	s.Require().NoError(err)
	myGroupID := groupRes.GroupId

	policy := group.NewThresholdDecisionPolicy(
		"8",
		gogotypes.Duration{Seconds: 20},
	)
	accountReq := &group.MsgCreateGroupAccount{
		Admin:    s.addrBls6.String(),
		GroupId:  myGroupID,
		Metadata: nil,
	}
	err = accountReq.SetDecisionPolicy(policy)
	s.Require().NoError(err)
	accountRes, err := s.msgClient.CreateGroupAccount(s.ctx, accountReq)
	s.Require().NoError(err)
	accountAddr := accountRes.Address
	groupAccount, err := sdk.AccAddressFromBech32(accountAddr)
	s.Require().NoError(err)
	s.Require().NotNil(groupAccount)

	s.Require().NoError(fundAccount(s.bankKeeper, s.sdkCtx, groupAccount, sdk.Coins{sdk.NewInt64Coin("test", 10000)}))

	req := &group.MsgCreateProposal{
		Address:   accountAddr,
		Metadata:  nil,
		Proposers: []string{s.addrBls1.String()},
		Msgs:      nil,
	}
	err = req.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
		FromAddress: accountAddr,
		ToAddress:   s.addrBls1.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}})
	s.Require().NoError(err)

	proposalRes, err := s.msgClient.CreateProposal(s.ctx, req)
	s.Require().NoError(err)
	myProposalID := proposalRes.ProposalId

	// proposals by group account
	proposalsRes, err := s.queryClient.ProposalsByGroupAccount(s.ctx, &group.QueryProposalsByGroupAccountRequest{
		Address: accountAddr,
	})
	s.Require().NoError(err)
	proposals := proposalsRes.Proposals
	s.Require().Equal(len(proposals), 1)
	s.Assert().Equal(req.Address, proposals[0].Address)
	s.Assert().Equal(req.Metadata, proposals[0].Metadata)
	s.Assert().Equal(req.Proposers, proposals[0].Proposers)

	submittedAt, err := gogotypes.TimestampFromProto(&proposals[0].SubmittedAt)
	s.Require().NoError(err)
	s.Assert().Equal(s.blockTime, submittedAt)

	s.Assert().Equal(uint64(1), proposals[0].GroupVersion)
	s.Assert().Equal(uint64(1), proposals[0].GroupAccountVersion)
	s.Assert().Equal(group.ProposalStatusSubmitted, proposals[0].Status)
	s.Assert().Equal(group.ProposalResultUnfinalized, proposals[0].Result)
	s.Assert().Equal(group.Tally{
		YesCount:     "0",
		NoCount:      "0",
		AbstainCount: "0",
		VetoCount:    "0",
	}, proposals[0].VoteState)

	voteAggTimeout, err := gogotypes.TimestampProto(submittedAt.Add(time.Second * 10))
	s.Require().NoError(err)

	type fullVote struct {
		Address string
		Choice  group.Choice
	}

	rawVotesAcc := []fullVote{
		{Address: s.addrBls1.String(), Choice: group.Choice_CHOICE_YES},
		{Address: s.addrBls2.String(), Choice: group.Choice_CHOICE_NO},
		{Address: s.addrBls3.String(), Choice: group.Choice_CHOICE_YES},
		{Address: s.addrBls4.String(), Choice: group.Choice_CHOICE_UNSPECIFIED},
		{Address: s.addrBls5.String(), Choice: group.Choice_CHOICE_YES},
	}
	sort.Slice(rawVotesAcc, func(i, j int) bool {
		addri, err := sdk.AccAddressFromBech32(rawVotesAcc[i].Address)
		s.Require().NoError(err)
		addrj, err := sdk.AccAddressFromBech32(rawVotesAcc[j].Address)
		s.Require().NoError(err)
		return bytes.Compare(addri, addrj) < 0
	})
	sortedVotesAcc := make([]group.Choice, len(rawVotesAcc))
	validVotesAcc := make([]fullVote, 0, len(rawVotesAcc))
	for i, v := range rawVotesAcc {
		sortedVotesAcc[i] = v.Choice
		if v.Choice != group.Choice_CHOICE_UNSPECIFIED {
			validVotesAcc = append(validVotesAcc, v)
		}
	}

	rawVotesRej := []fullVote{
		{Address: s.addrBls1.String(), Choice: group.Choice_CHOICE_YES},
		{Address: s.addrBls2.String(), Choice: group.Choice_CHOICE_ABSTAIN},
		{Address: s.addrBls3.String(), Choice: group.Choice_CHOICE_NO},
		{Address: s.addrBls4.String(), Choice: group.Choice_CHOICE_UNSPECIFIED},
		{Address: s.addrBls5.String(), Choice: group.Choice_CHOICE_VETO},
	}
	sort.Slice(rawVotesRej, func(i, j int) bool {
		addri, err := sdk.AccAddressFromBech32(rawVotesRej[i].Address)
		s.Require().NoError(err)
		addrj, err := sdk.AccAddressFromBech32(rawVotesRej[j].Address)
		s.Require().NoError(err)
		return bytes.Compare(addri, addrj) < 0
	})
	sortedVotesRej := make([]group.Choice, len(rawVotesRej))
	validVotesRej := make([]fullVote, 0, len(rawVotesRej))
	for i, v := range rawVotesRej {
		sortedVotesRej[i] = v.Choice
		if v.Choice != group.Choice_CHOICE_UNSPECIFIED {
			validVotesRej = append(validVotesRej, v)
		}
	}

	msgNo := &group.MsgVoteBasic{
		ProposalId: myProposalID,
		Choice:     group.Choice_CHOICE_NO,
		Expiry:     *voteAggTimeout,
	}
	signBytesNo := msgNo.GetSignBytes()

	msgYes := &group.MsgVoteBasic{
		ProposalId: myProposalID,
		Choice:     group.Choice_CHOICE_YES,
		Expiry:     *voteAggTimeout,
	}
	signBytesYes := msgYes.GetSignBytes()

	msgAbstain := &group.MsgVoteBasic{
		ProposalId: myProposalID,
		Choice:     group.Choice_CHOICE_ABSTAIN,
		Expiry:     *voteAggTimeout,
	}
	signBytesAbstain := msgAbstain.GetSignBytes()

	msgVeto := &group.MsgVoteBasic{
		ProposalId: myProposalID,
		Choice:     group.Choice_CHOICE_VETO,
		Expiry:     *voteAggTimeout,
	}
	signBytesVeto := msgVeto.GetSignBytes()

	sig1, err := s.skBls1.Sign(signBytesYes)
	s.Require().NoError(err)
	sig2, err := s.skBls2.Sign(signBytesNo)
	s.Require().NoError(err)
	sig3, err := s.skBls3.Sign(signBytesYes)
	s.Require().NoError(err)
	sig5, err := s.skBls5.Sign(signBytesYes)
	s.Require().NoError(err)
	sigmaAcc, err := bls12381.AggregateSignature([][]byte{sig1, sig2, sig3, sig5})
	s.Require().NoError(err)

	sig1, err = s.skBls1.Sign(signBytesYes)
	s.Require().NoError(err)
	sig2, err = s.skBls2.Sign(signBytesAbstain)
	s.Require().NoError(err)
	sig3, err = s.skBls3.Sign(signBytesNo)
	s.Require().NoError(err)
	sig5, err = s.skBls5.Sign(signBytesVeto)
	s.Require().NoError(err)
	sigmaRej, err := bls12381.AggregateSignature([][]byte{sig1, sig2, sig3, sig5})
	s.Require().NoError(err)

	specs := map[string]struct {
		expVoteState      group.Tally
		req               *group.MsgVoteAgg
		votes             []fullVote
		postRun           func(sdkCtx sdk.Context)
		expProposalStatus group.Proposal_Status
		expResult         group.Proposal_Result
		expExecutorResult group.Proposal_ExecutorResult
		expErr            bool
	}{
		"result accepted with exec": {
			req: &group.MsgVoteAgg{
				Sender:     s.addr1.String(),
				ProposalId: myProposalID,
				Votes:      sortedVotesAcc,
				Expiry:     *voteAggTimeout,
				AggSig:     sigmaAcc,
				Exec:       group.Exec_EXEC_TRY,
			},
			votes: validVotesAcc,
			expVoteState: group.Tally{
				YesCount:     "9",
				NoCount:      "2",
				AbstainCount: "0",
				VetoCount:    "0",
			},
			expProposalStatus: group.ProposalStatusClosed,
			expResult:         group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
			postRun: func(sdkCtx sdk.Context) {
				fromBalances := s.bankKeeper.GetAllBalances(sdkCtx, groupAccount)
				s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("test", 9900)}, fromBalances)
				toBalances := s.bankKeeper.GetAllBalances(sdkCtx, s.addrBls1)
				s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("test", 100)}, toBalances)
			},
		},
		"result rejected": {
			req: &group.MsgVoteAgg{
				Sender:     s.addr1.String(),
				ProposalId: myProposalID,
				Votes:      sortedVotesRej,
				Expiry:     *voteAggTimeout,
				AggSig:     sigmaRej,
				Exec:       group.Exec_EXEC_TRY,
			},
			votes: validVotesRej,
			expVoteState: group.Tally{
				YesCount:     "1",
				NoCount:      "3",
				AbstainCount: "2",
				VetoCount:    "5",
			},
			expProposalStatus: group.ProposalStatusClosed,
			expResult:         group.ProposalResultRejected,
			expExecutorResult: group.ProposalExecutorResultNotRun,
			postRun: func(sdkCtx sdk.Context) {
				fromBalances := s.bankKeeper.GetAllBalances(sdkCtx, groupAccount)
				s.Require().Equal(sdk.Coins{sdk.NewInt64Coin("test", 10000)}, fromBalances)
				toBalances := s.bankKeeper.GetAllBalances(sdkCtx, s.addrBls1)
				s.Require().Equal(sdk.Coins{}, toBalances)
			},
		},
		"invalid signature": {
			req: &group.MsgVoteAgg{
				Sender:     s.addr1.String(),
				ProposalId: myProposalID,
				Votes:      sortedVotesAcc,
				Expiry:     *voteAggTimeout,
				AggSig:     sigmaRej,
				Exec:       group.Exec_EXEC_TRY,
			},
			expErr: true,
		},
	}

	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx, _ := s.sdkCtx.CacheContext()
			ctx := types.Context{Context: sdkCtx}

			_, err := s.msgClient.VoteAgg(ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// query votes by proposal
			votesByProposalRes, err := s.queryClient.VotesByProposal(ctx, &group.QueryVotesByProposalRequest{
				ProposalId: spec.req.ProposalId,
			})
			s.Require().NoError(err)
			votesByProposal := votesByProposalRes.Votes
			s.Require().Equal(len(spec.votes), len(votesByProposal))

			for i, vote := range votesByProposal {
				s.Assert().Equal(spec.req.ProposalId, vote.ProposalId)
				s.Assert().Equal(spec.votes[i].Address, vote.Voter)
				s.Assert().Equal(spec.votes[i].Choice, vote.Choice)
				submittedAt, err = gogotypes.TimestampFromProto(&vote.SubmittedAt)
				s.Require().NoError(err)
				s.Assert().Equal(s.blockTime, submittedAt)
			}

			// query votes by voter
			for _, vote := range spec.votes {
				votesByVoterRes, err := s.queryClient.VotesByVoter(ctx, &group.QueryVotesByVoterRequest{
					Voter: vote.Address,
				})
				s.Require().NoError(err)
				votesByVoter := votesByVoterRes.Votes
				s.Require().Equal(1, len(votesByVoter))
				s.Assert().Equal(spec.req.ProposalId, votesByVoter[0].ProposalId)
				s.Assert().Equal(vote.Address, votesByVoter[0].Voter)
				s.Assert().Equal(vote.Choice, votesByVoter[0].Choice)
				submittedAt, err = gogotypes.TimestampFromProto(&votesByVoter[0].SubmittedAt)
				s.Require().NoError(err)
				s.Assert().Equal(s.blockTime, submittedAt)
			}

			// and proposal is updated
			proposalRes, err := s.queryClient.Proposal(ctx, &group.QueryProposalRequest{
				ProposalId: spec.req.ProposalId,
			})
			s.Require().NoError(err)
			proposal := proposalRes.Proposal
			s.Assert().Equal(spec.expVoteState, proposal.VoteState)
			s.Assert().Equal(spec.expResult, proposal.Result)
			s.Assert().Equal(spec.expProposalStatus, proposal.Status)
			s.Assert().Equal(spec.expExecutorResult, proposal.ExecutorResult)

			spec.postRun(sdkCtx)
		})
	}
}

func (s *IntegrationTestSuite) TestExecProposal() {
	msgSend1 := &banktypes.MsgSend{
		FromAddress: s.groupAccountAddr.String(),
		ToAddress:   s.addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 100)},
	}
	msgSend2 := &banktypes.MsgSend{
		FromAddress: s.groupAccountAddr.String(),
		ToAddress:   s.addr2.String(),
		Amount:      sdk.Coins{sdk.NewInt64Coin("test", 10001)},
	}
	proposers := []string{s.addr2.String()}

	specs := map[string]struct {
		srcBlockTime      time.Time
		setupProposal     func(ctx context.Context) uint64
		expErr            bool
		expProposalStatus group.Proposal_Status
		expProposalResult group.Proposal_Result
		expExecutorResult group.Proposal_ExecutorResult
		expFromBalances   sdk.Coins
		expToBalances     sdk.Coins
	}{
		"proposal executed when accepted": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_YES)
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
			expFromBalances:   sdk.Coins{sdk.NewInt64Coin("test", 9800)},
			expToBalances:     sdk.Coins{sdk.NewInt64Coin("test", 200)},
		},
		"proposal with multiple messages executed when accepted": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1, msgSend1}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_YES)
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
			expFromBalances:   sdk.Coins{sdk.NewInt64Coin("test", 9700)},
			expToBalances:     sdk.Coins{sdk.NewInt64Coin("test", 300)},
		},
		"proposal not executed when rejected": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_NO)
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultRejected,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"open proposal must not fail": {
			setupProposal: func(ctx context.Context) uint64 {
				return createProposal(ctx, s, []sdk.Msg{msgSend1}, proposers)
			},
			expProposalStatus: group.ProposalStatusSubmitted,
			expProposalResult: group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"existing proposal required": {
			setupProposal: func(ctx context.Context) uint64 {
				return 9999
			},
			expErr: true,
		},
		"Decision policy also applied on timeout": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_NO)
			},
			srcBlockTime:      s.blockTime.Add(time.Second),
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultRejected,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"Decision policy also applied after timeout": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_NO)
			},
			srcBlockTime:      s.blockTime.Add(time.Second).Add(time.Millisecond),
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultRejected,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"with group modified before tally": {
			setupProposal: func(ctx context.Context) uint64 {
				myProposalID := createProposal(ctx, s, []sdk.Msg{msgSend1}, proposers)

				// then modify group
				_, err := s.msgClient.UpdateGroupMetadata(ctx, &group.MsgUpdateGroupMetadata{
					Admin:    s.addr1.String(),
					GroupId:  s.groupID,
					Metadata: []byte{1, 2, 3},
				})
				s.Require().NoError(err)
				return myProposalID
			},
			expProposalStatus: group.ProposalStatusAborted,
			expProposalResult: group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"with group account modified before tally": {
			setupProposal: func(ctx context.Context) uint64 {
				myProposalID := createProposal(ctx, s, []sdk.Msg{msgSend1}, proposers)
				_, err := s.msgClient.UpdateGroupAccountMetadata(ctx, &group.MsgUpdateGroupAccountMetadata{
					Admin:    s.addr1.String(),
					Address:  s.groupAccountAddr.String(),
					Metadata: []byte("group account modified before tally"),
				})
				s.Require().NoError(err)
				return myProposalID
			},
			expProposalStatus: group.ProposalStatusAborted,
			expProposalResult: group.ProposalResultUnfinalized,
			expExecutorResult: group.ProposalExecutorResultNotRun,
		},
		"prevent double execution when successful": {
			setupProposal: func(ctx context.Context) uint64 {
				myProposalID := createProposalAndVote(ctx, s, []sdk.Msg{msgSend1}, proposers, group.Choice_CHOICE_YES)

				_, err := s.msgClient.Exec(ctx, &group.MsgExec{Signer: s.addr1.String(), ProposalId: myProposalID})
				s.Require().NoError(err)
				return myProposalID
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
			expFromBalances:   sdk.Coins{sdk.NewInt64Coin("test", 9800)},
			expToBalances:     sdk.Coins{sdk.NewInt64Coin("test", 200)},
		},
		"rollback all msg updates on failure": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend1, msgSend2}
				return createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_YES)
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultFailure,
		},
		"executable when failed before": {
			setupProposal: func(ctx context.Context) uint64 {
				msgs := []sdk.Msg{msgSend2}
				myProposalID := createProposalAndVote(ctx, s, msgs, proposers, group.Choice_CHOICE_YES)

				_, err := s.msgClient.Exec(ctx, &group.MsgExec{Signer: s.addr1.String(), ProposalId: myProposalID})
				s.Require().NoError(err)
				s.Require().NoError(fundAccount(s.bankKeeper, ctx.(types.Context).Context, s.groupAccountAddr, sdk.Coins{sdk.NewInt64Coin("test", 10002)}))

				return myProposalID
			},
			expProposalStatus: group.ProposalStatusClosed,
			expProposalResult: group.ProposalResultAccepted,
			expExecutorResult: group.ProposalExecutorResultSuccess,
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx, _ := s.sdkCtx.CacheContext()
			ctx := types.Context{Context: sdkCtx}

			proposalID := spec.setupProposal(ctx)

			if !spec.srcBlockTime.IsZero() {
				sdkCtx = sdkCtx.WithBlockTime(spec.srcBlockTime)
				ctx = types.Context{Context: sdkCtx}
			}

			_, err := s.msgClient.Exec(ctx, &group.MsgExec{Signer: s.addr1.String(), ProposalId: proposalID})
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// and proposal is updated
			res, err := s.queryClient.Proposal(ctx, &group.QueryProposalRequest{ProposalId: proposalID})
			s.Require().NoError(err)
			proposal := res.Proposal

			exp := group.Proposal_Result_name[int32(spec.expProposalResult)]
			got := group.Proposal_Result_name[int32(proposal.Result)]
			s.Assert().Equal(exp, got)

			exp = group.Proposal_Status_name[int32(spec.expProposalStatus)]
			got = group.Proposal_Status_name[int32(proposal.Status)]
			s.Assert().Equal(exp, got)

			exp = group.Proposal_ExecutorResult_name[int32(spec.expExecutorResult)]
			got = group.Proposal_ExecutorResult_name[int32(proposal.ExecutorResult)]
			s.Assert().Equal(exp, got)

			if spec.expFromBalances != nil {
				fromBalances := s.bankKeeper.GetAllBalances(sdkCtx, s.groupAccountAddr)
				s.Require().Equal(spec.expFromBalances, fromBalances)
			}
			if spec.expToBalances != nil {
				toBalances := s.bankKeeper.GetAllBalances(sdkCtx, s.addr2)
				s.Require().Equal(spec.expToBalances, toBalances)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestCreatePoll() {
	myGroupID := s.groupID
	now := s.blockTime
	endTime, err := gogotypes.TimestampProto(now.Add(time.Second * 100))
	s.Require().NoError(err)
	past := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	oldTime, err := gogotypes.TimestampProto(past)
	s.Require().NoError(err)

	defaultPoll := group.Poll{
		Status: group.PollStatusSubmitted,
	}

	specs := map[string]struct {
		req     *group.MsgCreatePoll
		expPoll group.Poll
		expErr  bool
	}{
		"all good": {
			req: &group.MsgCreatePoll{
				GroupId:   myGroupID,
				Title:     "2021 Election",
				Options:   group.Options{Titles: []string{"alice", "bob", "charlie"}},
				Creator:   s.addr2.String(),
				VoteLimit: 2,
				Timeout:   *endTime,
			},
			expPoll: defaultPoll,
		},
		"only group members can create a poll": {
			req: &group.MsgCreatePoll{
				GroupId:   myGroupID,
				Title:     "2021 Election",
				Options:   group.Options{Titles: []string{"alice", "bob", "charlie"}},
				Creator:   s.addr4.String(),
				VoteLimit: 2,
				Timeout:   *endTime,
			},
			expErr: true,
		},
		"admin that is not a group member can not create poll": {
			req: &group.MsgCreatePoll{
				GroupId:   myGroupID,
				Title:     "2021 Election",
				Options:   group.Options{Titles: []string{"alice", "bob", "charlie"}},
				Creator:   s.addr1.String(),
				VoteLimit: 2,
				Timeout:   *endTime,
			},
			expErr: true,
		},
		"poll expired": {
			req: &group.MsgCreatePoll{
				GroupId:   myGroupID,
				Title:     "2021 Election",
				Options:   group.Options{Titles: []string{"alice", "bob", "charlie"}},
				Creator:   s.addr2.String(),
				VoteLimit: 2,
				Timeout:   *oldTime,
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			res, err := s.msgClient.CreatePoll(s.ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			id := res.PollId

			// then all data persisted
			pollRes, err := s.queryClient.Poll(s.ctx, &group.QueryPollRequest{PollId: id})
			s.Require().NoError(err)
			poll := pollRes.Poll

			s.Assert().Equal(spec.req.GroupId, poll.GroupId)
			s.Assert().Equal(spec.req.Title, poll.Title)
			s.Assert().Equal(spec.req.Options, poll.Options)
			s.Assert().Equal(spec.req.Creator, poll.Creator)
			s.Assert().Equal(spec.req.VoteLimit, poll.VoteLimit)
			s.Assert().Equal(spec.req.Metadata, poll.Metadata)
			s.Assert().Equal(spec.req.Timeout, poll.Timeout)

			submittedAt, err := gogotypes.TimestampFromProto(&poll.SubmittedAt)
			s.Require().NoError(err)
			s.Assert().Equal(s.blockTime, submittedAt)

			s.Assert().Equal(uint64(1), poll.GroupVersion)
			s.Assert().Equal(spec.expPoll.Status, poll.Status)
		})
	}
}

func (s *IntegrationTestSuite) TestVotePoll() {
	members := []group.Member{
		{Address: s.addr4.String(), Weight: "1"},
		{Address: s.addr3.String(), Weight: "2"},
	}
	groupRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    s.addr1.String(),
		Members:  members,
		Metadata: nil,
	})
	s.Require().NoError(err)
	myGroupID := groupRes.GroupId

	now := s.blockTime
	endTime, err := gogotypes.TimestampProto(now.Add(time.Second * 100))
	s.Require().NoError(err)

	req := &group.MsgCreatePoll{
		GroupId:   myGroupID,
		Title:     "2021 Election",
		Options:   group.Options{Titles: []string{"alice", "bob", "charlie", "linda", "tom"}},
		Creator:   s.addr3.String(),
		VoteLimit: 2,
		Timeout:   *endTime,
	}

	pollRes, err := s.msgClient.CreatePoll(s.ctx, req)
	s.Require().NoError(err)
	myPollID := pollRes.PollId

	_, err = s.queryClient.Poll(s.ctx, &group.QueryPollRequest{PollId: myPollID})
	s.Require().NoError(err)

	specs := map[string]struct {
		srcCtx        sdk.Context
		expVoteState  group.TallyPoll
		req           *group.MsgVotePoll
		doBefore      func(ctx context.Context)
		expPollStatus group.Poll_Status
		expErr        bool
	}{
		"all good": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Voter:   s.addr4.String(),
				Options: group.Options{Titles: []string{"alice", "bob"}},
			},
			expVoteState: group.TallyPoll{
				Counts: map[string]string{
					"alice": "1",
					"bob":   "1",
				},
			},
			expPollStatus: group.PollStatusSubmitted,
		},
		"invalid option": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Voter:   s.addr4.String(),
				Options: group.Options{Titles: []string{"eva"}},
			},
			expErr: true,
		},
		"on vote limit": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Voter:   s.addr4.String(),
				Options: group.Options{Titles: []string{"alice", "bob", "charlie"}},
			},
			expErr: true,
		},
		"voter must be in group": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Voter:   s.addr2.String(),
				Options: group.Options{Titles: []string{"alice"}},
			},
			expErr: true,
		},
		"voter must not be empty": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Voter:   "",
				Options: group.Options{Titles: []string{"alice"}},
			},
			expErr: true,
		},
		"voters must not be nil": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Options: group.Options{Titles: []string{"alice"}},
			},
			expErr: true,
		},
		"admin that is not a group member can not vote": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Voter:   s.addr1.String(),
				Options: group.Options{Titles: []string{"alice"}},
			},
			expErr: true,
		},
		"on timeout": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Voter:   s.addr4.String(),
				Options: group.Options{Titles: []string{"alice"}},
			},
			srcCtx: s.sdkCtx.WithBlockTime(s.blockTime.Add(time.Second * 101)),
			expErr: true,
		},
		"multiple votes": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Voter:   s.addr3.String(),
				Options: group.Options{Titles: []string{"alice", "bob"}},
			},
			doBefore: func(ctx context.Context) {
				_, err := s.msgClient.VotePoll(ctx, &group.MsgVotePoll{
					PollId:  myPollID,
					Voter:   s.addr4.String(),
					Options: group.Options{Titles: []string{"bob"}},
				})
				s.Require().NoError(err)
			},
			expVoteState: group.TallyPoll{
				Counts: map[string]string{
					"alice": "2",
					"bob":   "3",
				},
			},
			expPollStatus: group.PollStatusSubmitted,
		},
		"voted already": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Voter:   s.addr4.String(),
				Options: group.Options{Titles: []string{"alice"}},
			},
			doBefore: func(ctx context.Context) {
				_, err := s.msgClient.VotePoll(ctx, &group.MsgVotePoll{
					PollId:  myPollID,
					Voter:   s.addr4.String(),
					Options: group.Options{Titles: []string{"bob"}},
				})
				s.Require().NoError(err)
			},
			expErr: true,
		},
		"with group modified": {
			req: &group.MsgVotePoll{
				PollId:  myPollID,
				Voter:   s.addr4.String(),
				Options: group.Options{Titles: []string{"alice"}},
			},
			doBefore: func(ctx context.Context) {
				_, err = s.msgClient.UpdateGroupMetadata(ctx, &group.MsgUpdateGroupMetadata{
					GroupId:  myGroupID,
					Admin:    s.addr1.String(),
					Metadata: []byte{1, 2, 3},
				})
				s.Require().NoError(err)
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx := s.sdkCtx
			if !spec.srcCtx.IsZero() {
				sdkCtx = spec.srcCtx
			}
			sdkCtx, _ = sdkCtx.CacheContext()
			ctx := types.Context{Context: sdkCtx}

			if spec.doBefore != nil {
				spec.doBefore(ctx)
			}
			_, err := s.msgClient.VotePoll(ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// vote is stored and all data persisted
			res, err := s.queryClient.VoteForPollByPollVoter(ctx, &group.QueryVoteForPollByPollVoterRequest{
				PollId: spec.req.PollId,
				Voter:  spec.req.Voter,
			})
			s.Require().NoError(err)
			loaded := res.Vote
			s.Assert().Equal(spec.req.PollId, loaded.PollId)
			s.Assert().Equal(spec.req.Voter, loaded.Voter)
			s.Assert().Equal(spec.req.Options, loaded.Options)
			s.Assert().Equal(spec.req.Metadata, loaded.Metadata)
			submittedAt, err := gogotypes.TimestampFromProto(&loaded.SubmittedAt)
			s.Require().NoError(err)
			s.Assert().Equal(s.blockTime, submittedAt)

			// query votes by proposal
			votesForPollByPollRes, err := s.queryClient.VotesForPollByPoll(ctx, &group.QueryVotesForPollByPollRequest{
				PollId: spec.req.PollId,
			})
			s.Require().NoError(err)
			votesByPoll := votesForPollByPollRes.Votes
			foundVoter := false
			for _, vote := range votesByPoll {
				if vote.Voter == spec.req.Voter {
					foundVoter = true
					s.Assert().Equal(spec.req.PollId, vote.PollId)
					s.Assert().Equal(spec.req.Voter, vote.Voter)
					s.Assert().Equal(spec.req.Options, vote.Options)
					s.Assert().Equal(spec.req.Metadata, vote.Metadata)
					submittedAt, err = gogotypes.TimestampFromProto(&vote.SubmittedAt)
					s.Require().NoError(err)
					s.Assert().Equal(s.blockTime, submittedAt)
				}
			}
			s.Require().True(foundVoter)

			// query votes by voter
			voter := spec.req.Voter
			votesByVoterRes, err := s.queryClient.VotesForPollByVoter(ctx, &group.QueryVotesForPollByVoterRequest{
				Voter: voter,
			})
			s.Require().NoError(err)
			votesByVoter := votesByVoterRes.Votes
			s.Require().Equal(1, len(votesByVoter))
			s.Assert().Equal(spec.req.PollId, votesByVoter[0].PollId)
			s.Assert().Equal(voter, votesByVoter[0].Voter)
			s.Assert().Equal(spec.req.Options, votesByVoter[0].Options)
			s.Assert().Equal(spec.req.Metadata, votesByVoter[0].Metadata)
			submittedAt, err = gogotypes.TimestampFromProto(&votesByVoter[0].SubmittedAt)
			s.Require().NoError(err)
			s.Assert().Equal(s.blockTime, submittedAt)

			// and poll is updated
			pollRes, err := s.queryClient.Poll(ctx, &group.QueryPollRequest{
				PollId: spec.req.PollId,
			})
			s.Require().NoError(err)
			poll := pollRes.Poll
			s.Assert().Equal(spec.expVoteState, poll.VoteState)
		})
	}
}

func (s *IntegrationTestSuite) TestVotePollAgg() {
	members := []group.Member{
		{Address: s.addrBls1.String(), Weight: "1"},
		{Address: s.addrBls2.String(), Weight: "2"},
		{Address: s.addrBls3.String(), Weight: "3"},
		{Address: s.addrBls4.String(), Weight: "4"},
		{Address: s.addrBls5.String(), Weight: "5"},
	}

	groupRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    s.addrBls6.String(),
		Members:  members,
		Metadata: nil,
		BlsOnly:  true,
	})
	s.Require().NoError(err)
	myGroupID := groupRes.GroupId

	now := s.blockTime
	endTime, err := gogotypes.TimestampProto(now.Add(time.Second * 20))
	s.Require().NoError(err)

	req := &group.MsgCreatePoll{
		GroupId:   myGroupID,
		Title:     "2021 Election",
		Options:   group.Options{Titles: []string{"alice", "bob", "charlie", "linda", "tom"}},
		Creator:   s.addrBls5.String(),
		VoteLimit: 2,
		Timeout:   *endTime,
	}

	pollRes, err := s.msgClient.CreatePoll(s.ctx, req)
	s.Require().NoError(err)
	myPollID := pollRes.PollId

	pollQuery, err := s.queryClient.Poll(s.ctx, &group.QueryPollRequest{PollId: myPollID})
	s.Require().NoError(err)
	submittedAt, err := gogotypes.TimestampFromProto(&pollQuery.Poll.SubmittedAt)
	s.Require().NoError(err)
	s.Assert().Equal(s.blockTime, submittedAt)

	s.Assert().Equal(uint64(1), pollQuery.Poll.GroupVersion)
	s.Assert().Equal(group.PollStatusSubmitted, pollQuery.Poll.Status)

	type fullVote struct {
		Address string
		Options group.Options
	}

	// valid votes
	rawVotes := []fullVote{
		{Address: s.addrBls1.String(), Options: group.Options{Titles: []string{"alice", "bob"}}},
		{Address: s.addrBls2.String()},
		{Address: s.addrBls3.String(), Options: group.Options{Titles: []string{"alice"}}},
		{Address: s.addrBls4.String()},
		{Address: s.addrBls5.String()},
	}
	sort.Slice(rawVotes, func(i, j int) bool {
		addri, err := sdk.AccAddressFromBech32(rawVotes[i].Address)
		s.Require().NoError(err)
		addrj, err := sdk.AccAddressFromBech32(rawVotes[j].Address)
		s.Require().NoError(err)
		return bytes.Compare(addri, addrj) < 0
	})
	sortedVotes := make([]group.Options, len(rawVotes))
	validVotes := make([]fullVote, 0, len(rawVotes))
	for i, v := range rawVotes {
		sortedVotes[i] = v.Options
		if len(v.Options.Titles) != 0 {
			validVotes = append(validVotes, v)
		}
	}

	voteExpiry, err := gogotypes.TimestampProto(submittedAt.Add(time.Second * 10))
	s.Require().NoError(err)

	msgs := make(map[string][]byte, len(req.Options.Titles))
	for _, opt := range req.Options.Titles {
		x := group.MsgVotePollBasic{
			PollId: myPollID,
			Option: opt,
			Expiry: *voteExpiry,
		}
		msgs[opt] = x.GetSignBytes()
	}

	sig11, err := s.skBls1.Sign(msgs["alice"])
	s.Require().NoError(err)
	sig12, err := s.skBls1.Sign(msgs["bob"])
	s.Require().NoError(err)
	sig1, err := bls12381.AggregateSignature([][]byte{sig11, sig12})
	s.Require().NoError(err)

	sig3, err := s.skBls3.Sign(msgs["alice"])
	s.Require().NoError(err)

	sigma, err := bls12381.AggregateSignature([][]byte{sig1, sig3})
	s.Require().NoError(err)

	// vote too late
	rawVotesLate := []fullVote{
		{Address: s.addrBls1.String(), Options: group.Options{Titles: []string{"alice", "bob"}}},
		{Address: s.addrBls2.String()},
		{Address: s.addrBls3.String(), Options: group.Options{Titles: []string{"alice"}}},
		{Address: s.addrBls4.String()},
		{Address: s.addrBls5.String()},
	}
	sort.Slice(rawVotesLate, func(i, j int) bool {
		addri, err := sdk.AccAddressFromBech32(rawVotesLate[i].Address)
		s.Require().NoError(err)
		addrj, err := sdk.AccAddressFromBech32(rawVotesLate[j].Address)
		s.Require().NoError(err)
		return bytes.Compare(addri, addrj) < 0
	})
	sortedVotesLate := make([]group.Options, len(rawVotesLate))
	for i, v := range rawVotesLate {
		sortedVotesLate[i] = v.Options
	}

	voteExpiryLate, err := gogotypes.TimestampProto(submittedAt.Add(time.Second * 30))
	s.Require().NoError(err)

	msgsLate := make(map[string][]byte, len(req.Options.Titles))
	for _, opt := range req.Options.Titles {
		x := group.MsgVotePollBasic{
			PollId: myPollID,
			Option: opt,
			Expiry: *voteExpiryLate,
		}
		msgsLate[opt] = x.GetSignBytes()
	}

	sigLate11, err := s.skBls1.Sign(msgsLate["alice"])
	s.Require().NoError(err)
	sigLate12, err := s.skBls1.Sign(msgsLate["bob"])
	s.Require().NoError(err)
	sigLate1, err := bls12381.AggregateSignature([][]byte{sigLate11, sigLate12})
	s.Require().NoError(err)

	sigLate3, err := s.skBls3.Sign(msgsLate["alice"])
	s.Require().NoError(err)

	sigmaLate, err := bls12381.AggregateSignature([][]byte{sigLate1, sigLate3})
	s.Require().NoError(err)

	// vote limit
	rawVotesLong := []fullVote{
		{Address: s.addrBls1.String(), Options: group.Options{Titles: []string{"alice", "bob", "linda"}}},
		{Address: s.addrBls2.String()},
		{Address: s.addrBls3.String(), Options: group.Options{Titles: []string{"alice"}}},
		{Address: s.addrBls4.String()},
		{Address: s.addrBls5.String()},
	}
	sort.Slice(rawVotesLong, func(i, j int) bool {
		addri, err := sdk.AccAddressFromBech32(rawVotesLong[i].Address)
		s.Require().NoError(err)
		addrj, err := sdk.AccAddressFromBech32(rawVotesLong[j].Address)
		s.Require().NoError(err)
		return bytes.Compare(addri, addrj) < 0
	})
	sortedVotesLong := make([]group.Options, len(rawVotesLong))
	for i, v := range rawVotesLong {
		sortedVotesLong[i] = v.Options
	}

	msgsLong := make(map[string][]byte, len(req.Options.Titles))
	for _, opt := range req.Options.Titles {
		x := group.MsgVotePollBasic{
			PollId: myPollID,
			Option: opt,
			Expiry: *voteExpiry,
		}
		msgsLong[opt] = x.GetSignBytes()
	}

	sigLong11, err := s.skBls1.Sign(msgsLong["alice"])
	s.Require().NoError(err)
	sigLong12, err := s.skBls1.Sign(msgsLong["bob"])
	s.Require().NoError(err)
	sigLong13, err := s.skBls1.Sign(msgsLong["linda"])
	s.Require().NoError(err)
	sigLong1, err := bls12381.AggregateSignature([][]byte{sigLong11, sigLong12, sigLong13})
	s.Require().NoError(err)

	sigLong3, err := s.skBls3.Sign(msgsLong["alice"])
	s.Require().NoError(err)

	sigmaLong, err := bls12381.AggregateSignature([][]byte{sigLong1, sigLong3})
	s.Require().NoError(err)

	// vote invalid option
	rawVotesInvalid := []fullVote{
		{Address: s.addrBls1.String(), Options: group.Options{Titles: []string{"alice", "eva"}}},
		{Address: s.addrBls2.String()},
		{Address: s.addrBls3.String(), Options: group.Options{Titles: []string{"alice"}}},
		{Address: s.addrBls4.String()},
		{Address: s.addrBls5.String()},
	}
	sort.Slice(rawVotesInvalid, func(i, j int) bool {
		addri, err := sdk.AccAddressFromBech32(rawVotesInvalid[i].Address)
		s.Require().NoError(err)
		addrj, err := sdk.AccAddressFromBech32(rawVotesInvalid[j].Address)
		s.Require().NoError(err)
		return bytes.Compare(addri, addrj) < 0
	})
	sortedVotesInvalid := make([]group.Options, len(rawVotesInvalid))
	for i, v := range rawVotesInvalid {
		sortedVotesInvalid[i] = v.Options
	}

	msgsInvalid := make(map[string][]byte, len(req.Options.Titles))
	for _, opt := range req.Options.Titles {
		x := group.MsgVotePollBasic{
			PollId: myPollID,
			Option: opt,
			Expiry: *voteExpiry,
		}
		msgsInvalid[opt] = x.GetSignBytes()
	}

	y := group.MsgVotePollBasic{
		PollId: myPollID,
		Option: "eva",
		Expiry: *voteExpiry,
	}
	msgsInvalid["eva"] = y.GetSignBytes()

	sigInvalid11, err := s.skBls1.Sign(msgsInvalid["alice"])
	s.Require().NoError(err)
	sigInvalid12, err := s.skBls1.Sign(msgsInvalid["eva"])
	s.Require().NoError(err)
	sigInvalid1, err := bls12381.AggregateSignature([][]byte{sigInvalid11, sigInvalid12})
	s.Require().NoError(err)

	sigInvalid3, err := s.skBls3.Sign(msgsInvalid["alice"])
	s.Require().NoError(err)

	sigmaInvalid, err := bls12381.AggregateSignature([][]byte{sigInvalid1, sigInvalid3})
	s.Require().NoError(err)

	// skip already voted
	validVotesSkip := []fullVote{
		{Address: s.addrBls1.String(), Options: group.Options{Titles: []string{"charlie"}}},
		{Address: s.addrBls3.String(), Options: group.Options{Titles: []string{"alice"}}},
	}
	sort.Slice(validVotesSkip, func(i, j int) bool {
		addri, err := sdk.AccAddressFromBech32(validVotesSkip[i].Address)
		s.Require().NoError(err)
		addrj, err := sdk.AccAddressFromBech32(validVotesSkip[j].Address)
		s.Require().NoError(err)
		return bytes.Compare(addri, addrj) < 0
	})

	specs := map[string]struct {
		srcCtx        sdk.Context
		expVoteState  group.TallyPoll
		req           *group.MsgVotePollAgg
		votes         []fullVote
		doBefore      func(ctx context.Context)
		expPollStatus group.Poll_Status
		expErr        bool
	}{
		"all good": {
			req: &group.MsgVotePollAgg{
				Sender:   s.addr1.String(),
				PollId:   myPollID,
				Votes:    sortedVotes,
				Expiry:   *voteExpiry,
				AggSig:   sigma,
				Metadata: []byte(fmt.Sprintf("aggregated votes submitted by %s", s.addr1.String())),
			},
			votes: validVotes,
			expVoteState: group.TallyPoll{
				Counts: map[string]string{
					"alice": "4",
					"bob":   "1",
				},
			},
			expPollStatus: group.PollStatusSubmitted,
		},
		"skip already voted": {
			req: &group.MsgVotePollAgg{
				Sender:   s.addr1.String(),
				PollId:   myPollID,
				Votes:    sortedVotes,
				Expiry:   *voteExpiry,
				AggSig:   sigma,
				Metadata: []byte(fmt.Sprintf("aggregated votes submitted by %s", s.addr1.String())),
			},
			votes: validVotesSkip,
			doBefore: func(ctx context.Context) {
				_, err := s.msgClient.VotePoll(ctx, &group.MsgVotePoll{
					PollId:  myPollID,
					Voter:   s.addrBls1.String(),
					Options: group.Options{Titles: []string{"charlie"}},
				})
				s.Require().NoError(err)
			},
			expVoteState: group.TallyPoll{
				Counts: map[string]string{
					"alice":   "3",
					"bob":     "0",
					"charlie": "1",
				},
			},
			expPollStatus: group.PollStatusSubmitted,
		},
		"on vote expiry": {
			req: &group.MsgVotePollAgg{
				Sender:   s.addr1.String(),
				PollId:   myPollID,
				Votes:    sortedVotes,
				Expiry:   *voteExpiry,
				AggSig:   sigma,
				Metadata: []byte(fmt.Sprintf("aggregated votes submitted by %s", s.addr1.String())),
			},
			srcCtx: s.sdkCtx.WithBlockTime(s.blockTime.Add(time.Second * 15)),
			expErr: true,
		},
		"on vote late": {
			req: &group.MsgVotePollAgg{
				Sender:   s.addr1.String(),
				PollId:   myPollID,
				Votes:    sortedVotesLate,
				Expiry:   *voteExpiryLate,
				AggSig:   sigmaLate,
				Metadata: []byte(fmt.Sprintf("aggregated votes submitted by %s", s.addr1.String())),
			},
			srcCtx: s.sdkCtx.WithBlockTime(s.blockTime.Add(time.Second * 35)),
			expErr: true,
		},
		"on vote limit": {
			req: &group.MsgVotePollAgg{
				Sender:   s.addr1.String(),
				PollId:   myPollID,
				Votes:    sortedVotesLong,
				Expiry:   *voteExpiry,
				AggSig:   sigmaLong,
				Metadata: []byte(fmt.Sprintf("aggregated votes submitted by %s", s.addr1.String())),
			},
			expErr: true,
		},
		"on vote option": {
			req: &group.MsgVotePollAgg{
				Sender:   s.addr1.String(),
				PollId:   myPollID,
				Votes:    sortedVotesInvalid,
				Expiry:   *voteExpiry,
				AggSig:   sigmaInvalid,
				Metadata: []byte(fmt.Sprintf("aggregated votes submitted by %s", s.addr1.String())),
			},
			expErr: true,
		},
		"on poll expiry": {
			req: &group.MsgVotePollAgg{
				Sender:   s.addr1.String(),
				PollId:   myPollID,
				Votes:    sortedVotesLate,
				Expiry:   *voteExpiryLate,
				AggSig:   sigmaLate,
				Metadata: []byte(fmt.Sprintf("aggregated votes submitted by %s", s.addr1.String())),
			},
			srcCtx: s.sdkCtx.WithBlockTime(s.blockTime.Add(time.Second * 25)),
			expErr: true,
		},
	}

	for msg, spec := range specs {
		spec := spec
		s.Run(msg, func() {
			sdkCtx := s.sdkCtx
			if !spec.srcCtx.IsZero() {
				sdkCtx = spec.srcCtx
			}
			sdkCtx, _ = sdkCtx.CacheContext()
			ctx := types.Context{Context: sdkCtx}

			if spec.doBefore != nil {
				spec.doBefore(ctx)
			}
			_, err := s.msgClient.VotePollAgg(ctx, spec.req)
			if spec.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// query votes by poll
			votesByPollRes, err := s.queryClient.VotesForPollByPoll(ctx, &group.QueryVotesForPollByPollRequest{
				PollId: myPollID,
			})
			s.Require().NoError(err)
			votesByPoll := votesByPollRes.Votes
			s.Require().Equal(len(spec.votes), len(votesByPoll))

			for i, vote := range votesByPoll {
				s.Assert().Equal(spec.req.PollId, vote.PollId)
				s.Assert().Equal(spec.votes[i].Address, vote.Voter)
				s.Assert().Equal(spec.votes[i].Options.Titles, vote.Options.Titles)
				submittedAt, err = gogotypes.TimestampFromProto(&vote.SubmittedAt)
				s.Require().NoError(err)
				s.Assert().Equal(s.blockTime, submittedAt)
			}

			// query votes by voter
			for _, vote := range spec.votes {
				votesByVoterRes, err := s.queryClient.VotesForPollByVoter(ctx, &group.QueryVotesForPollByVoterRequest{
					Voter: vote.Address,
				})
				s.Require().NoError(err)
				votesByVoter := votesByVoterRes.Votes
				s.Require().Equal(1, len(votesByVoter))
				s.Assert().Equal(spec.req.PollId, votesByVoter[0].PollId)
				s.Assert().Equal(vote.Address, votesByVoter[0].Voter)
				s.Assert().Equal(vote.Options.Titles, votesByVoter[0].Options.Titles)
				submittedAt, err = gogotypes.TimestampFromProto(&votesByVoter[0].SubmittedAt)
				s.Require().NoError(err)
				s.Assert().Equal(s.blockTime, submittedAt)
			}

			// and poll is updated
			pollRes, err := s.queryClient.Poll(ctx, &group.QueryPollRequest{
				PollId: spec.req.PollId,
			})
			s.Require().NoError(err)
			poll := pollRes.Poll
			s.Assert().Equal(spec.expVoteState, poll.VoteState)
			s.Assert().Equal(spec.expPollStatus, poll.Status)
		})
	}
}

func createProposal(
	ctx context.Context, s *IntegrationTestSuite, msgs []sdk.Msg,
	proposers []string) uint64 {
	proposalReq := &group.MsgCreateProposal{
		Address:   s.groupAccountAddr.String(),
		Proposers: proposers,
		Metadata:  nil,
	}
	err := proposalReq.SetMsgs(msgs)
	s.Require().NoError(err)

	proposalRes, err := s.msgClient.CreateProposal(ctx, proposalReq)
	s.Require().NoError(err)
	return proposalRes.ProposalId
}

func createProposalAndVote(
	ctx context.Context, s *IntegrationTestSuite, msgs []sdk.Msg,
	proposers []string, choice group.Choice) uint64 {
	s.Require().Greater(len(proposers), 0)
	myProposalID := createProposal(ctx, s, msgs, proposers)

	_, err := s.msgClient.Vote(ctx, &group.MsgVote{
		ProposalId: myProposalID,
		Voter:      proposers[0],
		Choice:     choice,
	})
	s.Require().NoError(err)
	return myProposalID
}

func createGroupAndGroupAccount(
	admin sdk.AccAddress,
	s *IntegrationTestSuite,
) (string, uint64, group.DecisionPolicy, []byte) {
	groupRes, err := s.msgClient.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:    admin.String(),
		Members:  nil,
		Metadata: nil,
	})
	s.Require().NoError(err)

	myGroupID := groupRes.GroupId
	groupAccount := &group.MsgCreateGroupAccount{
		Admin:    admin.String(),
		GroupId:  myGroupID,
		Metadata: nil,
	}

	policy := group.NewThresholdDecisionPolicy(
		"1",
		gogotypes.Duration{Seconds: 1},
	)
	err = groupAccount.SetDecisionPolicy(policy)
	s.Require().NoError(err)

	groupAccountRes, err := s.msgClient.CreateGroupAccount(s.ctx, groupAccount)
	s.Require().NoError(err)

	res, err := s.queryClient.GroupAccountInfo(s.ctx, &group.QueryGroupAccountInfoRequest{Address: groupAccountRes.Address})
	s.Require().NoError(err)

	return groupAccountRes.Address, myGroupID, policy, res.Info.DerivationKey
}

package app

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	capability "github.com/cosmos/cosmos-sdk/x/capability/types"
	crisis "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidence "github.com/cosmos/cosmos-sdk/x/evidence/types"
	feegrant "github.com/cosmos/cosmos-sdk/x/feegrant"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	mint "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	DefaultStakingBondDenom     string = "afet"
	DefaultStakingMaxValidators uint32 = 60

	DefaultMintParams = mint.Params{
		MintDenom:           DefaultStakingBondDenom,
		InflationRateChange: sdk.ZeroDec(),
		InflationMax:        sdk.NewDecWithPrec(3, 2),
		InflationMin:        sdk.NewDecWithPrec(3, 2),
		GoalBonded:          sdk.NewDecWithPrec(67, 2),  // default
		BlocksPerYear:       uint64(60 * 60 * 8766 / 5), // default, assuming 5 second block times
	}

	DefaultGovStartingProposalID uint64 = 1
	DefaultGovDepositParams             = govv1.NewDepositParams(
		sdk.NewCoins(sdk.NewCoin(DefaultStakingBondDenom, sdk.NewInt(2048).Mul(sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))))),
		5*24*time.Hour,
	)
	DefaultGovVotingParams = govv1.NewVotingParams(5 * 24 * time.Hour)
	DefaultGovTallyParams  = govv1.NewTallyParams(
		sdk.NewDecWithPrec(4, 1),   // quorum
		sdk.NewDecWithPrec(5, 1),   // threshold
		sdk.NewDecWithPrec(334, 3), // veto threshold
	)

	DefaultCrisisConstantFee = sdk.NewCoin(DefaultStakingBondDenom, sdk.NewInt(1000))

	DefaultSlashingParams = slashing.Params{
		SignedBlocksWindow:      10000,
		MinSignedPerWindow:      sdk.NewDecWithPrec(5, 2),
		DowntimeJailDuration:    10 * time.Minute,
		SlashFractionDoubleSign: sdk.NewDecWithPrec(1, 1),
		SlashFractionDowntime:   sdk.NewDecWithPrec(1, 4),
	}

	DefaultConsensusBlockParams = tmproto.BlockParams{
		MaxBytes:   300000,
		MaxGas:     3000000,
		TimeIotaMs: 1000,
	}

	DefaultEvidenceParams = tmproto.EvidenceParams{
		MaxAgeNumBlocks: 100000,
		MaxAgeDuration:  48 * time.Hour,
		MaxBytes:        200000,
	}

	DefaultDistributionWithdrawAddrEnabled = false
)

// The genesis state of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	distributionState := distribution.DefaultGenesisState()
	distributionState.Params.WithdrawAddrEnabled = DefaultDistributionWithdrawAddrEnabled

	genState := map[string]json.RawMessage{
		auth.ModuleName:       cdc.MustMarshalJSON(auth.DefaultGenesisState()),
		genutil.ModuleName:    cdc.MustMarshalJSON(genutil.DefaultGenesisState()),
		bank.ModuleName:       cdc.MustMarshalJSON(bank.DefaultGenesisState()),
		capability.ModuleName: cdc.MustMarshalJSON(capability.DefaultGenesis()),
		staking.ModuleName: cdc.MustMarshalJSON(staking.NewGenesisState(staking.Params{
			UnbondingTime:     staking.DefaultUnbondingTime,
			MaxValidators:     DefaultStakingMaxValidators,
			MaxEntries:        staking.DefaultMaxEntries,
			HistoricalEntries: staking.DefaultHistoricalEntries,
			BondDenom:         DefaultStakingBondDenom,
			MinCommissionRate: staking.DefaultMinCommissionRate,
		}, nil, nil)),
		mint.ModuleName: cdc.MustMarshalJSON(mint.NewGenesisState(mint.InitialMinter(
			DefaultMintParams.InflationMax,
		), DefaultMintParams)),
		distribution.ModuleName: cdc.MustMarshalJSON(distributionState),
		gov.ModuleName: cdc.MustMarshalJSON(govv1.NewGenesisState(
			DefaultGovStartingProposalID,
			DefaultGovDepositParams,
			DefaultGovVotingParams,
			DefaultGovTallyParams,
		)),
		crisis.ModuleName: cdc.MustMarshalJSON(crisis.NewGenesisState(
			DefaultCrisisConstantFee,
		)),
		slashing.ModuleName: cdc.MustMarshalJSON(slashing.NewGenesisState(
			DefaultSlashingParams,
			[]slashing.SigningInfo{},
			[]slashing.ValidatorMissedBlocks{},
		)),
		feegrant.ModuleName: cdc.MustMarshalJSON(feegrant.DefaultGenesisState()),
		upgrade.ModuleName:  []byte("{}"),
		evidence.ModuleName: cdc.MustMarshalJSON(evidence.DefaultGenesisState()),
		authz.ModuleName:    cdc.MustMarshalJSON(authz.DefaultGenesisState()),
		group.ModuleName:    cdc.MustMarshalJSON(group.NewGenesisState()),
		vesting.ModuleName:  []byte("{}"),
	}

	return genState
}

func NewDefaultConsensusParams() *tmproto.ConsensusParams {
	consensusParams := tmtypes.DefaultConsensusParams()
	consensusParams.Block = DefaultConsensusBlockParams
	consensusParams.Evidence = DefaultEvidenceParams
	return consensusParams
}

package app

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"io"
	"os"
	"path"
)

const manifestFilenameBase = "upgrade_manifest.json"

type UpgradeManifest struct {
	MovedBalances   []UpgradeBalances `json:"moved_balances,omitempty"`
	InitialBalances []UpgradeBalances `json:"initial_balances,omitempty"`

	// Following 2 hash data members are intentionally without `omitempty` parameter in `json:...` decorator
	GenesisFileSha256           string                `json:"genesis_file_sha256"`
	NetworkConfigFileSha256     string                `json:"network_config_file_sha256"`
	MergeSourceChainID          string                `json:"merge_source_chain_id"`
	DestinationChainID          string                `json:"destination_chain_id"`
	SourceChainBlockHeight      int64                 `json:"source_chain_block_height"`
	DestinationChainBlockHeight int64                 `json:"destination_chain_block_height"`
	GovProposalUpgradePlanName  string                `json:"gov_proposal_upgrade_plan_name"`
	MaxValidatorsChange         *ParamsChange[uint32] `json:"max_validators_change,omitempty"`

	Reconciliation     *UpgradeReconciliation     `json:"reconciliation,omitempty"`
	Contracts          *Contracts                 `json:"contracts,omitempty"`
	IBC                *UpgradeIBCTransfers       `json:"ibc,omitempty"`
	Migration          *UpgradeMigation           `json:"migration,omitempty"`
	MoveGenesisBalance *UpgradeMoveGenesisBalance `json:"move_genesis_balance,omitempty"`
	Delegate           *UpgradeDelegate           `json:"delegate,omitempty"`
	MoveMintedBalance  *UpgradeMoveMintedBalance  `json:"move_minted_balance,omitempty"`
	VestingCollision   *UpgradeVestingCollision   `json:"vesting_collision,omitempty"`
	MoveDelegations    *UpgradeMoveDelegations    `json:"move_delegation,omitempty"`
	CreatedAccounts    *UpgradeCreatedAccounts    `json:"created_accounts,omitempty"`
}

func NewUpgradeManifest() *UpgradeManifest {
	return &UpgradeManifest{}
}

type ParamsChange[T any] struct {
	OriginalVal T `json:"original_val"`
	NewVal      T `json:"new_val"`
}

type Contracts struct {
	StateCleaned   []string                `json:"contracts_state_cleaned,omitempty"`
	AdminUpdated   []ContractValueUpdate   `json:"contracts_admin_updated,omitempty"`
	LabelUpdated   []ContractValueUpdate   `json:"contracts_label_updated,omitempty"`
	VersionUpdated []ContractVersionUpdate `json:"version_updated,omitempty"`
}

type ContractValueUpdate struct {
	Address string `json:"address"`
	From    string `json:"from"`
	To      string `json:"to"`
}

type ContractVersionUpdate struct {
	Address string              `json:"address"`
	From    *CW2ContractVersion `json:"from,omitempty"`
	To      *CW2ContractVersion `json:"to"`
}

type ValueUpdate struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type UpgradeReconciliation struct {
	Transfers     *UpgradeReconciliationTransfers     `json:"transfers,omitempty"`
	ContractState *UpgradeReconciliationContractState `json:"contract_state,omitempty"`
}

type UpgradeReconciliationTransfer struct {
	From    string      `json:"from"`
	EthAddr string      `json:"eth_addr"`
	Amount  types.Coins `json:"amount"`
}

type UpgradeReconciliationTransfers struct {
	Transfers                   []UpgradeReconciliationTransfer `json:"transfers"`
	To                          string                          `json:"to"`
	AggregatedTransferredAmount types.Coins                     `json:"aggregated_transferred_amount"`
	NumberOfTransfers           int                             `json:"number_of_transfers"`
}

type UpgradeReconciliationContractStateBalanceRecord struct {
	EthAddr  string      `json:"eth_addr"`
	Balances types.Coins `json:"balances"`
}

type UpgradeReconciliationContractState struct {
	Balances                 []UpgradeReconciliationContractStateBalanceRecord `json:"balances"`
	AggregatedBalancesAmount types.Coins                                       `json:"aggregated_balances_amount"`
	NumberOfBalanceRecords   int                                               `json:"number_of_balance_records"`
}

type UpgradeIBCTransfer struct {
	From      string      `json:"from"`
	ChannelID string      `json:"channel_id"`
	Amount    types.Coins `json:"amount"`
}

type UpgradeBalanceMovement struct {
	From          string      `json:"from"`
	To            string      `json:"to"`
	SourceBalance types.Coins `json:"source_balance,omitempty"`
	DestBalance   types.Coins `json:"dest_balance,omitempty"`
	Memo          string      `json:"memo,omitempty"`
}

type UpgradeIBCTransfers struct {
	Transfers                   []UpgradeIBCTransfer `json:"transfer"`
	To                          string               `json:"to"`
	AggregatedTransferredAmount types.Coins          `json:"aggregated_transferred_amount"`
	NumberOfTransfers           int                  `json:"number_of_transfers"`
}

type UpgradeMigation struct {
	Migrations               []UpgradeBalanceMovement `json:"migration"`
	AggregatedMigratedAmount types.Coins              `json:"aggregated_migrated_amount"`
	NumberOfMigrations       int                      `json:"number_of_migrations"`
}

type UpgradeDelegationMovements struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Validator string    `json:"validator"`
	Tokens    types.Int `json:"tokens"`
	Memo      string    `json:"memo,omitempty"`
}

type UpgradeMoveGenesisBalance struct {
	Movements             []UpgradeBalanceMovement `json:"movements"`
	AggregatedMovedAmount types.Coins              `json:"aggregated_moved_amount"`
	NumberOfMovements     int                      `json:"number_of_movements"`
}

type UpgradeVestingCollision struct {
	Collisions         []VestingCollision `json:"collisions"`
	NumberOfCollisions int                `json:"number_of_collisions"`
}

type UpgradeDelegate struct {
	Delegations               []UpgradeDelegation `json:"delegation"`
	AggregatedDelegatedAmount *types.Int          `json:"aggregated_delegated_amount"`
	NumberOfDelegations       int                 `json:"number_of_delegations"`
}

type UpgradeMoveDelegations struct {
	Movements         []UpgradeDelegationMovements `json:"delegation_movements"`
	NumberOfMovements int                          `json:"number_of_movements"`
}

type UpgradeDelegation struct {
	OriginalValidator string    `json:"original_validator"`
	NewValidator      string    `json:"new_validator"`
	NewDelegator      string    `json:"new_delegator"`
	OriginalTokens    types.Int `json:"original_tokens"`
	NewTokens         types.Int `json:"new_tokens"`
	NewShares         types.Dec `json:"new_shares"`
}

type VestingCollision struct {
	OriginalAccount      any         `json:"original_account"`
	OriginalAccountFunds types.Coins `json:"original_account_funds"`
	TargetAccount        any         `json:"target_account,omitempty"`
	TargetAccountFunds   types.Coins `json:"target_account_funds,omitempty"`
}

type UpgradeMoveMintedBalance struct {
	Movements []UpgradeBalanceMovement `json:"movements"`
}

type UpgradeCreatedAccounts struct {
	Accounts          []UpgradeAccountCreation `json:"accounts,omitempty"`
	NumberOfCreations int                      `json:"number_of_creations"`
}

type UpgradeAccountCreation struct {
	Address string `json:"address"`
	Reason  string `json:"reason"`
}

type ValidatorBalance struct {
	Validator string      `json:"validator"`
	Balance   types.Coins `json:"balance"`
}

type UpgradeBalances struct {
	Address                      string      `json:"address"`
	BankBalance                  types.Coins `json:"bank_balance"`
	VestedBalance                types.Coins `json:"vested_balance,omitempty"`
	BondedStakingBalancesAggr    types.Coins `json:"bonded_staking_balances_aggr,omitempty"`
	UnbondedStakingBalancesAggr  types.Coins `json:"unbonded_staking_balances_aggr,omitempty"`
	UnbondingStakingBalancesAggr types.Coins `json:"unbonding_staking_balances_aggr,omitempty"`
	DelegatorRewardsAggr         types.Coins `json:"delegator_rewards_aggr,omitempty"`
	ValidatorRewards             types.Coins `json:"validator_rewards,omitempty"`

	BondedStakingBalances    []ValidatorBalance `json:"bonded_staking_balances,omitempty"`
	UnbondedStakingBalances  []ValidatorBalance `json:"unbonded_staking_balances,omitempty"`
	UnbondingStakingBalances []ValidatorBalance `json:"unbonding_staking_balances,omitempty"`
	DelegatorRewards         []ValidatorBalance `json:"delegator_rewards,omitempty"`
}

func getManifestFilePath(app *App, prefix string) (string, error) {
	var upgradeFilePath string
	var err error

	if upgradeFilePath, err = app.UpgradeKeeper.GetUpgradeInfoPath(); err != nil {
		return "", err
	}

	upgradeDir := path.Dir(upgradeFilePath)

	manifestFileName := manifestFilenameBase
	if prefix != "" {
		manifestFileName = fmt.Sprintf("%s_%s", prefix, manifestFilenameBase)
	}

	manifestFilePath := path.Join(upgradeDir, manifestFileName)

	return manifestFilePath, nil
}

func SaveManifest(app *App, manifest *UpgradeManifest, upgradeLabel string) error {

	manifestFilePath, err := getManifestFilePath(app, upgradeLabel)
	if err != nil {
		return err
	}

	return SaveManifestToPath(manifest, manifestFilePath)
}

func SaveManifestToPath(manifest *UpgradeManifest, manifestFilePath string) error {
	var serialisedManifest []byte
	var err error

	if serialisedManifest, err = json.MarshalIndent(manifest, "", "\t"); err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	var f *os.File
	if f, err = os.Create(manifestFilePath); err != nil {
		return fmt.Errorf("failed to create file \"%s\": %w", manifestFilePath, err)
	}
	defer f.Close()

	if _, err = f.Write(serialisedManifest); err != nil {
		return fmt.Errorf("failed to write manifest to the \"%s\" file : %w", manifestFilePath, err)
	}

	return nil
}

func LoadManifestFromPath(manifestFilePath string) (*UpgradeManifest, error) {
	var manifest UpgradeManifest

	// Open the file
	file, err := os.Open(manifestFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file \"%s\": %w", manifestFilePath, err)
	}
	defer file.Close()

	// Read the file contents
	fileContents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file \"%s\": %w", manifestFilePath, err)
	}

	// Unmarshal the JSON content into the manifest
	if err := json.Unmarshal(fileContents, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest from file \"%s\": %w", manifestFilePath, err)
	}

	return &manifest, nil
}

func RegisterVestingCollision(manifest *UpgradeManifest, originalAccount *AccountInfo, targetAccountFunds types.Coins, targetAccount authtypes.AccountI) error {
	if manifest.VestingCollision == nil {
		manifest.VestingCollision = &UpgradeVestingCollision{}
	}

	collision := VestingCollision{
		OriginalAccountFunds: originalAccount.Balance,
		OriginalAccount:      originalAccount.RawAddress,
	}
	if targetAccount != nil {
		res, err := codec.MarshalJSONIndent(authtypes.ModuleCdc.LegacyAmino, targetAccount)
		if err != nil {
			return err
		}

		var resultMap map[string]interface{}

		// Unmarshal the JSON into the map
		err = json.Unmarshal(res, &resultMap)
		if err != nil {
			return err
		}
		collision.TargetAccount = resultMap

		collision.TargetAccountFunds = targetAccountFunds
	}

	manifest.VestingCollision.Collisions = append(manifest.VestingCollision.Collisions, collision)

	manifest.VestingCollision.NumberOfCollisions = len(manifest.VestingCollision.Collisions)
	return nil
}

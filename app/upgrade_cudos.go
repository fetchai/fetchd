package app

import (
	"encoding/base64"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibccore "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/spf13/cast"
)

const (
	Bech32Chars        = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	AddrDataLength     = 32
	WasmAddrDataLength = 52
	MaxAddrDataLength  = 100
	AddrChecksumLength = 6

	AccAddressPrefix  = ""
	ValAddressPrefix  = "valoper"
	ConsAddressPrefix = "valcons"

	NewAddrPrefix = "fetch"

	FlagGenesisTime = "genesis-time"

	ModuleAccount   = "/cosmos.auth.v1beta1.ModuleAccount"
	BaseAccount     = "/cosmos.auth.v1beta1.BaseAccount"
	UnbondedStatus  = "BOND_STATUS_UNBONDED"
	UnbondingStatus = "BOND_STATUS_UNBONDING"
	BondedStatus    = "BOND_STATUS_BONDED"

	// Modules with balance
	BondedPoolAccName    = "bonded_tokens_pool"
	NotBondedPoolAccName = "not_bonded_tokens_pool"
	GravityAccName       = "gravity"

	// Modules without balance
	MintAccName         = "cudoMint"
	GovAccName          = "gov"
	DistributionAccName = "distribution"
	MarketplaceAccName  = "marketplace"
	FeeCollectorAccName = "fee_collector"
)

func convertAddressToFetch(addr string, addressPrefix string) (string, error) {
	_, decodedAddrData, err := bech32.DecodeAndConvert(addr)
	if err != nil {
		return "", err
	}

	newAddress, err := bech32.ConvertAndEncode(NewAddrPrefix+addressPrefix, decodedAddrData)
	if err != nil {
		return "", err
	}

	err = sdk.VerifyAddressFormat(decodedAddrData)
	if err != nil {
		return "", err
	}

	return newAddress, nil
}
func convertAddressPrefix(addr string, newPrefix string) (string, error) {
	_, decodedAddrData, err := bech32.DecodeAndConvert(addr)
	if err != nil {
		return "", err
	}

	newAddress, err := bech32.ConvertAndEncode(newPrefix, decodedAddrData)
	if err != nil {
		return "", err
	}

	return newAddress, nil
}

func convertAddressToRaw(addr string, networkInfo NetworkConfig) (sdk.AccAddress, error) {
	prefix, decodedAddrData, err := bech32.DecodeAndConvert(addr)

	if prefix != networkInfo.oldAddrPrefix {
		return nil, fmt.Errorf("Unknown prefix: %s", prefix)
	}

	if err != nil {
		return nil, err
	}

	return decodedAddrData, nil
}

type AccountType string

const (
	BaseAccountType     AccountType = "base_acc"
	ModuleAccountType   AccountType = "module_acc"
	ContractAccountType AccountType = "contract_acc"
	IBCAccountType      AccountType = "IBC_acc"
)

type GenesisData struct {
	totalSupply sdk.Coins

	accounts    OrderedMap[string, AccountInfo]
	contracts   OrderedMap[string, ContractInfo]
	IBCAccounts OrderedMap[string, IBCInfo]

	validators           OrderedMap[string, ValidatorInfo]
	bondedPoolAddress    string
	notBondedPoolAddress string

	distributionInfo DistributionInfo

	gravityModuleAccountAddress string
}

func parseGenesisData(jsonData map[string]interface{}, networkInfo NetworkConfig) (*GenesisData, error) {
	genesisData := GenesisData{}

	totalSupply, err := parseTotalSupply(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to total supply: %w", err)
	}
	genesisData.totalSupply = totalSupply

	contracts, err := parseWasmContracts(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to get contracts: %w", err)
	}
	genesisData.contracts = *contracts

	IBCAccounts, err := parseIBCAccounts(jsonData, networkInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get ibc accounts: %w", err)
	}
	genesisData.IBCAccounts = *IBCAccounts

	// Get all accounts and balances into map
	accounts, err := parseGenesisAccounts(jsonData, contracts, IBCAccounts, networkInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts map: %w", err)
	}
	genesisData.accounts = *accounts

	// Staking module
	bondedPoolAddress, err := GetAddressByName(accounts, BondedPoolAccName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bonded pool account %w", err)
	}
	genesisData.bondedPoolAddress = bondedPoolAddress

	notBondedPoolAddress, err := GetAddressByName(accounts, NotBondedPoolAccName)
	if err != nil {
		return nil, fmt.Errorf("failed to get not-bonded pool account %w", err)
	}
	genesisData.notBondedPoolAddress = notBondedPoolAddress

	validators, err := parseGenesisValidators(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to get validators map: %w", err)
	}
	genesisData.validators = *validators

	distributionInfo, err := parseGenesisDistribution(jsonData, accounts)
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution module map: %w", err)
	}
	genesisData.distributionInfo = *distributionInfo

	gravityModuleAccountAddress, err := GetAddressByName(accounts, GravityAccName)
	if err != nil {
		return nil, fmt.Errorf("failed to get gravity module account: %w", err)
	}
	genesisData.gravityModuleAccountAddress = gravityModuleAccountAddress

	return &genesisData, nil
}

type AccountInfo struct {
	name        string
	pubkey      cryptotypes.PubKey
	balance     sdk.Coins
	migrated    bool
	accountType AccountType
	rawAddress  sdk.AccAddress
}

func parseGenesisAccounts(jsonData map[string]interface{}, contractAccountMap *OrderedMap[string, ContractInfo], IBCAccountsMap *OrderedMap[string, IBCInfo], networkInfo NetworkConfig) (*OrderedMap[string, AccountInfo], error) {
	var err error

	// Map to verify that account exists in auth module
	auth := jsonData[authtypes.ModuleName].(map[string]interface{})
	accounts := auth["accounts"].([]interface{})

	accountMap := NewOrderedMap[string, AccountInfo]()

	for _, acc := range accounts {
		accMap := acc.(map[string]interface{})
		accType := accMap["@type"]

		var name string
		accData := acc
		if accType == ModuleAccount {
			accData = accMap["base_account"]
			name = accMap["name"].(string)

		}

		accDataMap := accData.(map[string]interface{})
		addr := accDataMap["address"].(string)

		// Check that public keys are the same
		var AccPubKey cryptotypes.PubKey
		if pk, ok := accDataMap["pub_key"]; ok {
			if pk != nil {
				AccPubKey, err = decodePubKeyFromMap(pk.(map[string]interface{}))
				if err != nil {
					return nil, err
				}
			}
		}

		var accountType AccountType
		if accType == ModuleAccount {
			accountType = ModuleAccountType
		} else if _, exists := contractAccountMap.Get(addr); exists {
			accountType = ContractAccountType
		} else if _, exists := IBCAccountsMap.Get(addr); exists {
			accountType = IBCAccountType
		} else {
			accountType = BaseAccountType
		}

		// Get raw address
		accRawAddr, err := convertAddressToRaw(addr, networkInfo)
		if err != nil {
			return nil, err
		}

		accountMap.SetNew(addr, AccountInfo{name: name, pubkey: AccPubKey, balance: sdk.NewCoins(), migrated: false, accountType: accountType, rawAddress: accRawAddr})
	}

	// Add balances to accounts map
	err = fillGenesisBalancesToAccountsMap(jsonData, accountMap)
	if err != nil {
		return nil, err
	}

	return accountMap, nil
}

func getConsAddressFromValidator(validatorData map[string]interface{}) (sdk.ConsAddress, error) {
	consensusPubkey := validatorData["consensus_pubkey"].(map[string]interface{})
	decodedConsensusPubkey, err := decodePubKeyFromMap(consensusPubkey)
	if err != nil {
		return nil, err
	}
	return sdk.ConsAddress(decodedConsensusPubkey.Address()), nil
}

type DelegationInfo struct {
	delegatorAddress string
	shares           sdk.Dec
}

type UnbondingDelegationInfo struct {
	delegatorAddress string
	entries          []UnbondingDelegationEntry
}

type UnbondingDelegationEntry struct {
	balance        sdk.Int
	initialBalance sdk.Int
	creationHeight uint64
	completionTime string
}

type ValidatorInfo struct {
	stake                sdk.Int
	shares               sdk.Dec
	status               string
	operatorAddress      string
	consensusPubkey      cryptotypes.PubKey
	delegations          OrderedMap[string, DelegationInfo]
	unbondingDelegations OrderedMap[string, UnbondingDelegationInfo]
}

func (v ValidatorInfo) TokensFromShares(shares sdk.Dec) sdk.Dec {
	return (shares.MulInt(v.stake)).Quo(v.shares)
}

func parseGenesisValidators(jsonData map[string]interface{}) (*OrderedMap[string, ValidatorInfo], error) {
	// Validator pubkey hex -> ValidatorInfo
	validatorInfoMap := NewOrderedMap[string, ValidatorInfo]()

	staking := jsonData[stakingtypes.ModuleName].(map[string]interface{})
	validators := staking["validators"].([]interface{})

	for _, validator := range validators {

		validatorMap := validator.(map[string]interface{})
		tokens := validatorMap["tokens"].(string)
		operatorAddress := validator.(map[string]interface{})["operator_address"].(string)

		consensusPubkey := validator.(map[string]interface{})["consensus_pubkey"].(map[string]interface{})
		decodedConsensusPubkey, err := decodePubKeyFromMap(consensusPubkey)
		if err != nil {
			return nil, err
		}

		// Convert amount to big.Int
		tokensInt, ok := sdk.NewIntFromString(tokens)
		if !ok {
			panic("Failed to convert validator tokens to big.Int")
		}

		status := validatorMap["status"].(string)

		validatorShares := validatorMap["delegator_shares"].(string)
		validatorSharesDec, err := sdk.NewDecFromStr(validatorShares)
		if err != nil {
			return nil, err
		}

		validatorInfoMap.SetNew(operatorAddress, ValidatorInfo{
			stake:                tokensInt,
			shares:               validatorSharesDec,
			status:               status,
			operatorAddress:      operatorAddress,
			consensusPubkey:      decodedConsensusPubkey,
			delegations:          *NewOrderedMap[string, DelegationInfo](),
			unbondingDelegations: *NewOrderedMap[string, UnbondingDelegationInfo](),
		})

	}

	// Map of delegatorAddress -> validatorPubkey -> sdk.coins balance
	delegations := staking["delegations"].([]interface{})
	for _, delegation := range delegations {
		delegationMap := delegation.(map[string]interface{})
		delegatorAddress := delegationMap["delegator_address"].(string)
		validatorAddress := delegationMap["validator_address"].(string)

		delegatorSharesDec, err := sdk.NewDecFromStr(delegationMap["shares"].(string))
		if err != nil {
			return nil, err
		}

		validator := validatorInfoMap.MustGet(validatorAddress)
		validator.delegations.SetNew(delegatorAddress, DelegationInfo{delegatorAddress: delegatorAddress, shares: delegatorSharesDec})
	}

	unbondingDelegations := staking["unbonding_delegations"].([]interface{})
	for _, unbondingDelegation := range unbondingDelegations {
		unbondingDelegationMap := unbondingDelegation.(map[string]interface{})
		delegatorAddress := unbondingDelegationMap["delegator_address"].(string)
		validatorAddress := unbondingDelegationMap["validator_address"].(string)

		entriesMap := unbondingDelegationMap["entries"].([]interface{})

		var unbondingDelegationEntries []UnbondingDelegationEntry

		for _, entry := range entriesMap {
			etnryMap := entry.(map[string]interface{})
			balance, ok := sdk.NewIntFromString(etnryMap["balance"].(string))
			if !ok {
				return nil, fmt.Errorf("failed to convert unbonding delegation balance to int")
			}

			initialBalance, ok := sdk.NewIntFromString(etnryMap["initial_balance"].(string))
			if !ok {
				return nil, fmt.Errorf("failed to convert unbonding delegation initial balance to int")
			}

			creationHeight := cast.ToUint64(etnryMap["creation_height"].(string))

			completionTime := etnryMap["completion_time"].(string)

			unbondingDelegationEntries = append(unbondingDelegationEntries, UnbondingDelegationEntry{balance: balance, initialBalance: initialBalance, creationHeight: creationHeight, completionTime: completionTime})
		}

		validator := validatorInfoMap.MustGet(validatorAddress)
		validator.unbondingDelegations.SetNew(delegatorAddress, UnbondingDelegationInfo{delegatorAddress: delegatorAddress, entries: unbondingDelegationEntries})
	}

	return validatorInfoMap, nil
}

func withdrawGenesisStakingDelegations(genesisData *GenesisData, networkInfo NetworkConfig, manifest *UpgradeManifest) (*OrderedMap[string, OrderedMap[string, sdk.Coins]], error) {
	// Handle delegations
	delegatedBalanceMap := NewOrderedMap[string, OrderedMap[string, sdk.Coins]]()
	for _, validatorOperatorAddress := range *genesisData.validators.Keys() {
		validator := genesisData.validators.MustGet(validatorOperatorAddress)
		for _, delegatorAddress := range *validator.delegations.Keys() {
			delegation := validator.delegations.MustGet(delegatorAddress)
			resolvedDelegatorAddress := resolveIfContractAddress(delegatorAddress, genesisData)

			currentValidatorInfo := genesisData.validators.MustGet(validatorOperatorAddress)
			delegatorTokens := currentValidatorInfo.TokensFromShares(delegation.shares).TruncateInt()

			// Move balance to delegator address
			delegatorBalance := sdk.NewCoins(sdk.NewCoin(networkInfo.originalDenom, delegatorTokens))

			if delegatorTokens.IsZero() {
				// This usually happens when number of shares is less than 1
				continue
			}

			// Subtract balance from bonded or not-bonded pool
			if currentValidatorInfo.status == BondedStatus {

				// Store delegation to delegated map
				if _, exists := delegatedBalanceMap.Get(resolvedDelegatorAddress); !exists {
					delegatedBalanceMap.Set(resolvedDelegatorAddress, *NewOrderedMap[string, sdk.Coins]())
				}

				resolvedDelegatorMap := delegatedBalanceMap.MustGet(resolvedDelegatorAddress)

				if _, exists := resolvedDelegatorMap.Get(validatorOperatorAddress); !exists {
					resolvedDelegatorMap.Set(validatorOperatorAddress, sdk.NewCoins())
				}
				resolvedDelegator := resolvedDelegatorMap.MustGet(validatorOperatorAddress)

				resolvedDelegatorMap.Set(validatorOperatorAddress, resolvedDelegator.Add(delegatorBalance...))

				delegatedBalanceMap.Set(resolvedDelegatorAddress, *resolvedDelegatorMap)

				// Move balance from bonded pool to delegator
				err := MoveGenesisBalance(genesisData, genesisData.bondedPoolAddress, resolvedDelegatorAddress, delegatorBalance, manifest)
				if err != nil {
					return nil, err
				}

			} else {
				// Delegations to unbonded/jailed/tombstoned validators are not re-delegated

				// Move balance from not-bonded pool to delegator
				err := MoveGenesisBalance(genesisData, genesisData.notBondedPoolAddress, resolvedDelegatorAddress, delegatorBalance, manifest)
				if err != nil {
					return nil, err
				}
			}

		}

		// Handle unbonding delegations
		for _, delegatorAddress := range *validator.unbondingDelegations.Keys() {
			unbondingDelegation := validator.unbondingDelegations.MustGet(delegatorAddress)
			resolvedDelegatorAddress := resolveIfContractAddress(delegatorAddress, genesisData)

			for _, entry := range unbondingDelegation.entries {
				unbondingDelegationBalance := sdk.NewCoins(sdk.NewCoin(networkInfo.originalDenom, entry.balance))

				// Move unbonding balance from not-bonded pool to delegator address
				err := MoveGenesisBalance(genesisData, genesisData.notBondedPoolAddress, resolvedDelegatorAddress, unbondingDelegationBalance, manifest)
				if err != nil {
					return nil, err
				}

			}
		}
	}

	// Handle remaining pool balances

	// Handle remaining bonded pool balance
	bondedPool := genesisData.accounts.MustGet(genesisData.bondedPoolAddress)

	// TODO: Write to manifest?
	err := checkTolerance(bondedPool.balance, maxToleratedRemainingStakingBalance)
	if err != nil {
		return nil, fmt.Errorf("Remaining bonded pool balance %s is too high", bondedPool.balance.String())
	}

	println("Remaining bonded pool balance: ", bondedPool.balance.String())
	err = MoveGenesisBalance(genesisData, genesisData.bondedPoolAddress, networkInfo.remainingStakingBalanceAddr, bondedPool.balance, manifest)
	if err != nil {
		return nil, err
	}

	// Handle remaining not-bonded pool balance
	notBondedPool := genesisData.accounts.MustGet(genesisData.notBondedPoolAddress)

	// TODO: Write to manifest?
	err = checkTolerance(notBondedPool.balance, maxToleratedRemainingStakingBalance)
	if err != nil {
		return nil, fmt.Errorf("Remaining not-bonded pool balance %s is too high", notBondedPool.balance.String())
	}

	println("Remaining not-bonded pool balance: ", notBondedPool.balance.String())
	err = MoveGenesisBalance(genesisData, genesisData.notBondedPoolAddress, networkInfo.remainingStakingBalanceAddr, notBondedPool.balance, manifest)
	if err != nil {
		return nil, err
	}

	return delegatedBalanceMap, nil
}

func resolveDestinationValidator(ctx sdk.Context, app *App, operatorAddress string, networkInfo NetworkConfig) (*stakingtypes.Validator, error) {

	/*
		vals := app.StakingKeeper.GetValidators(ctx, 1234)
		for _, val := range vals {
			if val.Status.String() == BondedStatus {
				println(val.GetOperator().String())
			}
		}
	*/

	if targetOperatorStringAddress, exists := networkInfo.validatorsMap[operatorAddress]; exists {
		targetOperatorAddress, err := sdk.ValAddressFromBech32(targetOperatorStringAddress)
		if err != nil {
			return nil, err
		}

		if targetValidator, found := app.StakingKeeper.GetValidator(ctx, targetOperatorAddress); found {
			if targetValidator.Status.String() == BondedStatus {
				return &targetValidator, nil
			}
		}

	}

	for _, targetOperatorStringAddress := range networkInfo.backupValidators {
		targetOperatorAddress, err := sdk.ValAddressFromBech32(targetOperatorStringAddress)
		if err != nil {
			return nil, err
		}

		if targetValidator, found := app.StakingKeeper.GetValidator(ctx, targetOperatorAddress); found {
			if targetValidator.Status.String() == BondedStatus {
				return &targetValidator, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to resolve validator")
}

func getIntAmountFromCoins(balance sdk.Coins, expectedDenom string) (*sdk.Int, error) {

	for _, coin := range balance {
		if coin.Denom == expectedDenom {
			return &coin.Amount, nil
		}
	}
	return nil, fmt.Errorf("denom %s not found in balance", expectedDenom)
}

func createDelegation(ctx sdk.Context, app *App, OriginalDelegator string, NewDelegatorRawAddr sdk.AccAddress, validator stakingtypes.Validator, tokensToDelegate sdk.Int, manifest *UpgradeManifest) error {

	newShares, err := app.StakingKeeper.Delegate(ctx, NewDelegatorRawAddr, tokensToDelegate, stakingtypes.Unbonded, validator, true)
	if err != nil {
		return err
	}

	if manifest.Delegate == nil {
		manifest.Delegate = &UpgradeDelegate{}
	}

	delegation := UpgradeDelegation{
		Delegator:         NewDelegatorRawAddr.String(),
		Validator:         validator.OperatorAddress,
		Tokens:            tokensToDelegate,
		NewShares:         newShares,
		OriginalDelegator: OriginalDelegator,
	}
	manifest.Delegate.Delegations = append(manifest.Delegate.Delegations, delegation)

	if manifest.Delegate.AggregatedDelegatedAmount == nil {
		manifest.Delegate.AggregatedDelegatedAmount = &tokensToDelegate
	} else {
		*manifest.Delegate.AggregatedDelegatedAmount = manifest.Delegate.AggregatedDelegatedAmount.Add(tokensToDelegate)
	}

	manifest.Delegate.NumberOfDelegations += 1

	return nil
}

func fundCommunityPool(ctx sdk.Context, app *App, genesisData *GenesisData, networkInfo NetworkConfig, manifest *UpgradeManifest) error {
	// Fund community pool
	communityPoolBalance, _ := genesisData.distributionInfo.feePool.communityPool.TruncateDecimal()
	convertedCommunityPoolBalance, err := convertBalance(communityPoolBalance, networkInfo)
	if err != nil {
		return err
	}

	communityPoolSourceAccountRawAddress := genesisData.accounts.MustGet(networkInfo.remainingDistributionBalanceAddr).rawAddress
	err = app.DistrKeeper.FundCommunityPool(ctx, convertedCommunityPoolBalance, communityPoolSourceAccountRawAddress)
	if err != nil {
		return err
	}
	// TODO: Update manifest

	return nil
}

func createGenesisDelegations(ctx sdk.Context, app *App, delegatedBalanceMap *OrderedMap[string, OrderedMap[string, sdk.Coins]], networkInfo NetworkConfig, manifest *UpgradeManifest) error {

	for _, delegatorAddr := range *delegatedBalanceMap.Keys() {
		delegatorAddrMap := delegatedBalanceMap.MustGet(delegatorAddr)

		for _, validatorOperatorStringAddr := range *delegatorAddrMap.Keys() {
			delegatedBalance := delegatorAddrMap.MustGet(validatorOperatorStringAddr)

			destValidator, err := resolveDestinationValidator(ctx, app, validatorOperatorStringAddr, networkInfo)
			if err != nil {
				return err
			}

			// Get int amount in native tokens
			convertedBalance, err := convertBalance(*delegatedBalance, networkInfo)
			if err != nil {
				return err
			}

			if convertedBalance.Empty() {
				// Very small balance gets truncated to 0 during conversion
				continue
			}

			tokensToDelegate, err := getIntAmountFromCoins(convertedBalance, networkInfo.stakingDenom)
			if err != nil {
				return err
			}
			NewDelegatorRawAddr, err := convertAddressToRaw(delegatorAddr, networkInfo)
			if err != nil {
				return err
			}

			// Create delegation
			err = createDelegation(ctx, app, validatorOperatorStringAddr, NewDelegatorRawAddr, *destValidator, *tokensToDelegate, manifest)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getCoinsFromInterfaceSlice(coins []interface{}) (sdk.Coins, error) {
	var resBalance sdk.Coins
	for _, coin := range coins {

		amount := coin.(map[string]interface{})["amount"].(string)

		denom := coin.(map[string]interface{})["denom"].(string)

		sdkAmount, ok := sdk.NewIntFromString(amount)
		if !ok {
			return nil, fmt.Errorf("Failed to convert amount to sdk.Int")
		}

		sdkCoin := sdk.NewCoin(denom, sdkAmount)
		resBalance = resBalance.Add(sdkCoin)

	}

	return resBalance, nil
}

func getDecCoinsFromInterfaceSlice(coins []interface{}) (sdk.DecCoins, error) {
	var resBalance sdk.DecCoins
	for _, coin := range coins {

		amount := coin.(map[string]interface{})["amount"].(string)

		denom := coin.(map[string]interface{})["denom"].(string)

		sdkAmount, err := sdk.NewDecFromStr(amount)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert amount to sdk.Int")
		}

		sdkCoin := sdk.NewDecCoinFromDec(denom, sdkAmount)
		resBalance = resBalance.Add(sdkCoin)

	}

	return resBalance, nil
}

func getInterfaceSliceFromCoins(coins sdk.Coins) []interface{} {
	var balance []interface{}
	for _, coin := range coins {
		balance = append(balance, map[string]interface{}{
			"denom":  coin.Denom,
			"amount": coin.Amount.String(),
		})
	}
	return balance
}

func withdrawGenesisContractBalances(genesisData *GenesisData, manifest *UpgradeManifest) error {

	for _, contractAddress := range *genesisData.contracts.Keys() {
		resolvedAddress := resolveIfContractAddress(contractAddress, genesisData)
		if resolvedAddress == contractAddress {
			return fmt.Errorf("Failed to resolve contract admin/owner for contract %s", contractAddress)
		}

		contractBalance, contractBalancePresent := genesisData.accounts.Get(contractAddress)
		if contractBalancePresent {
			err := MoveGenesisBalance(genesisData, contractAddress, resolvedAddress, contractBalance.balance, manifest)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func convertBalance(balance sdk.Coins, networkInfo NetworkConfig) (sdk.Coins, error) {
	var resBalance sdk.Coins

	for _, coin := range balance {
		if conversionConstant, ok := networkInfo.balanceConversionConstants[coin.Denom]; ok {
			newAmount := coin.Amount.ToDec().Mul(conversionConstant).TruncateInt()
			sdkCoin := sdk.NewCoin(networkInfo.convertedDenom, newAmount)
			resBalance = resBalance.Add(sdkCoin)
		} else {
			//println("Unknown denom: ", coin.Denom)
			// Ignore unlisted tokens
			continue
			/*
				// Just add without conversion
				newAmount, ok := sdk.NewIntFromString(amount)
				if !ok {
					panic("Failed to convert amount to big.Int")
				}

				sdkCoin := sdk.NewCoin(denom, newAmount)

				resBalance = resBalance.Add(sdkCoin)
			*/
		}
	}

	return resBalance, nil
}

func fillGenesisBalancesToAccountsMap(jsonData map[string]interface{}, genesisAccountsMap *OrderedMap[string, AccountInfo]) error {
	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	balances := bank["balances"].([]interface{})

	for _, balance := range balances {

		addr := balance.(map[string]interface{})["address"]
		if addr == nil {
			panic("Failed to get address")
		}
		addrStr := addr.(string)

		// Verify that account exists in auth module
		if _, exists := genesisAccountsMap.Get(addrStr); !exists {
			panic("Account not registered in auth module")
		}

		coins := balance.(map[string]interface{})["coins"]

		sdkBalance, err := getCoinsFromInterfaceSlice(coins.([]interface{}))
		if err != nil {
			panic(err)
		}

		if !sdkBalance.Empty() {
			accountInfoEntry := genesisAccountsMap.MustGet(addrStr)
			accountInfoEntry.balance = sdkBalance

			genesisAccountsMap.Set(addrStr, *accountInfoEntry)
		}

	}
	return nil
}

func GenesisUpgradeWithdrawIBCChannelsBalances(genesisData *GenesisData, networkInfo NetworkConfig, manifest *UpgradeManifest) error {
	if networkInfo.ibcTargetAddr == "" {

		panic("No IBC withdrawal address set")
	}

	ibcWithdrawalAddress := networkInfo.ibcTargetAddr

	manifest.IBC = &UpgradeIBCTransfers{
		To: ibcWithdrawalAddress,
	}

	for _, IBCaccountAddress := range *genesisData.IBCAccounts.Keys() {

		IBCaccount, IBCAccountExists := genesisData.accounts.Get(IBCaccountAddress)
		IBCinfo := genesisData.IBCAccounts.MustGet(IBCaccountAddress)

		var channelBalance sdk.Coins
		if IBCAccountExists {

			channelBalance = IBCaccount.balance
			err := MoveGenesisBalance(genesisData, IBCaccountAddress, ibcWithdrawalAddress, channelBalance, manifest)
			if err != nil {
				return err
			}
		}

		manifest.IBC.Transfers = append(manifest.IBC.Transfers, UpgradeIBCTransfer{From: IBCaccountAddress, ChannelID: fmt.Sprintf("%s/%s", IBCinfo.portId, IBCinfo.channelId), Amount: channelBalance})
		manifest.IBC.AggregatedTransferredAmount = manifest.IBC.AggregatedTransferredAmount.Add(channelBalance...)
		manifest.IBC.NumberOfTransfers += 1

	}
	return nil
}

type IBCInfo struct {
	channelId string
	portId    string
}

func parseIBCAccounts(jsonData map[string]interface{}, networkInfo NetworkConfig) (*OrderedMap[string, IBCInfo], error) {
	ibcAccountMap := NewOrderedMap[string, IBCInfo]()

	ibc, ok := jsonData[ibccore.ModuleName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("IBC module data not found in genesis")
	}

	channelGenesis, ok := ibc["channel_genesis"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("channel genesis data not found in IBC module")
	}

	ibcChannels, ok := channelGenesis["channels"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("channels data not found in channel genesis")
	}

	for _, channel := range ibcChannels {
		channelMap, ok := channel.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid channel format in genesis")
		}

		channelId, ok := channelMap["channel_id"].(string)
		if !ok {
			return nil, fmt.Errorf("channel_id not found or invalid in channel")
		}

		portId, ok := channelMap["port_id"].(string)
		if !ok {
			return nil, fmt.Errorf("port_id not found or invalid in channel")
		}

		rawAddr := ibctransfertypes.GetEscrowAddress(portId, channelId)
		channelAddr, err := sdk.Bech32ifyAddressBytes(networkInfo.oldAddrPrefix+AccAddressPrefix, rawAddr)
		if err != nil {
			return nil, err
		}

		ibcAccountMap.Set(channelAddr, IBCInfo{channelId: channelId, portId: portId})
	}

	return ibcAccountMap, nil
}

type ContractInfo struct {
	Admin   string
	Creator string
}

func parseWasmContracts(jsonData map[string]interface{}) (*OrderedMap[string, ContractInfo], error) {
	contractAccountMap := NewOrderedMap[string, ContractInfo]()

	// Navigate to the "wasm" module
	wasm, ok := jsonData["wasm"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("wasm module data not found in genesis")
	}

	// Navigate to the "contracts" section
	contracts, ok := wasm["contracts"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("contracts data not found in wasm module")
	}

	// Iterate over each contract to get the "contract_address"
	for _, contract := range contracts {
		contractMap, ok := contract.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid contract format in genesis")
		}

		contractAddr, ok := contractMap["contract_address"].(string)
		if !ok {
			return nil, fmt.Errorf("contract_address not found or invalid in contract")
		}

		contractInfo, ok := contractMap["contract_info"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("contract_info not found or invalid in contract")
		}

		admin := contractInfo["admin"].(string)
		creator := contractInfo["creator"].(string)

		contractAccountMap.Set(contractAddr, ContractInfo{Admin: admin, Creator: creator})
	}

	return contractAccountMap, nil
}

func resolveIfContractAddress(address string, genesisData *GenesisData) string {
	if contractInfo, exists := genesisData.contracts.Get(address); exists {
		if contractInfo.Admin != "" {
			return resolveIfContractAddress(contractInfo.Admin, genesisData)
		} else if contractInfo.Creator != "" {
			return resolveIfContractAddress(contractInfo.Creator, genesisData)
		} else {
			panic(fmt.Errorf("contract %s has no admin nor creator", address))
		}
	} else {
		return address
	}
}

func decodePubKeyFromMap(pubKeyMap map[string]interface{}) (cryptotypes.PubKey, error) {
	keyType, ok := pubKeyMap["@type"].(string)
	if !ok {
		return nil, fmt.Errorf("@type field not found or is not a string in pubKeyMap")
	}

	switch keyType {
	case "/cosmos.crypto.secp256k1.PubKey":
		keyStr, ok := pubKeyMap["key"].(string)
		if !ok {
			return nil, fmt.Errorf("key field not found or is not a string in pubKeyMap")
		}

		keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 key: %w", err)
		}

		// Ensure the byte slice is the correct length for a secp256k1 public key
		if len(keyBytes) != secp256k1.PubKeySize {
			return nil, fmt.Errorf("invalid pubkey length: got %d, expected %d", len(keyBytes), secp256k1.PubKeySize)
		}

		pubKey := secp256k1.PubKey{
			Key: keyBytes,
		}
		return &pubKey, nil

	case "/cosmos.crypto.ed25519.PubKey":
		keyStr, ok := pubKeyMap["key"].(string)
		if !ok {
			return nil, fmt.Errorf("key field not found or is not a string in pubKeyMap")
		}

		keyBytes, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 key: %w", err)
		}

		// Ensure the byte slice is the correct length for an ed25519 public key
		if len(keyBytes) != ed25519.PubKeySize {
			return nil, fmt.Errorf("invalid pubkey length: got %d, expected %d", len(keyBytes), ed25519.PubKeySize)
		}

		pubKey := ed25519.PubKey{
			Key: keyBytes,
		}
		return &pubKey, nil

	case "/cosmos.crypto.multisig.LegacyAminoPubKey":
		threshold, ok := pubKeyMap["threshold"].(float64) // JSON numbers are float64
		if !ok {
			return nil, fmt.Errorf("threshold field not found or is not a number in pubKeyMap")
		}

		pubKeysInterface, ok := pubKeyMap["public_keys"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("public_keys field not found or is not an array in pubKeyMap")
		}

		var pubKeys []cryptotypes.PubKey
		for _, pubKeyInterface := range pubKeysInterface {
			pubKeyMap, ok := pubKeyInterface.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("public key entry is not a valid map")
			}

			pubKey, err := decodePubKeyFromMap(pubKeyMap)
			if err != nil {
				return nil, fmt.Errorf("failed to decode public key: %w", err)
			}

			pubKeys = append(pubKeys, pubKey)
		}

		legacyAminoPubKey := multisig.NewLegacyAminoPubKey(int(threshold), pubKeys)
		return legacyAminoPubKey, nil

	default:
		return nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

func getNewBaseAccount(ctx sdk.Context, app *App, accountInfo AccountInfo) (*authtypes.BaseAccount, error) {
	// Create new account
	newAccNumber := app.AccountKeeper.GetNextAccountNumber(ctx)
	newBaseAccount := authtypes.NewBaseAccount(accountInfo.rawAddress, accountInfo.pubkey, newAccNumber, 0)
	return newBaseAccount, nil
}

func createNewVestingAccountFromBaseAccount(ctx sdk.Context, app *App, account *authtypes.BaseAccount, vestedCoins sdk.Coins, startTime int64, endTime int64) error {
	newBaseVestingAcc := authvesting.NewBaseVestingAccount(account, vestedCoins, endTime)
	newContinuousVestingAcc := authvesting.NewContinuousVestingAccountRaw(newBaseVestingAcc, startTime)

	app.AccountKeeper.SetAccount(ctx, newContinuousVestingAcc)

	return nil
}

func createNewNormalAccountFromBaseAccount(ctx sdk.Context, app *App, account *authtypes.BaseAccount) error {
	app.AccountKeeper.SetAccount(ctx, account)

	return nil
}

func mintToAccount(ctx sdk.Context, app *App, fromAddress string, toAddress sdk.AccAddress, newCoins sdk.Coins, manifest *UpgradeManifest) error {

	err := app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, toAddress, newCoins)
	if err != nil {
		return err
	}

	if manifest.Minting == nil {
		manifest.Minting = &UpgradeMinting{}
	}

	mint := UpgradeMint{
		From:   fromAddress,
		To:     toAddress.String(),
		Amount: newCoins,
	}
	manifest.Minting.Mints = append(manifest.Minting.Mints, mint)

	manifest.Minting.AggregatedMintedAmount = manifest.Minting.AggregatedMintedAmount.Add(newCoins...)
	manifest.Minting.NumberOfMints += 1

	return nil
}

func MarkAccountAsMigrated(genesisData *GenesisData, accountAddress string) error {
	AccountInfoRecord, exists := genesisData.accounts.Get(accountAddress)
	if !exists {
		return fmt.Errorf("Genesis account %s not found", accountAddress)
	}

	if AccountInfoRecord.migrated {
		return fmt.Errorf("Genesis account %s already migrated", accountAddress)
	}

	AccountInfoRecord.migrated = true

	genesisData.accounts.Set(accountAddress, *AccountInfoRecord)

	return nil
}

func MoveGenesisBalance(genesisData *GenesisData, fromAddress, toAddress string, amount sdk.Coins, manifest *UpgradeManifest) error {
	// Check if fromAddress exists
	if _, ok := genesisData.accounts.Get(fromAddress); !ok {
		return fmt.Errorf("fromAddress %s does not exist in genesis balances", fromAddress)
	}

	if _, ok := genesisData.accounts.Get(toAddress); !ok {
		return fmt.Errorf("toAddress %s does not exist in genesis balances", toAddress)
	}

	if manifest.MoveBalance == nil {
		manifest.MoveBalance = &UpgradeMoveBalance{}
	}

	movement := UpgradeBalanceMovement{
		From:   fromAddress,
		To:     toAddress,
		Amount: amount,
	}
	manifest.MoveBalance.Movements = append(manifest.MoveBalance.Movements, movement)

	manifest.MoveBalance.AggregatedMovedAmount = manifest.MoveBalance.AggregatedMovedAmount.Add(amount...)
	manifest.MoveBalance.NumberOfMovements += 1

	if toAcc := genesisData.accounts.MustGet(toAddress); toAcc.migrated {
		return fmt.Errorf("Genesis account %s already migrated", toAddress)
	}
	if fromAcc := genesisData.accounts.MustGet(fromAddress); fromAcc.migrated {
		return fmt.Errorf("Genesis account %s already migrated", fromAddress)
	}

	genesisToBalance := genesisData.accounts.MustGet(toAddress)
	genesisFromBalance := genesisData.accounts.MustGet(fromAddress)

	genesisToBalance.balance = genesisToBalance.balance.Add(amount...)
	genesisFromBalance.balance = genesisFromBalance.balance.Sub(amount)

	genesisData.accounts.Set(toAddress, *genesisToBalance)
	genesisData.accounts.Set(fromAddress, *genesisFromBalance)

	return nil
}

func GetAddressByName(genesisAccounts *OrderedMap[string, AccountInfo], name string) (string, error) {

	for _, accAddress := range *genesisAccounts.Keys() {
		acc := genesisAccounts.MustGet(accAddress)

		if acc.name == name {
			return accAddress, nil
		}

	}

	return "", fmt.Errorf("address not found")
}

func checkDecTolerance(coins sdk.DecCoins, maxToleratedDiff sdk.Int) error {
	for _, coin := range coins {
		if coin.Amount.TruncateInt().GT(maxToleratedDiff) {
			return fmt.Errorf("Remaining balance %s is too high", coin.String())
		}
	}
	return nil
}

func WithdrawGenesisGravity(genesisData *GenesisData, networkInfo NetworkConfig, manifest *UpgradeManifest) error {

	gravityBalance := genesisData.accounts.MustGet(genesisData.gravityModuleAccountAddress).balance
	err := MoveGenesisBalance(genesisData, genesisData.gravityModuleAccountAddress, networkInfo.remainingGravityBalanceAddr, gravityBalance, manifest)
	if err != nil {
		return err
	}

	return nil
}

func MigrateGenesisAccounts(genesisData *GenesisData, ctx sdk.Context, app *App, networkInfo NetworkConfig, manifest *UpgradeManifest) error {
	// Mint donor chain total supply
	totalSupplyToMint, err := convertBalance(genesisData.totalSupply, networkInfo)
	if err != nil {
		return err
	}

	err = app.MintKeeper.MintCoins(ctx, totalSupplyToMint)
	if err != nil {
		return err
	}

	for _, genesisAccountAddress := range *genesisData.accounts.Keys() {
		genesisAccount := genesisData.accounts.MustGet(genesisAccountAddress)

		if genesisAccount.accountType == ContractAccountType {
			// All contracts balance should be handled already
			if genesisAccount.balance.Empty() {
				err = MarkAccountAsMigrated(genesisData, genesisAccountAddress)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Unresolved contract balance: %s %s", genesisAccountAddress, genesisAccount.balance.String())
			}
			continue
		}
		if genesisAccount.accountType == ModuleAccountType {
			if genesisAccount.balance.Empty() {
				err = MarkAccountAsMigrated(genesisData, genesisAccountAddress)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Unresolved module balance: %s %s %s", genesisAccountAddress, genesisAccount.balance.String(), genesisAccount.name)
			}
			continue
		}

		if genesisAccount.accountType == IBCAccountType {
			// All IBC balances should be handled already
			if genesisAccount.balance.Empty() {
				err = MarkAccountAsMigrated(genesisData, genesisAccountAddress)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Unresolved contract balance: %s %s", genesisAccountAddress, genesisAccount.balance.String())
			}
			continue
		}

		// Get balance to mint
		newBalance, err := convertBalance(genesisAccount.balance, networkInfo)
		if err != nil {
			return err
		}

		var newBaseAccount *authtypes.BaseAccount

		// Check for collision
		existingAccount := app.AccountKeeper.GetAccount(ctx, genesisAccount.rawAddress)
		if existingAccount != nil {
			// Handle collision

			existingAccountPubkey := existingAccount.GetPubKey()

			// Set pubkey from newAcc if is not in existingAccount
			if existingAccountPubkey == nil && genesisAccount.pubkey != nil {
				existingAccount.SetPubKey(genesisAccount.pubkey)
			}

			if genesisAccount.pubkey != nil && existingAccountPubkey != nil && !existingAccountPubkey.Equals(genesisAccount.pubkey) {
				return fmt.Errorf("account already exists with different pubkey: %s", genesisAccountAddress)
			}

			newBaseAccount = authtypes.NewBaseAccount(genesisAccount.rawAddress, existingAccount.GetPubKey(), existingAccount.GetAccountNumber(), existingAccount.GetSequence())

		} else {

			// Handle regular migration
			newBaseAccount, err = getNewBaseAccount(ctx, app, *genesisAccount)
			if err != nil {
				return err
			}

		}

		// If there is anything to mint
		if newBalance != nil {

			// Account is vesting
			if networkInfo.notVestedAccounts[genesisAccountAddress] {
				err := createNewNormalAccountFromBaseAccount(ctx, app, newBaseAccount)
				if err != nil {
					return err
				}
			} else {
				// Account is not vesting
				err := createNewVestingAccountFromBaseAccount(ctx, app, newBaseAccount, newBalance, networkInfo.mergeTime, networkInfo.mergeTime+networkInfo.vestingPeriod)
				if err != nil {
					return err
				}
			}

			err = mintToAccount(ctx, app, genesisAccountAddress, genesisAccount.rawAddress, newBalance, manifest)
			if err != nil {
				return err
			}
			// There is nothing to mint
		} else {
			// Just create account if it's base account, but there is no balance for vesting
			err := createNewNormalAccountFromBaseAccount(ctx, app, newBaseAccount)
			if err != nil {
				return err
			}
		}

		err = MarkAccountAsMigrated(genesisData, genesisAccountAddress)
		if err != nil {
			return err
		}

	}

	return nil
}

func parseTotalSupply(jsonData map[string]interface{}) (sdk.Coins, error) {
	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	supply := bank["supply"].([]interface{})
	totalSupply, err := getCoinsFromInterfaceSlice(supply)
	if err != nil {
		return nil, err
	}

	return totalSupply, nil

}

func VerifySupply(genesisData *GenesisData, networkInfo NetworkConfig, manifest *UpgradeManifest) error {

	expectedMintedSupply, err := convertBalance(genesisData.totalSupply, networkInfo)
	if err != nil {
		return err
	}

	mintedSupply := manifest.Minting.AggregatedMintedAmount

	maximumDifference, ok := sdk.NewIntFromString("10000000000")
	if !ok {
		return fmt.Errorf("invalid maximum difference value")
	}

	for _, expectedCoin := range expectedMintedSupply {
		for _, mintedCoin := range mintedSupply {
			if expectedCoin.Denom == mintedCoin.Denom {
				var difference sdk.Int
				if expectedCoin.Amount.GT(mintedCoin.Amount) {
					difference = expectedCoin.Amount.Sub(mintedCoin.Amount)
				} else {
					difference = mintedCoin.Amount.Sub(expectedCoin.Amount)
				}

				if difference.GT(maximumDifference) {
					return fmt.Errorf("Total supply is not correct, expected %s, got %s", expectedCoin.String(), mintedCoin.String())
				}

			}
		}

	}

	return nil
}

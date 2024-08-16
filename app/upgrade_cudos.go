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
	"math/big"
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

type AccountInfo struct {
	name        string
	pubkey      cryptotypes.PubKey
	balance     sdk.Coins
	migrated    bool
	accountType AccountType
	rawAddress  sdk.AccAddress
}

func getGenesisAccountMap(jsonData map[string]interface{}, contractAccountMap map[string]ContractInfo, IBCAccountsMap map[string]IBCInfo, networkInfo NetworkConfig) (map[string]AccountInfo, error) {
	var err error

	// Map to verify that account exists in auth module
	auth := jsonData[authtypes.ModuleName].(map[string]interface{})
	accounts := auth["accounts"].([]interface{})

	accountMap := make(map[string]AccountInfo)

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
		} else if _, exists := contractAccountMap[addr]; exists {
			accountType = ContractAccountType
		} else if _, exists := IBCAccountsMap[addr]; exists {
			accountType = IBCAccountType
		} else {
			accountType = BaseAccountType
		}

		// Get raw address
		accRawAddr, err := convertAddressToRaw(addr, networkInfo)
		if err != nil {
			return nil, err
		}

		accountMap[addr] = AccountInfo{name: name, pubkey: AccPubKey, balance: sdk.NewCoins(), migrated: false, accountType: accountType, rawAddress: accRawAddr}
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

type ValidatorInfo struct {
	stake                  sdk.Int
	shares                 sdk.Dec
	status                 string
	operatorAddress        string
	consensusPubkey        cryptotypes.PubKey
	validatorStringAddress string
}

func getGenesisValidatorsMap(jsonData map[string]interface{}) (map[string]ValidatorInfo, map[string]string, error) {
	// Validator pubkey hex -> ValidatorInfo
	validatorInfoMap := make(map[string]ValidatorInfo)
	validatorOperatorMap := make(map[string]string)

	staking := jsonData[stakingtypes.ModuleName].(map[string]interface{})
	validators := staking["validators"].([]interface{})

	for _, validator := range validators {

		validatorMap := validator.(map[string]interface{})
		tokens := validatorMap["tokens"].(string)
		operatorAddress := validator.(map[string]interface{})["operator_address"].(string)

		consensusPubkey := validator.(map[string]interface{})["consensus_pubkey"].(map[string]interface{})
		decodedConsensusPubkey, err := decodePubKeyFromMap(consensusPubkey)
		if err != nil {
			return nil, nil, err
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
			return nil, nil, err
		}

		validatorStringAddress := decodedConsensusPubkey.Address().String()

		validatorInfoMap[validatorStringAddress] = ValidatorInfo{
			stake:                  tokensInt,
			shares:                 validatorSharesDec,
			status:                 status,
			operatorAddress:        operatorAddress,
			consensusPubkey:        decodedConsensusPubkey,
			validatorStringAddress: validatorStringAddress,
		}
		validatorOperatorMap[operatorAddress] = validatorStringAddress

	}
	return validatorInfoMap, validatorOperatorMap, nil
}

func withdrawGenesisStakingDelegations(jsonData map[string]interface{}, genesisValidators map[string]ValidatorInfo, validatorOperatorMap map[string]string, genesisAccounts map[string]AccountInfo, contractAccountMap map[string]ContractInfo, networkInfo NetworkConfig, manifest *UpgradeManifest) (map[string]map[string]sdk.Coins, error) {
	staking := jsonData[stakingtypes.ModuleName].(map[string]interface{})

	bondedPoolAddress, err := GetAddressByName(jsonData, BondedPoolAccName)
	if err != nil {
		return nil, err
	}

	notBondedPoolAddress, err := GetAddressByName(jsonData, NotBondedPoolAccName)
	if err != nil {
		return nil, err
	}

	// Handle delegations

	// Map of delegatorAddress -> validatorPubkey -> sdk.coins balance
	delegatedBalanceMap := make(map[string]map[string]sdk.Coins)
	delegations := staking["delegations"].([]interface{})
	for _, delegation := range delegations {
		delegationMap := delegation.(map[string]interface{})
		delegatorAddress := delegationMap["delegator_address"].(string)
		resolvedDelegatorAddress := resolveIfContractAddress(delegatorAddress, contractAccountMap)
		validatorOperatorAddress := delegationMap["validator_address"].(string)
		delegatorSharesDec, err := sdk.NewDecFromStr(delegationMap["shares"].(string))
		if err != nil {
			return nil, err
		}

		validatorAddress := validatorOperatorMap[validatorOperatorAddress]
		currentValidatorInfo := genesisValidators[validatorAddress]

		delegatorTokens := (delegatorSharesDec.MulInt(currentValidatorInfo.stake)).Quo(currentValidatorInfo.shares).TruncateInt()

		// Move balance to delegator address
		delegatorBalance := sdk.NewCoins(sdk.NewCoin(networkInfo.originalDenom, delegatorTokens))

		// Subtract balance from bonded or not-bonded pool
		if currentValidatorInfo.status == BondedStatus {
			// Store delegation to delegated map
			if delegatedBalanceMap[resolvedDelegatorAddress] == nil {
				delegatedBalanceMap[resolvedDelegatorAddress] = make(map[string]sdk.Coins)
			}

			if delegatedBalanceMap[resolvedDelegatorAddress][validatorAddress] == nil {
				delegatedBalanceMap[resolvedDelegatorAddress][validatorAddress] = sdk.NewCoins()
			}

			delegatedBalanceMap[resolvedDelegatorAddress][validatorAddress] = delegatedBalanceMap[resolvedDelegatorAddress][validatorAddress].Add(delegatorBalance...)

			// Move balance from bonded pool to delegator
			err := moveGenesisBalance(genesisAccounts, bondedPoolAddress, resolvedDelegatorAddress, delegatorBalance, manifest)
			if err != nil {
				return nil, err
			}

		} else {
			// Delegations to unbonded/jailed/tombstoned validators are not re-delegated

			// Move balance from not-bonded pool to delegator
			err := moveGenesisBalance(genesisAccounts, notBondedPoolAddress, resolvedDelegatorAddress, delegatorBalance, manifest)
			if err != nil {
				return nil, err
			}
		}

		// TODO: This balance should be delegated to new validator, but it needs to be minted first!

	}

	// Handle unbonding delegations
	totalUnbondingDelegationsStake := sdk.NewInt(0)
	unbonding_delegations := staking["unbonding_delegations"].([]interface{})
	for _, unbondingDelegation := range unbonding_delegations {
		unbondingDelegationMap := unbondingDelegation.(map[string]interface{})

		entries := unbondingDelegationMap["entries"].([]interface{})
		delegatorAddress := unbondingDelegationMap["delegator_address"].(string)
		resolvedDelegatorAddress := resolveIfContractAddress(delegatorAddress, contractAccountMap)

		for _, entry := range entries {
			unbondingDelegationTokensInt, ok := sdk.NewIntFromString(entry.(map[string]interface{})["balance"].(string))
			if !ok {
				return nil, fmt.Errorf("Cannot parse balance to Int")
			}

			unbondingDelegationBalance := sdk.NewCoins(sdk.NewCoin(networkInfo.originalDenom, unbondingDelegationTokensInt))

			// Move unbonding balance from not-bonded pool to delegator address
			err := moveGenesisBalance(genesisAccounts, notBondedPoolAddress, resolvedDelegatorAddress, unbondingDelegationBalance, manifest)
			if err != nil {
				return nil, err
			}

			totalUnbondingDelegationsStake = totalUnbondingDelegationsStake.Add(unbondingDelegationTokensInt)
		}
	}

	// Handle remaining pool balances

	// Handle remaining bonded pool balance
	err = moveGenesisBalance(genesisAccounts, bondedPoolAddress, networkInfo.remainingBalanceAddr, genesisAccounts[bondedPoolAddress].balance, manifest)
	if err != nil {
		return nil, err
	}

	// Handle remaining not-bonded pool balance
	err = moveGenesisBalance(genesisAccounts, notBondedPoolAddress, networkInfo.remainingBalanceAddr, genesisAccounts[notBondedPoolAddress].balance, manifest)
	if err != nil {
		return nil, err
	}

	return delegatedBalanceMap, nil
}

func createGenesisDelegations(ctx sdk.Context, app *App, delegatedBalanceMap map[string]map[string]sdk.Coins, genesisValidatorsMap map[string]ValidatorInfo, validatorOperatorMap map[string]string, networkInfo NetworkConfig, manifest *UpgradeManifest) error {

	for delegatorAddr, delegatorAddrMap := range delegatedBalanceMap {
		for validatorStringAddr, delegatedBalance := range delegatorAddrMap {
			println(delegatorAddr, validatorStringAddr, delegatedBalance)

			operatorAddr := app.StakingKeeper.GetValidators(ctx, 234)[0].GetOperator()

			validator, found := app.StakingKeeper.GetValidator(ctx, operatorAddr)
			if !found {
				println("not found")
			}

			/*
				validatorAddr, err := sdk.ValAddressFromHex(networkInfo.backupValidators[0])
				if err != nil {
					return err
				}
				validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddr)
				if !found {
					println("not found")
				}

			*/

			println(validator.Status)
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

func withdrawGenesisContractBalances(genesisAccounts map[string]AccountInfo, contractAccountMap map[string]ContractInfo, networkInfo NetworkConfig, manifest *UpgradeManifest) error {

	for contractAddress := range contractAccountMap {
		resolvedAddress := resolveIfContractAddress(contractAddress, contractAccountMap)
		if resolvedAddress == contractAddress {
			return fmt.Errorf("Failed to resolve contract admin/owner for contract %s", contractAddress)
		}

		contractBalance, contractBalancePresent := genesisAccounts[contractAddress]
		if contractBalancePresent {
			err := moveGenesisBalance(genesisAccounts, contractAddress, resolvedAddress, contractBalance.balance, manifest)
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
		if divisionConst, ok := networkInfo.balanceConversionConstants[coin.Denom]; ok {
			divisionConstBigInt := big.NewInt(int64(divisionConst))
			newAmount := new(big.Int).Div(coin.Amount.BigInt(), divisionConstBigInt)

			sdkCoin := sdk.NewCoin(networkInfo.convertedDenom, sdk.NewIntFromBigInt(newAmount))
			resBalance = resBalance.Add(sdkCoin)
		} else {
			println("Unknown denom: ", coin.Denom)
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

func fillGenesisBalancesToAccountsMap(jsonData map[string]interface{}, genesisAccountsMap map[string]AccountInfo) error {
	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	balances := bank["balances"].([]interface{})

	for _, balance := range balances {

		addr := balance.(map[string]interface{})["address"]
		if addr == nil {
			panic("Failed to get address")
		}
		addrStr := addr.(string)

		// Verify that account exists in auth module
		if _, exists := genesisAccountsMap[addrStr]; !exists {
			panic("Account not registered in auth module")
		}

		coins := balance.(map[string]interface{})["coins"]

		sdkBalance, err := getCoinsFromInterfaceSlice(coins.([]interface{}))
		if err != nil {
			panic(err)
		}

		if !sdkBalance.Empty() {
			accountInfoEntry := genesisAccountsMap[addrStr]
			accountInfoEntry.balance = sdkBalance

			genesisAccountsMap[addrStr] = accountInfoEntry
		}

	}
	return nil
}

func GenesisUpgradeWithdrawIBCChannelsBalances(IBCAccountsMap map[string]IBCInfo, genesisAccounts map[string]AccountInfo, networkInfo NetworkConfig, manifest *UpgradeManifest) error {
	if networkInfo.ibcTargetAddr == "" {

		panic("No IBC withdrawal address set")
	}

	ibcWithdrawalAddress := networkInfo.ibcTargetAddr

	manifest.IBC = &UpgradeIBCTransfers{
		To: ibcWithdrawalAddress,
	}

	for IBCaccountAddress, IBCinfo := range IBCAccountsMap {

		IBCaccount, IBCAccountExists := genesisAccounts[IBCaccountAddress]

		var channelBalance sdk.Coins
		if IBCAccountExists {

			channelBalance = IBCaccount.balance
			err := moveGenesisBalance(genesisAccounts, IBCaccountAddress, ibcWithdrawalAddress, channelBalance, manifest)
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

func GetIBCAccountsMap(jsonData map[string]interface{}, networkInfo NetworkConfig) (map[string]IBCInfo, error) {
	ibcAccountSet := make(map[string]IBCInfo)

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

		ibcAccountSet[channelAddr] = IBCInfo{channelId: channelId, portId: portId}
	}

	return ibcAccountSet, nil
}

type ContractInfo struct {
	Admin   string
	Creator string
}

func GetWasmContractAccounts(jsonData map[string]interface{}) (map[string]ContractInfo, error) {
	contractAccountMap := make(map[string]ContractInfo)

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

		contractAccountMap[contractAddr] = ContractInfo{Admin: admin, Creator: creator}
	}

	return contractAccountMap, nil
}

func resolveIfContractAddress(address string, contractAccountMap map[string]ContractInfo) string {
	if contractInfo, exists := contractAccountMap[address]; exists {
		if contractInfo.Admin != "" {
			return resolveIfContractAddress(contractInfo.Admin, contractAccountMap)
		} else if contractInfo.Creator != "" {
			return resolveIfContractAddress(contractInfo.Creator, contractAccountMap)
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

	err := app.MintKeeper.MintCoins(ctx, newCoins)
	if err != nil {
		return err
	}
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, toAddress, newCoins)
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

func MarkAccountAsMigrated(genesisAccounts map[string]AccountInfo, accountAddress string) error {
	AccountInfoRecord, exists := genesisAccounts[accountAddress]
	if !exists {
		return fmt.Errorf("Genesis account %s not found", accountAddress)
	}

	if AccountInfoRecord.migrated {
		return fmt.Errorf("Genesis account %s already migrated", accountAddress)
	}

	AccountInfoRecord.migrated = true

	genesisAccounts[accountAddress] = AccountInfoRecord

	return nil
}

func moveGenesisBalance(genesisAccounts map[string]AccountInfo, fromAddress, toAddress string, amount sdk.Coins, manifest *UpgradeManifest) error {
	// Check if fromAddress exists
	if _, ok := genesisAccounts[fromAddress]; !ok {
		return fmt.Errorf("fromAddress %s does not exist in genesis balances", fromAddress)
	}

	if _, ok := genesisAccounts[toAddress]; !ok {
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

	if genesisAccounts[toAddress].migrated {
		return fmt.Errorf("Genesis account %s already migrated", toAddress)
	}
	if genesisAccounts[fromAddress].migrated {
		return fmt.Errorf("Genesis account %s already migrated", fromAddress)
	}

	genesisToBalance := genesisAccounts[toAddress]
	genesisFromBalance := genesisAccounts[fromAddress]

	genesisToBalance.balance = genesisToBalance.balance.Add(amount...)
	genesisFromBalance.balance = genesisFromBalance.balance.Sub(amount)

	genesisAccounts[toAddress] = genesisToBalance
	genesisAccounts[fromAddress] = genesisFromBalance

	return nil
}

func GetAddressByName(jsonData map[string]interface{}, name string) (string, error) {
	auth := jsonData[authtypes.ModuleName].(map[string]interface{})
	accounts := auth["accounts"].([]interface{})

	for _, acc := range accounts {
		accMap := acc.(map[string]interface{})
		accType := accMap["@type"]

		if accType == ModuleAccount {
			accName := accMap["name"].(string)
			if accName == name {

				baseAccData := accMap["base_account"].(map[string]interface{})
				accAddr := baseAccData["address"].(string)

				return accAddr, nil
			}
		}

	}

	return "", fmt.Errorf("address not found")
}

func MigrateGenesisAccounts(ctx sdk.Context, app *App, networkInfo NetworkConfig, manifest *UpgradeManifest, genesisAccountsMap map[string]AccountInfo) error {
	var err error
	for genesisAccountAddress, genesisAccount := range genesisAccountsMap {

		if genesisAccount.accountType == ContractAccountType {
			// All contracts balance should be handled already
			if genesisAccountsMap[genesisAccountAddress].balance.Empty() {
				err = MarkAccountAsMigrated(genesisAccountsMap, genesisAccountAddress)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("Unresolved contract balance: %s %s", genesisAccountAddress, genesisAccount.balance.String())
			}
			continue
		}
		if genesisAccount.accountType == ModuleAccountType {
			if genesisAccountsMap[genesisAccountAddress].balance.Empty() {
				err = MarkAccountAsMigrated(genesisAccountsMap, genesisAccountAddress)
				if err != nil {
					return err
				}
			} else {
				// TODO: Replace with return ERROR when all modules balances are handled
				println("Unresolved module balance: ", genesisAccountAddress, genesisAccount.balance.String(), genesisAccount.name)
			}
			continue
		}

		if genesisAccount.accountType == IBCAccountType {
			// All IBC balances should be handled already
			if genesisAccountsMap[genesisAccountAddress].balance.Empty() {
				err = MarkAccountAsMigrated(genesisAccountsMap, genesisAccountAddress)
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
			newBaseAccount, err = getNewBaseAccount(ctx, app, genesisAccount)
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

		err = MarkAccountAsMigrated(genesisAccountsMap, genesisAccountAddress)
		if err != nil {
			return err
		}

	}

	return nil
}

func VerifySupply(jsonData map[string]interface{}, genesisAccounts map[string]AccountInfo, networkInfo NetworkConfig, manifest *UpgradeManifest) error {

	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	supply := bank["supply"].([]interface{})
	totalSupply, err := getCoinsFromInterfaceSlice(supply)
	if err != nil {
		return err
	}

	expectedMintedSupply, err := convertBalance(totalSupply, networkInfo)
	if err != nil {
		return err
	}

	println(expectedMintedSupply.String())
	println(manifest.Minting.AggregatedMintedAmount.String())

	return nil
}

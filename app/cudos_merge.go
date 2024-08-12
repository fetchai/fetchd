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
	"strconv"
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
	OldAddrPrefix = "cudos"

	OriginalDenom  = "acudos"
	ConvertedDenom = "afet"

	MergeTime     = 123456                // Epoch time of merge
	VestingPeriod = 3 * 30 * 24 * 60 * 60 // 3 months period

	FlagGenesisTime = "genesis-time"

	ModuleAccount  = "/cosmos.auth.v1beta1.ModuleAccount"
	BaseAccount    = "/cosmos.auth.v1beta1.BaseAccount"
	UnbondedStatus = "BOND_STATUS_UNBONDED"
	BondedStatus   = "BOND_STATUS_BONDED"

	TransferAccName      = "transfer"
	BondedPoolAccName    = "bonded_tokens_pool"
	NotBondedPoolAccName = "not_bonded_tokens_pool"
	MintAccName          = "cudoMint"
	GovAccName           = "gov"
	DistributionAccName  = "distribution"
	GravityAccName       = "gravity"
	MarketplaceAccName   = "marketplace"
	FeeCollectorAccName  = "fee_collector"
)

var BalanceDivisionConstants = map[string]int{
	OriginalDenom: 11,
}

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

func convertAddressToRaw(addr string) (sdk.AccAddress, error) {
	prefix, decodedAddrData, err := bech32.DecodeAndConvert(addr)

	if prefix != OldAddrPrefix {
		return nil, fmt.Errorf("Unknown prefix: %s", prefix)
	}

	if err != nil {
		return nil, err
	}

	return decodedAddrData, nil
}

func getGenesisAccountSequenceMap(accounts []interface{}) map[string]int {
	accountMap := make(map[string]int)

	for _, acc := range accounts {
		accMap := acc.(map[string]interface{})
		accType := accMap["@type"]

		accData := acc
		if accType == ModuleAccount {
			accData = accMap["base_account"]
		}

		accDataMap := accData.(map[string]interface{})
		addr := accDataMap["address"].(string)
		sequence := accDataMap["sequence"].(string)

		sequenceInt, ok := strconv.Atoi(sequence)
		if ok != nil {
			panic("getGenesisAccountSequenceMap: failed to convert sequence to int")
		}
		accountMap[addr] = sequenceInt
	}

	return accountMap
}

func getGenesisBalancesMap(balances []interface{}) *map[string]int {
	balanceMap := make(map[string]int)

	for i, balance := range balances {
		addr := balance.(map[string]interface{})["address"]
		if addr == nil {
			fmt.Println(balance)
		}
		addrStr := addr.(string)
		balanceMap[addrStr] = i
	}

	return &balanceMap
}

func getConsAddressFromValidator(validatorData map[string]interface{}) (sdk.ConsAddress, error) {
	consensusPubkey := validatorData["consensus_pubkey"].(map[string]interface{})
	decodedConsensusPubkey, err := decodePubKeyFromMap(consensusPubkey)
	if err != nil {
		return nil, err
	}
	return sdk.ConsAddress(decodedConsensusPubkey.Address()), nil
}

func withdrawGenesisStakingRewards(jsonData map[string]interface{}, convertedBalances map[string]sdk.Coins) (map[string]map[string]sdk.Coins, error) {

	// Validator pubkey hex -> tokens int amount
	validatorStakeMap := make(map[string]sdk.Int)
	validatorSharesMap := make(map[string]sdk.Dec)

	// Operator address -> Validator pubkey hex
	validatorOperatorToPubkeyMap := make(map[string]string)

	staking := jsonData[stakingtypes.ModuleName].(map[string]interface{})
	validators := staking["validators"].([]interface{})

	// Prepare maps and total tokens amount
	totalStake := sdk.NewInt(0)
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
		totalStake = totalStake.Add(tokensInt)

		validatorStakeMap[decodedConsensusPubkey.String()] = tokensInt
		validatorOperatorToPubkeyMap[operatorAddress] = decodedConsensusPubkey.String()

		validatorShares := validatorMap["delegator_shares"].(string)

		validatorSharesDec, err := sdk.NewDecFromStr(validatorShares)
		if err != nil {
			return nil, err
		}
		validatorSharesMap[decodedConsensusPubkey.String()] = validatorSharesDec

	}

	println(totalStake.String())

	// Handle delegations

	// Map of delegatorAddress -> validatorPubkey -> sdk.coins balance
	delegatedBalanceMap := make(map[string]map[string]sdk.Coins)
	delegations := staking["delegations"].([]interface{})
	for _, delegation := range delegations {
		delegationMap := delegation.(map[string]interface{})
		delegatorAddress := delegationMap["delegator_address"].(string)
		validatorOperatorAddress := delegationMap["validator_address"].(string)
		delegatorShares := delegationMap["shares"].(string)

		delegatorSharesDec, err := sdk.NewDecFromStr(delegatorShares)
		if err != nil {
			return nil, err
		}

		validatorPubkey := validatorOperatorToPubkeyMap[validatorOperatorAddress]
		validatorTokens := validatorStakeMap[validatorPubkey]
		validatorShares := validatorSharesMap[validatorPubkey]

		var delegatorTokens sdk.Int
		if validatorTokens.String() != validatorShares.TruncateInt().String() {
			delegatorTokens = (delegatorSharesDec.MulInt(validatorTokens)).Quo(validatorShares).TruncateInt()
		} else {
			delegatorTokens = delegatorSharesDec.TruncateInt()
		}

		//println("("+delegatorSharesDec.String()+"/"+validatorShares.String()+")*"+validatorTokens.String()+"="+delegatorTokens.String(), delegatorAddress)

		// Add delegated balance to convertedBalances map

		delegatorBalance := sdk.NewCoins(sdk.NewCoin(OriginalDenom, delegatorTokens))

		// Convert acudos to afet
		convertedBalance, err := convertBalance(delegatorBalance)
		if err != nil {
			panic(err)
		}

		// Add balance to converted balances - panics if account doesn't already exist
		convertedBalances[delegatorAddress] = convertedBalances[delegatorAddress].Add(convertedBalance...)

		if delegatedBalanceMap[delegatorAddress] == nil {
			delegatedBalanceMap[delegatorAddress] = make(map[string]sdk.Coins)
		}

		if delegatedBalanceMap[delegatorAddress][validatorPubkey] == nil {
			delegatedBalanceMap[delegatorAddress][validatorPubkey] = sdk.NewCoins()
		}

		delegatedBalanceMap[delegatorAddress][validatorPubkey] = delegatedBalanceMap[delegatorAddress][validatorPubkey].Add(convertedBalance...)

		// TODO: This balance should be delegated to new validator, but it needs to be minted first!

	}

	// Handle unbonding delegations

	// Handle redelegations

	// Handle rewards

	return delegatedBalanceMap, nil
}

func parseGenesisBalance(coins []interface{}) (sdk.Coins, error) {
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

func convertBalance(balance sdk.Coins) (sdk.Coins, error) {
	var resBalance sdk.Coins

	for _, coin := range balance {
		if divisionConst, ok := BalanceDivisionConstants[coin.Denom]; ok {
			divisionConstBigInt := big.NewInt(int64(divisionConst))
			newAmount := new(big.Int).Div(coin.Amount.BigInt(), divisionConstBigInt)

			sdkCoin := sdk.NewCoin(ConvertedDenom, sdk.NewIntFromBigInt(newAmount))
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

func getConvertedGenesisBalancesMap(jsonData map[string]interface{}) map[string]sdk.Coins {
	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	balances := bank["balances"].([]interface{})

	// Map to verify that account exists in auth module
	auth := jsonData[authtypes.ModuleName].(map[string]interface{})
	accounts := auth["accounts"].([]interface{})
	accountsMap := getGenesisAccountSequenceMap(accounts)

	balanceMap := make(map[string]sdk.Coins)
	for _, balance := range balances {

		addr := balance.(map[string]interface{})["address"]
		if addr == nil {
			panic("Failed to get address")
		}
		addrStr := addr.(string)

		// Verify that account exists in auth module
		if _, exists := accountsMap[addrStr]; !exists {
			panic("Account not registered in auth module")
		}

		coins := balance.(map[string]interface{})["coins"]

		sdkBalance, err := parseGenesisBalance(coins.([]interface{}))
		if err != nil {
			panic(err)
		}

		convertedBalance, err := convertBalance(sdkBalance)
		if err != nil {
			panic(err)
		}
		if !convertedBalance.Empty() {
			balanceMap[addrStr] = convertedBalance
		}

	}

	return balanceMap
}

func getCoinsFromInterfaceSlice(data interface{}) sdk.Coins {
	balance := data.(map[string]interface{})["coins"]
	var balanceCoins sdk.Coins

	for _, coin := range balance.([]interface{}) {
		coinData := coin.(map[string]interface{})
		coinDenom := coinData["denom"].(string)
		coinAmount, ok := sdk.NewIntFromString(coinData["amount"].(string))
		if !ok {
			panic("ibc withdraw: failed to convert coin amount to int")
		}
		balanceCoins = append(balanceCoins, sdk.NewCoin(coinDenom, coinAmount))
	}

	balanceCoins = sdk.NewCoins(balanceCoins...)
	return balanceCoins
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

func GenesisUpgradeWithdrawIBCChannelsBalances(jsonData map[string]interface{}, networkInfo NetworkConfig, manifest *UpgradeManifest) {
	if networkInfo.IbcTargetAddr == "" {
		return
	}

	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	balances := bank["balances"].([]interface{})
	balanceMap := getGenesisBalancesMap(balances)
	ibcWithdrawalAddress := networkInfo.IbcTargetAddr

	manifest.IBC = &UpgradeIBCTransfers{
		To: ibcWithdrawalAddress,
	}
	withdrawalBalanceIdx, ok := (*balanceMap)[ibcWithdrawalAddress]
	if !ok {
		panic("failed to find ibc withdrawal address in genesis balances")
	}

	ibc := jsonData[ibccore.ModuleName].(map[string]interface{})
	channelGenesis := ibc["channel_genesis"].(map[string]interface{})
	ibcChannels := channelGenesis["channels"].([]interface{})

	for _, channel := range ibcChannels {
		channelMap := channel.(map[string]interface{})
		channelId := channelMap["channel_id"].(string)
		portId := channelMap["port_id"].(string)

		// close channel
		channelMap["state"] = "STATE_CLOSED"

		rawAddr := ibctransfertypes.GetEscrowAddress(portId, channelId)
		channelAddr, err := sdk.Bech32ifyAddressBytes(OldAddrPrefix+AccAddressPrefix, rawAddr)
		if err != nil {
			panic(err)
		}

		balanceIdx, ok := (*balanceMap)[channelAddr]
		if !ok {
			// channel address not found in genesis balances
			continue
		}

		channelBalanceCoins := getCoinsFromInterfaceSlice(balances[balanceIdx])
		withdrawalBalanceCoins := getCoinsFromInterfaceSlice(balances[withdrawalBalanceIdx])

		// add channel balance to withdrawal balance
		newWithdrawalBalanceCoins := withdrawalBalanceCoins.Add(channelBalanceCoins...)
		balances[withdrawalBalanceIdx].(map[string]interface{})["coins"] = getInterfaceSliceFromCoins(newWithdrawalBalanceCoins)

		// zero out the channel balance
		balances[balanceIdx].(map[string]interface{})["coins"] = []interface{}{}

		manifest.IBC.Transfers = append(manifest.IBC.Transfers, UpgradeIBCTransfer{From: channelAddr, ChannelID: fmt.Sprintf("%s/%s", portId, channelId), Amount: channelBalanceCoins})
		manifest.IBC.AggregatedTransferredAmount = manifest.IBC.AggregatedTransferredAmount.Add(channelBalanceCoins...)
		manifest.IBC.NumberOfTransfers += 1
	}
}

func GetIBCAccountAddresses(jsonData map[string]interface{}) (map[string]bool, error) {
	ibcAccountSet := make(map[string]bool)

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
		channelAddr, err := sdk.Bech32ifyAddressBytes(OldAddrPrefix+AccAddressPrefix, rawAddr)
		if err != nil {
			return nil, err
		}

		ibcAccountSet[channelAddr] = true
	}

	return ibcAccountSet, nil
}

func GetWasmContractAccounts(jsonData map[string]interface{}) (map[string]bool, error) {
	contractAccountSet := make(map[string]bool)

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

		contractAccountSet[contractAddr] = true
	}

	return contractAccountSet, nil
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

func getNewBaseAccount(ctx sdk.Context, app *App, accDataMap map[string]interface{}) (*authtypes.BaseAccount, error) {
	// Get raw address
	addr := accDataMap["address"].(string)
	accRawAddr, err := convertAddressToRaw(addr)
	if err != nil {
		return nil, err
	}

	// Set pubkey if present
	var pubKey cryptotypes.PubKey
	if pk, ok := accDataMap["pub_key"]; ok {
		if pk != nil {
			pubKey, err = decodePubKeyFromMap(pk.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
		}
	}

	// Create new account

	newAccNumber := app.AccountKeeper.GetNextAccountNumber(ctx)
	newBaseAccount := authtypes.NewBaseAccount(accRawAddr, pubKey, newAccNumber, 0)
	return newBaseAccount, nil
}

func createNewVestingAccountFromBaseAccount(ctx sdk.Context, app *App, account *authtypes.BaseAccount, vestedCoins sdk.Coins, startTime int64, endTime int64) error {
	newBaseVestingAcc := authvesting.NewBaseVestingAccount(account, vestedCoins, endTime)
	newContinuousVestingAcc := authvesting.NewContinuousVestingAccountRaw(newBaseVestingAcc, startTime)

	app.AccountKeeper.SetAccount(ctx, newContinuousVestingAcc)

	return nil
}

func mintToAccount(ctx sdk.Context, app *App, fromAddress string, toAddress sdk.AccAddress, newCoins sdk.Coins, manifest *UpgradeManifest) error {

	app.MintKeeper.MintCoins(ctx, newCoins)
	app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, toAddress, newCoins)

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

func GetAddressByName(name string, jsonData map[string]interface{}) (string, error) {
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

func ProcessBaseAccountsAndBalances(ctx sdk.Context, app *App, jsonData map[string]interface{}, networkInfo NetworkConfig, manifest *UpgradeManifest, convertedBalancesMap map[string]sdk.Coins) error {

	auth := jsonData[authtypes.ModuleName].(map[string]interface{})
	accounts := auth["accounts"].([]interface{})

	ibcAccountsSet, err := GetIBCAccountAddresses(jsonData)
	if err != nil {
		return err
	}

	contractAccountsSet, err := GetWasmContractAccounts(jsonData)
	if err != nil {
		return err
	}

	// Handle accounts
	for _, acc := range accounts {
		accMap := acc.(map[string]interface{})
		accType := accMap["@type"]

		accData := acc
		if accType == ModuleAccount {
			accData = accMap["base_account"]
		}

		accDataMap := accData.(map[string]interface{})
		addr := accDataMap["address"].(string)

		// Skip if account is not regular basic account
		if ibcAccountsSet[addr] || contractAccountsSet[addr] || accType != BaseAccount {
			continue
		}

		accRawAddr, err := convertAddressToRaw(addr)
		if err != nil {
			return err
		}

		// Get balance to mint
		newBalance := convertedBalancesMap[addr]
		var newBaseAccount *authtypes.BaseAccount

		// Check for collision
		existingAccount := app.AccountKeeper.GetAccount(ctx, accRawAddr)
		if existingAccount != nil {
			// Handle collision

			// Check that public keys are the same
			var newAccPubKey cryptotypes.PubKey
			if pk, ok := accDataMap["pub_key"]; ok {
				if pk != nil {
					newAccPubKey, err = decodePubKeyFromMap(pk.(map[string]interface{}))
					if err != nil {
						return err
					}
				}
			}
			existingAccountPubkey := existingAccount.GetPubKey()

			// Set pubkey from newAcc if is not in existingAccount
			if existingAccountPubkey == nil && newAccPubKey != nil {
				existingAccount.SetPubKey(newAccPubKey)
			}

			if newAccPubKey != nil && existingAccountPubkey != nil && !existingAccountPubkey.Equals(newAccPubKey) {
				return fmt.Errorf("account already exists with different pubkey: %s", addr)
			}

			newBaseAccount = authtypes.NewBaseAccount(accRawAddr, existingAccount.GetPubKey(), existingAccount.GetAccountNumber(), existingAccount.GetSequence())

		} else {

			// Handle regular migration
			newBaseAccount, err = getNewBaseAccount(ctx, app, accDataMap)
			if err != nil {
				return err
			}

		}

		createNewVestingAccountFromBaseAccount(ctx, app, newBaseAccount, newBalance, MergeTime, MergeTime+VestingPeriod)

		err = mintToAccount(ctx, app, addr, accRawAddr, newBalance, manifest)
		if err != nil {
			return err
		}

	}

	return nil
}

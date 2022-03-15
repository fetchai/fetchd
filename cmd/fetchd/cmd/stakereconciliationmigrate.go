package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/fetchai/fetchd/app"
)

type IneligibleRegistration struct {
	NativeAddr string
	Reason     error
}

func (i IneligibleRegistration) Print() {
	fmt.Fprintf(os.Stderr, "%q ineligible for reason: %s\n", i.NativeAddr, i.Reason.Error())
}

type Registration struct {
	EthAddress    string `json:"eth_address"`
	NativeAddress string `json:"native_address"`
}

type Registrations []Registration

type AddrString = string
type AddrMap map[AddrString]Registration

func (r Registrations) EthAddrMap() AddrMap {
	regMap := make(AddrMap)
	for _, reg := range r {
		regMap[normalizeEthAddr(reg.EthAddress)] = reg
	}
	return regMap
}

var (
	errNonZeroSeqNum   = fmt.Errorf("%s", "sequence number must be 0")
	errBalanceNotMatch = fmt.Errorf("%s", "old account balance must match staked export amount")
	stakesCSVPath,
	registrationsPath,
	coinDenom string
	debugFlag    bool
	skipValidate bool
)

func AddStakeReconciliationMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake-reconciliation-migrate <genesis-path> -s <stakes-csv-path> -r <registrations-path>",
		Short: "Migrate fetch genesis to reconcile legacy stakes according to exported legacy stake data and reconciliation contract registrations",
		Long: `Migrate fetch genesis to reconcile legacy stakes according to exported legacy stake data and reconciliation contract registrations.
Eligible accounts:
	- are present in reconciliation contract all registrations query result
	- have a sequence number of 0 for the original fetch account
	- have a balance on the original fetch account which matches the legacy stake amount
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			genesisPath := args[0]
			genDoc, appState, err := loadAppStateFromGenesis(genesisPath)
			if err != nil {
				return err
			}
			registrations, err := loadRegistrations()
			if err != nil {
				return err
			}

			bankGenesis := banktypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)
			authGenesis := authtypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)
			genesisAccounts, err := authtypes.UnpackAccounts(authGenesis.Accounts)
			if err != nil {
				return fmt.Errorf("unable to unpack accounts: %w", err)
			}

			stakesFile, err := os.Open(stakesCSVPath)
			if err != nil {
				return fmt.Errorf("unable to open staked export CSV %q: %w", stakesCSVPath, err)
			}
			stakesReader := csv.NewReader(stakesFile)

			var (
				ineligible         []IneligibleRegistration
				row                []string
				ethAddrString      string
				oldFetchAddrString string
				oldFetchAddr       sdk.AccAddress
				migratedEthAddrs   = make(map[AddrString]struct{})
			)
			for {
				row, err = stakesReader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("failed to read stakes CSV file %s: %w", stakesCSVPath, err)
				}

				ethAddrString = normalizeEthAddr(row[0])
				if _, ok := migratedEthAddrs[ethAddrString]; ok {
					// Skip row if ethereum address has been successfully process previously this invocation.
					continue
				}

				oldFetchAddrString = row[2]
				oldFetchAddr, err = sdk.AccAddressFromBech32(oldFetchAddrString)
				if err != nil {
					return fmt.Errorf("failed to parse bech32 address for old account %q: %w", oldFetchAddrString, err)
				}

				registration, ok := registrations.EthAddrMap()[ethAddrString]
				if !ok {
					ineligible = append(ineligible, IneligibleRegistration{
						NativeAddr: oldFetchAddr.String(),
						Reason:     fmt.Errorf("no registration found for old account address %q", ethAddrString),
					})
					continue
				}

				newFetchAddr, err := sdk.AccAddressFromBech32(registration.NativeAddress)
				if err != nil {
					return fmt.Errorf("failed to parse bech32 address for new account %q: %w", registration.NativeAddress, err)
				}

				var oldAcctBalance, newAcctBalance sdk.Coins
				var oldBalanceIndex, newBalanceIndex int
				for i, b := range bankGenesis.Balances {
					if b.GetAddress().Equals(oldFetchAddr) {
						oldAcctBalance = b.GetCoins()
						oldBalanceIndex = i
						if newAcctBalance == nil {
							continue
						}
						break
					}

					if b.GetAddress().Equals(newFetchAddr) {
						newAcctBalance = b.GetCoins()
						newBalanceIndex = i
						if oldAcctBalance == nil {
							continue
						}
						break
					}
				}

				if oldAcctBalance == nil {
					continue
				}

				var oldAccount, newAccount authtypes.GenesisAccount
				for _, _account := range genesisAccounts {
					if _account.GetAddress().Equals(oldFetchAddr) {
						oldAccount = _account
						if newAccount == nil {
							continue
						}
						break
					}

					if _account.GetAddress().Equals(newFetchAddr) {
						newAccount = _account
						if oldAccount == nil {
							continue
						}
						break
					}

				}

				if oldAccount == nil {
					ineligible = append(ineligible, IneligibleRegistration{
						NativeAddr: oldFetchAddr.String(),
						Reason:     fmt.Errorf("unable to find old account with address %q", oldFetchAddr.String()),
					})
					continue
				}

				if newAccount == nil {
					ineligible = append(ineligible, IneligibleRegistration{
						NativeAddr: oldFetchAddr.String(),
						Reason:     fmt.Errorf("unable to find new account with address %q", newFetchAddr.String()),
					})
					continue
				}

				if newAcctBalance == nil {
					return fmt.Errorf("new account with address %q does not have a balance", newFetchAddr.String())
				}

				if oldAccount.GetSequence() != 0 {
					ineligible = append(ineligible, IneligibleRegistration{
						NativeAddr: oldFetchAddr.String(),
						Reason:     errNonZeroSeqNum,
					})
					continue
				}

				migrateCoin, err := sdk.ParseCoinNormalized(fmt.Sprintf("%s%s", row[3], coinDenom))
				if err != nil {
					return fmt.Errorf("unable to parse amount of tokens to migrate for row %q: %w", row, err)
				}
				migrateCoins := sdk.NewCoins(migrateCoin)

				if !oldAcctBalance.IsEqual(migrateCoins) {
					ineligible = append(ineligible, IneligibleRegistration{
						NativeAddr: oldFetchAddr.String(),
						Reason:     errBalanceNotMatch,
					})
					continue
				}

				// Zero out old account balance
				bankGenesis.Balances[oldBalanceIndex].Coins = oldAcctBalance.Sub(migrateCoins)

				// Add migrated coins to new account balance
				bankGenesis.Balances[newBalanceIndex].Coins = newAcctBalance.Add(migrateCoins...)

				// Mark this ethereum address as migrated.
				migratedEthAddrs[ethAddrString] = struct{}{}

			}

			updatedBankGenesisJSON, err := clientCtx.Codec.MarshalJSON(bankGenesis)
			if err != nil {
				return fmt.Errorf("unable to marshal updated bank genesis state: %w", err)
			}

			authGenesis.Accounts, err = authtypes.PackAccounts(genesisAccounts)
			if err != nil {
				return fmt.Errorf("unable to pack accounts: %w", err)
			}
			updatedAuthGenesisJSON, err := clientCtx.Codec.MarshalJSON(&authGenesis)
			if err != nil {
				return fmt.Errorf("unable to marshal updated auth genesis state: %w", err)
			}

			appState[banktypes.ModuleName] = updatedBankGenesisJSON
			appState[authtypes.ModuleName] = updatedAuthGenesisJSON

			// validate state (same as fetchd validate-genesis cmd)
			if !skipValidate {
				if err := app.ModuleBasics.ValidateGenesis(clientCtx.Codec, clientCtx.TxConfig, appState); err != nil {
					return fmt.Errorf("failed to validate state: %w", err)
				}
			}

			// build and print the new genesis state
			genDoc.AppState, err = json.Marshal(appState)
			if err != nil {
				return errors.Wrap(err, "failed to JSON marshal migrated genesis state")
			}
			bz, err := tmjson.Marshal(genDoc)
			if err != nil {
				return errors.Wrap(err, "failed to marshal genesis doc")
			}
			sortedBz, err := sdk.SortJSON(bz)
			if err != nil {
				return errors.Wrap(err, "failed to sort JSON genesis doc")
			}

			fmt.Println(string(sortedBz))

			if debugFlag {
				for _, _ineligible := range ineligible {
					_ineligible.Print()
				}
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&stakesCSVPath, "stakes-csv", "s", "./staked_export.csv", `Path to a CSV file containing legacy stake-holder addresses; fields: ETH_ADDR,PUB_KEY,ORIG_FETCH_ADDR,MIGRATE_AMOUNT`)
	cmd.Flags().StringVarP(&registrationsPath, "registrations", "r", "./registrations.json", "Path to a JSON file containing the result of querying reconciliation contract registrations")
	cmd.Flags().StringVarP(&coinDenom, "coin-denom", "c", "afet", "coin denomination to use when checking eligibility and migrating balances")
	cmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "if true, prints ineligible old addresses to stderr")
	cmd.Flags().BoolVar(&skipValidate, "skip-validate", false, "if true, skips validating the resulting genesis file before printing")

	return cmd
}

// normalizeEthAddr drops the "0x" prefix if there is one and makes letters lower case.
func normalizeEthAddr(ethAddr string) string {
	return strings.ToLower(strings.TrimPrefix(ethAddr, "0x"))
}

func loadRegistrations() (Registrations, error) {
	registrationsJSON, err := ioutil.ReadFile(registrationsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read registrations file %q: %w", registrationsPath, err)
	}
	var registrations Registrations
	if err := json.Unmarshal(registrationsJSON, &registrations); err != nil {
		return nil, fmt.Errorf("unable to unmarshal registrations JSON %q: %w", registrationsPath, err)
	}
	return registrations, nil
}

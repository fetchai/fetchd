package cmd

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	iavltree "github.com/cosmos/iavl"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"os"
	"path/filepath"
)

const (
	FlagApply = "apply"
)

// formatHeights formats a slice of int64 heights into a string representation of the versions to be pruned.
// It returns a formatted string of the heights in the format "X..Y" or "X..Y, Z" for consecutive or non-consecutive heights, respectively.
// If the slice is empty, it returns the string "nothing".
func formatHeights(heights []int64) string {
	if len(heights) == 0 {
		return "nothing"
	}

	var result string
	var previousHeight int64

	for _, height := range heights {
		if previousHeight == 0 {
			result = fmt.Sprintf("%v..", height)
		} else if height != previousHeight+1 {
			result += fmt.Sprintf("%v, %v..", previousHeight, height)
		}
		previousHeight = height
	}

	result += fmt.Sprintf("%v", previousHeight)

	return result
}

// filterPruningHeights filters a slice of int64 heights to only include the heights that exist in the given iavl.Store.
// It returns a new slice of filtered heights.
func filterPruningHeights(store *iavl.Store, heights []int64) []int64 {
	var result []int64

	for _, height := range heights {
		if store.VersionExists(height) {
			result = append(result, height)
		}
	}

	return result
}

// AddPruningCmd prunes the sdk root multi store history versions based on the pruning options
// specified by command flags.
func AddPruningCmd(appCreator servertypes.AppCreator, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune app history states by keeping the recent heights and deleting old heights",
		Long: `Prune app history states by keeping the recent heights and deleting old heights.
The pruning option is provided via the '--pruning' flag or alternatively with '--pruning-keep-recent'
		
For '--pruning' the options are as follows:
		
default: the last 362880 states are kept
nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
everything: 2 latest states will be kept
custom: allow pruning options to be manually specified through 'pruning-keep-recent'.

If no pruning option is provided, the default option from the config file will be used.

Besides pruning options, database home directory can also be specified via flag '--home'.

If '--apply' is not provided, only the heights to be pruned will be printed out.`,
		Example: `prune --home './' --pruning 'custom' --pruning-keep-recent 100 --apply`,

		RunE: func(cmd *cobra.Command, _ []string) error {
			vp := viper.New()

			// bind flags to the Context's Viper so we can get pruning options.
			if err := vp.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			pruningOptions, err := server.GetPruningOptionsFromFlags(vp)
			if err != nil {
				return err
			}

			home := vp.GetString(flags.FlagHome)
			db, err := openDB(home)
			if err != nil {
				return err
			}

			logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
			app := appCreator(logger, db, nil, vp)
			cms := app.CommitMultiStore()

			rootMultiStore, ok := cms.(*rootmulti.Store)
			if !ok {
				return fmt.Errorf("currently only support the pruning of rootmulti.Store type")
			}
			latestHeight := rootmulti.GetLatestVersion(db)

			// valid heights should be greater than 0.
			if latestHeight <= 0 {
				return fmt.Errorf("the database has no valid heights to prune, the latest height: %v", latestHeight)
			}

			// get the heights to be pruned
			var pruningHeights []int64
			for height := int64(1); height < latestHeight; height++ {
				if height < latestHeight-int64(pruningOptions.KeepRecent) {
					pruningHeights = append(pruningHeights, height)
				}
			}
			if len(pruningHeights) == 0 || pruningOptions.KeepRecent == 0 {
				fmt.Printf("no heights to prune\n")
				return nil
			}

			// print out versions to be pruned
			fmt.Println("Versions to be pruned:")
			for key, store := range rootMultiStore.GetStores() {
				if store.GetStoreType() == storetypes.StoreTypeIAVL {
					var kvStore = rootMultiStore.GetCommitKVStore(key)
					pruningHeightsFiltered := filterPruningHeights(kvStore.(*iavl.Store), pruningHeights)
					fmt.Printf("%v: %v\n", key.Name(), formatHeights(pruningHeightsFiltered))

					if vp.GetBool(FlagApply) {
						if err := kvStore.(*iavl.Store).DeleteVersions(pruningHeights...); err != nil {
							if errCause := errors.Cause(err); errCause != nil && errCause != iavltree.ErrVersionDoesNotExist {
								panic(err)
							}
						}
					}

				}
			}

			// pruning was done if the flag was set.
			if vp.GetBool(FlagApply) {
				fmt.Printf("successfully pruned the application root multi stores\n")
			}

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The database home directory")
	cmd.Flags().String(server.FlagPruning, storetypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(server.FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Bool(FlagApply, false, "Do the pruning")

	return cmd
}

func openDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return sdk.NewLevelDB("application", dataDir)
}

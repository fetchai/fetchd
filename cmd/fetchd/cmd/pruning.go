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
		besides pruning options, database home directory should also be specified via flag
		'--home'.`,
		Example: `prune --home './' --pruning 'custom' --pruning-keep-recent 100 --
		pruning-keep-every 10, --pruning-interval 10`,

		RunE: func(cmd *cobra.Command, _ []string) error {
			vp := viper.New()

			// Bind flags to the Context's Viper so we can get pruning options.
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

			// Print out versions to be pruned
			fmt.Println("Versions to be pruned:")
			for key, store := range rootMultiStore.GetStores() {
				if store.GetStoreType() == storetypes.StoreTypeIAVL {
					var kvStore = rootMultiStore.GetCommitKVStore(key)
					pruningHeightsFiltered := filterPruningHeights(kvStore.(*iavl.Store), pruningHeights)
					fmt.Printf("%v: %v\n", key.Name(), formatHeights(pruningHeightsFiltered))
				}
			}

			// Do the pruning if the apply flag is set.
			if vp.GetBool(FlagApply) {
				fmt.Printf(
					"pruning heights start from %v, end at %v\n",
					pruningHeights[0],
					pruningHeights[len(pruningHeights)-1],
				)

				rootMultiStore.PruneStores(false, pruningHeights)
				if err != nil {
					return err
				}
				fmt.Printf("successfully pruned the application root multi stores\n")
			}

			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The database home directory")
	cmd.Flags().String(server.FlagPruning, storetypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(server.FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(server.FlagPruningKeepEvery, 0,
		`Offset heights to keep on disk after 'keep-every' (ignored if pruning is not 'custom'),
		this is not used by this command but kept for compatibility with the complete pruning options`)
	cmd.Flags().Uint64(server.FlagPruningInterval, 10,
		`Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom'), 
		this is not used by this command but kept for compatibility with the complete pruning options`)
	cmd.Flags().Bool(FlagApply, false, "Do the pruning")

	return cmd
}

func openDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return sdk.NewLevelDB("application", dataDir)
}

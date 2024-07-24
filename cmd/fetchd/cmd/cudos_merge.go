package cmd

import "github.com/spf13/cobra"

// Module init related flags
const (
	FlagCudosPath = "cudos-path"
)

func AddCudosFlags(startCmd *cobra.Command) {
	startCmd.Flags().String(FlagCudosPath, "", "Cudos genesis path")
}

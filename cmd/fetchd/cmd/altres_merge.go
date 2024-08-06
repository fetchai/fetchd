package cmd

import "github.com/spf13/cobra"

// Module init related flags
const (
	FlagAltresPath = "altres-path"
)

func AddAltresFlags(startCmd *cobra.Command) {
	startCmd.Flags().String(FlagAltresPath, "", "Altres genesis path")
}

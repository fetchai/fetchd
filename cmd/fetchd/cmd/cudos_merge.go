package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Module init related flags
const (
	FlagCudosGenesisPath           = "cudos-genesis-path"
	FlagCudosGenesisSha256         = "cudos-genesis-sha256"
	FlagCudosMigrationConfigPath   = "cudos-migration-config-path"
	FlagCudosMigrationConfigSha256 = "cudos-migration-config-sha256"
)

func AddCudosFlags(startCmd *cobra.Command) {
	startCmd.Flags().String(FlagCudosGenesisPath, "", "Cudos genesis file path. Required to be provided *exclusively* during cudos migration upgrade node start, *ignored* on all subsequent node starts.")
	startCmd.Flags().String(FlagCudosMigrationConfigPath, "", "Upgrade config file path. Required to be provided *exclusively* during cudos migration upgrade node start, *ignored* on all subsequent node starts.")
	startCmd.Flags().String(FlagCudosGenesisSha256, "", "Sha256 of the cudos genesis file. Optional to be provided *exclusively* during cudos migration upgrade node start, *ignored* on all subsequent node starts.")
	startCmd.Flags().String(FlagCudosMigrationConfigSha256, "", fmt.Sprintf("Sha256 of the upgrade config file. Required if to be provided *exclusively* during cudos migration upgrade node start and *only IF* \"%v\" flag has been provided, *ignored* on all subsequent node starts.", FlagCudosMigrationConfigPath))
}

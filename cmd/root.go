package cmd

import (
	"github.com/anomalyco/nook/config"
	"github.com/spf13/cobra"

	_ "github.com/anomalyco/nook/provider"
)

var cfg *config.GlobalConfig

var rootCmd = &cobra.Command{
	Use:   "nook",
	Short: "Workspace organizer CLI for developers",
	Long:  "Nook is a CLI tool to organize and launch project workspaces with a single command.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.LoadGlobalConfig()
		return err
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
}

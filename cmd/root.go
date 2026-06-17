package cmd

import (
	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/tui"
	"github.com/spf13/cobra"

	_ "github.com/lorenzo-vecchio/nook/provider"
)

var rootCmd = &cobra.Command{
	Use:   "nook",
	Short: "Workspace organizer CLI for developers",
	Long:  "Nook is a CLI tool to organize and launch project workspaces with a single command.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.LoadGlobalConfig()
		return err
	},
}

func init() {
	p := tui.NewPrompter()
	rootCmd.AddCommand(NewInitCmd(p))
	rootCmd.AddCommand(NewOpenCmd(p))
	rootCmd.AddCommand(NewListCmd())
	rootCmd.AddCommand(NewEditCmd())
	rootCmd.AddCommand(NewDeleteCmd())
	rootCmd.AddCommand(NewDetectCmd())
	rootCmd.AddCommand(NewScanCmd())
}

func Execute() error {
	return rootCmd.Execute()
}

package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "nook",
	Short: "Workspace organizer CLI for developers",
	Long:  "Nook is a CLI tool to organize and launch project workspaces with a single command.",
}

func Execute() error {
	return rootCmd.Execute()
}

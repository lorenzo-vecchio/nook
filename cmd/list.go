package cmd

import (
	"fmt"
	"os"

	"github.com/anomalyco/nook/config"
	"github.com/anomalyco/nook/detector"
	"github.com/anomalyco/nook/tui"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadGlobalConfig()
			if err != nil {
				return err
			}

			totalFound := 0
			for _, scanPath := range cfg.ScanPaths {
				workspaces, err := detector.ScanPath(scanPath)
				if err != nil {
					tui.PrintError(os.Stderr, fmt.Sprintf("Failed to scan %s: %s", scanPath, err))
					continue
				}
				if len(workspaces) == 0 {
					continue
				}
				totalFound += len(workspaces)
				tui.PrintHeader(os.Stdout, fmt.Sprintf("From %s", scanPath))
				items := make([]tui.WorkspaceListItem, 0, len(workspaces))
				for _, ws := range workspaces {
					items = append(items, tui.WorkspaceListItem{
						Name:        ws.Name,
						Description: ws.Description,
						Envs:        ws.EnvNames,
						Path:        ws.Path,
					})
				}
				tui.PrintWorkspaceList(os.Stdout, items)
			}

			if totalFound == 0 {
				tui.PrintWorkspaceList(os.Stdout, nil)
			}

			return nil
		},
	}
}

package cmd

import (
	"os"
	"path/filepath"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/detector"
	"github.com/lorenzo-vecchio/nook/tui"
	"github.com/spf13/cobra"
)

func NewScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Rescan all scan paths for workspace.yaml files",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadGlobalConfig()
			if err != nil {
				return err
			}

			var updated bool
			var validPaths []string
			var items []tui.WorkspaceListItem

			for _, path := range cfg.ScanPaths {
				info, err := os.Stat(path)
				if err != nil {
					if os.IsNotExist(err) {
						updated = true
						continue
					}
					return err
				}
				if !info.IsDir() {
					updated = true
					continue
				}

				validPaths = append(validPaths, path)

				workspaces, err := detector.ScanPath(path)
				if err != nil {
					continue
				}

				for _, ws := range workspaces {
					items = append(items, tui.WorkspaceListItem{
						Name:        ws.Name,
						Description: ws.Description,
						Envs:        ws.EnvNames,
						Path:        ws.Path,
					})
				}
			}

			if updated {
				relPaths := make([]string, len(validPaths))
				for i, p := range validPaths {
					relPaths[i] = filepath.Clean(p)
				}
				cfg.ScanPaths = relPaths
				if err := config.SaveGlobalConfig(cfg); err != nil {
					return err
				}
			}

			tui.PrintHeader(cmd.OutOrStdout(), "Scan Results")
			tui.PrintWorkspaceList(cmd.OutOrStdout(), items)

			return nil
		},
	}
}

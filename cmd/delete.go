package cmd

import (
	"fmt"
	"os"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/detector"
	"github.com/lorenzo-vecchio/nook/tui"
	"github.com/spf13/cobra"
)

var deletePrompter = tui.NewPrompter()

func NewDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a workspace",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadGlobalConfig()
			if err != nil {
				return err
			}

			var workspaceName string
			if len(args) > 0 {
				workspaceName = args[0]
			} else {
				var names []string
				for _, sp := range cfg.ScanPaths {
					workspaces, err := detector.ScanPath(sp)
					if err != nil {
						continue
					}
					for name := range workspaces {
						names = append(names, name)
					}
				}
				if len(names) == 0 {
					return fmt.Errorf("no workspaces found")
				}
				selected, err := deletePrompter.Select("Select workspace to delete", names, names[0])
				if err != nil {
					return err
				}
				workspaceName = selected
			}

			var wsPath string
			var parentScanPath string
			for _, sp := range cfg.ScanPaths {
				workspaces, err := detector.ScanPath(sp)
				if err != nil {
					continue
				}
				if ws, ok := workspaces[workspaceName]; ok {
					wsPath = ws.Path
					parentScanPath = sp
					break
				}
			}
			if wsPath == "" {
				return fmt.Errorf("workspace %q not found", workspaceName)
			}

			confirmed, err := deletePrompter.Confirm(fmt.Sprintf("Delete workspace %s?", workspaceName), false)
			if err != nil {
				return err
			}
			if !confirmed {
				return nil
			}

			if wsPath == parentScanPath {
				cfg.ScanPaths = removeStr(cfg.ScanPaths, parentScanPath)
				if err := config.SaveGlobalConfig(cfg); err != nil {
					return err
				}
				tui.PrintSuccess(os.Stdout, fmt.Sprintf("Removed %s from scan paths", parentScanPath))
			} else {
				if err := os.RemoveAll(wsPath); err != nil {
					return err
				}
				tui.PrintSuccess(os.Stdout, fmt.Sprintf("Deleted workspace %s", workspaceName))
			}

			return nil
		},
	}
}

func removeStr(slice []string, s string) []string {
	idx := 0
	for _, v := range slice {
		if v != s {
			slice[idx] = v
			idx++
		}
	}
	return slice[:idx]
}

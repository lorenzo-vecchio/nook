package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/detector"
	"github.com/lorenzo-vecchio/nook/tui"
	"github.com/lorenzo-vecchio/nook/utils"
	"github.com/spf13/cobra"
)

var editPrompter = tui.NewPrompter()

func NewEditCmd() *cobra.Command {
	var newName string
	var newDesc string

	cmd := &cobra.Command{
		Use:   "edit [name]",
		Short: "Edit a workspace configuration",
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
				selected, err := editPrompter.Select("Select workspace to edit", names, names[0])
				if err != nil {
					return err
				}
				workspaceName = selected
			}

			yamlPath := ""
			for _, sp := range cfg.ScanPaths {
				workspaces, err := detector.ScanPath(sp)
				if err != nil {
					continue
				}
				if ws, ok := workspaces[workspaceName]; ok {
					yamlPath = filepath.Join(ws.Path, "workspace.yaml")
					break
				}
			}
			if yamlPath == "" {
				return fmt.Errorf("workspace %q not found", workspaceName)
			}

			if newName != "" || newDesc != "" {
				ws, err := config.LoadWorkspace(yamlPath)
				if err != nil {
					return err
				}
				if newName != "" {
					ws.Name = newName
				}
				if newDesc != "" {
					ws.Description = newDesc
				}
				if err := config.Validate(ws); err != nil {
					return err
				}
				if err := config.SaveWorkspace(ws, yamlPath); err != nil {
					return err
				}
				tui.PrintSuccess(os.Stdout, fmt.Sprintf("Workspace %q updated", workspaceName))
				return nil
			}

			editor := utils.DefaultEditor()
			editorCmd := exec.Command(editor, yamlPath)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr
			return editorCmd.Run()
		},
	}

	cmd.Flags().StringVar(&newName, "name", "", "Set workspace name")
	cmd.Flags().StringVar(&newDesc, "description", "", "Set workspace description")

	return cmd
}

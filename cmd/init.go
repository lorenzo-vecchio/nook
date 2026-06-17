package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/tui"
	"github.com/lorenzo-vecchio/nook/utils"
	"github.com/spf13/cobra"
)

func NewInitCmd(p tui.Prompter) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a new workspace interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(p)
		},
	}
}

func runInit(p tui.Prompter) error {
	name, err := p.Input("Workspace name", "")
	if err != nil {
		return err
	}

	desc, err := p.Input("Description (optional)", "")
	if err != nil {
		return err
	}

	ws := &config.WorkspaceConfig{
		Name:         name,
		Description:  desc,
		Environments: make(map[string]config.Environment),
	}

	addEnv := true
	for addEnv {
		envName, err := p.Input("Environment name", "dev")
		if err != nil {
			return err
		}

		envFile, err := p.Input("Path to .env file (optional)", "")
		if err != nil {
			return err
		}

		serviceOptions := []string{"VS Code", "DBeaver", "Chrome", "Docker Compose", "Custom Command"}
		selected, err := p.MultiSelect("Select services", serviceOptions, nil)
		if err != nil {
			return err
		}

		var services []config.Service
		for _, s := range selected {
			svc, err := configureService(p, s)
			if err != nil {
				return err
			}
			services = append(services, *svc)
		}

		ws.Environments[envName] = config.Environment{
			EnvFile:  envFile,
			Services: services,
		}

		addEnv, err = p.Confirm("Add another environment?", false)
		if err != nil {
			return err
		}
	}

	if err := config.Validate(ws); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	wsDir := filepath.Join(home, ".nook", "workspaces", name)
	if err := utils.EnsureDir(wsDir); err != nil {
		return err
	}

	wsPath := filepath.Join(wsDir, "workspace.yaml")
	if err := config.SaveWorkspace(ws, wsPath); err != nil {
		return err
	}

	if err := utils.EnsureDir(filepath.Join(wsDir, ".workspace")); err != nil {
		return err
	}

	tui.PrintSuccess(os.Stdout, fmt.Sprintf("Workspace %q created at %s", name, wsPath))
	return nil
}

func configureService(p tui.Prompter, serviceType string) (*config.Service, error) {
	svc := &config.Service{
		Provider: serviceTypeToProvider(serviceType),
	}

	switch serviceType {
	case "VS Code":
		folder, err := p.Input("Folder path", "")
		if err != nil {
			return nil, err
		}
		svc.Folder = folder

		addTerminals, err := p.Confirm("Add terminals?", false)
		if err != nil {
			return nil, err
		}
		if addTerminals {
			for {
				name, err := p.Input("Terminal name", "")
				if err != nil {
					return nil, err
				}
				dir, err := p.Input("Terminal directory", "")
				if err != nil {
					return nil, err
				}
				cmdStr, err := p.Input("Terminal command (optional)", "")
				if err != nil {
					return nil, err
				}
				svc.Terminals = append(svc.Terminals, config.Terminal{
					Name:      name,
					Directory: dir,
					Command:   cmdStr,
				})

				more, err := p.Confirm("Add another terminal?", false)
				if err != nil {
					return nil, err
				}
				if !more {
					break
				}
			}
		}

	case "DBeaver":
		conn, err := p.Input("Connection string", "")
		if err != nil {
			return nil, err
		}
		svc.Connection = conn

	case "Chrome":
		urlsStr, err := p.Input("URLs (comma-separated)", "")
		if err != nil {
			return nil, err
		}
		if urlsStr != "" {
			parts := strings.Split(urlsStr, ",")
			for i, u := range parts {
				parts[i] = strings.TrimSpace(u)
			}
			svc.URLs = parts
		}

	case "Docker Compose":
		file, err := p.Input("Docker Compose file path", "")
		if err != nil {
			return nil, err
		}
		svc.File = file

		profile, err := p.Input("Profile (optional)", "")
		if err != nil {
			return nil, err
		}
		svc.Profile = profile

	case "Custom Command":
		cmdStr, err := p.Input("Command to run", "")
		if err != nil {
			return nil, err
		}
		svc.Cmd = cmdStr

		cwd, err := p.Input("Working directory (optional)", "")
		if err != nil {
			return nil, err
		}
		svc.Cwd = cwd
	}

	return svc, nil
}

func serviceTypeToProvider(s string) string {
	switch s {
	case "VS Code":
		return "vscode"
	case "DBeaver":
		return "dbeaver"
	case "Chrome":
		return "chrome"
	case "Docker Compose":
		return "docker"
	case "Custom Command":
		return "command"
	}
	return ""
}

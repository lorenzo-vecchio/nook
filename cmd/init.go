package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/detector"
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

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	locations := []string{"Default (~/.nook/workspaces)", "Current directory (" + cwd + ")", "Choose a scan path..."}
	choice, err := p.Select("Where to save the workspace?", locations, locations[0])
	if err != nil {
		return err
	}

	var wsDir string
	switch choice {
	case locations[0]:
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		wsDir = filepath.Join(home, ".nook", "workspaces", name)
	case locations[1]:
		wsDir = cwd
	case locations[2]:
		cfg, err := config.LoadGlobalConfig()
		if err != nil {
			return err
		}
		picked, err := p.Select("Choose a scan path", cfg.ScanPaths, cfg.ScanPaths[0])
		if err != nil {
			return err
		}
		wsDir = filepath.Join(picked, name)
	}

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

	if choice != locations[0] {
		if err := detector.TrustPath(wsDir); err != nil {
			return err
		}
	}

	tui.PrintSuccess(os.Stdout, fmt.Sprintf("Workspace %q created at %s", name, wsPath))
	return nil
}

func configureService(p tui.Prompter, serviceType string) (*config.Service, error) {
	svc := &config.Service{
		Provider: serviceTypeToProvider(serviceType),
	}
	label := tui.Dim("["+serviceType+"]") + " "

	switch serviceType {
	case "VS Code":
		folder, err := p.Input(label+"Project folder to open (e.g. ./backend or .)", "")
		if err != nil {
			return nil, err
		}
		svc.Folder = folder

		addTerminals, err := p.Confirm(label+"Add terminals?", false)
		if err != nil {
			return nil, err
		}
		if addTerminals {
			for {
				name, err := p.Input(label+"Terminal name", "")
				if err != nil {
					return nil, err
				}
				dir, err := p.Input(label+"Terminal directory", "")
				if err != nil {
					return nil, err
				}
				cmdStr, err := p.Input(label+"Terminal command (optional)", "")
				if err != nil {
					return nil, err
				}
				svc.Terminals = append(svc.Terminals, config.Terminal{
					Name:      name,
					Directory: dir,
					Command:   cmdStr,
				})

				more, err := p.Confirm(label+"Add another terminal?", false)
				if err != nil {
					return nil, err
				}
				if !more {
					break
				}
			}
		}

	case "DBeaver":
		useInteractive, err := p.Confirm(label+"Build connection interactively?", true)
		if err != nil {
			return nil, err
		}
		if useInteractive {
			drivers := []string{
				"postgresql", "mysql", "mariadb", "sqlite",
				"clickhouse", "oracle", "sqlserver", "db2",
				"firebird", "h2", "derby", "mongodb",
				"cassandra", "redis", "vertica", "bigquery",
			}
			driver, err := p.Select(label+"Driver", drivers, "postgresql")
			if err != nil {
				return nil, err
			}
			host, err := p.Input(label+"Host", "localhost")
			if err != nil {
				return nil, err
			}
			port, err := p.Input(label+"Port", "5432")
			if err != nil {
				return nil, err
			}
			db, err := p.Input(label+"Database", "")
			if err != nil {
				return nil, err
			}
			user, err := p.Input(label+"User (use ${VAR} for env vars)", "")
			if err != nil {
				return nil, err
			}
			pass, err := p.Input(label+"Password (use ${VAR} for env vars)", "")
			if err != nil {
				return nil, err
			}
			extra, err := p.Input(label+"Extra params (optional, e.g. sslmode=disable)", "")
			if err != nil {
				return nil, err
			}

			parts := []string{"driver=" + driver, "host=" + host, "port=" + port}
			if db != "" {
				parts = append(parts, "database="+db)
			}
			if user != "" {
				parts = append(parts, "user="+user)
			}
			if pass != "" {
				parts = append(parts, "password="+pass)
			}
			if extra != "" {
				parts = append(parts, extra)
			}
			svc.Connection = strings.Join(parts, "|")
		} else {
			conn, err := p.Input(label+"Connection string", "")
			if err != nil {
				return nil, err
			}
			svc.Connection = conn
		}

	case "Chrome":
		urlsStr, err := p.Input(label+"URLs (comma-separated)", "")
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
		file, err := p.Input(label+"Compose file path", "")
		if err != nil {
			return nil, err
		}
		svc.File = file

		profile, err := p.Input(label+"Profile (optional)", "")
		if err != nil {
			return nil, err
		}
		svc.Profile = profile

	case "Custom Command":
		cmdStr, err := p.Input(label+"Command to run", "")
		if err != nil {
			return nil, err
		}
		svc.Cmd = cmdStr

		cwd, err := p.Input(label+"Working directory (optional)", "")
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

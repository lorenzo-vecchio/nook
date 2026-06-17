package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/anomalyco/nook/config"
	"github.com/anomalyco/nook/detector"
	"github.com/anomalyco/nook/provider"
	"github.com/anomalyco/nook/tui"
	"github.com/spf13/cobra"
)

func NewOpenCmd(p tui.Prompter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [name]",
		Short: "Open a workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOpen(p, cmd, args)
		},
	}
	cmd.Flags().String("env", "", "Environment to use")
	return cmd
}

func runOpen(p tui.Prompter, cmd *cobra.Command, args []string) error {
	result, err := detector.ScanCurrentDir()
	if err != nil {
		return err
	}

	if result != nil {
		var untrusted []detector.WorkspaceInfo
		untrusted = append(untrusted, result.InCWD...)
		untrusted = append(untrusted, result.InSubdirs...)

		if len(untrusted) > 0 {
			globalCfg, err := config.LoadGlobalConfig()
			if err != nil {
				return err
			}

			for _, ws := range untrusted {
				if !detector.IsTrusted(ws.Path, globalCfg.ScanPaths) {
					trust, err := p.Confirm(
						fmt.Sprintf("Workspace %q at %s is not trusted. Trust it?", ws.Name, ws.Path),
						true,
					)
					if err != nil {
						return err
					}
					if trust {
						if err := detector.TrustPath(ws.Path); err != nil {
							tui.PrintError(os.Stderr, fmt.Sprintf("Failed to trust path: %v", err))
						}
					}
				}
			}
		}
	}

	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		return err
	}

	allWorkspaces := make(map[string]detector.WorkspaceInfo)
	for _, sp := range cfg.ScanPaths {
		workspaces, err := detector.ScanPath(sp)
		if err != nil {
			tui.PrintError(os.Stderr, fmt.Sprintf("Failed to scan %s: %v", sp, err))
			continue
		}
		for name, info := range workspaces {
			allWorkspaces[name] = *info
		}
	}

	var selected *detector.WorkspaceInfo

	if len(args) > 0 {
		name := args[0]
		info, ok := allWorkspaces[name]
		if !ok {
			return fmt.Errorf("workspace %q not found", name)
		}
		selected = &info
	} else {
		if len(allWorkspaces) == 0 {
			return fmt.Errorf("no workspaces found")
		}

		names := make([]string, 0, len(allWorkspaces))
		for name := range allWorkspaces {
			names = append(names, name)
		}
		sort.Strings(names)

		chosen, err := p.Select("Select workspace", names, names[0])
		if err != nil {
			return err
		}
		info := allWorkspaces[chosen]
		selected = &info
	}

	if len(selected.EnvNames) == 0 {
		return fmt.Errorf("workspace %q has no environments", selected.Name)
	}

	envFlag, _ := cmd.Flags().GetString("env")

	var envName string
	if envFlag != "" {
		envName = envFlag
	} else if len(selected.EnvNames) == 1 {
		envName = selected.EnvNames[0]
	} else {
		chosen, err := p.Select("Select environment", selected.EnvNames, selected.EnvNames[0])
		if err != nil {
			return err
		}
		envName = chosen
	}

	wsPath := filepath.Join(selected.Path, "workspace.yaml")
	ws, err := config.LoadWorkspace(wsPath)
	if err != nil {
		return err
	}

	env, ok := ws.Environments[envName]
	if !ok {
		return fmt.Errorf("environment %q not found", envName)
	}

	var envFileName string
	if env.EnvFile != "" {
		if filepath.IsAbs(env.EnvFile) {
			envFileName = env.EnvFile
		} else {
			envFileName = filepath.Join(selected.Path, env.EnvFile)
		}
	}

	var envVars map[string]string
	if envFileName != "" {
		envVars, err = config.LoadEnvFile(envFileName)
		if err != nil {
			tui.PrintError(os.Stderr, fmt.Sprintf("Failed to load env file: %v", err))
		}
	}

	ctx := context.Background()
	var wg sync.WaitGroup

	for i := range env.Services {
		svc := env.Services[i]

		svc.Folder = config.ResolveAllEnvVars(svc.Folder, envFileName, envVars)
		svc.Connection = config.ResolveAllEnvVars(svc.Connection, envFileName, envVars)
		svc.Cmd = config.ResolveAllEnvVars(svc.Cmd, envFileName, envVars)
		svc.Cwd = config.ResolveAllEnvVars(svc.Cwd, envFileName, envVars)
		svc.File = config.ResolveAllEnvVars(svc.File, envFileName, envVars)
		svc.Profile = config.ResolveAllEnvVars(svc.Profile, envFileName, envVars)
		for j := range svc.URLs {
			svc.URLs[j] = config.ResolveAllEnvVars(svc.URLs[j], envFileName, envVars)
		}
		for j := range svc.Terminals {
			svc.Terminals[j].Directory = config.ResolveAllEnvVars(svc.Terminals[j].Directory, envFileName, envVars)
			svc.Terminals[j].Command = config.ResolveAllEnvVars(svc.Terminals[j].Command, envFileName, envVars)
		}

		wg.Add(1)
		go func(svc config.Service) {
			defer wg.Done()

			prov, ok := provider.Get(svc.Provider)
			if !ok {
				tui.PrintError(os.Stderr, fmt.Sprintf("Provider %q not found", svc.Provider))
				return
			}

			tui.PrintProgress(os.Stdout, prov.Name(), "Launching...")

			if err := prov.Launch(ctx, svc, selected.Path, envVars); err != nil {
				tui.PrintError(os.Stderr, fmt.Sprintf("Failed to launch %s: %v", prov.Name(), err))
			}
		}(svc)
	}

	wg.Wait()
	tui.PrintSuccess(os.Stdout, "All services launched!")
	return nil
}

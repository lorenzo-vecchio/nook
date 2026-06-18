package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/detector"
	"github.com/lorenzo-vecchio/nook/provider"
	"github.com/lorenzo-vecchio/nook/tui"
	"github.com/lorenzo-vecchio/nook/utils"
	"github.com/spf13/cobra"
)

var execCommandContext = exec.CommandContext

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

	services := make([]config.Service, len(env.Services))
	for i, svc := range env.Services {
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
		services[i] = svc
	}

	var dockerSvc *config.Service
	var rest []config.Service
	for i := range services {
		if services[i].Provider == "docker" {
			dockerSvc = &services[i]
		} else {
			rest = append(rest, services[i])
		}
	}

	if dockerSvc != nil {
		launchDocker(ctx, dockerSvc, selected.Path, env.WaitForComposeHealthy)
	}

	sort.Slice(rest, func(i, j int) bool {
		oi := rest[i].Order
		oj := rest[j].Order
		if oi == 0 {
			oi = int(^uint(0) >> 1)
		}
		if oj == 0 {
			oj = int(^uint(0) >> 1)
		}
		return oi < oj
	})

	groups := make(map[int][]config.Service)
	allZero := true
	for _, svc := range rest {
		groups[svc.Order] = append(groups[svc.Order], svc)
		if svc.Order != 0 {
			allZero = false
		}
	}

	keys := make([]int, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ki := keys[i]
		kj := keys[j]
		if ki == 0 {
			ki = int(^uint(0) >> 1)
		}
		if kj == 0 {
			kj = int(^uint(0) >> 1)
		}
		return ki < kj
	})

	if allZero {
		var wg sync.WaitGroup
		for _, svc := range rest {
			wg.Add(1)
			go func(svc config.Service) {
				defer wg.Done()
				launchService(ctx, svc, selected.Path, envVars)
			}(svc)
		}
		wg.Wait()
	} else {
		for _, key := range keys {
			group := groups[key]
			var wg sync.WaitGroup
			for _, svc := range group {
				if svc.DelayMs > 0 {
					time.Sleep(time.Duration(svc.DelayMs) * time.Millisecond)
				}
				if svc.ReadyCheck != nil {
					waitForReady(ctx, svc)
				}
				wg.Add(1)
				go func(svc config.Service) {
					defer wg.Done()
					launchService(ctx, svc, selected.Path, envVars)
				}(svc)
			}
			wg.Wait()
		}
	}

	tui.PrintSuccess(os.Stdout, "All services launched!")
	return nil
}

func launchService(ctx context.Context, svc config.Service, basePath string, envVars map[string]string) {
	prov, ok := provider.Get(svc.Provider)
	if !ok {
		tui.PrintError(os.Stderr, fmt.Sprintf("Provider %q not found", svc.Provider))
		return
	}
	tui.PrintProgress(os.Stdout, prov.Name(), "Launching...")
	if err := prov.Launch(ctx, svc, basePath, envVars); err != nil {
		tui.PrintError(os.Stderr, fmt.Sprintf("Failed to launch %s: %v", prov.Name(), err))
	}
}

func launchDocker(ctx context.Context, svc *config.Service, baseDir string, waitHealthy bool) {
	tui.PrintProgress(os.Stdout, "docker", "Launching...")
	file := utils.ResolvePath(baseDir, svc.File)
	args := []string{"compose", "-f", file}
	if svc.Profile != "" {
		args = append(args, "--profile", svc.Profile)
	}
	args = append(args, "up", "-d")
	dockerCmd := execCommandContext(ctx, "docker", args...)
	dockerCmd.Stderr = os.Stderr
	if err := dockerCmd.Start(); err != nil {
		tui.PrintError(os.Stderr, fmt.Sprintf("Docker Compose failed to start: %v", err))
		return
	}
	if err := dockerCmd.Wait(); err != nil {
		tui.PrintError(os.Stderr, fmt.Sprintf("Docker Compose failed: %v", err))
		return
	}
	if waitHealthy {
		if err := waitDockerHealthy(ctx); err != nil {
			tui.PrintError(os.Stderr, "Docker Compose unhealthy after 120s, continuing...")
		} else {
			tui.PrintSuccess(os.Stdout, "Docker Compose healthy")
		}
	}
}

func waitForReady(ctx context.Context, svc config.Service) {
	interval := svc.ReadyCheck.IntervalMs
	if interval <= 0 {
		interval = 2000
	}
	timeout := svc.ReadyCheck.TimeoutMs
	if timeout <= 0 {
		timeout = 30000
	}
	name, shellArgs := utils.ShellCommand(svc.ReadyCheck.Cmd)
	deadline := time.Now().Add(time.Duration(timeout) * time.Millisecond)
	for time.Now().Before(deadline) {
		cmd := execCommandContext(ctx, name, shellArgs...)
		if cmd.Run() == nil {
			return
		}
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
	tui.PrintError(os.Stderr, svc.Provider+" readiness check timed out, continuing...")
}

func waitDockerHealthy(ctx context.Context) error {
	deadline := time.Now().Add(120 * time.Second)
	for time.Now().Before(deadline) {
		cmd := execCommandContext(ctx, "docker", "compose", "ps", "--format", "json")
		out, err := cmd.Output()
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		allHealthy := true
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			var c struct {
				Health string `json:"Health"`
			}
			if err := json.Unmarshal([]byte(line), &c); err != nil {
				continue
			}
			if c.Health == "starting" {
				allHealthy = false
				break
			}
		}
		if allHealthy {
			return nil
		}
		fmt.Fprintf(os.Stderr, "  %s Waiting for Docker Compose to be healthy...\n", tui.LoadingMark())
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("timeout")
}

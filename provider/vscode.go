package provider

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/anomalyco/nook/config"
	"github.com/anomalyco/nook/utils"
)

func vscodeCommonPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
		}
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		return []string{
			`C:\Program Files\Microsoft VS Code\bin\code.cmd`,
			filepath.Join(localAppData, "Programs", "Microsoft VS Code", "bin", "code.cmd"),
		}
	default:
		return []string{
			"/usr/share/code/bin/code",
			"/snap/bin/code",
		}
	}
}

var execCommandContext = exec.CommandContext

type VSCodeProvider struct{}

func (p *VSCodeProvider) Name() string {
	return "vscode"
}

func (p *VSCodeProvider) Detect() (bool, error) {
	if _, err := exec.LookPath("code"); err == nil {
		return true, nil
	}
	if _, err := exec.LookPath("code-insiders"); err == nil {
		return true, nil
	}

	for _, path := range vscodeCommonPaths() {
		if _, err := os.Stat(path); err == nil {
			return true, nil
		}
	}

	return false, nil
}

func (p *VSCodeProvider) Launch(ctx context.Context, svc config.Service, baseDir string, envVars map[string]string) error {
	folder := utils.ResolvePath(baseDir, svc.Folder)

	if len(svc.Terminals) == 0 {
		cmd := execCommandContext(ctx, "code", folder)
		return cmd.Start()
	}

	wsPath, err := p.generateWorkspaceFile(folder, svc.Terminals, baseDir)
	if err != nil {
		return err
	}

	cmd := execCommandContext(ctx, "code", wsPath)
	return cmd.Start()
}

type codeWorkspace struct {
	Folders []codeWorkspaceFolder `json:"folders"`
	Tasks   codeWorkspaceTasks    `json:"tasks"`
}

type codeWorkspaceFolder struct {
	Path string `json:"path"`
}

type codeWorkspaceTasks struct {
	Version string            `json:"version"`
	Tasks   []codeWorkspaceTask `json:"tasks"`
}

type codeWorkspaceTask struct {
	Label          string   `json:"label"`
	Type           string   `json:"type"`
	Command        string   `json:"command"`
	RunOn          string   `json:"runOn"`
	ProblemMatcher []string `json:"problemMatcher"`
}

func (p *VSCodeProvider) generateWorkspaceFile(folder string, terminals []config.Terminal, baseDir string) (string, error) {
	wsDir := filepath.Join(baseDir, ".workspace")
	if err := utils.EnsureDir(wsDir); err != nil {
		return "", err
	}

	name := filepath.Base(folder)
	wsPath := filepath.Join(wsDir, name+".code-workspace")

	tasks := make([]codeWorkspaceTask, 0, len(terminals))
	for _, t := range terminals {
		termDir := utils.ResolvePath(folder, t.Directory)
		cmd := "cd " + termDir
		if t.Command != "" {
			cmd += " && " + t.Command
		}
		tasks = append(tasks, codeWorkspaceTask{
			Label:          t.Name,
			Type:           "shell",
			Command:        cmd,
			RunOn:          "folderOpen",
			ProblemMatcher: []string{},
		})
	}

	ws := codeWorkspace{
		Folders: []codeWorkspaceFolder{{Path: folder}},
		Tasks: codeWorkspaceTasks{
			Version: "2.0.0",
			Tasks:   tasks,
		},
	}

	data, err := json.MarshalIndent(ws, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(wsPath, data, 0644); err != nil {
		return "", err
	}

	return wsPath, nil
}

func init() {
	Register(&VSCodeProvider{})
}

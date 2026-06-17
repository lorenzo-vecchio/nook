package provider

import (
	"context"
	"os/exec"

	"github.com/lorenzo-vecchio/nook/config"
	"github.com/lorenzo-vecchio/nook/utils"
)

type CommandProvider struct{}

func (p *CommandProvider) Name() string {
	return "command"
}

func (p *CommandProvider) Detect() (bool, error) {
	return true, nil
}

func (p *CommandProvider) Launch(ctx context.Context, svc config.Service, baseDir string, envVars map[string]string) error {
	cwd := ""
	if svc.Cwd != "" {
		cwd = utils.ResolvePath(baseDir, svc.Cwd)
	}

	cmd := commandCmd(ctx, svc.Cmd, cwd)
	return cmd.Start()
}

func commandCmd(ctx context.Context, cmdStr, cwd string) *exec.Cmd {
	shell, args := utils.ShellCommand(cmdStr)
	cmd := execCommandContext(ctx, shell, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	return cmd
}

func init() {
	Register(&CommandProvider{})
}
